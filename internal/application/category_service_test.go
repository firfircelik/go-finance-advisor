package application

import (
	"testing"
	"time"

	"go-finance-advisor/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupCategoryTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&domain.User{}, &domain.Category{}, &domain.Transaction{}, &domain.Budget{})
	require.NoError(t, err)

	return db
}

func createCategoryTestData(t *testing.T, db *gorm.DB) uint {
	// Create test user
	user := &domain.User{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	return user.ID
}

func TestCategoryService_InitializeDefaultCategories(t *testing.T) {
	db := setupCategoryTestDB(t)
	categoryService := &CategoryService{DB: db}

	t.Run("initialize default categories", func(t *testing.T) {
		err := categoryService.InitializeDefaultCategories()
		assert.NoError(t, err)

		// Verify default categories were created
		var categories []domain.Category
		err = db.Where("is_default = ?", true).Find(&categories).Error
		assert.NoError(t, err)
		assert.Greater(t, len(categories), 0)

		// Check for some expected default categories
		var foodCategory domain.Category
		err = db.Where("name = ? AND type = ? AND is_default = ?", "Food & Dining", "expense", true).First(&foodCategory).Error
		assert.NoError(t, err)
		assert.Equal(t, "Food & Dining", foodCategory.Name)
	})

	t.Run("don't duplicate existing categories", func(t *testing.T) {
		// Run initialization again
		err := categoryService.InitializeDefaultCategories()
		assert.NoError(t, err)

		// Count should remain the same
		var count1, count2 int64
		db.Model(&domain.Category{}).Where("is_default = ?", true).Count(&count1)
		
		// Run again
		err = categoryService.InitializeDefaultCategories()
		assert.NoError(t, err)
		db.Model(&domain.Category{}).Where("is_default = ?", true).Count(&count2)
		
		assert.Equal(t, count1, count2)
	})
}

func TestCategoryService_CreateCategory(t *testing.T) {
	db := setupCategoryTestDB(t)
	categoryService := &CategoryService{DB: db}

	t.Run("successful creation", func(t *testing.T) {
		category := &domain.Category{
			Name:        "Custom Category",
			Type:        "expense",
			Description: "A custom category for testing",
			Icon:        "custom-icon",
			Color:       "#FF5733",
		}

		err := categoryService.CreateCategory(category)
		assert.NoError(t, err)
		assert.NotZero(t, category.ID)
		assert.False(t, category.IsDefault)
	})

	t.Run("duplicate category error", func(t *testing.T) {
		// Create first category
		category1 := &domain.Category{
			Name: "Duplicate Test",
			Type: "income",
		}
		err := categoryService.CreateCategory(category1)
		require.NoError(t, err)

		// Try to create duplicate
		category2 := &domain.Category{
			Name: "Duplicate Test",
			Type: "income",
		}
		err = categoryService.CreateCategory(category2)
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrDuplicatedKey, err)
	})

	t.Run("duplicate name not allowed", func(t *testing.T) {
		// Create first category
		category1 := &domain.Category{
			Name: "Duplicate Name",
			Type: "expense",
		}
		err := categoryService.CreateCategory(category1)
		require.NoError(t, err)

		// Try to create category with same name
		category2 := &domain.Category{
			Name: "Duplicate Name",
			Type: "income",
		}
		err = categoryService.CreateCategory(category2)
		assert.Error(t, err)
	})
}

func TestCategoryService_GetAllCategories(t *testing.T) {
	db := setupCategoryTestDB(t)
	categoryService := &CategoryService{DB: db}

	// Create test categories
	categories := []*domain.Category{
		{Name: "Food", Type: "expense", IsDefault: false},
		{Name: "Salary", Type: "income", IsDefault: false},
		{Name: "Transport", Type: "expense", IsDefault: false},
	}

	for _, category := range categories {
		err := db.Create(category).Error
		require.NoError(t, err)
	}

	t.Run("get all categories", func(t *testing.T) {
		result, err := categoryService.GetAllCategories()
		assert.NoError(t, err)
		assert.Len(t, result, 3)
		
		// Should be ordered by type ASC, name ASC
		assert.Equal(t, "expense", result[0].Type)
		assert.Equal(t, "Food", result[0].Name)
		assert.Equal(t, "expense", result[1].Type)
		assert.Equal(t, "Transport", result[1].Name)
		assert.Equal(t, "income", result[2].Type)
		assert.Equal(t, "Salary", result[2].Name)
	})
}

func TestCategoryService_GetCategoriesByType(t *testing.T) {
	db := setupCategoryTestDB(t)
	categoryService := &CategoryService{DB: db}

	// Create test categories
	categories := []*domain.Category{
		{Name: "Food", Type: "expense"},
		{Name: "Transport", Type: "expense"},
		{Name: "Salary", Type: "income"},
		{Name: "Freelance", Type: "income"},
	}

	for _, category := range categories {
		err := db.Create(category).Error
		require.NoError(t, err)
	}

	t.Run("get expense categories", func(t *testing.T) {
		result, err := categoryService.GetCategoriesByType("expense")
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		for _, category := range result {
			assert.Equal(t, "expense", category.Type)
		}
		// Should be ordered by name ASC
		assert.Equal(t, "Food", result[0].Name)
		assert.Equal(t, "Transport", result[1].Name)
	})

	t.Run("get income categories", func(t *testing.T) {
		result, err := categoryService.GetCategoriesByType("income")
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		for _, category := range result {
			assert.Equal(t, "income", category.Type)
		}
	})

	t.Run("get non-existent type", func(t *testing.T) {
		result, err := categoryService.GetCategoriesByType("invalid")
		assert.NoError(t, err)
		assert.Len(t, result, 0)
	})
}

func TestCategoryService_GetCategoryByID(t *testing.T) {
	db := setupCategoryTestDB(t)
	categoryService := &CategoryService{DB: db}

	// Create test category
	category := &domain.Category{
		Name:        "Test Category",
		Type:        "expense",
		Description: "Test description",
	}
	err := db.Create(category).Error
	require.NoError(t, err)

	t.Run("get existing category", func(t *testing.T) {
		result, err := categoryService.GetCategoryByID(category.ID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, category.ID, result.ID)
		assert.Equal(t, "Test Category", result.Name)
		assert.Equal(t, "Test description", result.Description)
	})

	t.Run("get non-existent category", func(t *testing.T) {
		result, err := categoryService.GetCategoryByID(99999)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCategoryService_UpdateCategory(t *testing.T) {
	db := setupCategoryTestDB(t)
	categoryService := &CategoryService{DB: db}

	t.Run("update custom category", func(t *testing.T) {
		// Create custom category
		category := &domain.Category{
			Name:        "Original Name",
			Type:        "expense",
			Description: "Original description",
			IsDefault:   false,
		}
		err := db.Create(category).Error
		require.NoError(t, err)

		// Update category
		updates := &domain.Category{
			Name:        "Updated Name",
			Type:        "income",
			Description: "Updated description",
			Icon:        "new-icon",
			Color:       "#FF0000",
		}

		err = categoryService.UpdateCategory(category.ID, updates)
		assert.NoError(t, err)

		// Verify update
		updated, err := categoryService.GetCategoryByID(category.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.Name)
		assert.Equal(t, "income", updated.Type)
		assert.Equal(t, "Updated description", updated.Description)
		assert.Equal(t, "new-icon", updated.Icon)
		assert.Equal(t, "#FF0000", updated.Color)
	})

	t.Run("update default category (limited)", func(t *testing.T) {
		// Create default category
		category := &domain.Category{
			Name:        "Default Category",
			Type:        "expense",
			Description: "Original description",
			IsDefault:   true,
		}
		err := db.Create(category).Error
		require.NoError(t, err)

		// Try to update default category
		updates := &domain.Category{
			Name:        "Should Not Change",
			Type:        "income",
			Description: "Updated description",
			Color:       "#00FF00",
		}

		err = categoryService.UpdateCategory(category.ID, updates)
		assert.NoError(t, err)

		// Verify only description and color were updated
		updated, err := categoryService.GetCategoryByID(category.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Default Category", updated.Name) // Should not change
		assert.Equal(t, "expense", updated.Type)          // Should not change
		assert.Equal(t, "Updated description", updated.Description) // Should change
		assert.Equal(t, "#00FF00", updated.Color)         // Should change
	})

	t.Run("update non-existent category", func(t *testing.T) {
		updates := &domain.Category{Name: "New Name"}
		err := categoryService.UpdateCategory(99999, updates)
		assert.Error(t, err)
	})
}

func TestCategoryService_DeleteCategory(t *testing.T) {
	db := setupCategoryTestDB(t)
	categoryService := &CategoryService{DB: db}
	userID := createCategoryTestData(t, db)

	t.Run("delete custom category without references", func(t *testing.T) {
		// Create custom category
		category := &domain.Category{
			Name:      "To Delete",
			Type:      "expense",
			IsDefault: false,
		}
		err := db.Create(category).Error
		require.NoError(t, err)

		err = categoryService.DeleteCategory(category.ID)
		assert.NoError(t, err)

		// Verify deletion
		_, err = categoryService.GetCategoryByID(category.ID)
		assert.Error(t, err)
	})

	t.Run("cannot delete default category", func(t *testing.T) {
		// Create default category
		category := &domain.Category{
			Name:      "Default Category",
			Type:      "expense",
			IsDefault: true,
		}
		err := db.Create(category).Error
		require.NoError(t, err)

		err = categoryService.DeleteCategory(category.ID)
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrInvalidData, err)
	})

	t.Run("cannot delete category with transactions", func(t *testing.T) {
		// Create category
		category := &domain.Category{
			Name:      "Used Category",
			Type:      "expense",
			IsDefault: false,
		}
		err := db.Create(category).Error
		require.NoError(t, err)

		// Create transaction using this category
		transaction := &domain.Transaction{
			UserID:     userID,
			CategoryID: category.ID,
			Amount:     100.00,
			Type:       "expense",
			Date:       time.Now(),
		}
		err = db.Create(transaction).Error
		require.NoError(t, err)

		err = categoryService.DeleteCategory(category.ID)
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrForeignKeyViolated, err)
	})

	t.Run("cannot delete category with budgets", func(t *testing.T) {
		// Create category
		category := &domain.Category{
			Name:      "Budgeted Category",
			Type:      "expense",
			IsDefault: false,
		}
		err := db.Create(category).Error
		require.NoError(t, err)

		// Create budget using this category
		budget := &domain.Budget{
			UserID:     userID,
			CategoryID: category.ID,
			Amount:     500.00,
			StartDate:  time.Now(),
			EndDate:    time.Now().AddDate(0, 1, 0),
		}
		err = db.Create(budget).Error
		require.NoError(t, err)

		err = categoryService.DeleteCategory(category.ID)
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrForeignKeyViolated, err)
	})
}

func TestCategoryService_GetDefaultCategoryByName(t *testing.T) {
	db := setupCategoryTestDB(t)
	categoryService := &CategoryService{DB: db}

	// Create test categories with unique names
	defaultCategory := &domain.Category{
		Name:      "Test Default Category",
		Type:      "expense",
		IsDefault: true,
	}
	customCategory := &domain.Category{
		Name:      "Test Custom Category",
		Type:      "income",
		IsDefault: false,
	}

	err := db.Create(defaultCategory).Error
	require.NoError(t, err)
	err = db.Create(customCategory).Error
	require.NoError(t, err)

	t.Run("find default category", func(t *testing.T) {
		result, err := categoryService.GetDefaultCategoryByName("Test Default Category", "expense")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, defaultCategory.ID, result.ID)
		assert.True(t, result.IsDefault)
	})

	t.Run("default category not found", func(t *testing.T) {
		result, err := categoryService.GetDefaultCategoryByName("NonExistent", "expense")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("custom category not returned", func(t *testing.T) {
		result, err := categoryService.GetDefaultCategoryByName("Food", "income")
		assert.Error(t, err) // Should not find the custom category
		assert.Nil(t, result)
	})
}

func TestCategoryService_GetCategoryUsageStats(t *testing.T) {
	db := setupCategoryTestDB(t)
	categoryService := &CategoryService{DB: db}
	userID := createCategoryTestData(t, db)

	// Create test categories
	categories := []*domain.Category{
		{Name: "Food", Type: "expense"},
		{Name: "Transport", Type: "expense"},
		{Name: "Salary", Type: "income"},
	}

	for _, category := range categories {
		err := db.Create(category).Error
		require.NoError(t, err)
	}

	// Create test transactions
	now := time.Now()
	transactions := []*domain.Transaction{
		{
			UserID:     userID,
			CategoryID: categories[0].ID, // Food
			Amount:     50.00,
			Type:       "expense",
			Date:       now.AddDate(0, 0, -1),
		},
		{
			UserID:     userID,
			CategoryID: categories[0].ID, // Food
			Amount:     30.00,
			Type:       "expense",
			Date:       now.AddDate(0, 0, -2),
		},
		{
			UserID:     userID,
			CategoryID: categories[2].ID, // Salary
			Amount:     2000.00,
			Type:       "income",
			Date:       now.AddDate(0, 0, -3),
		},
		// Transport category has no transactions
	}

	for _, tx := range transactions {
		err := db.Create(tx).Error
		require.NoError(t, err)
	}

	t.Run("get category usage stats", func(t *testing.T) {
		stats, err := categoryService.GetCategoryUsageStats(userID)
		assert.NoError(t, err)
		assert.Len(t, stats, 3)

		// Should be ordered by transaction_count DESC, total_amount DESC
		// Food should be first (2 transactions)
		assert.Equal(t, "Food", stats[0].CategoryName)
		assert.Equal(t, 2, stats[0].TransactionCount)
		assert.Equal(t, 80.00, stats[0].TotalAmount)
		assert.Equal(t, 40.00, stats[0].AverageAmount)
		assert.NotNil(t, stats[0].LastUsed)

		// Salary should be second (1 transaction, higher amount)
		assert.Equal(t, "Salary", stats[1].CategoryName)
		assert.Equal(t, 1, stats[1].TransactionCount)
		assert.Equal(t, 2000.00, stats[1].TotalAmount)
		assert.Equal(t, 2000.00, stats[1].AverageAmount)

		// Transport should be last (0 transactions)
		assert.Equal(t, "Transport", stats[2].CategoryName)
		assert.Equal(t, 0, stats[2].TransactionCount)
		assert.Equal(t, 0.00, stats[2].TotalAmount)
		assert.Equal(t, 0.00, stats[2].AverageAmount)
		assert.Nil(t, stats[2].LastUsed)
	})
}