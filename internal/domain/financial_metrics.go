package domain

import (
	"time"
)

// FinancialMetrics represents comprehensive financial analysis
type FinancialMetrics struct {
	UserID              uint                    `json:"user_id"`
	Period              string                  `json:"period"` // "monthly", "quarterly", "yearly"
	StartDate           time.Time               `json:"start_date"`
	EndDate             time.Time               `json:"end_date"`
	TotalIncome         float64                 `json:"total_income"`
	TotalExpenses       float64                 `json:"total_expenses"`
	NetIncome           float64                 `json:"net_income"`
	SavingsRate         float64                 `json:"savings_rate"`
	ExpenseRatio        float64                 `json:"expense_ratio"`
	CashFlow            float64                 `json:"cash_flow"`
	CategoryBreakdown   []CategoryMetrics       `json:"category_breakdown"`
	MonthlyTrends       []MonthlyTrend          `json:"monthly_trends"`
	FinancialHealth     FinancialHealthScore    `json:"financial_health"`
	BudgetPerformance   BudgetPerformanceMetrics `json:"budget_performance"`
}

// CategoryMetrics represents spending analysis by category
type CategoryMetrics struct {
	CategoryID       uint    `json:"category_id"`
	CategoryName     string  `json:"category_name"`
	TotalAmount      float64 `json:"total_amount"`
	TransactionCount int     `json:"transaction_count"`
	PercentageOfTotal float64 `json:"percentage_of_total"`
	AverageAmount    float64 `json:"average_amount"`
	Trend            string  `json:"trend"` // "increasing", "decreasing", "stable"
}

// MonthlyTrend represents month-over-month financial trends
type MonthlyTrend struct {
	Month         string  `json:"month"`
	Year          int     `json:"year"`
	Income        float64 `json:"income"`
	Expenses      float64 `json:"expenses"`
	NetIncome     float64 `json:"net_income"`
	SavingsRate   float64 `json:"savings_rate"`
}

// FinancialHealthScore represents overall financial health assessment
type FinancialHealthScore struct {
	OverallScore      int    `json:"overall_score"` // 0-100
	SavingsScore      int    `json:"savings_score"`
	SpendingScore     int    `json:"spending_score"`
	BudgetScore       int    `json:"budget_score"`
	HealthStatus      string `json:"health_status"` // "excellent", "good", "fair", "poor"
	Recommendations   []string `json:"recommendations"`
}

// BudgetPerformanceMetrics represents budget vs actual spending analysis
type BudgetPerformanceMetrics struct {
	TotalBudgeted     float64 `json:"total_budgeted"`
	TotalSpent        float64 `json:"total_spent"`
	Variance          float64 `json:"variance"`
	VariancePercentage float64 `json:"variance_percentage"`
	CategoriesOverBudget int   `json:"categories_over_budget"`
	CategoriesUnderBudget int  `json:"categories_under_budget"`
}

// IncomeExpenseAnalysis represents detailed income vs expense analysis
type IncomeExpenseAnalysis struct {
	UserID            uint                    `json:"user_id"`
	Period            string                  `json:"period"`
	StartDate         time.Time               `json:"start_date"`
	EndDate           time.Time               `json:"end_date"`
	IncomeBreakdown   []CategoryMetrics       `json:"income_breakdown"`
	ExpenseBreakdown  []CategoryMetrics       `json:"expense_breakdown"`
	TopIncomeCategories []CategoryMetrics     `json:"top_income_categories"`
	TopExpenseCategories []CategoryMetrics    `json:"top_expense_categories"`
	DailyAverages     DailyAverages           `json:"daily_averages"`
	WeeklyTrends      []WeeklyTrend           `json:"weekly_trends"`
}

// DailyAverages represents daily spending and income averages
type DailyAverages struct {
	DailyIncome    float64 `json:"daily_income"`
	DailyExpenses  float64 `json:"daily_expenses"`
	DailyNetIncome float64 `json:"daily_net_income"`
}

// WeeklyTrend represents week-over-week trends
type WeeklyTrend struct {
	WeekStart     time.Time `json:"week_start"`
	WeekEnd       time.Time `json:"week_end"`
	WeekNumber    int       `json:"week_number"`
	Income        float64   `json:"income"`
	Expenses      float64   `json:"expenses"`
	NetIncome     float64   `json:"net_income"`
	TransactionCount int    `json:"transaction_count"`
}

// CalculateFinancialHealth calculates overall financial health score
func (fm *FinancialMetrics) CalculateFinancialHealth() {
	// Savings Rate Score (0-40 points)
	savingsScore := 0
	if fm.SavingsRate >= 0.20 { // 20% or more
		savingsScore = 40
	} else if fm.SavingsRate >= 0.15 {
		savingsScore = 30
	} else if fm.SavingsRate >= 0.10 {
		savingsScore = 20
	} else if fm.SavingsRate >= 0.05 {
		savingsScore = 10
	}

	// Expense Ratio Score (0-30 points)
	expenseScore := 0
	if fm.ExpenseRatio <= 0.50 { // 50% or less of income
		expenseScore = 30
	} else if fm.ExpenseRatio <= 0.70 {
		expenseScore = 20
	} else if fm.ExpenseRatio <= 0.85 {
		expenseScore = 10
	}

	// Budget Performance Score (0-30 points)
	budgetScore := 0
	if fm.BudgetPerformance.VariancePercentage >= -10 && fm.BudgetPerformance.VariancePercentage <= 5 {
		budgetScore = 30
	} else if fm.BudgetPerformance.VariancePercentage >= -20 && fm.BudgetPerformance.VariancePercentage <= 10 {
		budgetScore = 20
	} else if fm.BudgetPerformance.VariancePercentage >= -30 && fm.BudgetPerformance.VariancePercentage <= 15 {
		budgetScore = 10
	}

	overallScore := savingsScore + expenseScore + budgetScore

	// Determine health status
	healthStatus := "poor"
	if overallScore >= 80 {
		healthStatus = "excellent"
	} else if overallScore >= 60 {
		healthStatus = "good"
	} else if overallScore >= 40 {
		healthStatus = "fair"
	}

	// Generate recommendations
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

	fm.FinancialHealth = FinancialHealthScore{
		OverallScore:    overallScore,
		SavingsScore:    savingsScore,
		SpendingScore:   expenseScore,
		BudgetScore:     budgetScore,
		HealthStatus:    healthStatus,
		Recommendations: recommendations,
	}
}