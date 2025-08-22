package application

import (
	"testing"
	"time"

	"go-finance-advisor/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupAdvisorTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&domain.User{},
		&domain.Category{},
		&domain.Transaction{},
	)
	require.NoError(t, err)

	return db
}

func createAdvisorTestData(t *testing.T, db *gorm.DB) (uint, uint, uint) {
	// Create test user
	user := &domain.User{
		Email:         "advisor@example.com",
		FirstName:     "Advisor",
		LastName:      "Test",
		RiskTolerance: "moderate",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	// Create test categories
	incomeCategory := &domain.Category{
		Name: "Salary",
		Type: "income",
	}
	err = db.Create(incomeCategory).Error
	require.NoError(t, err)

	expenseCategory := &domain.Category{
		Name: "Food",
		Type: "expense",
	}
	err = db.Create(expenseCategory).Error
	require.NoError(t, err)

	return user.ID, incomeCategory.ID, expenseCategory.ID
}

func TestAdvisorService_CalculateMonthlySavings(t *testing.T) {
	db := setupAdvisorTestDB(t)
	advisorService := &AdvisorService{DB: db}
	userID, incomeCategoryID, expenseCategoryID := createAdvisorTestData(t, db)

	now := time.Now()

	t.Run("calculate savings with transactions", func(t *testing.T) {
		// Create test transactions for the last 3 months
		transactions := []*domain.Transaction{
			{
				UserID:      userID,
				CategoryID:  incomeCategoryID,
				Amount:      3000.00,
				Type:        "income",
				Description: "Salary Month 1",
				Date:        now.AddDate(0, -1, 0), // 1 month ago
			},
			{
				UserID:      userID,
				CategoryID:  incomeCategoryID,
				Amount:      3000.00,
				Type:        "income",
				Description: "Salary Month 2",
				Date:        now.AddDate(0, -2, 0), // 2 months ago
			},
			{
				UserID:      userID,
				CategoryID:  incomeCategoryID,
				Amount:      3000.00,
				Type:        "income",
				Description: "Salary Month 3",
				Date:        now.AddDate(0, -3, 5), // Just within 3 months
			},
			{
				UserID:      userID,
				CategoryID:  expenseCategoryID,
				Amount:      800.00,
				Type:        "expense",
				Description: "Food Month 1",
				Date:        now.AddDate(0, -1, 0),
			},
			{
				UserID:      userID,
				CategoryID:  expenseCategoryID,
				Amount:      750.00,
				Type:        "expense",
				Description: "Food Month 2",
				Date:        now.AddDate(0, -2, 0),
			},
			{
				UserID:      userID,
				CategoryID:  expenseCategoryID,
				Amount:      900.00,
				Type:        "expense",
				Description: "Food Month 3",
				Date:        now.AddDate(0, -3, 5),
			},
			// Transaction older than 3 months (should be excluded)
			{
				UserID:      userID,
				CategoryID:  incomeCategoryID,
				Amount:      2000.00,
				Type:        "income",
				Description: "Old Salary",
				Date:        now.AddDate(0, -4, 0), // 4 months ago
			},
		}

		for _, tx := range transactions {
			err := db.Create(tx).Error
			require.NoError(t, err)
		}

		savings, err := advisorService.CalculateMonthlySavings(userID)
		assert.NoError(t, err)

		// Total income: 9000, Total expense: 2450
		// Monthly savings: (9000 - 2450) / 3 = 2183.33
		expectedSavings := (9000.00 - 2450.00) / 3.0
		assert.InDelta(t, expectedSavings, savings, 0.01)
	})

	t.Run("calculate savings with no transactions", func(t *testing.T) {
		// Clean up previous transactions
		db.Where("user_id = ?", userID).Delete(&domain.Transaction{})

		savings, err := advisorService.CalculateMonthlySavings(userID)
		assert.NoError(t, err)
		assert.Equal(t, 0.0, savings)
	})

	t.Run("calculate savings with only expenses", func(t *testing.T) {
		// Clean up previous transactions
		db.Where("user_id = ?", userID).Delete(&domain.Transaction{})

		// Create only expense transactions
		expenseTransaction := &domain.Transaction{
			UserID:      userID,
			CategoryID:  expenseCategoryID,
			Amount:      1000.00,
			Type:        "expense",
			Description: "Only Expense",
			Date:        now.AddDate(0, -1, 0),
		}
		err := db.Create(expenseTransaction).Error
		require.NoError(t, err)

		savings, err := advisorService.CalculateMonthlySavings(userID)
		assert.NoError(t, err)
		// Should be negative: (0 - 1000) / 3 = -333.33
		expectedSavings := -1000.00 / 3.0
		assert.InDelta(t, expectedSavings, savings, 0.01)
	})

	t.Run("calculate savings with only income", func(t *testing.T) {
		// Clean up previous transactions
		db.Where("user_id = ?", userID).Delete(&domain.Transaction{})

		// Create only income transactions
		incomeTransaction := &domain.Transaction{
			UserID:      userID,
			CategoryID:  incomeCategoryID,
			Amount:      3000.00,
			Type:        "income",
			Description: "Only Income",
			Date:        now.AddDate(0, -1, 0),
		}
		err := db.Create(incomeTransaction).Error
		require.NoError(t, err)

		savings, err := advisorService.CalculateMonthlySavings(userID)
		assert.NoError(t, err)
		// Should be: 3000 / 3 = 1000
		expectedSavings := 3000.00 / 3.0
		assert.Equal(t, expectedSavings, savings)
	})

	t.Run("calculate savings for non-existent user", func(t *testing.T) {
		savings, err := advisorService.CalculateMonthlySavings(99999)
		assert.NoError(t, err)
		assert.Equal(t, 0.0, savings)
	})
}

func TestAdvisorService_GenerateAdvice(t *testing.T) {
	db := setupAdvisorTestDB(t)
	advisorService := &AdvisorService{DB: db}
	userID, incomeCategoryID, expenseCategoryID := createAdvisorTestData(t, db)

	now := time.Now()

	// Create base transactions for all tests
	baseTransactions := []*domain.Transaction{
		{
			UserID:      userID,
			CategoryID:  incomeCategoryID,
			Amount:      3000.00,
			Type:        "income",
			Description: "Monthly Salary",
			Date:        now.AddDate(0, -1, 0),
		},
		{
			UserID:      userID,
			CategoryID:  expenseCategoryID,
			Amount:      2000.00,
			Type:        "expense",
			Description: "Monthly Expenses",
			Date:        now.AddDate(0, -1, 0),
		},
	}

	for _, tx := range baseTransactions {
		err := db.Create(tx).Error
		require.NoError(t, err)
	}

	t.Run("generate advice for conservative user", func(t *testing.T) {
		user := domain.User{
			ID:            userID,
			Email:         "conservative@example.com",
			FirstName:     "Conservative",
			LastName:      "User",
			RiskTolerance: "conservative",
		}

		advice, err := advisorService.GenerateAdvice(user)
		assert.NoError(t, err)
		assert.NotNil(t, advice)

		// Monthly savings should be (3000 - 2000) / 3 = 333.33
		expectedSavings := 1000.00 / 3.0
		assert.InDelta(t, expectedSavings, advice.MonthlySavings, 0.01)
		assert.Equal(t, "conservative", advice.Risk)
		assert.Len(t, advice.Recommendations, 2)

		// Conservative allocation: 70% SPY, 30% BTC
		assert.Equal(t, "SPY", advice.Recommendations[0].Asset)
		assert.InDelta(t, expectedSavings*0.7, advice.Recommendations[0].Amount, 0.01)
		assert.Equal(t, 70.0, advice.Recommendations[0].Percent)

		assert.Equal(t, "BTC", advice.Recommendations[1].Asset)
		assert.InDelta(t, expectedSavings*0.3, advice.Recommendations[1].Amount, 0.01)
		assert.Equal(t, 30.0, advice.Recommendations[1].Percent)
	})

	t.Run("generate advice for aggressive user", func(t *testing.T) {
		user := domain.User{
			ID:            userID,
			Email:         "aggressive@example.com",
			FirstName:     "Aggressive",
			LastName:      "User",
			RiskTolerance: "aggressive",
		}

		advice, err := advisorService.GenerateAdvice(user)
		assert.NoError(t, err)
		assert.NotNil(t, advice)

		expectedSavings := 1000.00 / 3.0
		assert.InDelta(t, expectedSavings, advice.MonthlySavings, 0.01)
		assert.Equal(t, "aggressive", advice.Risk)
		assert.Len(t, advice.Recommendations, 2)

		// Aggressive allocation: 30% SPY, 70% BTC
		assert.Equal(t, "SPY", advice.Recommendations[0].Asset)
		assert.InDelta(t, expectedSavings*0.3, advice.Recommendations[0].Amount, 0.01)
		assert.Equal(t, 30.0, advice.Recommendations[0].Percent)

		assert.Equal(t, "BTC", advice.Recommendations[1].Asset)
		assert.InDelta(t, expectedSavings*0.7, advice.Recommendations[1].Amount, 0.01)
		assert.Equal(t, 70.0, advice.Recommendations[1].Percent)
	})

	t.Run("generate advice for moderate user", func(t *testing.T) {
		user := domain.User{
			ID:            userID,
			Email:         "moderate@example.com",
			FirstName:     "Moderate",
			LastName:      "User",
			RiskTolerance: "moderate",
		}

		advice, err := advisorService.GenerateAdvice(user)
		assert.NoError(t, err)
		assert.NotNil(t, advice)

		expectedSavings := 1000.00 / 3.0
		assert.InDelta(t, expectedSavings, advice.MonthlySavings, 0.01)
		assert.Equal(t, "moderate", advice.Risk)
		assert.Len(t, advice.Recommendations, 2)

		// Moderate allocation: 50% SPY, 50% BTC
		assert.Equal(t, "SPY", advice.Recommendations[0].Asset)
		assert.InDelta(t, expectedSavings*0.5, advice.Recommendations[0].Amount, 0.01)
		assert.Equal(t, 50.0, advice.Recommendations[0].Percent)

		assert.Equal(t, "BTC", advice.Recommendations[1].Asset)
		assert.InDelta(t, expectedSavings*0.5, advice.Recommendations[1].Amount, 0.01)
		assert.Equal(t, 50.0, advice.Recommendations[1].Percent)
	})

	t.Run("generate advice for unknown risk tolerance", func(t *testing.T) {
		user := domain.User{
			ID:            userID,
			Email:         "unknown@example.com",
			FirstName:     "Unknown",
			LastName:      "User",
			RiskTolerance: "unknown",
		}

		advice, err := advisorService.GenerateAdvice(user)
		assert.NoError(t, err)
		assert.NotNil(t, advice)

		// Should default to moderate allocation
		expectedSavings := 1000.00 / 3.0
		assert.InDelta(t, expectedSavings, advice.MonthlySavings, 0.01)
		assert.Equal(t, "unknown", advice.Risk)
		assert.Len(t, advice.Recommendations, 2)

		// Default (moderate) allocation: 50% SPY, 50% BTC
		assert.Equal(t, "SPY", advice.Recommendations[0].Asset)
		assert.InDelta(t, expectedSavings*0.5, advice.Recommendations[0].Amount, 0.01)
		assert.Equal(t, 50.0, advice.Recommendations[0].Percent)

		assert.Equal(t, "BTC", advice.Recommendations[1].Asset)
		assert.InDelta(t, expectedSavings*0.5, advice.Recommendations[1].Amount, 0.01)
		assert.Equal(t, 50.0, advice.Recommendations[1].Percent)
	})

	t.Run("generate advice with negative savings", func(t *testing.T) {
		// Clean up previous transactions
		db.Where("user_id = ?", userID).Delete(&domain.Transaction{})

		// Create transactions that result in negative savings
		negativeTransactions := []*domain.Transaction{
			{
				UserID:      userID,
				CategoryID:  incomeCategoryID,
				Amount:      1000.00,
				Type:        "income",
				Description: "Low Income",
				Date:        now.AddDate(0, -1, 0),
			},
			{
				UserID:      userID,
				CategoryID:  expenseCategoryID,
				Amount:      2000.00,
				Type:        "expense",
				Description: "High Expenses",
				Date:        now.AddDate(0, -1, 0),
			},
		}

		for _, tx := range negativeTransactions {
			err := db.Create(tx).Error
			require.NoError(t, err)
		}

		user := domain.User{
			ID:            userID,
			Email:         "negative@example.com",
			FirstName:     "Negative",
			LastName:      "User",
			RiskTolerance: "moderate",
		}

		advice, err := advisorService.GenerateAdvice(user)
		assert.NoError(t, err)
		assert.NotNil(t, advice)

		// Monthly savings should be negative: (1000 - 2000) / 3 = -333.33
		expectedSavings := -1000.00 / 3.0
		assert.InDelta(t, expectedSavings, advice.MonthlySavings, 0.01)
		assert.Equal(t, "moderate", advice.Risk)
		assert.Len(t, advice.Recommendations, 2)

		// Even with negative savings, recommendations should be calculated
		assert.Equal(t, "SPY", advice.Recommendations[0].Asset)
		assert.InDelta(t, expectedSavings*0.5, advice.Recommendations[0].Amount, 0.01)
		assert.Equal(t, 50.0, advice.Recommendations[0].Percent)

		assert.Equal(t, "BTC", advice.Recommendations[1].Asset)
		assert.InDelta(t, expectedSavings*0.5, advice.Recommendations[1].Amount, 0.01)
		assert.Equal(t, 50.0, advice.Recommendations[1].Percent)
	})
}
