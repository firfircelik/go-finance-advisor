package application

import (
	"math"
	"sort"
	"time"

	"go-finance-advisor/internal/domain"
	"gorm.io/gorm"
)

type AnalyticsService struct {
	DB *gorm.DB
}

func NewAnalyticsService(db *gorm.DB) *AnalyticsService {
	return &AnalyticsService{DB: db}
}

// GetFinancialMetrics calculates comprehensive financial metrics for a user
func (s *AnalyticsService) GetFinancialMetrics(userID uint, period string, startDate, endDate time.Time) (*domain.FinancialMetrics, error) {
	var transactions []domain.Transaction
	err := s.DB.Preload("Category").Where("user_id = ? AND date BETWEEN ? AND ?", userID, startDate, endDate).Find(&transactions).Error
	if err != nil {
		return nil, err
	}

	metrics := &domain.FinancialMetrics{
		UserID:    userID,
		Period:    period,
		StartDate: startDate,
		EndDate:   endDate,
	}

	// Calculate basic metrics
	s.calculateBasicMetrics(metrics, transactions)

	// Calculate category breakdown
	metrics.CategoryBreakdown = s.calculateCategoryBreakdown(transactions)

	// Calculate monthly trends
	metrics.MonthlyTrends = s.calculateMonthlyTrends(userID, startDate, endDate)

	// Calculate budget performance
	metrics.BudgetPerformance = s.calculateBudgetPerformance(userID, startDate, endDate)

	// Calculate financial health
	metrics.CalculateFinancialHealth()

	return metrics, nil
}

// GetIncomeExpenseAnalysis provides detailed income vs expense analysis
func (s *AnalyticsService) GetIncomeExpenseAnalysis(userID uint, period string, startDate, endDate time.Time) (*domain.IncomeExpenseAnalysis, error) {
	var transactions []domain.Transaction
	err := s.DB.Preload("Category").Where("user_id = ? AND date BETWEEN ? AND ?", userID, startDate, endDate).Find(&transactions).Error
	if err != nil {
		return nil, err
	}

	analysis := &domain.IncomeExpenseAnalysis{
		UserID:    userID,
		Period:    period,
		StartDate: startDate,
		EndDate:   endDate,
	}

	// Separate income and expense transactions
	incomeTransactions := []domain.Transaction{}
	expenseTransactions := []domain.Transaction{}

	for _, tx := range transactions {
		if tx.Type == "income" {
			incomeTransactions = append(incomeTransactions, tx)
		} else {
			expenseTransactions = append(expenseTransactions, tx)
		}
	}

	// Calculate breakdowns
	analysis.IncomeBreakdown = s.calculateCategoryBreakdown(incomeTransactions)
	analysis.ExpenseBreakdown = s.calculateCategoryBreakdown(expenseTransactions)

	// Get top categories
	analysis.TopIncomeCategories = s.getTopCategories(analysis.IncomeBreakdown, 5)
	analysis.TopExpenseCategories = s.getTopCategories(analysis.ExpenseBreakdown, 5)

	// Calculate daily averages
	analysis.DailyAverages = s.calculateDailyAverages(transactions, startDate, endDate)

	// Calculate weekly trends
	analysis.WeeklyTrends = s.calculateWeeklyTrends(userID, startDate, endDate)

	return analysis, nil
}

// GetCategoryAnalysis provides detailed analysis for a specific category
func (s *AnalyticsService) GetCategoryAnalysis(userID, categoryID uint, startDate, endDate time.Time) (*domain.CategoryMetrics, error) {
	var transactions []domain.Transaction
	err := s.DB.Preload("Category").Where("user_id = ? AND category_id = ? AND date BETWEEN ? AND ?", userID, categoryID, startDate, endDate).Find(&transactions).Error
	if err != nil {
		return nil, err
	}

	if len(transactions) == 0 {
		return &domain.CategoryMetrics{
			CategoryID:       categoryID,
			TotalAmount:      0,
			TransactionCount: 0,
			AverageAmount:    0,
			Trend:            "stable",
		}, nil
	}

	totalAmount := 0.0
	for _, tx := range transactions {
		totalAmount += tx.Amount
	}

	averageAmount := totalAmount / float64(len(transactions))

	// Calculate trend (simplified - compare with previous period)
	trend := s.calculateCategoryTrend(userID, categoryID, startDate, endDate)

	return &domain.CategoryMetrics{
		CategoryID:       categoryID,
		CategoryName:     transactions[0].Category.Name,
		TotalAmount:      totalAmount,
		TransactionCount: len(transactions),
		AverageAmount:    averageAmount,
		Trend:            trend,
	}, nil
}

// Helper functions

func (s *AnalyticsService) calculateBasicMetrics(metrics *domain.FinancialMetrics, transactions []domain.Transaction) {
	totalIncome := 0.0
	totalExpenses := 0.0

	for _, tx := range transactions {
		if tx.Type == "income" {
			totalIncome += tx.Amount
		} else {
			totalExpenses += tx.Amount
		}
	}

	metrics.TotalIncome = totalIncome
	metrics.TotalExpenses = totalExpenses
	metrics.NetIncome = totalIncome - totalExpenses
	metrics.CashFlow = metrics.NetIncome

	if totalIncome > 0 {
		metrics.SavingsRate = metrics.NetIncome / totalIncome
		metrics.ExpenseRatio = totalExpenses / totalIncome
	}
}

func (s *AnalyticsService) calculateCategoryBreakdown(transactions []domain.Transaction) []domain.CategoryMetrics {
	categoryMap := make(map[uint]*domain.CategoryMetrics)
	totalAmount := 0.0

	// Calculate totals for each category
	for _, tx := range transactions {
		totalAmount += tx.Amount
		if metrics, exists := categoryMap[tx.CategoryID]; exists {
			metrics.TotalAmount += tx.Amount
			metrics.TransactionCount++
		} else {
			categoryMap[tx.CategoryID] = &domain.CategoryMetrics{
				CategoryID:       tx.CategoryID,
				CategoryName:     tx.Category.Name,
				TotalAmount:      tx.Amount,
				TransactionCount: 1,
			}
		}
	}

	// Calculate percentages and averages
	var breakdown []domain.CategoryMetrics
	for _, metrics := range categoryMap {
		metrics.AverageAmount = metrics.TotalAmount / float64(metrics.TransactionCount)
		if totalAmount > 0 {
			metrics.PercentageOfTotal = (metrics.TotalAmount / totalAmount) * 100
		}
		breakdown = append(breakdown, *metrics)
	}

	// Sort by total amount (descending)
	sort.Slice(breakdown, func(i, j int) bool {
		return breakdown[i].TotalAmount > breakdown[j].TotalAmount
	})

	return breakdown
}

func (s *AnalyticsService) calculateMonthlyTrends(userID uint, startDate, endDate time.Time) []domain.MonthlyTrend {
	var trends []domain.MonthlyTrend

	// Iterate through each month in the date range
	current := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())
	for current.Before(endDate) || current.Equal(endDate) {
		monthStart := current
		monthEnd := monthStart.AddDate(0, 1, -1)

		var transactions []domain.Transaction
		s.DB.Where("user_id = ? AND date BETWEEN ? AND ?", userID, monthStart, monthEnd).Find(&transactions)

		income := 0.0
		expenses := 0.0
		for _, tx := range transactions {
			if tx.Type == "income" {
				income += tx.Amount
			} else {
				expenses += tx.Amount
			}
		}

		netIncome := income - expenses
		savingsRate := 0.0
		if income > 0 {
			savingsRate = netIncome / income
		}

		trends = append(trends, domain.MonthlyTrend{
			Month:       monthStart.Format("January"),
			Year:        monthStart.Year(),
			Income:      income,
			Expenses:    expenses,
			NetIncome:   netIncome,
			SavingsRate: savingsRate,
		})

		current = current.AddDate(0, 1, 0)
	}

	return trends
}

func (s *AnalyticsService) calculateBudgetPerformance(userID uint, startDate, endDate time.Time) domain.BudgetPerformanceMetrics {
	var budgets []domain.Budget
	s.DB.Preload("Category").Where("user_id = ? AND start_date <= ? AND end_date >= ?", userID, endDate, startDate).Find(&budgets)

	totalBudgeted := 0.0
	totalSpent := 0.0
	categoriesOverBudget := 0
	categoriesUnderBudget := 0

	for _, budget := range budgets {
		totalBudgeted += budget.Amount
		totalSpent += budget.Spent

		if budget.Spent > budget.Amount {
			categoriesOverBudget++
		} else {
			categoriesUnderBudget++
		}
	}

	variance := totalSpent - totalBudgeted
	variancePercentage := 0.0
	if totalBudgeted > 0 {
		variancePercentage = (variance / totalBudgeted) * 100
	}

	return domain.BudgetPerformanceMetrics{
		TotalBudgeted:         totalBudgeted,
		TotalSpent:            totalSpent,
		Variance:              variance,
		VariancePercentage:    variancePercentage,
		CategoriesOverBudget:  categoriesOverBudget,
		CategoriesUnderBudget: categoriesUnderBudget,
	}
}

func (s *AnalyticsService) getTopCategories(categories []domain.CategoryMetrics, limit int) []domain.CategoryMetrics {
	if len(categories) <= limit {
		return categories
	}
	return categories[:limit]
}

func (s *AnalyticsService) calculateDailyAverages(transactions []domain.Transaction, startDate, endDate time.Time) domain.DailyAverages {
	totalIncome := 0.0
	totalExpenses := 0.0

	for _, tx := range transactions {
		if tx.Type == "income" {
			totalIncome += tx.Amount
		} else {
			totalExpenses += tx.Amount
		}
	}

	days := math.Ceil(endDate.Sub(startDate).Hours() / 24)
	if days == 0 {
		days = 1
	}

	return domain.DailyAverages{
		DailyIncome:    totalIncome / days,
		DailyExpenses:  totalExpenses / days,
		DailyNetIncome: (totalIncome - totalExpenses) / days,
	}
}

func (s *AnalyticsService) calculateWeeklyTrends(userID uint, startDate, endDate time.Time) []domain.WeeklyTrend {
	var trends []domain.WeeklyTrend

	// Start from the beginning of the week
	current := startDate
	for current.Weekday() != time.Monday {
		current = current.AddDate(0, 0, -1)
	}

	weekNumber := 1
	for current.Before(endDate) {
		weekStart := current
		weekEnd := current.AddDate(0, 0, 6)
		if weekEnd.After(endDate) {
			weekEnd = endDate
		}

		var transactions []domain.Transaction
		s.DB.Where("user_id = ? AND date BETWEEN ? AND ?", userID, weekStart, weekEnd).Find(&transactions)

		income := 0.0
		expenses := 0.0
		for _, tx := range transactions {
			if tx.Type == "income" {
				income += tx.Amount
			} else {
				expenses += tx.Amount
			}
		}

		trends = append(trends, domain.WeeklyTrend{
			WeekStart:        weekStart,
			WeekEnd:          weekEnd,
			WeekNumber:       weekNumber,
			Income:           income,
			Expenses:         expenses,
			NetIncome:        income - expenses,
			TransactionCount: len(transactions),
		})

		current = current.AddDate(0, 0, 7)
		weekNumber++
	}

	return trends
}

// GetDashboardSummary returns a comprehensive dashboard overview
func (s *AnalyticsService) GetDashboardSummary(userID uint, period string) (*domain.DashboardSummary, error) {
	// Calculate date range based on period
	now := time.Now()
	var startDate, endDate time.Time

	switch period {
	case "week":
		startDate = now.AddDate(0, 0, -7)
		endDate = now
	case "month":
		startDate = now.AddDate(0, -1, 0)
		endDate = now
	case "quarter":
		startDate = now.AddDate(0, -3, 0)
		endDate = now
	case "year":
		startDate = now.AddDate(-1, 0, 0)
		endDate = now
	default:
		startDate = now.AddDate(0, -1, 0) // Default to last month
		endDate = now
	}

	// Get all transactions for the period
	var transactions []domain.Transaction
	err := s.DB.Preload("Category").Where("user_id = ? AND date BETWEEN ? AND ?", userID, startDate, endDate).Find(&transactions).Error
	if err != nil {
		return nil, err
	}

	// Calculate basic metrics
	totalIncome := 0.0
	totalExpenses := 0.0
	for _, tx := range transactions {
		if tx.Type == "income" {
			totalIncome += tx.Amount
		} else {
			totalExpenses += tx.Amount
		}
	}

	monthlySavings := totalIncome - totalExpenses
	savingsRate := 0.0
	if totalIncome > 0 {
		savingsRate = (monthlySavings / totalIncome) * 100
	}

	// Get category breakdown
	categoryBreakdown := s.calculateCategoryBreakdown(transactions)
	topExpenseCategories := s.getTopCategories(categoryBreakdown, 5)

	// Get recent transactions (last 10)
	var recentTransactions []domain.Transaction
	s.DB.Preload("Category").Where("user_id = ?", userID).Order("date DESC").Limit(10).Find(&recentTransactions)

	// Calculate budget alerts
	budgetAlerts := s.calculateBudgetAlerts(userID, startDate, endDate)

	// Get financial goals
	var financialGoals []domain.FinancialGoal
	s.DB.Where("user_id = ? AND status = ?", userID, "active").Find(&financialGoals)
	for i := range financialGoals {
		financialGoals[i].CalculateProgress()
	}

	// Calculate quick stats
	quickStats := s.calculateQuickStats(userID, transactions, startDate, endDate)

	dashboard := &domain.DashboardSummary{
		UserID:               userID,
		Period:               period,
		StartDate:            startDate,
		EndDate:              endDate,
		TotalBalance:         monthlySavings, // This could be calculated from account balances
		MonthlyIncome:        totalIncome,
		MonthlyExpenses:      totalExpenses,
		MonthlySavings:       monthlySavings,
		SavingsRate:          savingsRate,
		TopExpenseCategories: topExpenseCategories,
		RecentTransactions:   recentTransactions,
		BudgetAlerts:         budgetAlerts,
		FinancialGoals:       financialGoals,
		QuickStats:           quickStats,
	}

	return dashboard, nil
}

func (s *AnalyticsService) calculateBudgetAlerts(userID uint, startDate, endDate time.Time) []domain.BudgetAlert {
	var budgets []domain.Budget
	s.DB.Preload("Category").Where("user_id = ?", userID).Find(&budgets)

	var alerts []domain.BudgetAlert
	for _, budget := range budgets {
		// Calculate spent amount for this budget's category
		var spentAmount float64
		s.DB.Model(&domain.Transaction{}).Where("user_id = ? AND category_id = ? AND date BETWEEN ? AND ?", userID, budget.CategoryID, startDate, endDate).Select("COALESCE(SUM(amount), 0)").Scan(&spentAmount)

		percentageUsed := 0.0
		if budget.Amount > 0 {
			percentageUsed = (spentAmount / budget.Amount) * 100
		}

		// Only create alerts for budgets that are over 60% used
		if percentageUsed >= 60 {
			alert := domain.BudgetAlert{
				BudgetID:       budget.ID,
				CategoryName:   budget.Category.Name,
				BudgetAmount:   budget.Amount,
				SpentAmount:    spentAmount,
				PercentageUsed: percentageUsed,
				DaysRemaining:  int(time.Until(endDate).Hours() / 24),
			}
			alert.GetAlertLevel()
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

func (s *AnalyticsService) calculateQuickStats(userID uint, transactions []domain.Transaction, startDate, endDate time.Time) domain.QuickStats {
	totalTransactions := len(transactions)
	averageTransaction := 0.0
	largestExpense := 0.0
	mostUsedCategory := ""

	if totalTransactions > 0 {
		totalAmount := 0.0
		categoryCount := make(map[string]int)

		for _, tx := range transactions {
			totalAmount += tx.Amount
			if tx.Type == "expense" && tx.Amount > largestExpense {
				largestExpense = tx.Amount
			}
			if tx.Category.Name != "" {
				categoryCount[tx.Category.Name]++
			}
		}

		averageTransaction = totalAmount / float64(totalTransactions)

		// Find most used category
		maxCount := 0
		for category, count := range categoryCount {
			if count > maxCount {
				maxCount = count
				mostUsedCategory = category
			}
		}
	}

	// Calculate cash flow trend
	cashFlowTrend := "stable"
	if len(transactions) > 0 {
		// Simple trend calculation based on recent vs older transactions
		midPoint := len(transactions) / 2
		recentTotal := 0.0
		olderTotal := 0.0

		for i, tx := range transactions {
			if i < midPoint {
				olderTotal += tx.Amount
			} else {
				recentTotal += tx.Amount
			}
		}

		if recentTotal > olderTotal*1.1 {
			cashFlowTrend = "positive"
		} else if recentTotal < olderTotal*0.9 {
			cashFlowTrend = "negative"
		}
	}

	return domain.QuickStats{
		TotalTransactions:   totalTransactions,
		AverageTransaction:  averageTransaction,
		LargestExpense:      largestExpense,
		MostUsedCategory:    mostUsedCategory,
		DaysUntilNextBudget: 30, // This could be calculated based on budget periods
		CashFlowTrend:       cashFlowTrend,
	}
}

func (s *AnalyticsService) calculateCategoryTrend(userID, categoryID uint, startDate, endDate time.Time) string {
	// Calculate spending for current period
	var currentTransactions []domain.Transaction
	s.DB.Where("user_id = ? AND category_id = ? AND date BETWEEN ? AND ?", userID, categoryID, startDate, endDate).Find(&currentTransactions)

	currentTotal := 0.0
	for _, tx := range currentTransactions {
		currentTotal += tx.Amount
	}

	// Calculate spending for previous period (same duration)
	duration := endDate.Sub(startDate)
	prevStartDate := startDate.Add(-duration)
	prevEndDate := startDate

	var prevTransactions []domain.Transaction
	s.DB.Where("user_id = ? AND category_id = ? AND date BETWEEN ? AND ?", userID, categoryID, prevStartDate, prevEndDate).Find(&prevTransactions)

	prevTotal := 0.0
	for _, tx := range prevTransactions {
		prevTotal += tx.Amount
	}

	// Determine trend
	if currentTotal > prevTotal*1.1 { // 10% increase threshold
		return "increasing"
	} else if currentTotal < prevTotal*0.9 { // 10% decrease threshold
		return "decreasing"
	}
	return "stable"
}