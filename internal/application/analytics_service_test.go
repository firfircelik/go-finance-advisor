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

func setupAnalyticsTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate tables
	err = db.AutoMigrate(
		&domain.User{},
		&domain.Category{},
		&domain.Transaction{},
		&domain.Budget{},
		&domain.FinancialGoal{},
	)
	require.NoError(t, err)

	return db
}

func createAnalyticsTestData(t *testing.T, db *gorm.DB) (uint, uint, uint) {
	// Create test user
	user := &domain.User{
		Email:     "analytics@test.com",
		Password:  "hashedpassword",
		FirstName: "Analytics",
		LastName:  "User",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	// Create test categories
	incomeCategory := &domain.Category{
		Name: "Salary",
		Type: "income",
	}
	expenseCategory := &domain.Category{
		Name: "Food",
		Type: "expense",
	}
	err = db.Create(incomeCategory).Error
	require.NoError(t, err)
	err = db.Create(expenseCategory).Error
	require.NoError(t, err)

	return user.ID, incomeCategory.ID, expenseCategory.ID
}

func TestAnalyticsService_GetFinancialMetrics(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	analyticsService := &AnalyticsService{DB: db}
	userID, incomeCategoryID, expenseCategoryID := createAnalyticsTestData(t, db)

	now := time.Now()
	startDate := now.AddDate(0, -1, 0) // 1 month ago
	endDate := now

	// Create test transactions
	transactions := []*domain.Transaction{
		{
			UserID:      userID,
			CategoryID:  incomeCategoryID,
			Amount:      5000.00,
			Type:        "income",
			Description: "Salary",
			Date:        now.AddDate(0, 0, -15),
		},
		{
			UserID:      userID,
			CategoryID:  expenseCategoryID,
			Amount:      1500.00,
			Type:        "expense",
			Description: "Groceries",
			Date:        now.AddDate(0, 0, -10),
		},
		{
			UserID:      userID,
			CategoryID:  expenseCategoryID,
			Amount:      800.00,
			Type:        "expense",
			Description: "Restaurant",
			Date:        now.AddDate(0, 0, -5),
		},
	}

	for _, tx := range transactions {
		err := db.Create(tx).Error
		require.NoError(t, err)
	}

	t.Run("calculate financial metrics", func(t *testing.T) {
		metrics, err := analyticsService.GetFinancialMetrics(userID, "month", startDate, endDate)
		assert.NoError(t, err)
		assert.NotNil(t, metrics)
		assert.Equal(t, userID, metrics.UserID)
		assert.Equal(t, "month", metrics.Period)
		assert.Equal(t, 5000.00, metrics.TotalIncome)
		assert.Equal(t, 2300.00, metrics.TotalExpenses)
		assert.Equal(t, 2700.00, metrics.NetIncome)
		assert.Equal(t, 0.54, metrics.SavingsRate) // 2700/5000
		assert.Equal(t, 0.46, metrics.ExpenseRatio) // 2300/5000
		assert.NotEmpty(t, metrics.CategoryBreakdown)
		assert.NotEmpty(t, metrics.MonthlyTrends)
	})

	t.Run("no transactions", func(t *testing.T) {
		// Clean up transactions
		db.Where("user_id = ?", userID).Delete(&domain.Transaction{})
		
		metrics, err := analyticsService.GetFinancialMetrics(userID, "month", startDate, endDate)
		assert.NoError(t, err)
		assert.NotNil(t, metrics)
		assert.Equal(t, 0.0, metrics.TotalIncome)
		assert.Equal(t, 0.0, metrics.TotalExpenses)
		assert.Equal(t, 0.0, metrics.NetIncome)
	})
}

func TestAnalyticsService_GetIncomeExpenseAnalysis(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	analyticsService := &AnalyticsService{DB: db}
	userID, incomeCategoryID, expenseCategoryID := createAnalyticsTestData(t, db)

	now := time.Now()
	startDate := now.AddDate(0, -1, 0)
	endDate := now

	// Create test transactions
	transactions := []*domain.Transaction{
		{
			UserID:      userID,
			CategoryID:  incomeCategoryID,
			Amount:      3000.00,
			Type:        "income",
			Description: "Salary",
			Date:        now.AddDate(0, 0, -15),
		},
		{
			UserID:      userID,
			CategoryID:  incomeCategoryID,
			Amount:      500.00,
			Type:        "income",
			Description: "Freelance",
			Date:        now.AddDate(0, 0, -10),
		},
		{
			UserID:      userID,
			CategoryID:  expenseCategoryID,
			Amount:      1200.00,
			Type:        "expense",
			Description: "Groceries",
			Date:        now.AddDate(0, 0, -8),
		},
	}

	for _, tx := range transactions {
		err := db.Create(tx).Error
		require.NoError(t, err)
	}

	t.Run("income expense analysis", func(t *testing.T) {
		analysis, err := analyticsService.GetIncomeExpenseAnalysis(userID, "month", startDate, endDate)
		assert.NoError(t, err)
		assert.NotNil(t, analysis)
		assert.Equal(t, userID, analysis.UserID)
		assert.Equal(t, "month", analysis.Period)
		assert.NotEmpty(t, analysis.IncomeBreakdown)
		assert.NotEmpty(t, analysis.ExpenseBreakdown)
		assert.NotEmpty(t, analysis.TopIncomeCategories)
		assert.NotEmpty(t, analysis.TopExpenseCategories)
		assert.NotNil(t, analysis.DailyAverages)
		assert.NotEmpty(t, analysis.WeeklyTrends)
	})
}

func TestAnalyticsService_GetCategoryAnalysis(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	analyticsService := &AnalyticsService{DB: db}
	userID, _, expenseCategoryID := createAnalyticsTestData(t, db)

	now := time.Now()
	startDate := now.AddDate(0, -1, 0)
	endDate := now

	// Create test transactions for specific category
	transactions := []*domain.Transaction{
		{
			UserID:      userID,
			CategoryID:  expenseCategoryID,
			Amount:      100.00,
			Type:        "expense",
			Description: "Lunch",
			Date:        now.AddDate(0, 0, -15),
		},
		{
			UserID:      userID,
			CategoryID:  expenseCategoryID,
			Amount:      200.00,
			Type:        "expense",
			Description: "Dinner",
			Date:        now.AddDate(0, 0, -10),
		},
		{
			UserID:      userID,
			CategoryID:  expenseCategoryID,
			Amount:      150.00,
			Type:        "expense",
			Description: "Groceries",
			Date:        now.AddDate(0, 0, -5),
		},
	}

	for _, tx := range transactions {
		err := db.Create(tx).Error
		require.NoError(t, err)
	}

	t.Run("category analysis with transactions", func(t *testing.T) {
		metrics, err := analyticsService.GetCategoryAnalysis(userID, expenseCategoryID, startDate, endDate)
		assert.NoError(t, err)
		assert.NotNil(t, metrics)
		assert.Equal(t, expenseCategoryID, metrics.CategoryID)
		assert.Equal(t, "Food", metrics.CategoryName)
		assert.Equal(t, 450.00, metrics.TotalAmount)
		assert.Equal(t, 3, metrics.TransactionCount)
		assert.Equal(t, 150.00, metrics.AverageAmount)
		assert.NotEmpty(t, metrics.Trend)
	})

	t.Run("category analysis no transactions", func(t *testing.T) {
		// Use a non-existent category ID
		metrics, err := analyticsService.GetCategoryAnalysis(userID, 999, startDate, endDate)
		assert.NoError(t, err)
		assert.NotNil(t, metrics)
		assert.Equal(t, uint(999), metrics.CategoryID)
		assert.Equal(t, 0.0, metrics.TotalAmount)
		assert.Equal(t, 0, metrics.TransactionCount)
		assert.Equal(t, 0.0, metrics.AverageAmount)
		assert.Equal(t, "stable", metrics.Trend)
	})
}

func TestAnalyticsService_GetDashboardSummary(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	analyticsService := &AnalyticsService{DB: db}
	userID, incomeCategoryID, expenseCategoryID := createAnalyticsTestData(t, db)

	now := time.Now()

	// Create test transactions
	transactions := []*domain.Transaction{
		{
			UserID:      userID,
			CategoryID:  incomeCategoryID,
			Amount:      4000.00,
			Type:        "income",
			Description: "Salary",
			Date:        now.AddDate(0, 0, -15),
		},
		{
			UserID:      userID,
			CategoryID:  expenseCategoryID,
			Amount:      1000.00,
			Type:        "expense",
			Description: "Groceries",
			Date:        now.AddDate(0, 0, -10),
		},
		{
			UserID:      userID,
			CategoryID:  expenseCategoryID,
			Amount:      500.00,
			Type:        "expense",
			Description: "Restaurant",
			Date:        now.AddDate(0, 0, -5),
		},
	}

	for _, tx := range transactions {
		err := db.Create(tx).Error
		require.NoError(t, err)
	}

	// Create test budget
	budget := &domain.Budget{
		UserID:     userID,
		CategoryID: expenseCategoryID,
		Amount:     2000.00,
		Spent:      1500.00,
		StartDate:  now.AddDate(0, 0, -30),
		EndDate:    now.AddDate(0, 0, 30),
		IsActive:   true,
	}
	err := db.Create(budget).Error
	require.NoError(t, err)

	// Create test financial goal
	goal := &domain.FinancialGoal{
		UserID:        userID,
		Title:         "Emergency Fund",
		Description:   "Emergency savings goal",
		TargetAmount:  10000.00,
		CurrentAmount: 5000.00,
		TargetDate:    now.AddDate(1, 0, 0),
		GoalType:      "savings",
		Status:        "active",
	}
	err = db.Create(goal).Error
	require.NoError(t, err)

	t.Run("dashboard summary month", func(t *testing.T) {
		dashboard, err := analyticsService.GetDashboardSummary(userID, "month")
		assert.NoError(t, err)
		assert.NotNil(t, dashboard)
		assert.Equal(t, userID, dashboard.UserID)
		assert.Equal(t, "month", dashboard.Period)
		assert.Equal(t, 4000.00, dashboard.MonthlyIncome)
		assert.Equal(t, 1500.00, dashboard.MonthlyExpenses)
		assert.Equal(t, 2500.00, dashboard.MonthlySavings)
		assert.Equal(t, 62.5, dashboard.SavingsRate) // (2500/4000)*100
		assert.NotEmpty(t, dashboard.TopExpenseCategories)
		assert.NotEmpty(t, dashboard.RecentTransactions)
		assert.NotEmpty(t, dashboard.BudgetAlerts)
		assert.NotEmpty(t, dashboard.FinancialGoals)
		assert.NotNil(t, dashboard.QuickStats)
	})

	t.Run("dashboard summary week", func(t *testing.T) {
		dashboard, err := analyticsService.GetDashboardSummary(userID, "week")
		assert.NoError(t, err)
		assert.NotNil(t, dashboard)
		assert.Equal(t, "week", dashboard.Period)
	})

	t.Run("dashboard summary year", func(t *testing.T) {
		dashboard, err := analyticsService.GetDashboardSummary(userID, "year")
		assert.NoError(t, err)
		assert.NotNil(t, dashboard)
		assert.Equal(t, "year", dashboard.Period)
	})

	t.Run("dashboard summary default period", func(t *testing.T) {
		dashboard, err := analyticsService.GetDashboardSummary(userID, "invalid")
		assert.NoError(t, err)
		assert.NotNil(t, dashboard)
		// Should default to month
	})
}

func TestAnalyticsService_HelperMethods(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	analyticsService := &AnalyticsService{DB: db}
	userID, incomeCategoryID, expenseCategoryID := createAnalyticsTestData(t, db)

	now := time.Now()
	startDate := now.AddDate(0, -1, 0)
	endDate := now

	// Create test transactions
	transactions := []*domain.Transaction{
		{
			UserID:      userID,
			CategoryID:  incomeCategoryID,
			Amount:      2000.00,
			Type:        "income",
			Description: "Salary",
			Date:        now.AddDate(0, 0, -15),
		},
		{
			UserID:      userID,
			CategoryID:  expenseCategoryID,
			Amount:      800.00,
			Type:        "expense",
			Description: "Food",
			Date:        now.AddDate(0, 0, -10),
		},
	}

	for _, tx := range transactions {
		err := db.Create(tx).Error
		require.NoError(t, err)
	}

	t.Run("calculate category breakdown", func(t *testing.T) {
		var allTransactions []domain.Transaction
		db.Preload("Category").Where("user_id = ?", userID).Find(&allTransactions)
		
		breakdown := analyticsService.calculateCategoryBreakdown(allTransactions)
		assert.NotEmpty(t, breakdown)
		
		// Should have 2 categories
		assert.Len(t, breakdown, 2)
		
		// Check that percentages are calculated
		for _, category := range breakdown {
			assert.True(t, category.PercentageOfTotal > 0)
			assert.True(t, category.AverageAmount > 0)
		}
	})

	t.Run("calculate monthly trends", func(t *testing.T) {
		trends := analyticsService.calculateMonthlyTrends(userID, startDate, endDate)
		assert.NotEmpty(t, trends)
		
		// Should have at least one month
		assert.True(t, len(trends) >= 1)
		
		for _, trend := range trends {
			assert.NotEmpty(t, trend.Month)
			assert.True(t, trend.Year > 0)
		}
	})

	t.Run("calculate daily averages", func(t *testing.T) {
		var allTransactions []domain.Transaction
		db.Where("user_id = ?", userID).Find(&allTransactions)
		
		averages := analyticsService.calculateDailyAverages(allTransactions, startDate, endDate)
		assert.True(t, averages.DailyIncome > 0)
		assert.True(t, averages.DailyExpenses > 0)
		assert.True(t, averages.DailyNetIncome > 0)
	})

	t.Run("calculate weekly trends", func(t *testing.T) {
		trends := analyticsService.calculateWeeklyTrends(userID, startDate, endDate)
		assert.NotEmpty(t, trends)
		
		for _, trend := range trends {
			assert.True(t, trend.WeekNumber > 0)
			assert.False(t, trend.WeekStart.IsZero())
			assert.False(t, trend.WeekEnd.IsZero())
		}
	})

	t.Run("calculate category trend", func(t *testing.T) {
		trend := analyticsService.calculateCategoryTrend(userID, expenseCategoryID, startDate, endDate)
		assert.Contains(t, []string{"increasing", "decreasing", "stable"}, trend)
	})
}