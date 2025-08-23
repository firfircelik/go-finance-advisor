package domain

import (
	"time"
)

// FinancialMetrics represents comprehensive financial analysis
type FinancialMetrics struct {
	UserID            uint                     `json:"user_id"`
	Period            string                   `json:"period"` // "monthly", "quarterly", "yearly"
	StartDate         time.Time                `json:"start_date"`
	EndDate           time.Time                `json:"end_date"`
	TotalIncome       float64                  `json:"total_income"`
	TotalExpenses     float64                  `json:"total_expenses"`
	NetIncome         float64                  `json:"net_income"`
	SavingsRate       float64                  `json:"savings_rate"`
	ExpenseRatio      float64                  `json:"expense_ratio"`
	CashFlow          float64                  `json:"cash_flow"`
	CategoryBreakdown []CategoryMetrics        `json:"category_breakdown"`
	MonthlyTrends     []MonthlyTrend           `json:"monthly_trends"`
	FinancialHealth   FinancialHealthScore     `json:"financial_health"`
	BudgetPerformance BudgetPerformanceMetrics `json:"budget_performance"`
}

// CategoryMetrics represents spending analysis by category
type CategoryMetrics struct {
	CategoryID        uint    `json:"category_id"`
	CategoryName      string  `json:"category_name"`
	TotalAmount       float64 `json:"total_amount"`
	TransactionCount  int     `json:"transaction_count"`
	PercentageOfTotal float64 `json:"percentage_of_total"`
	AverageAmount     float64 `json:"average_amount"`
	Trend             string  `json:"trend"` // "increasing", "decreasing", "stable"
}

// MonthlyTrend represents month-over-month financial trends
type MonthlyTrend struct {
	Month       string  `json:"month"`
	Year        int     `json:"year"`
	Income      float64 `json:"income"`
	Expenses    float64 `json:"expenses"`
	NetIncome   float64 `json:"net_income"`
	SavingsRate float64 `json:"savings_rate"`
}

// FinancialHealthScore represents overall financial health assessment
type FinancialHealthScore struct {
	OverallScore    int      `json:"overall_score"` // 0-100
	SavingsScore    int      `json:"savings_score"`
	SpendingScore   int      `json:"spending_score"`
	BudgetScore     int      `json:"budget_score"`
	HealthStatus    string   `json:"health_status"` // "excellent", "good", "fair", "poor"
	Recommendations []string `json:"recommendations"`
}

// BudgetPerformanceMetrics represents budget vs actual spending analysis
type BudgetPerformanceMetrics struct {
	TotalBudgeted         float64 `json:"total_budgeted"`
	TotalSpent            float64 `json:"total_spent"`
	Variance              float64 `json:"variance"`
	VariancePercentage    float64 `json:"variance_percentage"`
	CategoriesOverBudget  int     `json:"categories_over_budget"`
	CategoriesUnderBudget int     `json:"categories_under_budget"`
}

// IncomeExpenseAnalysis represents detailed income vs expense analysis
type IncomeExpenseAnalysis struct {
	UserID               uint              `json:"user_id"`
	Period               string            `json:"period"`
	StartDate            time.Time         `json:"start_date"`
	EndDate              time.Time         `json:"end_date"`
	IncomeBreakdown      []CategoryMetrics `json:"income_breakdown"`
	ExpenseBreakdown     []CategoryMetrics `json:"expense_breakdown"`
	TopIncomeCategories  []CategoryMetrics `json:"top_income_categories"`
	TopExpenseCategories []CategoryMetrics `json:"top_expense_categories"`
	DailyAverages        DailyAverages     `json:"daily_averages"`
	WeeklyTrends         []WeeklyTrend     `json:"weekly_trends"`
}

// DailyAverages represents daily spending and income averages
type DailyAverages struct {
	DailyIncome    float64 `json:"daily_income"`
	DailyExpenses  float64 `json:"daily_expenses"`
	DailyNetIncome float64 `json:"daily_net_income"`
}

// WeeklyTrend represents week-over-week trends
type WeeklyTrend struct {
	WeekStart        time.Time `json:"week_start"`
	WeekEnd          time.Time `json:"week_end"`
	WeekNumber       int       `json:"week_number"`
	Income           float64   `json:"income"`
	Expenses         float64   `json:"expenses"`
	NetIncome        float64   `json:"net_income"`
	TransactionCount int       `json:"transaction_count"`
}

// CalculateFinancialHealth calculates overall financial health score
func (fm *FinancialMetrics) CalculateFinancialHealth() {
	savingsScore := fm.calculateSavingsScore()
	expenseScore := fm.calculateExpenseScore()
	budgetScore := fm.calculateBudgetScore()
	overallScore := savingsScore + expenseScore + budgetScore
	healthStatus := fm.determineHealthStatus(overallScore)
	recommendations := fm.generateRecommendations()

	fm.FinancialHealth = FinancialHealthScore{
		OverallScore:    overallScore,
		SavingsScore:    savingsScore,
		SpendingScore:   expenseScore,
		BudgetScore:     budgetScore,
		HealthStatus:    healthStatus,
		Recommendations: recommendations,
	}
}

// calculateSavingsScore calculates savings rate score (0-40 points)
func (fm *FinancialMetrics) calculateSavingsScore() int {
	switch {
	case fm.SavingsRate >= 0.20:
		return 40
	case fm.SavingsRate >= 0.15:
		return 30
	case fm.SavingsRate >= 0.10:
		return 20
	case fm.SavingsRate >= 0.05:
		return 10
	default:
		return 0
	}
}

// calculateExpenseScore calculates expense ratio score (0-30 points)
func (fm *FinancialMetrics) calculateExpenseScore() int {
	switch {
	case fm.ExpenseRatio <= 0.50:
		return 30
	case fm.ExpenseRatio <= 0.70:
		return 20
	case fm.ExpenseRatio <= 0.85:
		return 10
	default:
		return 0
	}
}

// calculateBudgetScore calculates budget performance score (0-30 points)
func (fm *FinancialMetrics) calculateBudgetScore() int {
	variance := fm.BudgetPerformance.VariancePercentage
	switch {
	case variance >= -10 && variance <= 5:
		return 30
	case variance >= -20 && variance <= 10:
		return 20
	case variance >= -30 && variance <= 15:
		return 10
	default:
		return 0
	}
}

// determineHealthStatus determines health status based on overall score
func (fm *FinancialMetrics) determineHealthStatus(overallScore int) string {
	switch {
	case overallScore >= 80:
		return "excellent"
	case overallScore >= 60:
		return "good"
	case overallScore >= 40:
		return "fair"
	default:
		return "poor"
	}
}

// generateRecommendations generates financial recommendations
func (fm *FinancialMetrics) generateRecommendations() []string {
	recommendations := []string{}
	if fm.SavingsRate < 0.10 {
		recommendations = append(recommendations, "Consider increasing your savings rate to at least 10% of income")
	}
	if fm.ExpenseRatio > 0.80 {
		recommendations = append(recommendations, "Your expenses are high relative to income. Look for areas to reduce spending")
	}
	if fm.BudgetPerformance.CategoriesOverBudget > fm.BudgetPerformance.CategoriesUnderBudget {
		recommendations = append(recommendations, "Review and adjust budgets for categories where you're overspending")
	}
	return recommendations
}
