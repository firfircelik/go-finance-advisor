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

func setupTransactionTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&domain.User{}, &domain.Category{}, &domain.Transaction{})
	require.NoError(t, err)

	return db
}

func createTestData(t *testing.T, db *gorm.DB) (uint, uint, uint) {
	// Create test user
	user := &domain.User{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
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

func TestTransactionService_Create(t *testing.T) {
	db := setupTransactionTestDB(t)
	txService := &TransactionService{DB: db}
	userID, categoryID, _ := createTestData(t, db)

	t.Run("successful creation", func(t *testing.T) {
		transaction := &domain.Transaction{
			UserID:      userID,
			CategoryID:  categoryID,
			Amount:      100.50,
			Description: "Test transaction",
			Type:        "income",
			Date:        time.Now(),
		}

		err := txService.Create(transaction)
		assert.NoError(t, err)
		assert.NotZero(t, transaction.ID)
	})

	t.Run("create with invalid user ID", func(t *testing.T) {
		transaction := &domain.Transaction{
			UserID:      99999, // Non-existent user
			CategoryID:  categoryID,
			Amount:      100.50,
			Description: "Test transaction",
			Type:        "income",
			Date:        time.Now(),
		}

		// GORM doesn't enforce foreign key constraints by default in SQLite
		err := txService.Create(transaction)
		assert.NoError(t, err) // Will succeed but with invalid reference
	})
}

func TestTransactionService_List(t *testing.T) {
	db := setupTransactionTestDB(t)
	txService := &TransactionService{DB: db}
	userID, categoryID, _ := createTestData(t, db)

	// Create test transactions
	transactions := []*domain.Transaction{
		{
			UserID:      userID,
			CategoryID:  categoryID,
			Amount:      100.00,
			Description: "Transaction 1",
			Type:        "income",
			Date:        time.Now().AddDate(0, 0, -2),
		},
		{
			UserID:      userID,
			CategoryID:  categoryID,
			Amount:      50.00,
			Description: "Transaction 2",
			Type:        "expense",
			Date:        time.Now().AddDate(0, 0, -1),
		},
	}

	for _, tx := range transactions {
		err := txService.Create(tx)
		require.NoError(t, err)
	}

	t.Run("list user transactions", func(t *testing.T) {
		result, err := txService.List(userID)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		// Should be ordered by date DESC
		assert.Equal(t, "Transaction 2", result[0].Description)
		assert.Equal(t, "Transaction 1", result[1].Description)
	})

	t.Run("list for non-existent user", func(t *testing.T) {
		result, err := txService.List(99999)
		assert.NoError(t, err)
		assert.Len(t, result, 0)
	})
}

func TestTransactionService_ListWithFilters(t *testing.T) {
	db := setupTransactionTestDB(t)
	txService := &TransactionService{DB: db}
	userID, incomeCategoryID, expenseCategoryID := createTestData(t, db)

	// Create test transactions
	now := time.Now()
	transactions := []*domain.Transaction{
		{
			UserID:      userID,
			CategoryID:  incomeCategoryID,
			Amount:      1000.00,
			Description: "Salary",
			Type:        "income",
			Date:        now.AddDate(0, 0, -5),
		},
		{
			UserID:      userID,
			CategoryID:  expenseCategoryID,
			Amount:      50.00,
			Description: "Lunch",
			Type:        "expense",
			Date:        now.AddDate(0, 0, -3),
		},
		{
			UserID:      userID,
			CategoryID:  expenseCategoryID,
			Amount:      30.00,
			Description: "Coffee",
			Type:        "expense",
			Date:        now.AddDate(0, 0, -1),
		},
	}

	for _, tx := range transactions {
		err := txService.Create(tx)
		require.NoError(t, err)
	}

	t.Run("filter by type", func(t *testing.T) {
		txType := "expense"
		result, err := txService.ListWithFilters(userID, &txType, nil, nil, nil, 10, 0)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		for _, tx := range result {
			assert.Equal(t, "expense", tx.Type)
		}
	})

	t.Run("filter by category", func(t *testing.T) {
		result, err := txService.ListWithFilters(userID, nil, &incomeCategoryID, nil, nil, 10, 0)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "Salary", result[0].Description)
	})

	t.Run("filter by date range", func(t *testing.T) {
		startDate := now.AddDate(0, 0, -4)
		endDate := now.AddDate(0, 0, -2)
		result, err := txService.ListWithFilters(userID, nil, nil, &startDate, &endDate, 10, 0)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "Lunch", result[0].Description)
	})

	t.Run("pagination", func(t *testing.T) {
		result, err := txService.ListWithFilters(userID, nil, nil, nil, nil, 2, 0)
		assert.NoError(t, err)
		assert.Len(t, result, 2)

		result, err = txService.ListWithFilters(userID, nil, nil, nil, nil, 2, 2)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})
}

func TestTransactionService_GetByID(t *testing.T) {
	db := setupTransactionTestDB(t)
	txService := &TransactionService{DB: db}
	userID, categoryID, _ := createTestData(t, db)

	// Create test transaction
	transaction := &domain.Transaction{
		UserID:      userID,
		CategoryID:  categoryID,
		Amount:      100.00,
		Description: "Test transaction",
		Type:        "income",
		Date:        time.Now(),
	}
	err := txService.Create(transaction)
	require.NoError(t, err)

	t.Run("get existing transaction", func(t *testing.T) {
		result, err := txService.GetByID(transaction.ID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, transaction.ID, result.ID)
		assert.Equal(t, "Test transaction", result.Description)
		assert.NotNil(t, result.Category) // Should be preloaded
	})

	t.Run("get non-existent transaction", func(t *testing.T) {
		result, err := txService.GetByID(99999)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestTransactionService_Update(t *testing.T) {
	db := setupTransactionTestDB(t)
	txService := &TransactionService{DB: db}
	userID, categoryID, _ := createTestData(t, db)

	// Create test transaction
	transaction := &domain.Transaction{
		UserID:      userID,
		CategoryID:  categoryID,
		Amount:      100.00,
		Description: "Original description",
		Type:        "income",
		Date:        time.Now(),
	}
	err := txService.Create(transaction)
	require.NoError(t, err)

	t.Run("update transaction", func(t *testing.T) {
		transaction.Amount = 150.00
		transaction.Description = "Updated description"

		err := txService.Update(transaction)
		assert.NoError(t, err)

		// Verify update
		updated, err := txService.GetByID(transaction.ID)
		assert.NoError(t, err)
		assert.Equal(t, 150.00, updated.Amount)
		assert.Equal(t, "Updated description", updated.Description)
	})
}

func TestTransactionService_Delete(t *testing.T) {
	db := setupTransactionTestDB(t)
	txService := &TransactionService{DB: db}
	userID, categoryID, _ := createTestData(t, db)

	// Create test transaction
	transaction := &domain.Transaction{
		UserID:      userID,
		CategoryID:  categoryID,
		Amount:      100.00,
		Description: "To be deleted",
		Type:        "income",
		Date:        time.Now(),
	}
	err := txService.Create(transaction)
	require.NoError(t, err)

	t.Run("delete existing transaction", func(t *testing.T) {
		err := txService.Delete(transaction.ID)
		assert.NoError(t, err)

		// Verify deletion
		_, err = txService.GetByID(transaction.ID)
		assert.Error(t, err)
	})

	t.Run("delete non-existent transaction", func(t *testing.T) {
		err := txService.Delete(99999)
		assert.NoError(t, err) // GORM doesn't return error for non-existent records
	})
}

func TestTransactionService_GetTotalByType(t *testing.T) {
	db := setupTransactionTestDB(t)
	txService := &TransactionService{DB: db}
	userID, categoryID, _ := createTestData(t, db)

	now := time.Now()
	startDate := now.AddDate(0, 0, -7)
	endDate := now

	// Create test transactions
	transactions := []*domain.Transaction{
		{
			UserID:      userID,
			CategoryID:  categoryID,
			Amount:      100.00,
			Type:        "income",
			Date:        now.AddDate(0, 0, -3),
		},
		{
			UserID:      userID,
			CategoryID:  categoryID,
			Amount:      200.00,
			Type:        "income",
			Date:        now.AddDate(0, 0, -2),
		},
		{
			UserID:      userID,
			CategoryID:  categoryID,
			Amount:      50.00,
			Type:        "expense",
			Date:        now.AddDate(0, 0, -1),
		},
	}

	for _, tx := range transactions {
		err := txService.Create(tx)
		require.NoError(t, err)
	}

	t.Run("get income total", func(t *testing.T) {
		total, err := txService.GetTotalByType(userID, "income", startDate, endDate)
		assert.NoError(t, err)
		assert.Equal(t, 300.00, total)
	})

	t.Run("get expense total", func(t *testing.T) {
		total, err := txService.GetTotalByType(userID, "expense", startDate, endDate)
		assert.NoError(t, err)
		assert.Equal(t, 50.00, total)
	})

	t.Run("no transactions in range", func(t *testing.T) {
		futureStart := now.AddDate(0, 0, 1)
		futureEnd := now.AddDate(0, 0, 7)
		total, err := txService.GetTotalByType(userID, "income", futureStart, futureEnd)
		assert.NoError(t, err)
		assert.Equal(t, 0.00, total)
	})
}

func TestTransactionService_GetDailyAverages(t *testing.T) {
	db := setupTransactionTestDB(t)
	txService := &TransactionService{DB: db}
	userID, categoryID, _ := createTestData(t, db)

	now := time.Now()
	startDate := now.AddDate(0, 0, -9) // 10 days period
	endDate := now

	// Create test transactions
	transactions := []*domain.Transaction{
		{
			UserID:      userID,
			CategoryID:  categoryID,
			Amount:      1000.00, // Total income: 1000
			Type:        "income",
			Date:        now.AddDate(0, 0, -5),
		},
		{
			UserID:      userID,
			CategoryID:  categoryID,
			Amount:      300.00, // Total expense: 300
			Type:        "expense",
			Date:        now.AddDate(0, 0, -3),
		},
	}

	for _, tx := range transactions {
		err := txService.Create(tx)
		require.NoError(t, err)
	}

	t.Run("calculate daily averages", func(t *testing.T) {
		averages, err := txService.GetDailyAverages(userID, startDate, endDate)
		assert.NoError(t, err)
		assert.Contains(t, averages, "daily_income")
		assert.Contains(t, averages, "daily_expense")
		assert.Contains(t, averages, "daily_net")

		// 10 days period
		assert.Equal(t, 100.00, averages["daily_income"])  // 1000/10
		assert.Equal(t, 30.00, averages["daily_expense"])   // 300/10
		assert.Equal(t, 70.00, averages["daily_net"])       // (1000-300)/10
	})

	t.Run("same day range", func(t *testing.T) {
		sameDay := now
		averages, err := txService.GetDailyAverages(userID, sameDay, sameDay)
		assert.NoError(t, err)
		// Should handle single day correctly
		assert.NotNil(t, averages)
	})
}