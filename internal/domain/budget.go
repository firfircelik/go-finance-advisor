package domain

import "time"

// Budget represents a user's budget for a specific category and period
type Budget struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `json:"user_id"`
	CategoryID uint      `json:"category_id"`
	Category   Category  `gorm:"foreignKey:CategoryID" json:"category"`
	Amount     float64   `json:"amount"`
	Period     string    `gorm:"type:varchar(20);default:'monthly'" json:"period"` // "weekly", "monthly", "yearly"
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Spent      float64   `gorm:"default:0" json:"spent"`
	Remaining  float64   `gorm:"default:0" json:"remaining"`
	IsActive   bool      `gorm:"default:true" json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// BudgetSummary represents budget overview for a user
type BudgetSummary struct {
	TotalBudget    float64 `json:"total_budget"`
	TotalSpent     float64 `json:"total_spent"`
	TotalRemaining float64 `json:"total_remaining"`
	PercentageUsed float64 `json:"percentage_used"`
	BudgetStatus   string  `json:"budget_status"` // "on_track", "warning", "over_budget"
}

// CalculateRemaining calculates the remaining budget amount
func (b *Budget) CalculateRemaining() {
	b.Remaining = b.Amount - b.Spent
}

// GetBudgetStatus returns the status based on spending percentage
func (b *Budget) GetBudgetStatus() string {
	if b.Amount == 0 {
		return "no_budget"
	}

	percentage := (b.Spent / b.Amount) * 100

	if percentage >= 100 {
		return "over_budget"
	} else if percentage >= 80 {
		return "warning"
	}
	return "on_track"
}
