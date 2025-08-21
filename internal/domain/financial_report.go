package domain

import (
	"fmt"
	"time"
)

// FinancialReport represents a comprehensive financial report
type FinancialReport struct {
	ID                      uint                    `json:"id" gorm:"primaryKey"`
	UserID                  uint                    `json:"user_id" gorm:"not null;index"`
	ReportType              string                  `json:"report_type" gorm:"not null"` // "monthly", "quarterly", "yearly", "custom"
	StartDate               time.Time               `json:"start_date" gorm:"not null"`
	EndDate                 time.Time               `json:"end_date" gorm:"not null"`
	TotalIncome             float64                 `json:"total_income"`
	TotalExpenses           float64                 `json:"total_expenses"`
	NetIncome               float64                 `json:"net_income"`
	SavingsRate             float64                 `json:"savings_rate"`
	TransactionCount        int                     `json:"transaction_count"`
	CategoryBreakdown       []CategoryMetrics       `json:"category_breakdown" gorm:"-"`
	MonthlyTrends           []MonthlyTrend          `json:"monthly_trends" gorm:"-"`
	BudgetPerformance       BudgetPerformanceMetrics `json:"budget_performance" gorm:"-"`
	TopIncomeCategories     []CategoryMetrics       `json:"top_income_categories" gorm:"-"`
	TopExpenseCategories    []CategoryMetrics       `json:"top_expense_categories" gorm:"-"`
	Insights                []string                `json:"insights" gorm:"type:text"`
	Recommendations         []string                `json:"recommendations" gorm:"type:text"`
	GeneratedAt             time.Time               `json:"generated_at"`
	CreatedAt               time.Time               `json:"created_at"`
	UpdatedAt               time.Time               `json:"updated_at"`
}

// ReportSummary represents a lightweight report summary for listing
type ReportSummary struct {
	ID           uint      `json:"id"`
	ReportType   string    `json:"report_type"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	TotalIncome  float64   `json:"total_income"`
	TotalExpenses float64  `json:"total_expenses"`
	NetIncome    float64   `json:"net_income"`
	SavingsRate  float64   `json:"savings_rate"`
	GeneratedAt  time.Time `json:"generated_at"`
}

// ReportExportData represents data structure for report exports
type ReportExportData struct {
	Report      *FinancialReport `json:"report"`
	Transactions []Transaction   `json:"transactions,omitempty"`
	Budgets     []Budget        `json:"budgets,omitempty"`
	Categories  []Category      `json:"categories,omitempty"`
	ExportedAt  time.Time       `json:"exported_at"`
	Format      string          `json:"format"`
}

// GetReportTitle returns a formatted title for the report
func (r *FinancialReport) GetReportTitle() string {
	switch r.ReportType {
	case "monthly":
		return r.StartDate.Format("January 2006") + " Monthly Report"
	case "quarterly":
		quarter := (r.StartDate.Month()-1)/3 + 1
		return fmt.Sprintf("Q%d %d Quarterly Report", quarter, r.StartDate.Year())
	case "yearly":
		return fmt.Sprintf("%d Annual Report", r.StartDate.Year())
	case "custom":
		return fmt.Sprintf("Custom Report (%s - %s)", r.StartDate.Format("Jan 2, 2006"), r.EndDate.Format("Jan 2, 2006"))
	default:
		return "Financial Report"
	}
}

// GetPeriodDays returns the number of days in the report period
func (r *FinancialReport) GetPeriodDays() int {
	return int(r.EndDate.Sub(r.StartDate).Hours() / 24)
}

// GetDailyAverageIncome returns the average daily income
func (r *FinancialReport) GetDailyAverageIncome() float64 {
	days := r.GetPeriodDays()
	if days > 0 {
		return r.TotalIncome / float64(days)
	}
	return 0
}

// GetDailyAverageExpenses returns the average daily expenses
func (r *FinancialReport) GetDailyAverageExpenses() float64 {
	days := r.GetPeriodDays()
	if days > 0 {
		return r.TotalExpenses / float64(days)
	}
	return 0
}

// GetExpenseRatio returns the expense to income ratio
func (r *FinancialReport) GetExpenseRatio() float64 {
	if r.TotalIncome > 0 {
		return (r.TotalExpenses / r.TotalIncome) * 100
	}
	return 0
}

// GetFinancialHealthStatus returns a health status based on savings rate
func (r *FinancialReport) GetFinancialHealthStatus() string {
	if r.SavingsRate >= 20 {
		return "Excellent"
	} else if r.SavingsRate >= 10 {
		return "Good"
	} else if r.SavingsRate >= 0 {
		return "Fair"
	} else {
		return "Poor"
	}
}

// ToSummary converts the full report to a summary
func (r *FinancialReport) ToSummary() ReportSummary {
	return ReportSummary{
		ID:            r.ID,
		ReportType:    r.ReportType,
		StartDate:     r.StartDate,
		EndDate:       r.EndDate,
		TotalIncome:   r.TotalIncome,
		TotalExpenses: r.TotalExpenses,
		NetIncome:     r.NetIncome,
		SavingsRate:   r.SavingsRate,
		GeneratedAt:   r.GeneratedAt,
	}
}