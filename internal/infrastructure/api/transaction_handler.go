package api

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go-finance-advisor/internal/application"
	"go-finance-advisor/internal/domain"

	"github.com/gin-gonic/gin"
)

// TransactionServiceInterface defines the interface for transaction service
type TransactionServiceInterface interface {
	Create(transaction *domain.Transaction) error
	GetByID(id uint) (*domain.Transaction, error)
	ListWithFilters(
		userID uint,
		transactionType *string,
		categoryID *uint,
		startDate, endDate *time.Time,
		limit, offset int,
	) ([]domain.Transaction, error)
	Update(transaction *domain.Transaction) error
	Delete(id uint) error
}

type TransactionHandler struct {
	Service TransactionServiceInterface
}

func NewTransactionHandler(service *application.TransactionService) *TransactionHandler {
	return &TransactionHandler{Service: service}
}

type CreateTransactionRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Type        string  `json:"type" binding:"required,oneof=income expense"`
	Description string  `json:"description" binding:"required,min=1,max=255"`
	CategoryID  uint    `json:"category_id" binding:"required"`
	Date        string  `json:"date,omitempty"`
}

func (h *TransactionHandler) Create(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req CreateTransactionRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": bindErr.Error()})
		return
	}

	// Parse date or use current date
	var transactionDate time.Time
	if req.Date != "" {
		transactionDate, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
			return
		}
	} else {
		transactionDate = time.Now()
	}

	transaction := &domain.Transaction{
		UserID:      uint(userID),
		Amount:      req.Amount,
		Type:        req.Type,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		Date:        transactionDate,
	}

	if err := h.Service.Create(transaction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	c.JSON(http.StatusCreated, transaction)
}

func (h *TransactionHandler) List(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get query parameters for filtering
	transactionTypeStr := c.Query("type")
	categoryIDStr := c.Query("category_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	var transactionType *string
	if transactionTypeStr != "" {
		transactionType = &transactionTypeStr
	}

	var categoryID *uint
	if categoryIDStr != "" {
		catID, parseErr := strconv.ParseUint(categoryIDStr, 10, 32)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
			return
		}
		catIDUint := uint(catID)
		categoryID = &catIDUint
	}

	var startDate, endDate *time.Time
	if startDateStr != "" {
		start, parseErr := time.Parse("2006-01-02", startDateStr)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format. Use YYYY-MM-DD"})
			return
		}
		startDate = &start
	}

	if endDateStr != "" {
		end, parseErr := time.Parse("2006-01-02", endDateStr)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format. Use YYYY-MM-DD"})
			return
		}
		endDate = &end
	}

	transactions, err := h.Service.ListWithFilters(uint(userID), transactionType, categoryID, startDate, endDate, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transactions"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

// GetByID retrieves a transaction by ID
func (h *TransactionHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	transaction, err := h.Service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// Update updates an existing transaction
func (h *TransactionHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	// Get existing transaction
	existingTransaction, err := h.Service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	var req CreateTransactionRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": bindErr.Error()})
		return
	}

	// Parse date if provided
	var transactionDate time.Time
	if req.Date != "" {
		transactionDate, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
			return
		}
	} else {
		transactionDate = existingTransaction.Date
	}

	// Update transaction fields
	existingTransaction.Amount = req.Amount
	existingTransaction.Type = req.Type
	existingTransaction.Description = req.Description
	existingTransaction.CategoryID = req.CategoryID
	existingTransaction.Date = transactionDate

	if err := h.Service.Update(existingTransaction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}

	c.JSON(http.StatusOK, existingTransaction)
}

// Delete deletes a transaction
func (h *TransactionHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	if err := h.Service.Delete(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction deleted successfully"})
}

// ExportCSV exports transactions as CSV
func (h *TransactionHandler) ExportCSV(c *gin.Context) {
	filters, err := h.parseExportFilters(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transactions, err := h.Service.ListWithFilters(
		filters.UserID,
		filters.TransactionType,
		filters.CategoryID,
		filters.StartDate,
		filters.EndDate,
		filters.Limit,
		filters.Offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transactions"})
		return
	}

	buf, err := h.generateCSVContent(transactions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate CSV"})
		return
	}

	// Set headers for file download
	filename := fmt.Sprintf("transactions_%d_%s.csv", filters.UserID, time.Now().Format("20060102_150405"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Length", fmt.Sprintf("%d", buf.Len()))

	c.Data(http.StatusOK, "text/csv", buf.Bytes())
}

// ExportFilters holds export filter parameters
type ExportFilters struct {
	UserID          uint
	TransactionType *string
	CategoryID      *uint
	StartDate       *time.Time
	EndDate         *time.Time
	Limit           int
	Offset          int
}

// parseExportFilters parses and validates export filter parameters
func (h *TransactionHandler) parseExportFilters(c *gin.Context) (*ExportFilters, error) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	transactionTypeStr := c.Query("type")
	categoryIDStr := c.Query("category_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	filters := &ExportFilters{
		UserID: uint(userID),
		Limit:  limit,
		Offset: offset,
	}

	if transactionTypeStr != "" {
		filters.TransactionType = &transactionTypeStr
	}

	if categoryIDStr != "" {
		catID, parseErr := strconv.ParseUint(categoryIDStr, 10, 32)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid category ID")
		}
		catIDUint := uint(catID)
		filters.CategoryID = &catIDUint
	}

	if startDateStr != "" {
		start, parseErr := time.Parse("2006-01-02", startDateStr)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid start date format. Use YYYY-MM-DD")
		}
		filters.StartDate = &start
	}

	if endDateStr != "" {
		end, parseErr := time.Parse("2006-01-02", endDateStr)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid end date format. Use YYYY-MM-DD")
		}
		filters.EndDate = &end
	}

	return filters, nil
}

// generateCSVContent generates CSV content for export
func (h *TransactionHandler) generateCSVContent(transactions []domain.Transaction) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write CSV header
	header := []string{"ID", "Amount", "Type", "Description", "Category ID", "Date", "Created At"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	// Write transaction data
	for i := range transactions {
		transaction := &transactions[i]
		record := []string{
			fmt.Sprintf("%d", transaction.ID),
			fmt.Sprintf("%.2f", transaction.Amount),
			transaction.Type,
			transaction.Description,
			fmt.Sprintf("%d", transaction.CategoryID),
			transaction.Date.Format("2006-01-02"),
			transaction.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return &buf, nil
}

// generateHTMLContent generates HTML content for PDF export
func (h *TransactionHandler) generateHTMLContent(transactions []domain.Transaction) string {
	var htmlContent bytes.Buffer
	htmlContent.WriteString(`<!DOCTYPE html>
<html>
<head>
    <title>Transaction Report</title>
    <style>
        body { font-family: Arial, sans-serif; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .header { text-align: center; margin-bottom: 20px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Transaction Report</h1>
        <p>Generated on: ` + time.Now().Format("2006-01-02 15:04:05") + `</p>
    </div>
    <table>
        <tr>
            <th>ID</th>
            <th>Amount</th>
            <th>Type</th>
            <th>Description</th>
            <th>Category ID</th>
            <th>Date</th>
        </tr>`)

	for i := range transactions {
		transaction := &transactions[i]
		htmlContent.WriteString(fmt.Sprintf(`
        <tr>
            <td>%d</td>
            <td>%.2f</td>
            <td>%s</td>
            <td>%s</td>
            <td>%d</td>
            <td>%s</td>
        </tr>`,
			transaction.ID,
			transaction.Amount,
			transaction.Type,
			transaction.Description,
			transaction.CategoryID,
			transaction.Date.Format("2006-01-02")))
	}

	htmlContent.WriteString(`
    </table>
</body>
</html>`)
	return htmlContent.String()
}

// ExportPDF exports transactions as PDF (simplified version)
func (h *TransactionHandler) ExportPDF(c *gin.Context) {
	filters, err := h.parseExportFilters(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transactions, err := h.Service.ListWithFilters(
		filters.UserID,
		filters.TransactionType,
		filters.CategoryID,
		filters.StartDate,
		filters.EndDate,
		filters.Limit,
		filters.Offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transactions"})
		return
	}

	htmlContent := h.generateHTMLContent(transactions)

	// Set headers for file download
	filename := fmt.Sprintf("transactions_%d_%s.html", filters.UserID, time.Now().Format("20060102_150405"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "text/html")
	c.Header("Content-Length", fmt.Sprintf("%d", len(htmlContent)))

	c.Data(http.StatusOK, "text/html", []byte(htmlContent))
}
