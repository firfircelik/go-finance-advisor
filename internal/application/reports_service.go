package application

import (
	"fmt"
	"time"

	"go-finance-advisor/internal/domain"
	"gorm.io/gorm"
)

type ReportsService struct {
	DB *gorm.DB
}

func NewReportsService(db *gorm.DB) *ReportsService {
	return &ReportsService{DB: db}
}

// GenerateMonthlyReport generates a comprehensive monthly financial report
func (s *ReportsService) GenerateMonthlyReport(userID uint, year int, month int) (*domain.FinancialReport, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	return s.generateReport(userID, "monthly", startDate, endDate)
}

// GenerateQuarterlyReport generates a comprehensive quarterly financial report
func (s *ReportsService) GenerateQuarterlyReport(userID uint, year int, quarter int) (*domain.FinancialReport, error) {
	var startMonth int
	switch quarter {
	case 1:
		startMonth = 1
	case 2:
		startMonth = 4
	case 3:
		startMonth = 7
	case 4:
		startMonth = 10
	default:
		return nil, fmt.Errorf("invalid quarter: %d", quarter)
	}

	startDate := time.Date(year, time.Month(startMonth), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 3, 0).Add(-time.Second)

	return s.generateReport(userID, "quarterly", startDate, endDate)
}

// GenerateYearlyReport generates a comprehensive yearly financial report
func (s *ReportsService) GenerateYearlyReport(userID uint, year int) (*domain.FinancialReport, error) {
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Second)

	return s.generateReport(userID, "yearly", startDate, endDate)
}

// GenerateCustomReport generates a report for a custom date range
func (s *ReportsService) GenerateCustomReport(userID uint, startDate, endDate time.Time) (*domain.FinancialReport, error) {
	return s.generateReport(userID, "custom", startDate, endDate)
}

func (s *ReportsService) generateReport(userID uint, reportType string, startDate, endDate time.Time) (*domain.FinancialReport, error) {
	// Get all transactions for the period
	var transactions []domain.Transaction
	err := s.DB.Preload("Category").Where("user_id = ? AND date BETWEEN ? AND ?", userID, startDate, endDate).Find(&transactions).Error
	if err != nil {
		return nil, err
	}

	// Calculate basic metrics
	totalIncome := 0.0
	totalExpenses := 0.0
	transactionCount := len(transactions)

	for _, tx := range transactions {
		if tx.Type == "income" {
			totalIncome += tx.Amount
		} else {
			totalExpenses += tx.Amount
		}
	}

	netIncome := totalIncome - totalExpenses
	savingsRate := 0.0
	if totalIncome > 0 {
		savingsRate = (netIncome / totalIncome) * 100
	}

	// Calculate category breakdown
	categoryBreakdown := s.calculateCategoryBreakdown(transactions)

	// Calculate monthly trends (for quarterly and yearly reports)
	monthlyTrends := s.calculateMonthlyTrends(userID, startDate, endDate)

	// Get budget performance
	budgetPerformance := s.calculateBudgetPerformance(userID, startDate, endDate)

	// Calculate top categories
	topIncomeCategories := s.getTopCategories(categoryBreakdown, "income", 5)
	topExpenseCategories := s.getTopCategories(categoryBreakdown, "expense", 5)

	// Generate insights and recommendations
	insights := s.generateInsights(totalIncome, totalExpenses, savingsRate, categoryBreakdown, budgetPerformance)
	recommendations := s.generateRecommendations(savingsRate, categoryBreakdown, budgetPerformance)

	report := &domain.FinancialReport{
		UserID:                  userID,
		ReportType:              reportType,
		StartDate:               startDate,
		EndDate:                 endDate,
		TotalIncome:             totalIncome,
		TotalExpenses:           totalExpenses,
		NetIncome:               netIncome,
		SavingsRate:             savingsRate,
		TransactionCount:        transactionCount,
		CategoryBreakdown:       categoryBreakdown,
		MonthlyTrends:           monthlyTrends,
		BudgetPerformance:       budgetPerformance,
		TopIncomeCategories:     topIncomeCategories,
		TopExpenseCategories:    topExpenseCategories,
		Insights:                insights,
		Recommendations:         recommendations,
		GeneratedAt:             time.Now(),
	}

	return report, nil
}

func (s *ReportsService) calculateCategoryBreakdown(transactions []domain.Transaction) []domain.CategoryMetrics {
	categoryMap := make(map[uint]*domain.CategoryMetrics)

	for _, tx := range transactions {
		if tx.CategoryID == 0 {
			continue
		}

		if _, exists := categoryMap[tx.CategoryID]; !exists {
			categoryMap[tx.CategoryID] = &domain.CategoryMetrics{
			CategoryID:       tx.CategoryID,
			CategoryName:     tx.Category.Name,
			TotalAmount:      0,
			PercentageOfTotal: 0,
			TransactionCount: 0,
			AverageAmount:    0,
			Trend:           "stable",
		}
		}

		categoryMap[tx.CategoryID].TotalAmount += tx.Amount
		categoryMap[tx.CategoryID].TransactionCount++
	}

	// Convert map to slice and calculate percentages
	var categories []domain.CategoryMetrics
	totalAmount := 0.0

	for _, category := range categoryMap {
		totalAmount += category.TotalAmount
		categories = append(categories, *category)
	}

	// Calculate percentages and average amounts
	for i := range categories {
		if totalAmount > 0 {
			categories[i].PercentageOfTotal = (categories[i].TotalAmount / totalAmount) * 100
		}
		if categories[i].TransactionCount > 0 {
			categories[i].AverageAmount = categories[i].TotalAmount / float64(categories[i].TransactionCount)
		}
	}

	return categories
}

func (s *ReportsService) calculateMonthlyTrends(userID uint, startDate, endDate time.Time) []domain.MonthlyTrend {
	var trends []domain.MonthlyTrend

	current := startDate
	for current.Before(endDate) {
		nextMonth := current.AddDate(0, 1, 0)
		if nextMonth.After(endDate) {
			nextMonth = endDate
		}

		var transactions []domain.Transaction
		s.DB.Where("user_id = ? AND date BETWEEN ? AND ?", userID, current, nextMonth).Find(&transactions)

		income := 0.0
		expenses := 0.0
		for _, tx := range transactions {
			if tx.Type == "income" {
				income += tx.Amount
			} else {
				expenses += tx.Amount
			}
		}

		trend := domain.MonthlyTrend{
			Month:           current.Format("2006-01"),
			Income:          income,
			Expenses:        expenses,
			NetIncome:       income - expenses,
			SavingsRate:     0,
		}

		trends = append(trends, trend)
		current = nextMonth
	}

	return trends
}

func (s *ReportsService) calculateBudgetPerformance(userID uint, startDate, endDate time.Time) domain.BudgetPerformanceMetrics {
	var budgets []domain.Budget
	s.DB.Preload("Category").Where("user_id = ?", userID).Find(&budgets)

	totalBudget := 0.0
	totalSpent := 0.0
	budgetsOnTrack := 0
	budgetsOverspent := 0

	for _, budget := range budgets {
		totalBudget += budget.Amount

		// Calculate spent amount for this budget's category
		var spentAmount float64
		s.DB.Model(&domain.Transaction{}).Where("user_id = ? AND category_id = ? AND date BETWEEN ? AND ?", userID, budget.CategoryID, startDate, endDate).Select("COALESCE(SUM(amount), 0)").Scan(&spentAmount)

		totalSpent += spentAmount

		if spentAmount <= budget.Amount {
			budgetsOnTrack++
		} else {
			budgetsOverspent++
		}
	}

	performanceScore := 0.0
	if totalBudget > 0 {
		performanceScore = ((totalBudget - totalSpent) / totalBudget) * 100
		if performanceScore < 0 {
			performanceScore = 0
		}
	}

	return domain.BudgetPerformanceMetrics{
		TotalBudgeted:        totalBudget,
		TotalSpent:           totalSpent,
		Variance:             totalBudget - totalSpent,
		VariancePercentage:   ((totalBudget - totalSpent) / totalBudget) * 100,
		CategoriesUnderBudget: budgetsOnTrack,
		CategoriesOverBudget:  budgetsOverspent,
	}
}

func (s *ReportsService) getTopCategories(categories []domain.CategoryMetrics, categoryType string, limit int) []domain.CategoryMetrics {
	filtered := append([]domain.CategoryMetrics{}, categories...)

	// Sort by total amount (descending)
	for i := 0; i < len(filtered)-1; i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[i].TotalAmount < filtered[j].TotalAmount {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered
}

func (s *ReportsService) generateInsights(totalIncome, totalExpenses, savingsRate float64, categories []domain.CategoryMetrics, budgetPerf domain.BudgetPerformanceMetrics) []string {
	var insights []string

	// Savings rate insights
	if savingsRate > 20 {
		insights = append(insights, "Excellent savings rate! You're saving more than 20% of your income.")
	} else if savingsRate > 10 {
		insights = append(insights, "Good savings rate. Consider increasing to 20% for better financial security.")
	} else if savingsRate > 0 {
		insights = append(insights, "You're saving money, but there's room for improvement.")
	} else {
		insights = append(insights, "Warning: You're spending more than you earn. Review your expenses.")
	}

	// Budget performance insights
	performanceScore := 100.0
	if budgetPerf.TotalBudgeted > 0 {
		performanceScore = (budgetPerf.TotalBudgeted - budgetPerf.TotalSpent) / budgetPerf.TotalBudgeted * 100
	}
	if performanceScore > 80 {
		insights = append(insights, "Great budget management! You're staying within your limits.")
	} else if performanceScore > 60 {
		insights = append(insights, "Good budget control with some room for improvement.")
	} else {
		insights = append(insights, "Budget management needs attention. Consider reviewing your spending habits.")
	}

	// Category insights
	for _, category := range categories {
		if category.PercentageOfTotal > 30 {
			insights = append(insights, fmt.Sprintf("High spending in %s category (%.1f%% of total expenses).", category.CategoryName, category.PercentageOfTotal))
		}
	}

	return insights
}

func (s *ReportsService) generateRecommendations(savingsRate float64, categories []domain.CategoryMetrics, budgetPerf domain.BudgetPerformanceMetrics) []string {
	var recommendations []string

	// Savings recommendations
	if savingsRate < 10 {
		recommendations = append(recommendations, "Try to save at least 10% of your income each month.")
	}
	if savingsRate < 20 {
		recommendations = append(recommendations, "Consider automating your savings to reach a 20% savings rate.")
	}

	// Budget recommendations
	if budgetPerf.CategoriesOverBudget > 0 {
		recommendations = append(recommendations, "Review and adjust budgets for overspent categories.")
	}

	// Category-specific recommendations
	for _, category := range categories {
		if category.PercentageOfTotal > 25 {
			recommendations = append(recommendations, fmt.Sprintf("Consider reducing spending in %s category.", category.CategoryName))
		}
	}

	// General recommendations
	recommendations = append(recommendations, "Track your expenses daily for better financial awareness.")
	recommendations = append(recommendations, "Review and update your budgets monthly.")
	recommendations = append(recommendations, "Consider setting up an emergency fund if you haven't already.")

	return recommendations
}