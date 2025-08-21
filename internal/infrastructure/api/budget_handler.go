package api

import (
	"net/http"
	"strconv"
	"time"

	"go-finance-advisor/internal/domain"

	"github.com/gin-gonic/gin"
)

type BudgetServiceInterface interface {
	CreateBudget(budget *domain.Budget) error
	GetBudgetsByUser(userID uint) ([]domain.Budget, error)
	GetBudgetByID(budgetID uint) (*domain.Budget, error)
	UpdateBudget(budgetID uint, updates *domain.Budget) error
	DeleteBudget(budgetID uint) error
	GetBudgetSummary(userID uint) (*domain.BudgetSummary, error)
}

type BudgetHandler struct {
	Service BudgetServiceInterface
}

type CreateBudgetRequest struct {
	CategoryID uint    `json:"category_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Period     string  `json:"period" binding:"required,oneof=weekly monthly quarterly yearly"`
	StartDate  string  `json:"start_date" binding:"required"`
	EndDate    string  `json:"end_date"`
}

type UpdateBudgetRequest struct {
	Amount    *float64 `json:"amount,omitempty"`
	Period    *string  `json:"period,omitempty"`
	StartDate *string  `json:"start_date,omitempty"`
	EndDate   *string  `json:"end_date,omitempty"`
	IsActive  *bool    `json:"is_active,omitempty"`
}

// CreateBudget creates a new budget for a user
func (h *BudgetHandler) CreateBudget(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse start date
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format. Use YYYY-MM-DD"})
		return
	}

	// Parse end date or calculate based on period
	var endDate time.Time
	if req.EndDate != "" {
		endDate, err = time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format. Use YYYY-MM-DD"})
			return
		}
	} else {
		// Calculate end date based on period
		switch req.Period {
		case "weekly":
			endDate = startDate.AddDate(0, 0, 7)
		case "monthly":
			endDate = startDate.AddDate(0, 1, 0)
		case "quarterly":
			endDate = startDate.AddDate(0, 3, 0)
		case "yearly":
			endDate = startDate.AddDate(1, 0, 0)
		}
	}

	budget := &domain.Budget{
		UserID:     uint(userID),
		CategoryID: req.CategoryID,
		Amount:     req.Amount,
		Period:     req.Period,
		StartDate:  startDate,
		EndDate:    endDate,
		Spent:      0,
		IsActive:   true,
	}

	err = h.Service.CreateBudget(budget)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create budget"})
		return
	}

	c.JSON(http.StatusCreated, budget)
}

// GetBudgets returns all budgets for a user
func (h *BudgetHandler) GetBudgets(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	budgets, err := h.Service.GetBudgetsByUser(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve budgets"})
		return
	}

	c.JSON(http.StatusOK, budgets)
}

// GetBudget returns a specific budget
func (h *BudgetHandler) GetBudget(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	budgetIDStr := c.Param("budgetId")
	budgetID, err := strconv.ParseUint(budgetIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid budget ID"})
		return
	}

	budget, err := h.Service.GetBudgetByID(uint(budgetID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
		return
	}

	// Verify budget belongs to user
	if budget.UserID != uint(userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, budget)
}

// UpdateBudget updates an existing budget
func (h *BudgetHandler) UpdateBudget(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	budgetIDStr := c.Param("budgetId")
	budgetID, err := strconv.ParseUint(budgetIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid budget ID"})
		return
	}

	var req UpdateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing budget
	budget, err := h.Service.GetBudgetByID(uint(budgetID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
		return
	}

	// Verify budget belongs to user
	if budget.UserID != uint(userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Update fields
	if req.Amount != nil {
		budget.Amount = *req.Amount
	}
	if req.Period != nil {
		budget.Period = *req.Period
	}
	if req.StartDate != nil {
		startDate, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format. Use YYYY-MM-DD"})
			return
		}
		budget.StartDate = startDate
	}
	if req.EndDate != nil {
		endDate, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format. Use YYYY-MM-DD"})
			return
		}
		budget.EndDate = endDate
	}
	if req.IsActive != nil {
		budget.IsActive = *req.IsActive
	}

	err = h.Service.UpdateBudget(uint(budgetID), budget)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update budget"})
		return
	}

	// Get the updated budget to return
	updatedBudget, err := h.Service.GetBudgetByID(uint(budgetID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated budget"})
		return
	}

	c.JSON(http.StatusOK, updatedBudget)
}

// DeleteBudget deletes a budget
func (h *BudgetHandler) DeleteBudget(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	budgetIDStr := c.Param("budgetId")
	budgetID, err := strconv.ParseUint(budgetIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid budget ID"})
		return
	}

	// Get existing budget to verify ownership
	budget, err := h.Service.GetBudgetByID(uint(budgetID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
		return
	}

	// Verify budget belongs to user
	if budget.UserID != uint(userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	err = h.Service.DeleteBudget(uint(budgetID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete budget"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Budget deleted successfully"})
}

// GetBudgetSummary returns budget performance summary
func (h *BudgetHandler) GetBudgetSummary(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	summary, err := h.Service.GetBudgetSummary(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate budget summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}