// Package api provides HTTP handlers for the finance advisor application.
package api

import (
	"net/http"
	"strconv"

	"go-finance-advisor/internal/application"
	"go-finance-advisor/internal/domain"

	"github.com/gin-gonic/gin"
)

type CategoryServiceInterface interface {
	InitializeDefaultCategories() error
	GetAllCategories() ([]domain.Category, error)
	GetCategoryByID(categoryID uint) (*domain.Category, error)
	CreateCategory(category *domain.Category) error
	UpdateCategory(categoryID uint, category *domain.Category) error
	DeleteCategory(categoryID uint) error
	GetCategoryUsageStats(userID uint) ([]application.CategoryUsageStats, error)
	GetCategoriesByType(categoryType string) ([]domain.Category, error)
}

type CategoryHandler struct {
	Service CategoryServiceInterface
}

type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Type        string `json:"type" binding:"required,oneof=income expense"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Color       string `json:"color,omitempty"`
}

type UpdateCategoryRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	Color       *string `json:"color,omitempty"`
}

// InitializeDefaultCategories initializes default categories for the system
func (h *CategoryHandler) InitializeDefaultCategories(c *gin.Context) {
	err := h.Service.InitializeDefaultCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize default categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Default categories initialized successfully"})
}

// GetCategories returns all categories with optional filtering
func (h *CategoryHandler) GetCategories(c *gin.Context) {
	categories, err := h.Service.GetAllCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve categories"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// GetCategory returns a specific category by ID
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	categoryIDStr := c.Param("categoryId")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	category, err := h.Service.GetCategoryByID(uint(categoryID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// CreateCategory creates a new custom category
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category := &domain.Category{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Icon:        req.Icon,
		Color:       req.Color,
		IsDefault:   false, // Custom categories are never default
	}

	err := h.Service.CreateCategory(category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// UpdateCategory updates an existing category
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	categoryIDStr := c.Param("categoryId")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	var req UpdateCategoryRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": bindErr.Error()})
		return
	}

	// Get existing category
	category, err := h.Service.GetCategoryByID(uint(categoryID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	// Check if it's a default category (cannot be modified)
	if category.IsDefault {
		c.JSON(http.StatusForbidden, gin.H{"error": "Default categories cannot be modified"})
		return
	}

	// Update fields
	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Description != nil {
		category.Description = *req.Description
	}
	if req.Icon != nil {
		category.Icon = *req.Icon
	}
	if req.Color != nil {
		category.Color = *req.Color
	}

	err = h.Service.UpdateCategory(uint(categoryID), category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}

	// Get the updated category to return
	updatedCategory, err := h.Service.GetCategoryByID(uint(categoryID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated category"})
		return
	}

	c.JSON(http.StatusOK, updatedCategory)
}

// DeleteCategory deletes a category
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	categoryIDStr := c.Param("categoryId")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	// Get existing category
	category, err := h.Service.GetCategoryByID(uint(categoryID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	// Check if it's a default category (cannot be deleted)
	if category.IsDefault {
		c.JSON(http.StatusForbidden, gin.H{"error": "Default categories cannot be deleted"})
		return
	}

	err = h.Service.DeleteCategory(uint(categoryID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}

// GetCategoryUsage returns usage statistics for categories
func (h *CategoryHandler) GetCategoryUsage(c *gin.Context) {
	// Get query parameters
	userIDStr := c.Query("user_id")

	var userID uint
	if userIDStr != "" {
		uID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			return
		}
		userID = uint(uID)
	}

	stats, err := h.Service.GetCategoryUsageStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve category usage"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetIncomeCategories returns all income categories
func (h *CategoryHandler) GetIncomeCategories(c *gin.Context) {
	categories, err := h.Service.GetCategoriesByType("income")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve income categories"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// GetExpenseCategories returns all expense categories
func (h *CategoryHandler) GetExpenseCategories(c *gin.Context) {
	categories, err := h.Service.GetCategoriesByType("expense")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve expense categories"})
		return
	}

	c.JSON(http.StatusOK, categories)
}
