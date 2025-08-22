package application

import (
	"time"

	"go-finance-advisor/internal/domain"
)

// AnalyticsServiceInterface defines the contract for analytics operations
type AnalyticsServiceInterface interface {
	GetFinancialMetrics(userID uint, period string, startDate, endDate time.Time) (*domain.FinancialMetrics, error)
	GetIncomeExpenseAnalysis(userID uint, period string, startDate, endDate time.Time) (*domain.IncomeExpenseAnalysis, error)
	GetCategoryAnalysis(userID, categoryID uint, startDate, endDate time.Time) (*domain.CategoryMetrics, error)
	GetDashboardSummary(userID uint, period string) (*domain.DashboardSummary, error)
}
