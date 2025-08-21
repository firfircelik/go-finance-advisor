package api

import (
	"net/http"
	"strconv"
	"time"

	"go-finance-advisor/internal/domain"

	"github.com/gin-gonic/gin"
)

// ReportsServiceInterface defines the contract for reports service operations
type ReportsServiceInterface interface {
	GenerateMonthlyReport(userID uint, year, month int) (*domain.FinancialReport, error)
	GenerateQuarterlyReport(userID uint, year, quarter int) (*domain.FinancialReport, error)
	GenerateYearlyReport(userID uint, year int) (*domain.FinancialReport, error)
	GenerateCustomReport(userID uint, startDate, endDate time.Time) (*domain.FinancialReport, error)
}

type ReportsHandler struct {
	Service ReportsServiceInterface
}

func NewReportsHandler(service ReportsServiceInterface) *ReportsHandler {
	return &ReportsHandler{Service: service}
}

// GenerateMonthlyReport generates a monthly financial report
func (h *ReportsHandler) GenerateMonthlyReport(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	monthStr := c.Param("month")
	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month"})
		return
	}

	report, err := h.Service.GenerateMonthlyReport(uint(userID), year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate monthly report"})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GenerateQuarterlyReport generates a quarterly financial report
func (h *ReportsHandler) GenerateQuarterlyReport(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	quarterStr := c.Param("quarter")
	quarter, err := strconv.Atoi(quarterStr)
	if err != nil || quarter < 1 || quarter > 4 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quarter"})
		return
	}

	report, err := h.Service.GenerateQuarterlyReport(uint(userID), year, quarter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate quarterly report"})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GenerateYearlyReport generates a yearly financial report
func (h *ReportsHandler) GenerateYearlyReport(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	report, err := h.Service.GenerateYearlyReport(uint(userID), year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate yearly report"})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GenerateCustomReport generates a custom date range financial report
func (h *ReportsHandler) GenerateCustomReport(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD"})
		return
	}

	if startDate.After(endDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date must be before end_date"})
		return
	}

	report, err := h.Service.GenerateCustomReport(uint(userID), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate custom report"})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetReportsList returns a list of generated reports for a user
func (h *ReportsHandler) GetReportsList(c *gin.Context) {
	userIDStr := c.Param("userId")
	_, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// For now, return an empty list since we're not storing reports in DB
	// In a real implementation, you would query the database for saved reports
	reports := []interface{}{}

	c.JSON(http.StatusOK, gin.H{
		"reports": reports,
		"message": "Reports are generated on-demand. Use the generate endpoints to create new reports.",
	})
}