package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-finance-advisor/internal/application"
	"go-finance-advisor/internal/domain"
	"go-finance-advisor/internal/infrastructure/api"
	"go-finance-advisor/internal/infrastructure/persistence"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCreateTransaction(t *testing.T) {
	db := setupTestDB()
	transactionRepo := persistence.NewTransactionRepository(db)
	transactionService := &application.TransactionService{DB: db}
	transactionHandler := api.NewTransactionHandler(transactionService)

	// Create test category
	category := &domain.Category{
		ID:   1,
		Name: "Food",
	}
	db.Create(category)

	// Create test user
	user := &domain.User{
		ID:        1,
		FirstName: "Test",
		LastName:  "User",
		Email:     "test@example.com",
	}
	db.Create(user)

	transactionData := map[string]interface{}{
		"user_id":     1,
		"amount":      50.0,
		"type":        "expense",
		"description": "Test transaction",
		"category_id": 1,
		"date":        "2024-01-15",
	}

	jsonData, _ := json.Marshal(transactionData)
	req := httptest.NewRequest("POST", "/users/1/transactions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/users/:userId/transactions", transactionHandler.Create)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify transaction was created
	transactions, err := transactionRepo.ListByUser(1)
	assert.NoError(t, err)
	assert.Len(t, transactions, 1)
	assert.Equal(t, 50.0, transactions[0].Amount)
	assert.Equal(t, "Test transaction", transactions[0].Description)
}

func TestListTransactions(t *testing.T) {
	db := setupTestDB()
	transactionRepo := persistence.NewTransactionRepository(db)
	transactionService := &application.TransactionService{DB: db}
	transactionHandler := api.NewTransactionHandler(transactionService)

	// Create test category
	category := &domain.Category{
		ID:   1,
		Name: "Food",
	}
	db.Create(category)

	// Create test user
	user := &domain.User{
		ID:        1,
		FirstName: "Test",
		LastName:  "User",
		Email:     "test@example.com",
	}
	db.Create(user)

	// Create test transactions
	transactions := []*domain.Transaction{
		{
			UserID:      1,
			Amount:      -50.0,
			Description: "Grocery shopping",
			CategoryID:  1,
			Date:        time.Now(),
		},
		{
			UserID:      1,
			Amount:      -30.0,
			Description: "Restaurant",
			CategoryID:  1,
			Date:        time.Now().AddDate(0, 0, -1),
		},
	}

	for _, transaction := range transactions {
		err := transactionRepo.Create(transaction)
		assert.NoError(t, err)
	}

	// Test listing transactions
	req := httptest.NewRequest("GET", "/users/1/transactions", nil)
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/users/:userId/transactions", transactionHandler.List)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []domain.Transaction
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "Grocery shopping", response[0].Description)
	assert.Equal(t, "Restaurant", response[1].Description)
}
