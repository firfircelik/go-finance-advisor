// Package domain contains the core business entities and domain models
// for the Go Finance Advisor application.
package domain

import "time"

// Transaction represents a financial transaction for a user
// Type can be "income" or "expense"
type Transaction struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `json:"user_id"`
	CategoryID  uint      `json:"category_id"`
	Category    Category  `gorm:"foreignKey:CategoryID" json:"category"`
	Type        string    `gorm:"type:varchar(10);default:'expense'" json:"type"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
