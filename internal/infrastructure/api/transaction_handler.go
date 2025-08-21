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

type TransactionHandler struct {
    Service *application.TransactionService
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
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    var req CreateTransactionRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    // Get query parameters for filtering
    transactionType := c.Query("type")
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

    var categoryID *uint
    if categoryIDStr != "" {
        catID, err := strconv.ParseUint(categoryIDStr, 10, 32)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
            return
        }
        catIDUint := uint(catID)
        categoryID = &catIDUint
    }

    var startDate, endDate *time.Time
    if startDateStr != "" {
        start, err := time.Parse("2006-01-02", startDateStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format. Use YYYY-MM-DD"})
            return
        }
        startDate = &start
    }

    if endDateStr != "" {
        end, err := time.Parse("2006-01-02", endDateStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format. Use YYYY-MM-DD"})
            return
        }
        endDate = &end
    }

    transactions, err := h.Service.ListWithFilters(uint(userID), &transactionType, categoryID, startDate, endDate, limit, offset)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transactions"})
        return
    }

    c.JSON(http.StatusOK, transactions)
}

// ExportCSV exports transactions as CSV
func (h *TransactionHandler) ExportCSV(c *gin.Context) {
    userIDStr := c.Param("userId")
    userID, err := strconv.ParseUint(userIDStr, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    // Get query parameters for filtering
    transactionType := c.Query("type")
    categoryIDStr := c.Query("category_id")
    startDateStr := c.Query("start_date")
    endDateStr := c.Query("end_date")
    limitStr := c.DefaultQuery("limit", "100")
    offsetStr := c.DefaultQuery("offset", "0")

    limit, _ := strconv.Atoi(limitStr)
    offset, _ := strconv.Atoi(offsetStr)

    var categoryID *uint
    if categoryIDStr != "" {
        catID, err := strconv.ParseUint(categoryIDStr, 10, 32)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
            return
        }
        catIDUint := uint(catID)
        categoryID = &catIDUint
    }

    var startDate, endDate *time.Time
    if startDateStr != "" {
        start, err := time.Parse("2006-01-02", startDateStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format. Use YYYY-MM-DD"})
            return
        }
        startDate = &start
    }

    if endDateStr != "" {
        end, err := time.Parse("2006-01-02", endDateStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format. Use YYYY-MM-DD"})
            return
        }
        endDate = &end
    }

    transactions, err := h.Service.ListWithFilters(uint(userID), &transactionType, categoryID, startDate, endDate, limit, offset)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transactions"})
        return
    }

    // Create CSV buffer
    var buf bytes.Buffer
    writer := csv.NewWriter(&buf)

    // Write CSV header
    header := []string{"ID", "Amount", "Type", "Description", "Category ID", "Date", "Created At"}
    if err := writer.Write(header); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write CSV header"})
        return
    }

    // Write transaction data
    for _, transaction := range transactions {
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
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write CSV record"})
            return
        }
    }

    writer.Flush()
    if err := writer.Error(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to flush CSV writer"})
        return
    }

    // Set headers for file download
    filename := fmt.Sprintf("transactions_%d_%s.csv", userID, time.Now().Format("20060102_150405"))
    c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
    c.Header("Content-Type", "text/csv")
    c.Header("Content-Length", fmt.Sprintf("%d", buf.Len()))

    c.Data(http.StatusOK, "text/csv", buf.Bytes())
}

// ExportPDF exports transactions as PDF (simplified version)
func (h *TransactionHandler) ExportPDF(c *gin.Context) {
    userIDStr := c.Param("userId")
    userID, err := strconv.ParseUint(userIDStr, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    // Get query parameters for filtering
    transactionType := c.Query("type")
    categoryIDStr := c.Query("category_id")
    startDateStr := c.Query("start_date")
    endDateStr := c.Query("end_date")
    limitStr := c.DefaultQuery("limit", "100")
    offsetStr := c.DefaultQuery("offset", "0")

    limit, _ := strconv.Atoi(limitStr)
    offset, _ := strconv.Atoi(offsetStr)

    var categoryID *uint
    if categoryIDStr != "" {
        catID, err := strconv.ParseUint(categoryIDStr, 10, 32)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
            return
        }
        catIDUint := uint(catID)
        categoryID = &catIDUint
    }

    var startDate, endDate *time.Time
    if startDateStr != "" {
        start, err := time.Parse("2006-01-02", startDateStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format. Use YYYY-MM-DD"})
            return
        }
        startDate = &start
    }

    if endDateStr != "" {
        end, err := time.Parse("2006-01-02", endDateStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format. Use YYYY-MM-DD"})
            return
        }
        endDate = &end
    }

    transactions, err := h.Service.ListWithFilters(uint(userID), &transactionType, categoryID, startDate, endDate, limit, offset)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transactions"})
        return
    }

    // Create simple HTML content for PDF (this is a basic implementation)
    // In a real application, you would use a proper PDF library like gofpdf
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

    for _, transaction := range transactions {
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

    // Set headers for file download
    filename := fmt.Sprintf("transactions_%d_%s.html", userID, time.Now().Format("20060102_150405"))
    c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
    c.Header("Content-Type", "text/html")
    c.Header("Content-Length", fmt.Sprintf("%d", htmlContent.Len()))

    c.Data(http.StatusOK, "text/html", htmlContent.Bytes())
}