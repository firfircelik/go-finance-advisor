package application

import (
	"go-finance-advisor/internal/domain"
	"gorm.io/gorm"
)

type CategoryService struct {
	DB *gorm.DB
}

type CategoryUsageStats struct {
	CategoryID       uint    `json:"category_id"`
	CategoryName     string  `json:"category_name"`
	CategoryType     string  `json:"category_type"`
	TransactionCount int     `json:"transaction_count"`
	TotalAmount      float64 `json:"total_amount"`
	AverageAmount    float64 `json:"average_amount"`
	LastUsed         *string `json:"last_used"`
}

func NewCategoryService(db *gorm.DB) *CategoryService {
	return &CategoryService{DB: db}
}

// InitializeDefaultCategories creates default categories if they don't exist
func (s *CategoryService) InitializeDefaultCategories() error {
	defaultCategories := domain.GetDefaultCategories()

	for _, category := range defaultCategories {
		var existingCategory domain.Category
		err := s.DB.Where("name = ? AND type = ?", category.Name, category.Type).First(&existingCategory).Error

		if err == gorm.ErrRecordNotFound {
			// Category doesn't exist, create it
			if err := s.DB.Create(&category).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	return nil
}

// CreateCategory creates a new custom category
func (s *CategoryService) CreateCategory(category *domain.Category) error {
	// Check if category with same name and type already exists
	var existingCategory domain.Category
	err := s.DB.Where("name = ? AND type = ?", category.Name, category.Type).First(&existingCategory).Error

	if err == nil {
		return gorm.ErrDuplicatedKey
	}

	if err != gorm.ErrRecordNotFound {
		return err
	}

	category.IsDefault = false
	return s.DB.Create(category).Error
}

// GetAllCategories retrieves all categories
func (s *CategoryService) GetAllCategories() ([]domain.Category, error) {
	var categories []domain.Category
	err := s.DB.Order("type ASC, name ASC").Find(&categories).Error
	return categories, err
}

// GetCategoriesByType retrieves categories by type (income/expense)
func (s *CategoryService) GetCategoriesByType(categoryType string) ([]domain.Category, error) {
	var categories []domain.Category
	err := s.DB.Where("type = ?", categoryType).Order("name ASC").Find(&categories).Error
	return categories, err
}

// GetCategoryByID retrieves a category by ID
func (s *CategoryService) GetCategoryByID(categoryID uint) (*domain.Category, error) {
	var category domain.Category
	err := s.DB.First(&category, categoryID).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// UpdateCategory updates an existing category
func (s *CategoryService) UpdateCategory(categoryID uint, updates *domain.Category) error {
	var category domain.Category
	err := s.DB.First(&category, categoryID).Error
	if err != nil {
		return err
	}

	// Don't allow updating default categories' core properties
	if category.IsDefault {
		// Only allow updating description and color for default categories
		category.Description = updates.Description
		category.Color = updates.Color
	} else {
		// Allow full updates for custom categories
		category.Name = updates.Name
		category.Type = updates.Type
		category.Description = updates.Description
		category.Icon = updates.Icon
		category.Color = updates.Color
	}

	return s.DB.Save(&category).Error
}

// DeleteCategory deletes a category (only custom categories)
func (s *CategoryService) DeleteCategory(categoryID uint) error {
	var category domain.Category
	err := s.DB.First(&category, categoryID).Error
	if err != nil {
		return err
	}

	// Don't allow deleting default categories
	if category.IsDefault {
		return gorm.ErrInvalidData
	}

	// Check if category is being used by any transactions
	var transactionCount int64
	s.DB.Model(&domain.Transaction{}).Where("category_id = ?", categoryID).Count(&transactionCount)

	if transactionCount > 0 {
		return gorm.ErrForeignKeyViolated
	}

	// Check if category is being used by any budgets
	var budgetCount int64
	s.DB.Model(&domain.Budget{}).Where("category_id = ?", categoryID).Count(&budgetCount)

	if budgetCount > 0 {
		return gorm.ErrForeignKeyViolated
	}

	return s.DB.Delete(&category).Error
}

// GetDefaultCategoryByName finds a default category by name and type
func (s *CategoryService) GetDefaultCategoryByName(name, categoryType string) (*domain.Category, error) {
	var category domain.Category
	err := s.DB.Where("name = ? AND type = ? AND is_default = ?", name, categoryType, true).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// GetCategoryUsageStats returns usage statistics for categories
func (s *CategoryService) GetCategoryUsageStats(userID uint) ([]CategoryUsageStats, error) {
	var stats []CategoryUsageStats

	query := `
		SELECT 
			c.id as category_id,
			c.name as category_name,
			c.type as category_type,
			COUNT(t.id) as transaction_count,
			COALESCE(SUM(t.amount), 0) as total_amount,
			COALESCE(AVG(t.amount), 0) as average_amount,
			MAX(t.date) as last_used
		FROM categories c
		LEFT JOIN transactions t ON c.id = t.category_id AND t.user_id = ?
		GROUP BY c.id, c.name, c.type
		ORDER BY transaction_count DESC, total_amount DESC
	`

	err := s.DB.Raw(query, userID).Scan(&stats).Error
	return stats, err
}
