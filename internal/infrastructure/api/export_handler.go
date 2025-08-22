package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go-finance-advisor/internal/application"
	"go-finance-advisor/internal/domain"
)

// ExportServiceInterface defines the interface for export service
type ExportServiceInterface interface {
	ExportTransactions(userID uint, format domain.ExportFormat, startDate, endDate *time.Time) ([]byte, string, error)
	ExportFinancialReport(userID uint, reportType string, year, month int, format domain.ExportFormat) ([]byte, string, error)
	ExportAllData(userID uint, format domain.ExportFormat) ([]byte, string, error)
	ExportBudgets(userID uint, format domain.ExportFormat) ([]byte, string, error)
}

type ExportHandler struct {
	Service ExportServiceInterface
}

func NewExportHandler(service *application.ExportService) *ExportHandler {
	return &ExportHandler{Service: service}
}

// ExportTransactions exports user transactions
// @Summary Export transactions
// @Description Export user transactions in CSV or JSON format
// @Tags export
// @Accept json
// @Produce application/octet-stream
// @Param format query string true "Export format (csv, json)"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {file} file "Exported file"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/export/transactions [get]
func (h *ExportHandler) ExportTransactions(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse format
	formatStr := c.Query("format")
	if formatStr == "" {
		formatStr = "csv"
	}
	format := domain.ExportFormat(formatStr)
	if !format.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid export format"})
		return
	}

	// Parse date range
	var startDate, endDate *time.Time
	if startStr := c.Query("start_date"); startStr != "" {
		if parsed, err := time.Parse("2006-01-02", startStr); err == nil {
			startDate = &parsed
		}
	}
	if endStr := c.Query("end_date"); endStr != "" {
		if parsed, err := time.Parse("2006-01-02", endStr); err == nil {
			endDate = &parsed
		}
	}

	// Export data
	data, filename, err := h.Service.ExportTransactions(userID.(uint), format, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set response headers
	c.Header("Content-Type", format.GetContentType())
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", strconv.Itoa(len(data)))

	// Send file
	c.Data(http.StatusOK, format.GetContentType(), data)
}

// ExportBudgets exports user budgets
// @Summary Export budgets
// @Description Export user budgets in CSV or JSON format
// @Tags export
// @Accept json
// @Produce application/octet-stream
// @Param format query string true "Export format (csv, json)"
// @Success 200 {file} file "Exported file"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/export/budgets [get]
func (h *ExportHandler) ExportBudgets(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse format
	formatStr := c.Query("format")
	if formatStr == "" {
		formatStr = "csv"
	}
	format := domain.ExportFormat(formatStr)
	if !format.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid export format"})
		return
	}

	// Export data
	data, filename, err := h.Service.ExportBudgets(userID.(uint), format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set response headers
	c.Header("Content-Type", format.GetContentType())
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", strconv.Itoa(len(data)))

	// Send file
	c.Data(http.StatusOK, format.GetContentType(), data)
}

// ExportFinancialReport exports a financial report
// @Summary Export financial report
// @Description Export a financial report in CSV or JSON format
// @Tags export
// @Accept json
// @Produce application/octet-stream
// @Param type query string true "Report type (monthly, yearly)"
// @Param year query int true "Year"
// @Param month query int false "Month (required for monthly reports)"
// @Param format query string true "Export format (csv, json)"
// @Success 200 {file} file "Exported file"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/export/reports [get]
func (h *ExportHandler) ExportFinancialReport(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse parameters
	reportType := c.Query("type")
	if reportType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Report type is required"})
		return
	}

	yearStr := c.Query("year")
	if yearStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Year is required"})
		return
	}
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	month := 0
	if monthStr := c.Query("month"); monthStr != "" {
		month, err = strconv.Atoi(monthStr)
		if err != nil || month < 1 || month > 12 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month"})
			return
		}
	}

	if reportType == "monthly" && month == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Month is required for monthly reports"})
		return
	}

	// Parse format
	formatStr := c.Query("format")
	if formatStr == "" {
		formatStr = "json"
	}
	format := domain.ExportFormat(formatStr)
	if !format.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid export format"})
		return
	}

	// Export data
	data, filename, err := h.Service.ExportFinancialReport(userID.(uint), reportType, year, month, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set response headers
	c.Header("Content-Type", format.GetContentType())
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", strconv.Itoa(len(data)))

	// Send file
	c.Data(http.StatusOK, format.GetContentType(), data)
}

// ExportAllData exports all user financial data
// @Summary Export all data
// @Description Export all user financial data in JSON format
// @Tags export
// @Accept json
// @Produce application/octet-stream
// @Param format query string false "Export format (json only)"
// @Success 200 {file} file "Exported file"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/export/all [get]
func (h *ExportHandler) ExportAllData(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse format (only JSON supported for all data export)
	formatStr := c.Query("format")
	if formatStr == "" {
		formatStr = "json"
	}
	format := domain.ExportFormat(formatStr)
	if format != domain.ExportFormatJSON {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only JSON format is supported for all data export"})
		return
	}

	// Export data
	data, filename, err := h.Service.ExportAllData(userID.(uint), format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set response headers
	c.Header("Content-Type", format.GetContentType())
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", strconv.Itoa(len(data)))

	// Send file
	c.Data(http.StatusOK, format.GetContentType(), data)
}

// GetExportFormats returns available export formats
// @Summary Get export formats
// @Description Get list of available export formats
// @Tags export
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/export/formats [get]
func (h *ExportHandler) GetExportFormats(c *gin.Context) {
	formats := map[string]interface{}{
		"formats": []map[string]string{
			{
				"value":       "csv",
				"label":       "CSV",
				"description": "Comma-separated values format",
				"mime_type":   "text/csv",
				"extension":   ".csv",
			},
			{
				"value":       "json",
				"label":       "JSON",
				"description": "JavaScript Object Notation format",
				"mime_type":   "application/json",
				"extension":   ".json",
			},
		},
		"data_types": []map[string]interface{}{
			{
				"value":               "transactions",
				"label":               "Transactions",
				"description":         "Export transaction data",
				"supported_formats":   []string{"csv", "json"},
				"supports_date_range": true,
			},
			{
				"value":               "budgets",
				"label":               "Budgets",
				"description":         "Export budget data",
				"supported_formats":   []string{"csv", "json"},
				"supports_date_range": false,
			},
			{
				"value":               "reports",
				"label":               "Financial Reports",
				"description":         "Export financial reports",
				"supported_formats":   []string{"csv", "json"},
				"supports_date_range": false,
			},
			{
				"value":               "all",
				"label":               "All Data",
				"description":         "Export all financial data",
				"supported_formats":   []string{"json"},
				"supports_date_range": false,
			},
		},
	}

	c.JSON(http.StatusOK, formats)
}
