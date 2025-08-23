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

// validateUserID validates user ID parameter
func (h *ReportsHandler) validateUserID(c *gin.Context) (uint, bool) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return 0, false
	}
	return uint(userID), true
}

// validateUserIDAndYear validates common parameters for report generation
func (h *ReportsHandler) validateUserIDAndYear(c *gin.Context) (userID uint, year int, valid bool) {
	userID, valid = h.validateUserID(c)
	if !valid {
		return 0, 0, false
	}

	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return 0, 0, false
	}

	return userID, year, true
}

// validatePeriodParam validates and parses a period parameter with given constraints
func (h *ReportsHandler) validatePeriodParam(
	c *gin.Context, paramName string, minVal, maxVal int, errorMsg string,
) (value int, valid bool) {
	paramStr := c.Param(paramName)
	value, err := strconv.Atoi(paramStr)
	if err != nil || value < minVal || value > maxVal {
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
		return 0, false
	}
	return value, true
}

// GenerateMonthlyReport generates a monthly financial report
func (h *ReportsHandler) GenerateMonthlyReport(c *gin.Context) {
	userID, year, valid := h.validateUserIDAndYear(c)
	if !valid {
		return
	}

	month, valid := h.validatePeriodParam(c, "month", 1, 12, "Invalid month")
	if !valid {
		return
	}

	report, err := h.Service.GenerateMonthlyReport(userID, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate monthly report"})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GenerateQuarterlyReport generates a quarterly financial report
func (h *ReportsHandler) GenerateQuarterlyReport(c *gin.Context) {
	userID, year, valid := h.validateUserIDAndYear(c)
	if !valid {
		return
	}

	quarter, valid := h.validatePeriodParam(c, "quarter", 1, 4, "Invalid quarter")
	if !valid {
		return
	}

	report, err := h.Service.GenerateQuarterlyReport(userID, year, quarter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate quarterly report"})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GenerateYearlyReport generates a yearly financial report
func (h *ReportsHandler) GenerateYearlyReport(c *gin.Context) {
	userID, year, valid := h.validateUserIDAndYear(c)
	if !valid {
		return
	}

	report, err := h.Service.GenerateYearlyReport(userID, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate yearly report"})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GenerateCustomReport generates a custom date range financial report
func (h *ReportsHandler) GenerateCustomReport(c *gin.Context) {
	userID, valid := h.validateUserID(c)
	if !valid {
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

	report, err := h.Service.GenerateCustomReport(userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate custom report"})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetReportsList returns a list of generated reports for a user
func (h *ReportsHandler) GetReportsList(c *gin.Context) {
	_, valid := h.validateUserID(c)
	if !valid {
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
