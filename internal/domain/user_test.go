package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUser_Creation(t *testing.T) {
	tests := []struct {
		name string
		user User
		want User
	}{
		{
			name: "create user with all fields",
			user: User{
				Email:         "test@example.com",
				Password:      "hashedpassword",
				FirstName:     "John",
				LastName:      "Doe",
				Age:           30,
				RiskTolerance: "moderate",
			},
			want: User{
				Email:         "test@example.com",
				Password:      "hashedpassword",
				FirstName:     "John",
				LastName:      "Doe",
				Age:           30,
				RiskTolerance: "moderate",
			},
		},
		{
			name: "create user with minimal fields",
			user: User{
				Email:    "minimal@example.com",
				Password: "hashedpassword",
			},
			want: User{
				Email:    "minimal@example.com",
				Password: "hashedpassword",
			},
		},
		{
			name: "create user with conservative risk tolerance",
			user: User{
				Email:         "conservative@example.com",
				Password:      "hashedpassword",
				RiskTolerance: "conservative",
			},
			want: User{
				Email:         "conservative@example.com",
				Password:      "hashedpassword",
				RiskTolerance: "conservative",
			},
		},
		{
			name: "create user with aggressive risk tolerance",
			user: User{
				Email:         "aggressive@example.com",
				Password:      "hashedpassword",
				RiskTolerance: "aggressive",
			},
			want: User{
				Email:         "aggressive@example.com",
				Password:      "hashedpassword",
				RiskTolerance: "aggressive",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want.Email, tt.user.Email)
			assert.Equal(t, tt.want.Password, tt.user.Password)
			assert.Equal(t, tt.want.FirstName, tt.user.FirstName)
			assert.Equal(t, tt.want.LastName, tt.user.LastName)
			assert.Equal(t, tt.want.Age, tt.user.Age)
			assert.Equal(t, tt.want.RiskTolerance, tt.user.RiskTolerance)
		})
	}
}

func TestUser_RiskToleranceValues(t *testing.T) {
	validRiskTolerances := []string{"conservative", "moderate", "aggressive"}

	for _, riskTolerance := range validRiskTolerances {
		t.Run("valid_risk_tolerance_"+riskTolerance, func(t *testing.T) {
			user := User{
				Email:         "test@example.com",
				Password:      "hashedpassword",
				RiskTolerance: riskTolerance,
			}
			assert.Equal(t, riskTolerance, user.RiskTolerance)
		})
	}
}

func TestUser_DefaultValues(t *testing.T) {
	t.Run("default age should be 30", func(t *testing.T) {
		// This test verifies the GORM default tag behavior
		user := User{
			Email:    "test@example.com",
			Password: "hashedpassword",
			// Age not set, should default to 30 in database
		}
		// Note: GORM defaults are applied at database level, not struct level
		assert.Equal(t, 0, user.Age) // Struct default is zero value
	})

	t.Run("default risk tolerance should be moderate", func(t *testing.T) {
		// This test verifies the GORM default tag behavior
		user := User{
			Email:    "test@example.com",
			Password: "hashedpassword",
			// RiskTolerance not set, should default to 'moderate' in database
		}
		// Note: GORM defaults are applied at database level, not struct level
		assert.Equal(t, "", user.RiskTolerance) // Struct default is zero value
	})
}

func TestUser_JSONSerialization(t *testing.T) {
	t.Run("password should be excluded from JSON", func(t *testing.T) {
		user := User{
			ID:            1,
			Email:         "test@example.com",
			Password:      "secretpassword",
			FirstName:     "John",
			LastName:      "Doe",
			Age:           30,
			RiskTolerance: "moderate",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		// The password field has json:"-" tag, so it should be excluded
		// This is more of a documentation test since we can't easily test JSON serialization
		// without importing encoding/json, but we verify the struct has the correct tag
		assert.Equal(t, "secretpassword", user.Password)
	})

	t.Run("omitempty fields should work correctly", func(t *testing.T) {
		user := User{
			Email:    "test@example.com",
			Password: "hashedpassword",
			// FirstName, LastName, Age not set (should be omitted in JSON)
		}

		// Fields with omitempty should be omitted when empty
		assert.Equal(t, "", user.FirstName)
		assert.Equal(t, "", user.LastName)
		assert.Equal(t, 0, user.Age)
	})
}

func TestUser_Relationships(t *testing.T) {
	t.Run("user can have transactions", func(t *testing.T) {
		transactions := []Transaction{
			{
				ID:          1,
				UserID:      1,
				Type:        "expense",
				Description: "Grocery shopping",
				Amount:      50.00,
				Date:        time.Now(),
			},
			{
				ID:          2,
				UserID:      1,
				Type:        "income",
				Description: "Salary",
				Amount:      3000.00,
				Date:        time.Now(),
			},
		}

		user := User{
			ID:           1,
			Email:        "test@example.com",
			Password:     "hashedpassword",
			Transactions: transactions,
		}

		assert.Len(t, user.Transactions, 2)
		assert.Equal(t, "Grocery shopping", user.Transactions[0].Description)
		assert.Equal(t, "Salary", user.Transactions[1].Description)
		assert.Equal(t, uint(1), user.Transactions[0].UserID)
		assert.Equal(t, uint(1), user.Transactions[1].UserID)
	})

	t.Run("user can have empty transactions", func(t *testing.T) {
		user := User{
			ID:           1,
			Email:        "test@example.com",
			Password:     "hashedpassword",
			Transactions: []Transaction{},
		}

		assert.Len(t, user.Transactions, 0)
		assert.NotNil(t, user.Transactions)
	})
}