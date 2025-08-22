package application

import (
	"testing"
	"time"

	"go-finance-advisor/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupReportsTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto migrate the schema
	err = db.AutoMigrate(&domain.User{}, &domain.Category{}, &domain.Transaction{}, &domain.Budget{}, &domain.FinancialReport{})
	if err != nil {
		panic("failed to migrate database")
	}

	return db
}

func createReportsTestData(db *gorm.DB) (uint, uint, uint) {
	// Create test user
	user := domain.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Password:  "hashedpassword",
	}
	db.Create(&user)

	// Create test categories
	incomeCategory := domain.Category{
		Name:        "Salary",
		Type:        "income",
		Description: "Monthly salary",
	}
	db.Create(&incomeCategory)

	expenseCategory := domain.Category{
		Name:        "Groceries",
		Type:        "expense",
		Description: "Food expenses",
	}
	db.Create(&expenseCategory)

	// Create test transactions for January 2024
	transactions := []domain.Transaction{
		{
			UserID:      user.ID,
			CategoryID:  incomeCategory.ID,
			Amount:      5000.00,
			Type:        "income",
			Description: "January Salary",
			Date:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			UserID:      user.ID,
			CategoryID:  expenseCategory.ID,
			Amount:      300.00,
			Type:        "expense",
			Description: "Weekly groceries",
			Date:        time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
		},
		{
			UserID:      user.ID,
			CategoryID:  expenseCategory.ID,
			Amount:      250.00,
			Type:        "expense",
			Description: "More groceries",
			Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			UserID:      user.ID,
			CategoryID:  incomeCategory.ID,
			Amount:      1000.00,
			Type:        "income",
			Description: "Bonus",
			Date:        time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tx := range transactions {
		db.Create(&tx)
	}

	// Create test budget
	budget := domain.Budget{
		UserID:     user.ID,
		CategoryID: expenseCategory.ID,
		Amount:     600.00,
		Period:     "monthly",
		StartDate:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:    time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		IsActive:   true,
	}
	db.Create(&budget)

	return user.ID, incomeCategory.ID, expenseCategory.ID
}

func TestReportsService_GenerateMonthlyReport(t *testing.T) {
	db := setupReportsTestDB()
	service := NewReportsService(db)
	userID, _, _ := createReportsTestData(db)

	t.Run("generate monthly report with data", func(t *testing.T) {
		report, err := service.GenerateMonthlyReport(userID, 2024, 1)

		require.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, "monthly", report.ReportType)
		assert.Equal(t, userID, report.UserID)
		assert.Equal(t, 6000.00, report.TotalIncome)  // 5000 + 1000
		assert.Equal(t, 550.00, report.TotalExpenses) // 300 + 250
		assert.Equal(t, 5450.00, report.NetIncome)    // 6000 - 550
		assert.Equal(t, 4, report.TransactionCount)
		assert.Greater(t, report.SavingsRate, 90.0) // Should be around 90.8%
		assert.NotEmpty(t, report.CategoryBreakdown)
		assert.NotZero(t, report.GeneratedAt)
	})

	t.Run("generate monthly report for period with no data", func(t *testing.T) {
		report, err := service.GenerateMonthlyReport(userID, 2024, 2)

		require.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, "monthly", report.ReportType)
		assert.Equal(t, 0.0, report.TotalIncome)
		assert.Equal(t, 0.0, report.TotalExpenses)
		assert.Equal(t, 0.0, report.NetIncome)
		assert.Equal(t, 0, report.TransactionCount)
		assert.Equal(t, 0.0, report.SavingsRate)
	})

	t.Run("generate monthly report for non-existent user", func(t *testing.T) {
		report, err := service.GenerateMonthlyReport(99999, 2024, 1)

		require.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, 0, report.TransactionCount)
	})
}

func TestReportsService_GenerateQuarterlyReport(t *testing.T) {
	db := setupReportsTestDB()
	service := NewReportsService(db)
	userID, _, _ := createReportsTestData(db)

	t.Run("generate Q1 quarterly report", func(t *testing.T) {
		report, err := service.GenerateQuarterlyReport(userID, 2024, 1)

		require.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, "quarterly", report.ReportType)
		assert.Equal(t, userID, report.UserID)
		assert.Equal(t, 6000.00, report.TotalIncome)
		assert.Equal(t, 550.00, report.TotalExpenses)
		assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), report.StartDate)
		assert.True(t, report.EndDate.After(time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC)))
	})

	t.Run("generate quarterly report with invalid quarter", func(t *testing.T) {
		report, err := service.GenerateQuarterlyReport(userID, 2024, 5)

		assert.Error(t, err)
		assert.Nil(t, report)
		assert.Contains(t, err.Error(), "invalid quarter")
	})

	t.Run("generate Q2 quarterly report", func(t *testing.T) {
		report, err := service.GenerateQuarterlyReport(userID, 2024, 2)

		require.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, "quarterly", report.ReportType)
		assert.Equal(t, time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC), report.StartDate)
	})

	t.Run("generate Q3 quarterly report", func(t *testing.T) {
		report, err := service.GenerateQuarterlyReport(userID, 2024, 3)

		require.NoError(t, err)
		assert.Equal(t, time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC), report.StartDate)
	})

	t.Run("generate Q4 quarterly report", func(t *testing.T) {
		report, err := service.GenerateQuarterlyReport(userID, 2024, 4)

		require.NoError(t, err)
		assert.Equal(t, time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC), report.StartDate)
	})
}

func TestReportsService_GenerateYearlyReport(t *testing.T) {
	db := setupReportsTestDB()
	service := NewReportsService(db)
	userID, _, _ := createReportsTestData(db)

	t.Run("generate yearly report with data", func(t *testing.T) {
		report, err := service.GenerateYearlyReport(userID, 2024)

		require.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, "yearly", report.ReportType)
		assert.Equal(t, userID, report.UserID)
		assert.Equal(t, 6000.00, report.TotalIncome)
		assert.Equal(t, 550.00, report.TotalExpenses)
		assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), report.StartDate)
		assert.True(t, report.EndDate.Year() == 2024)
		assert.True(t, report.EndDate.Month() == 12)
	})

	t.Run("generate yearly report for different year", func(t *testing.T) {
		report, err := service.GenerateYearlyReport(userID, 2023)

		require.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, "yearly", report.ReportType)
		assert.Equal(t, 0, report.TransactionCount) // No data for 2023
		assert.Equal(t, time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), report.StartDate)
	})
}

func TestReportsService_GenerateCustomReport(t *testing.T) {
	db := setupReportsTestDB()
	service := NewReportsService(db)
	userID, _, _ := createReportsTestData(db)

	t.Run("generate custom report for specific date range", func(t *testing.T) {
		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 1, 15, 23, 59, 59, 0, time.UTC)

		report, err := service.GenerateCustomReport(userID, startDate, endDate)

		require.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, "custom", report.ReportType)
		assert.Equal(t, userID, report.UserID)
		assert.Equal(t, startDate, report.StartDate)
		assert.Equal(t, endDate, report.EndDate)
		// Should include transactions from Jan 1, 5, and 15 (3 transactions)
		assert.Equal(t, 3, report.TransactionCount)
		assert.Equal(t, 5000.00, report.TotalIncome)  // 5000 from Jan 1
		assert.Equal(t, 550.00, report.TotalExpenses) // 300 + 250
	})

	t.Run("generate custom report for single day", func(t *testing.T) {
		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC)

		report, err := service.GenerateCustomReport(userID, startDate, endDate)

		require.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, 1, report.TransactionCount) // Only Jan 1 transaction
		assert.Equal(t, 5000.00, report.TotalIncome)
		assert.Equal(t, 0.0, report.TotalExpenses)
	})

	t.Run("generate custom report for range with no data", func(t *testing.T) {
		startDate := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 2, 28, 23, 59, 59, 0, time.UTC)

		report, err := service.GenerateCustomReport(userID, startDate, endDate)

		require.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, 0, report.TransactionCount)
		assert.Equal(t, 0.0, report.TotalIncome)
		assert.Equal(t, 0.0, report.TotalExpenses)
	})
}

func TestReportsService_Integration(t *testing.T) {
	db := setupReportsTestDB()
	service := NewReportsService(db)
	userID, incomeID, expenseID := createReportsTestData(db)

	t.Run("report includes category breakdown", func(t *testing.T) {
		report, err := service.GenerateMonthlyReport(userID, 2024, 1)

		require.NoError(t, err)
		assert.NotEmpty(t, report.CategoryBreakdown)

		// Find categories in breakdown
		var salaryCategory, groceriesCategory *domain.CategoryMetrics
		for i := range report.CategoryBreakdown {
			if report.CategoryBreakdown[i].CategoryID == incomeID {
				salaryCategory = &report.CategoryBreakdown[i]
			} else if report.CategoryBreakdown[i].CategoryID == expenseID {
				groceriesCategory = &report.CategoryBreakdown[i]
			}
		}

		assert.NotNil(t, salaryCategory)
		assert.NotNil(t, groceriesCategory)
		assert.Equal(t, 6000.00, salaryCategory.TotalAmount)
		assert.Equal(t, 550.00, groceriesCategory.TotalAmount)
		assert.Equal(t, 2, salaryCategory.TransactionCount)
		assert.Equal(t, 2, groceriesCategory.TransactionCount)
	})

	t.Run("report calculates savings rate correctly", func(t *testing.T) {
		report, err := service.GenerateMonthlyReport(userID, 2024, 1)

		require.NoError(t, err)
		expectedSavingsRate := ((6000.0 - 550.0) / 6000.0) * 100
		assert.InDelta(t, expectedSavingsRate, report.SavingsRate, 0.01)
	})

	t.Run("report includes insights and recommendations", func(t *testing.T) {
		report, err := service.GenerateMonthlyReport(userID, 2024, 1)

		require.NoError(t, err)
		assert.NotEmpty(t, report.Insights)
		assert.NotEmpty(t, report.Recommendations)
	})

	t.Run("report includes budget performance", func(t *testing.T) {
		report, err := service.GenerateMonthlyReport(userID, 2024, 1)

		require.NoError(t, err)
		assert.NotEmpty(t, report.BudgetPerformance)

		// Verify budget performance metrics
		assert.GreaterOrEqual(t, report.BudgetPerformance.TotalBudgeted, 0.0)
		assert.GreaterOrEqual(t, report.BudgetPerformance.TotalSpent, 0.0)
		assert.NotZero(t, report.BudgetPerformance.CategoriesUnderBudget+report.BudgetPerformance.CategoriesOverBudget)
	})
}
