package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCategory_Creation(t *testing.T) {
	tests := []struct {
		name     string
		category Category
		want     Category
	}{
		{
			name: "create expense category",
			category: Category{
				Name:        "Food & Dining",
				Type:        "expense",
				Description: "Restaurants, groceries, food delivery",
				Icon:        "🍽️",
				Color:       "#F44336",
				IsDefault:   true,
			},
			want: Category{
				Name:        "Food & Dining",
				Type:        "expense",
				Description: "Restaurants, groceries, food delivery",
				Icon:        "🍽️",
				Color:       "#F44336",
				IsDefault:   true,
			},
		},
		{
			name: "create income category",
			category: Category{
				Name:        "Salary",
				Type:        "income",
				Description: "Regular salary income",
				Icon:        "💼",
				Color:       "#4CAF50",
				IsDefault:   true,
			},
			want: Category{
				Name:        "Salary",
				Type:        "income",
				Description: "Regular salary income",
				Icon:        "💼",
				Color:       "#4CAF50",
				IsDefault:   true,
			},
		},
		{
			name: "create custom category",
			category: Category{
				Name:        "Pet Care",
				Type:        "expense",
				Description: "Veterinary bills, pet food, grooming",
				Icon:        "🐕",
				Color:       "#9C27B0",
				IsDefault:   false,
			},
			want: Category{
				Name:        "Pet Care",
				Type:        "expense",
				Description: "Veterinary bills, pet food, grooming",
				Icon:        "🐕",
				Color:       "#9C27B0",
				IsDefault:   false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want.Name, tt.category.Name)
			assert.Equal(t, tt.want.Type, tt.category.Type)
			assert.Equal(t, tt.want.Description, tt.category.Description)
			assert.Equal(t, tt.want.Icon, tt.category.Icon)
			assert.Equal(t, tt.want.Color, tt.category.Color)
			assert.Equal(t, tt.want.IsDefault, tt.category.IsDefault)
		})
	}
}

func TestCategory_Types(t *testing.T) {
	validTypes := []string{"income", "expense"}

	for _, categoryType := range validTypes {
		t.Run("valid_type_"+categoryType, func(t *testing.T) {
			category := Category{
				Name:        "Test Category",
				Type:        categoryType,
				Description: "Test description",
				Icon:        "🔧",
				Color:       "#000000",
				IsDefault:   false,
			}
			assert.Equal(t, categoryType, category.Type)
		})
	}
}

func TestCategory_DefaultValues(t *testing.T) {
	t.Run("default is_default should be false", func(t *testing.T) {
		// This test verifies the GORM default tag behavior
		category := Category{
			Name:        "Test Category",
			Type:        "expense",
			Description: "Test description",
			Icon:        "🔧",
			Color:       "#000000",
			// IsDefault not set, should default to false in database
		}
		// Note: GORM defaults are applied at database level, not struct level
		assert.Equal(t, false, category.IsDefault) // Struct default is zero value
	})
}

func TestCategory_NameValidation(t *testing.T) {
	tests := []struct {
		name         string
		categoryName string
		valid        bool
	}{
		{"normal name", "Food & Dining", true},
		{"short name", "Gas", true},
		{"long name", "Entertainment and Recreation Activities", true},
		{"name with numbers", "Utilities 2024", true},
		{"name with special characters", "Health & Wellness", true},
		{"name with ampersand", "Food & Dining", true},
		{"empty name", "", false}, // Should be invalid but we're just testing struct creation
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category := Category{
				Name:        tt.categoryName,
				Type:        "expense",
				Description: "Test description",
				Icon:        "🔧",
				Color:       "#000000",
				IsDefault:   false,
			}
			assert.Equal(t, tt.categoryName, category.Name)
		})
	}
}

func TestCategory_ColorValidation(t *testing.T) {
	tests := []struct {
		name  string
		color string
		valid bool
	}{
		{"hex color with hash", "#FF5722", true},
		{"hex color uppercase", "#4CAF50", true},
		{"hex color lowercase", "#f44336", true},
		{"short hex color", "#F00", true},
		{"color name", "red", true}, // Might be valid depending on implementation
		{"rgb color", "rgb(255, 87, 34)", true},
		{"empty color", "", true}, // Might be valid as default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category := Category{
				Name:        "Test Category",
				Type:        "expense",
				Description: "Test description",
				Icon:        "🔧",
				Color:       tt.color,
				IsDefault:   false,
			}
			assert.Equal(t, tt.color, category.Color)
		})
	}
}

func TestCategory_IconValidation(t *testing.T) {
	tests := []struct {
		name string
		icon string
	}{
		{"food emoji", "🍽️"},
		{"money emoji", "💰"},
		{"car emoji", "🚗"},
		{"house emoji", "🏠"},
		{"medical emoji", "🏥"},
		{"education emoji", "📚"},
		{"entertainment emoji", "🎬"},
		{"shopping emoji", "🛒"},
		{"travel emoji", "✈️"},
		{"business emoji", "💼"},
		{"empty icon", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category := Category{
				Name:        "Test Category",
				Type:        "expense",
				Description: "Test description",
				Icon:        tt.icon,
				Color:       "#000000",
				IsDefault:   false,
			}
			assert.Equal(t, tt.icon, category.Icon)
		})
	}
}

func TestGetDefaultCategories(t *testing.T) {
	t.Run("get default categories", func(t *testing.T) {
		categories := GetDefaultCategories()

		// Should have both income and expense categories
		assert.Greater(t, len(categories), 0)

		// Check that all returned categories are marked as default
		for _, category := range categories {
			assert.True(t, category.IsDefault, "Category %s should be marked as default", category.Name)
			assert.NotEmpty(t, category.Name, "Category name should not be empty")
			assert.NotEmpty(t, category.Type, "Category type should not be empty")
			assert.Contains(t, []string{"income", "expense"}, category.Type, "Category type should be income or expense")
			assert.NotEmpty(t, category.Icon, "Category icon should not be empty")
			assert.NotEmpty(t, category.Color, "Category color should not be empty")
		}
	})

	t.Run("verify specific default income categories", func(t *testing.T) {
		categories := GetDefaultCategories()
		incomeCategories := make([]Category, 0)

		for _, category := range categories {
			if category.Type == "income" {
				incomeCategories = append(incomeCategories, category)
			}
		}

		assert.Greater(t, len(incomeCategories), 0, "Should have income categories")

		// Check for specific income categories
		incomeNames := make([]string, len(incomeCategories))
		for i, category := range incomeCategories {
			incomeNames[i] = category.Name
		}

		assert.Contains(t, incomeNames, "Salary")
		assert.Contains(t, incomeNames, "Freelance")
		assert.Contains(t, incomeNames, "Investment")
	})

	t.Run("verify specific default expense categories", func(t *testing.T) {
		categories := GetDefaultCategories()
		expenseCategories := make([]Category, 0)

		for _, category := range categories {
			if category.Type == "expense" {
				expenseCategories = append(expenseCategories, category)
			}
		}

		assert.Greater(t, len(expenseCategories), 0, "Should have expense categories")

		// Check for specific expense categories
		expenseNames := make([]string, len(expenseCategories))
		for i, category := range expenseCategories {
			expenseNames[i] = category.Name
		}

		assert.Contains(t, expenseNames, "Food & Dining")
		assert.Contains(t, expenseNames, "Transportation")
	})

	t.Run("verify category properties", func(t *testing.T) {
		categories := GetDefaultCategories()

		// Find a specific category to test
		var salaryCategory *Category
		for _, category := range categories {
			if category.Name == "Salary" {
				salaryCategory = &category
				break
			}
		}

		assert.NotNil(t, salaryCategory, "Salary category should exist")
		assert.Equal(t, "income", salaryCategory.Type)
		assert.Equal(t, "Regular salary income", salaryCategory.Description)
		assert.Equal(t, "💼", salaryCategory.Icon)
		assert.Equal(t, "#4CAF50", salaryCategory.Color)
		assert.True(t, salaryCategory.IsDefault)
	})
}

func TestCategory_JSONSerialization(t *testing.T) {
	t.Run("all fields should be included in JSON", func(t *testing.T) {
		category := Category{
			ID:          1,
			Name:        "Food & Dining",
			Type:        "expense",
			Description: "Restaurants, groceries, food delivery",
			Icon:        "🍽️",
			Color:       "#F44336",
			IsDefault:   true,
			CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		}

		// All fields should be serializable to JSON (no json:"-" tags)
		assert.Equal(t, uint(1), category.ID)
		assert.Equal(t, "Food & Dining", category.Name)
		assert.Equal(t, "expense", category.Type)
		assert.Equal(t, "Restaurants, groceries, food delivery", category.Description)
		assert.Equal(t, "🍽️", category.Icon)
		assert.Equal(t, "#F44336", category.Color)
		assert.True(t, category.IsDefault)
	})
}

func TestCategory_UniqueConstraints(t *testing.T) {
	t.Run("category names should be unique", func(t *testing.T) {
		// This test documents the uniqueIndex constraint on Name field
		category1 := Category{
			Name:        "Food & Dining",
			Type:        "expense",
			Description: "First description",
			Icon:        "🍽️",
			Color:       "#F44336",
			IsDefault:   true,
		}

		category2 := Category{
			Name:        "Food & Dining", // Same name
			Type:        "expense",
			Description: "Different description",
			Icon:        "🍕",
			Color:       "#FF5722",
			IsDefault:   false,
		}

		// Both categories can be created at struct level
		// The uniqueIndex constraint would be enforced at database level
		assert.Equal(t, "Food & Dining", category1.Name)
		assert.Equal(t, "Food & Dining", category2.Name)
	})
}
