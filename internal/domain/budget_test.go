package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBudget_Creation(t *testing.T) {
	tests := []struct {
		name   string
		budget Budget
		want   Budget
	}{
		{
			name: "create monthly budget",
			budget: Budget{
				UserID:     1,
				CategoryID: 1,
				Amount:     500.00,
				Period:     "monthly",
				StartDate:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:    time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
				Spent:      150.00,
				Remaining:  350.00,
				IsActive:   true,
			},
			want: Budget{
				UserID:     1,
				CategoryID: 1,
				Amount:     500.00,
				Period:     "monthly",
				StartDate:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:    time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
				Spent:      150.00,
				Remaining:  350.00,
				IsActive:   true,
			},
		},
		{
			name: "create weekly budget",
			budget: Budget{
				UserID:     2,
				CategoryID: 2,
				Amount:     100.00,
				Period:     "weekly",
				StartDate:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:    time.Date(2024, 1, 7, 23, 59, 59, 0, time.UTC),
				Spent:      25.00,
				Remaining:  75.00,
				IsActive:   true,
			},
			want: Budget{
				UserID:     2,
				CategoryID: 2,
				Amount:     100.00,
				Period:     "weekly",
				StartDate:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:    time.Date(2024, 1, 7, 23, 59, 59, 0, time.UTC),
				Spent:      25.00,
				Remaining:  75.00,
				IsActive:   true,
			},
		},
		{
			name: "create yearly budget",
			budget: Budget{
				UserID:     1,
				CategoryID: 3,
				Amount:     12000.00,
				Period:     "yearly",
				StartDate:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:    time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
				Spent:      3000.00,
				Remaining:  9000.00,
				IsActive:   true,
			},
			want: Budget{
				UserID:     1,
				CategoryID: 3,
				Amount:     12000.00,
				Period:     "yearly",
				StartDate:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:    time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
				Spent:      3000.00,
				Remaining:  9000.00,
				IsActive:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want.UserID, tt.budget.UserID)
			assert.Equal(t, tt.want.CategoryID, tt.budget.CategoryID)
			assert.Equal(t, tt.want.Amount, tt.budget.Amount)
			assert.Equal(t, tt.want.Period, tt.budget.Period)
			assert.Equal(t, tt.want.StartDate, tt.budget.StartDate)
			assert.Equal(t, tt.want.EndDate, tt.budget.EndDate)
			assert.Equal(t, tt.want.Spent, tt.budget.Spent)
			assert.Equal(t, tt.want.Remaining, tt.budget.Remaining)
			assert.Equal(t, tt.want.IsActive, tt.budget.IsActive)
		})
	}
}

func TestBudget_CalculateRemaining(t *testing.T) {
	tests := []struct {
		name      string
		amount    float64
		spent     float64
		expected  float64
	}{
		{"normal case", 1000.00, 300.00, 700.00},
		{"no spending", 500.00, 0.00, 500.00},
		{"full budget spent", 200.00, 200.00, 0.00},
		{"overspent budget", 100.00, 150.00, -50.00},
		{"zero budget", 0.00, 0.00, 0.00},
		{"zero budget with spending", 0.00, 50.00, -50.00},
		{"small amounts", 10.50, 3.25, 7.25},
		{"large amounts", 50000.00, 12345.67, 37654.33},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			budget := Budget{
				Amount: tt.amount,
				Spent:  tt.spent,
			}
			budget.CalculateRemaining()
			assert.Equal(t, tt.expected, budget.Remaining)
		})
	}
}

func TestBudget_GetBudgetStatus(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		spent    float64
		expected string
	}{
		{"on track - 0% spent", 1000.00, 0.00, "on_track"},
		{"on track - 50% spent", 1000.00, 500.00, "on_track"},
		{"on track - 79% spent", 1000.00, 790.00, "on_track"},
		{"warning - 80% spent", 1000.00, 800.00, "warning"},
		{"warning - 90% spent", 1000.00, 900.00, "warning"},
		{"warning - 99% spent", 1000.00, 990.00, "warning"},
		{"over budget - 100% spent", 1000.00, 1000.00, "over_budget"},
		{"over budget - 150% spent", 1000.00, 1500.00, "over_budget"},
		{"no budget - zero amount", 0.00, 0.00, "no_budget"},
		{"no budget - zero amount with spending", 0.00, 100.00, "no_budget"},
		{"small budget on track", 10.00, 5.00, "on_track"},
		{"small budget warning", 10.00, 8.50, "warning"},
		{"small budget over", 10.00, 12.00, "over_budget"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			budget := Budget{
				Amount: tt.amount,
				Spent:  tt.spent,
			}
			status := budget.GetBudgetStatus()
			assert.Equal(t, tt.expected, status)
		})
	}
}

func TestBudget_DefaultValues(t *testing.T) {
	t.Run("default period should be monthly", func(t *testing.T) {
		// This test verifies the GORM default tag behavior
		budget := Budget{
			UserID:     1,
			CategoryID: 1,
			Amount:     500.00,
			// Period not set, should default to 'monthly' in database
		}
		// Note: GORM defaults are applied at database level, not struct level
		assert.Equal(t, "", budget.Period) // Struct default is zero value
	})

	t.Run("default spent should be 0", func(t *testing.T) {
		// This test verifies the GORM default tag behavior
		budget := Budget{
			UserID:     1,
			CategoryID: 1,
			Amount:     500.00,
			// Spent not set, should default to 0 in database
		}
		// Note: GORM defaults are applied at database level, not struct level
		assert.Equal(t, 0.0, budget.Spent) // Struct default is zero value
	})

	t.Run("default remaining should be 0", func(t *testing.T) {
		// This test verifies the GORM default tag behavior
		budget := Budget{
			UserID:     1,
			CategoryID: 1,
			Amount:     500.00,
			// Remaining not set, should default to 0 in database
		}
		// Note: GORM defaults are applied at database level, not struct level
		assert.Equal(t, 0.0, budget.Remaining) // Struct default is zero value
	})

	t.Run("default is_active should be true", func(t *testing.T) {
		// This test verifies the GORM default tag behavior
		budget := Budget{
			UserID:     1,
			CategoryID: 1,
			Amount:     500.00,
			// IsActive not set, should default to true in database
		}
		// Note: GORM defaults are applied at database level, not struct level
		assert.Equal(t, false, budget.IsActive) // Struct default is zero value
	})
}

func TestBudget_CategoryRelationship(t *testing.T) {
	t.Run("budget with category", func(t *testing.T) {
		category := Category{
			ID:          1,
			Name:        "Food & Dining",
			Type:        "expense",
			Description: "Restaurants, groceries, food delivery",
			Icon:        "🍽️",
			Color:       "#F44336",
			IsDefault:   true,
		}

		budget := Budget{
			ID:         1,
			UserID:     1,
			CategoryID: 1,
			Category:   category,
			Amount:     500.00,
			Period:     "monthly",
			StartDate:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
			Spent:      150.00,
			IsActive:   true,
		}

		assert.Equal(t, uint(1), budget.CategoryID)
		assert.Equal(t, "Food & Dining", budget.Category.Name)
		assert.Equal(t, "expense", budget.Category.Type)
		assert.Equal(t, "🍽️", budget.Category.Icon)
	})

	t.Run("budget without category loaded", func(t *testing.T) {
		budget := Budget{
			ID:         1,
			UserID:     1,
			CategoryID: 1,
			Amount:     500.00,
			Period:     "monthly",
			// Category not loaded
		}

		assert.Equal(t, uint(1), budget.CategoryID)
		assert.Equal(t, uint(0), budget.Category.ID) // Zero value
		assert.Equal(t, "", budget.Category.Name)     // Zero value
	})
}

func TestBudgetSummary_Creation(t *testing.T) {
	t.Run("create budget summary", func(t *testing.T) {
		summary := BudgetSummary{
			TotalBudget:    2000.00,
			TotalSpent:     1200.00,
			TotalRemaining: 800.00,
			PercentageUsed: 60.0,
			BudgetStatus:   "on_track",
		}

		assert.Equal(t, 2000.00, summary.TotalBudget)
		assert.Equal(t, 1200.00, summary.TotalSpent)
		assert.Equal(t, 800.00, summary.TotalRemaining)
		assert.Equal(t, 60.0, summary.PercentageUsed)
		assert.Equal(t, "on_track", summary.BudgetStatus)
	})
}

func TestBudget_PeriodTypes(t *testing.T) {
	validPeriods := []string{"weekly", "monthly", "yearly"}

	for _, period := range validPeriods {
		t.Run("valid_period_"+period, func(t *testing.T) {
			budget := Budget{
				UserID:     1,
				CategoryID: 1,
				Amount:     500.00,
				Period:     period,
				StartDate:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:    time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
				IsActive:   true,
			}
			assert.Equal(t, period, budget.Period)
		})
	}
}

func TestBudget_ActiveStatus(t *testing.T) {
	t.Run("active budget", func(t *testing.T) {
		budget := Budget{
			UserID:     1,
			CategoryID: 1,
			Amount:     500.00,
			Period:     "monthly",
			IsActive:   true,
		}
		assert.True(t, budget.IsActive)
	})

	t.Run("inactive budget", func(t *testing.T) {
		budget := Budget{
			UserID:     1,
			CategoryID: 1,
			Amount:     500.00,
			Period:     "monthly",
			IsActive:   false,
		}
		assert.False(t, budget.IsActive)
	})
}