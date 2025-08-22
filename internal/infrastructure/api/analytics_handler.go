package api

import (
	"net/http"
	"strconv"
	"time"

	"go-finance-advisor/internal/application"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	Service application.AnalyticsServiceInterface
}

// GetFinancialMetrics returns comprehensive financial metrics for a user
func (h *AnalyticsHandler) GetFinancialMetrics(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get query parameters
	period := c.DefaultQuery("period", "monthly")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Parse dates or use defaults
	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format. Use YYYY-MM-DD"})
			return
		}
	} else {
		// Default to current month
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format. Use YYYY-MM-DD"})
			return
		}
	} else {
		// Default to end of current month
		now := time.Now()
		endDate = time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, now.Location())
	}

	metrics, err := h.Service.GetFinancialMetrics(uint(userID), period, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate financial metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetIncomeExpenseAnalysis returns detailed income vs expense analysis
func (h *AnalyticsHandler) GetIncomeExpenseAnalysis(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get query parameters
	period := c.DefaultQuery("period", "monthly")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Parse dates or use defaults
	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format. Use YYYY-MM-DD"})
			return
		}
	} else {
		// Default to current month
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format. Use YYYY-MM-DD"})
			return
		}
	} else {
		// Default to end of current month
		now := time.Now()
		endDate = time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, now.Location())
	}

	analysis, err := h.Service.GetIncomeExpenseAnalysis(uint(userID), period, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate income-expense analysis"})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// GetCategoryAnalysis returns detailed analysis for a specific category
func (h *AnalyticsHandler) GetCategoryAnalysis(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	categoryIDStr := c.Param("categoryId")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	// Get query parameters
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Parse dates or use defaults
	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format. Use YYYY-MM-DD"})
			return
		}
	} else {
		// Default to current month
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format. Use YYYY-MM-DD"})
			return
		}
	} else {
		// Default to end of current month
		now := time.Now()
		endDate = time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, now.Location())
	}

	analysis, err := h.Service.GetCategoryAnalysis(uint(userID), uint(categoryID), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate category analysis"})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// GetDashboardSummary returns a comprehensive dashboard summary
func (h *AnalyticsHandler) GetDashboardSummary(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get query parameters
	period := c.DefaultQuery("period", "month")

	// Get dashboard summary using the new service method
	dashboard, err := h.Service.GetDashboardSummary(uint(userID), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate dashboard summary"})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}
