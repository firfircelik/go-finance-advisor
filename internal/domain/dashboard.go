package domain

import "time"

// DashboardSummary represents a comprehensive dashboard overview
type DashboardSummary struct {
	UserID               uint              `json:"user_id"`
	Period               string            `json:"period"`
	StartDate            time.Time         `json:"start_date"`
	EndDate              time.Time         `json:"end_date"`
	TotalBalance         float64           `json:"total_balance"`
	MonthlyIncome        float64           `json:"monthly_income"`
	MonthlyExpenses      float64           `json:"monthly_expenses"`
	MonthlySavings       float64           `json:"monthly_savings"`
	SavingsRate          float64           `json:"savings_rate"`
	TopExpenseCategories []CategoryMetrics `json:"top_expense_categories"`
	RecentTransactions   []Transaction     `json:"recent_transactions"`
	BudgetAlerts         []BudgetAlert     `json:"budget_alerts"`
	FinancialGoals       []FinancialGoal   `json:"financial_goals"`
	QuickStats           QuickStats        `json:"quick_stats"`
}

// BudgetAlert represents budget overspending alerts
type BudgetAlert struct {
	BudgetID       uint    `json:"budget_id"`
	CategoryName   string  `json:"category_name"`
	BudgetAmount   float64 `json:"budget_amount"`
	SpentAmount    float64 `json:"spent_amount"`
	PercentageUsed float64 `json:"percentage_used"`
	AlertLevel     string  `json:"alert_level"` // "warning", "danger", "critical"
	DaysRemaining  int     `json:"days_remaining"`
}

// FinancialGoal represents user's financial goals
type FinancialGoal struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UserID        uint      `json:"user_id" gorm:"not null;index"`
	Title         string    `json:"title" gorm:"not null"`
	Description   string    `json:"description"`
	TargetAmount  float64   `json:"target_amount" gorm:"not null"`
	CurrentAmount float64   `json:"current_amount" gorm:"default:0"`
	TargetDate    time.Time `json:"target_date"`
	GoalType      string    `json:"goal_type" gorm:"not null"`    // "savings", "debt_payoff", "investment"
	Status        string    `json:"status" gorm:"default:active"` // "active", "completed", "paused"
	Progress      float64   `json:"progress" gorm:"-"`            // Calculated field
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// QuickStats represents quick financial statistics
type QuickStats struct {
	TotalTransactions   int     `json:"total_transactions"`
	AverageTransaction  float64 `json:"average_transaction"`
	LargestExpense      float64 `json:"largest_expense"`
	MostUsedCategory    string  `json:"most_used_category"`
	DaysUntilNextBudget int     `json:"days_until_next_budget"`
	CashFlowTrend       string  `json:"cash_flow_trend"` // "positive", "negative", "stable"
}

// CalculateProgress calculates the progress percentage for a financial goal
func (fg *FinancialGoal) CalculateProgress() {
	if fg.TargetAmount > 0 {
		fg.Progress = (fg.CurrentAmount / fg.TargetAmount) * 100
		if fg.Progress > 100 {
			fg.Progress = 100
		}
	}
}

// GetAlertLevel determines the alert level based on percentage used
func (ba *BudgetAlert) GetAlertLevel() {
	if ba.PercentageUsed >= 100 {
		ba.AlertLevel = "critical"
	} else if ba.PercentageUsed >= 80 {
		ba.AlertLevel = "danger"
	} else if ba.PercentageUsed >= 60 {
		ba.AlertLevel = "warning"
	} else {
		ba.AlertLevel = "normal"
	}
}
