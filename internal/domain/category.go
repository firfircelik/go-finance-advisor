// Package domain contains the core business entities and domain models
// for the Go Finance Advisor application.
package domain

import "time"

// Category represents a transaction category
type Category struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(50);uniqueIndex" json:"name"`
	Type        string    `gorm:"type:varchar(10)" json:"type"` // "income" or "expense"
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Color       string    `json:"color"`
	IsDefault   bool      `gorm:"default:false" json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GetDefaultCategories returns predefined categories
func GetDefaultCategories() []Category {
	return []Category{
		// Income Categories
		{Name: "Salary", Type: "income", Description: "Regular salary income", Icon: "💼", Color: "#4CAF50", IsDefault: true},
		{Name: "Freelance", Type: "income", Description: "Freelance work income", Icon: "💻", Color: "#2196F3", IsDefault: true},
		{Name: "Investment", Type: "income", Description: "Investment returns", Icon: "📈", Color: "#FF9800", IsDefault: true},
		{Name: "Business", Type: "income", Description: "Business income", Icon: "🏢", Color: "#9C27B0", IsDefault: true},
		{Name: "Other Income", Type: "income", Description: "Other sources of income", Icon: "💰", Color: "#607D8B", IsDefault: true},

		// Expense Categories
		{
			Name: "Food & Dining", Type: "expense",
			Description: "Restaurants, groceries, food delivery",
			Icon:        "🍽️", Color: "#F44336", IsDefault: true,
		},
		{
			Name: "Transportation", Type: "expense",
			Description: "Gas, public transport, car maintenance",
			Icon:        "🚗", Color: "#FF5722", IsDefault: true,
		},
		{
			Name: "Shopping", Type: "expense",
			Description: "Clothing, electronics, general shopping",
			Icon:        "🛍️", Color: "#E91E63", IsDefault: true,
		},
		{
			Name: "Entertainment", Type: "expense",
			Description: "Movies, games, hobbies",
			Icon:        "🎬", Color: "#9C27B0", IsDefault: true,
		},
		{
			Name: "Bills & Utilities", Type: "expense",
			Description: "Electricity, water, internet, phone",
			Icon:        "📄", Color: "#FF9800", IsDefault: true,
		},
		{Name: "Healthcare", Type: "expense", Description: "Medical expenses, insurance", Icon: "🏥", Color: "#4CAF50", IsDefault: true},
		{Name: "Education", Type: "expense", Description: "Courses, books, training", Icon: "📚", Color: "#2196F3", IsDefault: true},
		{Name: "Travel", Type: "expense", Description: "Vacation, business trips", Icon: "✈️", Color: "#00BCD4", IsDefault: true},
		{Name: "Housing", Type: "expense", Description: "Rent, mortgage, home maintenance", Icon: "🏠", Color: "#795548", IsDefault: true},
		{Name: "Other Expenses", Type: "expense", Description: "Miscellaneous expenses", Icon: "💸", Color: "#607D8B", IsDefault: true},
	}
}
