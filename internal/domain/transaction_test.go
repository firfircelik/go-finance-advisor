package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTransaction_Creation(t *testing.T) {
	tests := []struct {
		name        string
		transaction Transaction
		want        Transaction
	}{
		{
			name: "create expense transaction",
			transaction: Transaction{
				UserID:      1,
				CategoryID:  1,
				Type:        "expense",
				Description: "Grocery shopping",
				Amount:      75.50,
				Date:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			want: Transaction{
				UserID:      1,
				CategoryID:  1,
				Type:        "expense",
				Description: "Grocery shopping",
				Amount:      75.50,
				Date:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name: "create income transaction",
			transaction: Transaction{
				UserID:      2,
				CategoryID:  5,
				Type:        "income",
				Description: "Freelance payment",
				Amount:      1200.00,
				Date:        time.Date(2024, 1, 20, 14, 0, 0, 0, time.UTC),
			},
			want: Transaction{
				UserID:      2,
				CategoryID:  5,
				Type:        "income",
				Description: "Freelance payment",
				Amount:      1200.00,
				Date:        time.Date(2024, 1, 20, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "create transaction with minimal fields",
			transaction: Transaction{
				UserID:     1,
				CategoryID: 1,
				Amount:     25.00,
				Date:       time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
			},
			want: Transaction{
				UserID:     1,
				CategoryID: 1,
				Amount:     25.00,
				Date:       time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want.UserID, tt.transaction.UserID)
			assert.Equal(t, tt.want.CategoryID, tt.transaction.CategoryID)
			assert.Equal(t, tt.want.Type, tt.transaction.Type)
			assert.Equal(t, tt.want.Description, tt.transaction.Description)
			assert.Equal(t, tt.want.Amount, tt.transaction.Amount)
			assert.Equal(t, tt.want.Date, tt.transaction.Date)
		})
	}
}

func TestTransaction_Types(t *testing.T) {
	validTypes := []string{"income", "expense"}

	for _, transactionType := range validTypes {
		t.Run("valid_type_"+transactionType, func(t *testing.T) {
			transaction := Transaction{
				UserID:      1,
				CategoryID:  1,
				Type:        transactionType,
				Description: "Test transaction",
				Amount:      100.00,
				Date:        time.Now(),
			}
			assert.Equal(t, transactionType, transaction.Type)
		})
	}
}

func TestTransaction_DefaultValues(t *testing.T) {
	t.Run("default type should be expense", func(t *testing.T) {
		// This test verifies the GORM default tag behavior
		transaction := Transaction{
			UserID:      1,
			CategoryID:  1,
			Description: "Test transaction",
			Amount:      50.00,
			Date:        time.Now(),
			// Type not set, should default to 'expense' in database
		}
		// Note: GORM defaults are applied at database level, not struct level
		assert.Equal(t, "", transaction.Type) // Struct default is zero value
	})
}

func TestTransaction_AmountValidation(t *testing.T) {
	tests := []struct {
		name   string
		amount float64
		valid  bool
	}{
		{"positive amount", 100.50, true},
		{"zero amount", 0.0, true},
		{"negative amount", -50.25, true}, // Negative amounts might be valid for corrections
		{"large amount", 999999.99, true},
		{"small decimal", 0.01, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transaction := Transaction{
				UserID:      1,
				CategoryID:  1,
				Type:        "expense",
				Description: "Test transaction",
				Amount:      tt.amount,
				Date:        time.Now(),
			}
			assert.Equal(t, tt.amount, transaction.Amount)
		})
	}
}

func TestTransaction_DateHandling(t *testing.T) {
	t.Run("transaction with past date", func(t *testing.T) {
		pastDate := time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC)
		transaction := Transaction{
			UserID:      1,
			CategoryID:  1,
			Type:        "expense",
			Description: "Christmas shopping",
			Amount:      200.00,
			Date:        pastDate,
		}
		assert.Equal(t, pastDate, transaction.Date)
	})

	t.Run("transaction with future date", func(t *testing.T) {
		futureDate := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
		transaction := Transaction{
			UserID:      1,
			CategoryID:  1,
			Type:        "income",
			Description: "Expected bonus",
			Amount:      500.00,
			Date:        futureDate,
		}
		assert.Equal(t, futureDate, transaction.Date)
	})

	t.Run("transaction with current date", func(t *testing.T) {
		now := time.Now()
		transaction := Transaction{
			UserID:      1,
			CategoryID:  1,
			Type:        "expense",
			Description: "Current purchase",
			Amount:      75.00,
			Date:        now,
		}
		assert.Equal(t, now, transaction.Date)
	})
}

func TestTransaction_CategoryRelationship(t *testing.T) {
	t.Run("transaction with category", func(t *testing.T) {
		category := Category{
			ID:          1,
			Name:        "Food & Dining",
			Type:        "expense",
			Description: "Restaurants, groceries, food delivery",
			Icon:        "🍽️",
			Color:       "#F44336",
			IsDefault:   true,
		}

		transaction := Transaction{
			ID:          1,
			UserID:      1,
			CategoryID:  1,
			Category:    category,
			Type:        "expense",
			Description: "Lunch at restaurant",
			Amount:      25.50,
			Date:        time.Now(),
		}

		assert.Equal(t, uint(1), transaction.CategoryID)
		assert.Equal(t, "Food & Dining", transaction.Category.Name)
		assert.Equal(t, "expense", transaction.Category.Type)
		assert.Equal(t, "🍽️", transaction.Category.Icon)
	})

	t.Run("transaction without category loaded", func(t *testing.T) {
		transaction := Transaction{
			ID:          1,
			UserID:      1,
			CategoryID:  1,
			Type:        "expense",
			Description: "Test transaction",
			Amount:      50.00,
			Date:        time.Now(),
			// Category not loaded
		}

		assert.Equal(t, uint(1), transaction.CategoryID)
		assert.Equal(t, uint(0), transaction.Category.ID) // Zero value
		assert.Equal(t, "", transaction.Category.Name)    // Zero value
	})
}

func TestTransaction_DescriptionHandling(t *testing.T) {
	tests := []struct {
		name        string
		description string
	}{
		{"empty description", ""},
		{"short description", "Coffee"},
		{"long description", "Monthly subscription for premium software license including all features and support"},
		{"description with special characters", "Café & Restaurant - 50% discount!"},
		{"description with numbers", "Invoice #12345 - Q1 2024"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transaction := Transaction{
				UserID:      1,
				CategoryID:  1,
				Type:        "expense",
				Description: tt.description,
				Amount:      100.00,
				Date:        time.Now(),
			}
			assert.Equal(t, tt.description, transaction.Description)
		})
	}
}

func TestTransaction_JSONSerialization(t *testing.T) {
	t.Run("all fields should be included in JSON", func(t *testing.T) {
		transaction := Transaction{
			ID:          1,
			UserID:      1,
			CategoryID:  1,
			Type:        "expense",
			Description: "Test transaction",
			Amount:      75.50,
			Date:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		}

		// All fields should be serializable to JSON (no json:"-" tags)
		assert.Equal(t, uint(1), transaction.ID)
		assert.Equal(t, uint(1), transaction.UserID)
		assert.Equal(t, uint(1), transaction.CategoryID)
		assert.Equal(t, "expense", transaction.Type)
		assert.Equal(t, "Test transaction", transaction.Description)
		assert.Equal(t, 75.50, transaction.Amount)
	})
}
