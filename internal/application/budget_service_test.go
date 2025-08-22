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

func setupBudgetTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&domain.User{}, &domain.Category{}, &domain.Budget{}, &domain.Transaction{})
	require.NoError(t, err)

	return db
}

func createBudgetTestData(t *testing.T, db *gorm.DB) (uint, uint) {
	// Create test user
	user := &domain.User{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	// Create test category
	category := &domain.Category{
		Name: "Food",
		Type: "expense",
	}
	err = db.Create(category).Error
	require.NoError(t, err)

	return user.ID, category.ID
}

func TestBudgetService_CreateBudget(t *testing.T) {
	db := setupBudgetTestDB(t)
	budgetService := &BudgetService{DB: db}
	userID, categoryID := createBudgetTestData(t, db)

	t.Run("successful creation with defaults", func(t *testing.T) {
		budget := &domain.Budget{
			UserID:     userID,
			CategoryID: categoryID,
			Amount:     500.00,
		}

		err := budgetService.CreateBudget(budget)
		assert.NoError(t, err)
		assert.NotZero(t, budget.ID)
		assert.Equal(t, "monthly", budget.Period)
		assert.True(t, budget.IsActive)
		assert.Equal(t, 500.00, budget.Remaining)
		assert.False(t, budget.StartDate.IsZero())
		assert.False(t, budget.EndDate.IsZero())
	})

	t.Run("creation with custom period", func(t *testing.T) {
		startDate := time.Now().Truncate(24 * time.Hour)
		budget := &domain.Budget{
			UserID:     userID,
			CategoryID: categoryID + 1, // Different category to avoid conflict
			Amount:     1000.00,
			Period:     "weekly",
			StartDate:  startDate,
		}

		// Create another category for this test
		category2 := &domain.Category{Name: "Transport", Type: "expense"}
		db.Create(category2)
		budget.CategoryID = category2.ID

		err := budgetService.CreateBudget(budget)
		assert.NoError(t, err)
		assert.Equal(t, "weekly", budget.Period)
		expectedEndDate := startDate.AddDate(0, 0, 7)
		assert.Equal(t, expectedEndDate.Format("2006-01-02"), budget.EndDate.Format("2006-01-02"))
	})

	t.Run("duplicate budget error", func(t *testing.T) {
		// Clean up previous test data
		db.Where("user_id = ?", userID).Delete(&domain.Budget{})

		// Create first budget
		budget1 := &domain.Budget{
			UserID:     userID,
			CategoryID: categoryID,
			Amount:     300.00,
			StartDate:  time.Now().Truncate(24 * time.Hour),
			EndDate:    time.Now().AddDate(0, 1, 0),
		}
		err := budgetService.CreateBudget(budget1)
		require.NoError(t, err)

		// Try to create overlapping budget
		budget2 := &domain.Budget{
			UserID:     userID,
			CategoryID: categoryID,
			Amount:     400.00,
			StartDate:  time.Now().AddDate(0, 0, 15), // Overlaps with first budget
			EndDate:    time.Now().AddDate(0, 2, 0),
		}
		err = budgetService.CreateBudget(budget2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "budget already exists")
	})
}

func TestBudgetService_UpdateBudget(t *testing.T) {
	db := setupBudgetTestDB(t)
	budgetService := &BudgetService{DB: db}
	userID, categoryID := createBudgetTestData(t, db)

	// Create test budget
	budget := &domain.Budget{
		UserID:     userID,
		CategoryID: categoryID,
		Amount:     500.00,
		Period:     "monthly",
	}
	err := budgetService.CreateBudget(budget)
	require.NoError(t, err)

	t.Run("successful update", func(t *testing.T) {
		updates := &domain.Budget{
			Amount:   750.00,
			Period:   "weekly",
			IsActive: false,
		}

		err := budgetService.UpdateBudget(budget.ID, updates)
		assert.NoError(t, err)

		// Verify update
		updated, err := budgetService.GetBudgetByID(budget.ID)
		assert.NoError(t, err)
		assert.Equal(t, 750.00, updated.Amount)
		assert.Equal(t, "weekly", updated.Period)
		assert.False(t, updated.IsActive)
	})

	t.Run("update non-existent budget", func(t *testing.T) {
		updates := &domain.Budget{Amount: 1000.00}
		err := budgetService.UpdateBudget(99999, updates)
		assert.Error(t, err)
	})
}

func TestBudgetService_GetBudgetsByUser(t *testing.T) {
	db := setupBudgetTestDB(t)
	budgetService := &BudgetService{DB: db}
	userID, categoryID := createBudgetTestData(t, db)

	// Create test budgets
	budgets := []*domain.Budget{
		{
			UserID:     userID,
			CategoryID: categoryID,
			Amount:     500.00,
			Period:     "monthly",
		},
		{
			UserID:     userID,
			CategoryID: categoryID,
			Amount:     200.00,
			Period:     "weekly",
			StartDate:  time.Now().AddDate(0, 1, 0), // Future budget
			EndDate:    time.Now().AddDate(0, 1, 7),
		},
	}

	for _, budget := range budgets {
		err := budgetService.CreateBudget(budget)
		require.NoError(t, err)
	}

	t.Run("get all user budgets", func(t *testing.T) {
		result, err := budgetService.GetBudgetsByUser(userID)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		// Should include Category preload
		for _, budget := range result {
			assert.NotNil(t, budget.Category)
			assert.Equal(t, "Food", budget.Category.Name)
		}
	})

	t.Run("get budgets for non-existent user", func(t *testing.T) {
		result, err := budgetService.GetBudgetsByUser(99999)
		assert.NoError(t, err)
		assert.Len(t, result, 0)
	})
}

func TestBudgetService_GetActiveBudgetsByUser(t *testing.T) {
	db := setupBudgetTestDB(t)
	budgetService := &BudgetService{DB: db}
	userID, categoryID := createBudgetTestData(t, db)

	now := time.Now()

	// Create test budgets
	budgets := []*domain.Budget{
		{
			UserID:     userID,
			CategoryID: categoryID,
			Amount:     500.00,
			StartDate:  now.AddDate(0, 0, -7),
			EndDate:    now.AddDate(0, 0, 7), // Active
			IsActive:   true,
		},
		{
			UserID:     userID,
			CategoryID: categoryID,
			Amount:     300.00,
			StartDate:  now.AddDate(0, 0, -30),
			EndDate:    now.AddDate(0, 0, -1), // Expired
			IsActive:   true,
		},
		{
			UserID:     userID,
			CategoryID: categoryID,
			Amount:     200.00,
			StartDate:  now.AddDate(0, 0, -7),
			EndDate:    now.AddDate(0, 0, 7),
			IsActive:   false, // Inactive
		},
	}

	for _, budget := range budgets {
		err := db.Create(budget).Error
		require.NoError(t, err)
	}

	t.Run("get only active budgets", func(t *testing.T) {
		// Clean up previous test data
		db.Where("user_id = ?", userID).Delete(&domain.Budget{})

		// Create test budgets
		budgets := []*domain.Budget{
			{
				UserID:     userID,
				CategoryID: categoryID,
				Amount:     500.00,
				StartDate:  now.AddDate(0, 0, -7),
				EndDate:    now.AddDate(0, 0, 7),
				IsActive:   true, // Active and current
			},
		}

		for _, budget := range budgets {
			err := db.Create(budget).Error
			require.NoError(t, err)
		}

		result, err := budgetService.GetActiveBudgetsByUser(userID)
		assert.NoError(t, err)
		assert.Len(t, result, 1) // Only the first budget should be returned
		assert.Equal(t, 500.00, result[0].Amount)
		assert.True(t, result[0].IsActive)
	})
}

func TestBudgetService_UpdateBudgetSpending(t *testing.T) {
	db := setupBudgetTestDB(t)
	budgetService := &BudgetService{DB: db}
	userID, categoryID := createBudgetTestData(t, db)

	now := time.Now()

	// Create test budget
	budget := &domain.Budget{
		UserID:     userID,
		CategoryID: categoryID,
		Amount:     500.00,
		StartDate:  now.AddDate(0, 0, -7),
		EndDate:    now.AddDate(0, 0, 7),
		IsActive:   true,
		Spent:      100.00,
		Remaining:  400.00,
	}
	err := db.Create(budget).Error
	require.NoError(t, err)

	t.Run("update spending within budget period", func(t *testing.T) {
		transactionDate := now.AddDate(0, 0, -2) // Within budget period
		err := budgetService.UpdateBudgetSpending(userID, categoryID, 50.00, transactionDate)
		assert.NoError(t, err)

		// Verify budget was updated
		updated, err := budgetService.GetBudgetByID(budget.ID)
		assert.NoError(t, err)
		assert.Equal(t, 150.00, updated.Spent)     // 100 + 50
		assert.Equal(t, 350.00, updated.Remaining) // 500 - 150
	})

	t.Run("transaction outside budget period", func(t *testing.T) {
		transactionDate := now.AddDate(0, 0, 10) // Outside budget period
		err := budgetService.UpdateBudgetSpending(userID, categoryID, 25.00, transactionDate)
		assert.NoError(t, err)

		// Budget should not be updated
		updated, err := budgetService.GetBudgetByID(budget.ID)
		assert.NoError(t, err)
		assert.Equal(t, 150.00, updated.Spent) // Should remain the same
	})
}

func TestBudgetService_GetBudgetSummary(t *testing.T) {
	db := setupBudgetTestDB(t)
	budgetService := &BudgetService{DB: db}
	userID, categoryID := createBudgetTestData(t, db)

	now := time.Now()

	// Create test budgets
	budgets := []*domain.Budget{
		{
			UserID:     userID,
			CategoryID: categoryID,
			Amount:     500.00,
			Spent:      300.00, // 60% used
			StartDate:  now.AddDate(0, 0, -7),
			EndDate:    now.AddDate(0, 0, 7),
			IsActive:   true,
		},
		{
			UserID:     userID,
			CategoryID: categoryID,
			Amount:     300.00,
			Spent:      350.00, // Over budget
			StartDate:  now.AddDate(0, 0, -7),
			EndDate:    now.AddDate(0, 0, 7),
			IsActive:   true,
		},
	}

	for _, budget := range budgets {
		err := db.Create(budget).Error
		require.NoError(t, err)
	}

	t.Run("calculate budget summary", func(t *testing.T) {
		summary, err := budgetService.GetBudgetSummary(userID)
		assert.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, 800.00, summary.TotalBudget)     // 500 + 300
		assert.Equal(t, 650.00, summary.TotalSpent)      // 300 + 350
		assert.Equal(t, 150.00, summary.TotalRemaining)  // 800 - 650
		assert.Equal(t, 81.25, summary.PercentageUsed)   // (650/800)*100
		assert.Equal(t, "warning", summary.BudgetStatus) // >80% but <100%
	})

	t.Run("over budget status", func(t *testing.T) {
		// Clean up previous test data
		db.Where("user_id = ?", userID).Delete(&domain.Budget{})

		// Create budget that's 100% over
		overBudget := &domain.Budget{
			UserID:     userID,
			CategoryID: categoryID,
			Amount:     100.00,
			Spent:      200.00, // 200% used
			StartDate:  now.AddDate(0, 0, -7),
			EndDate:    now.AddDate(0, 0, 7),
			IsActive:   true,
		}
		db.Create(overBudget)

		summary, err := budgetService.GetBudgetSummary(userID)
		assert.NoError(t, err)
		assert.Equal(t, "over_budget", summary.BudgetStatus)
	})
}

func TestBudgetService_RefreshBudgetSpending(t *testing.T) {
	db := setupBudgetTestDB(t)
	budgetService := &BudgetService{DB: db}
	userID, categoryID := createBudgetTestData(t, db)

	now := time.Now()

	// Create test budget
	budget := &domain.Budget{
		UserID:     userID,
		CategoryID: categoryID,
		Amount:     500.00,
		Spent:      0.00, // Will be recalculated
		StartDate:  now.AddDate(0, 0, -7),
		EndDate:    now.AddDate(0, 0, 7),
		IsActive:   true,
	}
	err := db.Create(budget).Error
	require.NoError(t, err)

	// Create test transactions
	transactions := []*domain.Transaction{
		{
			UserID:     userID,
			CategoryID: categoryID,
			Amount:     100.00,
			Type:       "expense",
			Date:       now.AddDate(0, 0, -3), // Within budget period
		},
		{
			UserID:     userID,
			CategoryID: categoryID,
			Amount:     50.00,
			Type:       "expense",
			Date:       now.AddDate(0, 0, -1), // Within budget period
		},
		{
			UserID:     userID,
			CategoryID: categoryID,
			Amount:     25.00,
			Type:       "expense",
			Date:       now.AddDate(0, 0, -10), // Outside budget period
		},
	}

	for _, tx := range transactions {
		err := db.Create(tx).Error
		require.NoError(t, err)
	}

	t.Run("refresh budget spending", func(t *testing.T) {
		err := budgetService.RefreshBudgetSpending(userID)
		assert.NoError(t, err)

		// Verify budget was updated with correct spent amount
		updated, err := budgetService.GetBudgetByID(budget.ID)
		assert.NoError(t, err)
		assert.Equal(t, 150.00, updated.Spent)     // Only transactions within period: 100 + 50
		assert.Equal(t, 350.00, updated.Remaining) // 500 - 150
	})
}

func TestBudgetService_DeleteBudget(t *testing.T) {
	db := setupBudgetTestDB(t)
	budgetService := &BudgetService{DB: db}
	userID, categoryID := createBudgetTestData(t, db)

	// Create test budget
	budget := &domain.Budget{
		UserID:     userID,
		CategoryID: categoryID,
		Amount:     500.00,
	}
	err := budgetService.CreateBudget(budget)
	require.NoError(t, err)

	t.Run("delete existing budget", func(t *testing.T) {
		err := budgetService.DeleteBudget(budget.ID)
		assert.NoError(t, err)

		// Verify deletion
		_, err = budgetService.GetBudgetByID(budget.ID)
		assert.Error(t, err)
	})

	t.Run("delete non-existent budget", func(t *testing.T) {
		err := budgetService.DeleteBudget(99999)
		assert.NoError(t, err) // GORM doesn't return error for non-existent records
	})
}
