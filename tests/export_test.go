package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-finance-advisor/internal/application"
	"go-finance-advisor/internal/domain"
	"go-finance-advisor/internal/infrastructure/api"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestExportCSV(t *testing.T) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto migrate
	db.AutoMigrate(&domain.User{}, &domain.Transaction{})

	// Create test user
	user := domain.User{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}
	db.Create(&user)

	// Create test category
	category := domain.Category{
		Name: "Food",
		Type: "expense",
	}
	db.Create(&category)

	// Create test transaction
	transaction := domain.Transaction{
		UserID:      user.ID,
		CategoryID:  category.ID,
		Amount:      100.50,
		Type:        "expense",
		Description: "Test transaction",
		Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}
	db.Create(&transaction)

	// Setup service and handler
	txSvc := &application.TransactionService{DB: db}
	txHandler := &api.TransactionHandler{Service: txSvc}

	// Setup router
	router := gin.New()
	router.GET("/users/:userId/transactions/export/csv", txHandler.ExportCSV)

	// Create request
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/users/1/transactions/export/csv", http.NoBody)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/csv")
	assert.Contains(t, w.Body.String(), "ID,Amount,Type,Description,Category ID,Date,Created At")
	assert.Contains(t, w.Body.String(), "Test transaction")
}

func TestExportPDF(t *testing.T) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto migrate
	db.AutoMigrate(&domain.User{}, &domain.Transaction{})

	// Create test user
	user := domain.User{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}
	db.Create(&user)

	// Create test category
	category := domain.Category{
		Name: "Food",
		Type: "expense",
	}
	db.Create(&category)

	// Create test transaction
	transaction := domain.Transaction{
		UserID:      user.ID,
		CategoryID:  category.ID,
		Amount:      100.50,
		Type:        "expense",
		Description: "Test transaction",
		Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}
	db.Create(&transaction)

	// Setup service and handler
	txSvc := &application.TransactionService{DB: db}
	txHandler := &api.TransactionHandler{Service: txSvc}

	// Setup router
	router := gin.New()
	router.GET("/users/:userId/transactions/export/pdf", txHandler.ExportPDF)

	// Create request
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/users/1/transactions/export/pdf", http.NoBody)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, w.Body.String(), "Transaction Report")
	assert.Contains(t, w.Body.String(), "Test transaction")
}
