package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-finance-advisor/internal/application"
	"go-finance-advisor/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)



// MockTransactionService is a mock implementation of TransactionService
type MockTransactionService struct {
	mock.Mock
}

func (m *MockTransactionService) Create(transaction *domain.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockTransactionService) GetByID(id uint) (*domain.Transaction, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

func (m *MockTransactionService) GetByUserID(userID uint) ([]domain.Transaction, error) {
	args := m.Called(userID)
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func (m *MockTransactionService) ListWithFilters(userID uint, transactionType *string, categoryID *uint, startDate, endDate *time.Time, limit, offset int) ([]domain.Transaction, error) {
	args := m.Called(userID, transactionType, categoryID, startDate, endDate, limit, offset)
	return args.Get(0).([]domain.Transaction), args.Error(1)
}



func (m *MockTransactionService) Update(transaction *domain.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockTransactionService) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTransactionService) GetTransactionsByDateRange(userID uint, startDate, endDate time.Time) ([]domain.Transaction, error) {
	args := m.Called(userID, startDate, endDate)
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func (m *MockTransactionService) GetTransactionsByCategory(userID, categoryID uint) ([]domain.Transaction, error) {
	args := m.Called(userID, categoryID)
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func (m *MockTransactionService) GetTransactionsByType(userID uint, transactionType string) ([]domain.Transaction, error) {
	args := m.Called(userID, transactionType)
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func setupTransactionHandler() (*TransactionHandler, *MockTransactionService) {
	mockService := new(MockTransactionService)
	handler := &TransactionHandler{Service: mockService}
	return handler, mockService
}

func TestNewTransactionHandler(t *testing.T) {
	t.Run("should create new transaction handler", func(t *testing.T) {
		service := &application.TransactionService{}
		handler := NewTransactionHandler(service)

		assert.NotNil(t, handler)
		assert.Equal(t, service, handler.Service)
	})
}

func TestTransactionHandler_Create(t *testing.T) {
	t.Run("should create transaction successfully", func(t *testing.T) {
		handler, mockService := setupTransactionHandler()
		router := setupGin()
		router.POST("/users/:userId/transactions", handler.Create)

		createReq := CreateTransactionRequest{
			Amount:      100.50,
			Type:        "expense",
			Description: "Grocery shopping",
			CategoryID:  1,
			Date:        "2024-01-15",
		}

		mockService.On("Create", mock.MatchedBy(func(t *domain.Transaction) bool {
			return t.Amount == 100.50 && t.Type == "expense" && t.UserID == 1
		})).Return(nil)

		requestBody, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/users/1/transactions", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should use current date when date not provided", func(t *testing.T) {
		handler, mockService := setupTransactionHandler()
		router := setupGin()
		router.POST("/users/:userId/transactions", handler.Create)

		createReq := CreateTransactionRequest{
			Amount:      50.00,
			Type:        "income",
			Description: "Freelance work",
			CategoryID:  2,
			// Date not provided
		}

		mockService.On("Create", mock.MatchedBy(func(t *domain.Transaction) bool {
			// Check that date is set to today
			now := time.Now()
			return t.Date.Year() == now.Year() && t.Date.Month() == now.Month() && t.Date.Day() == now.Day()
		})).Return(nil)

		requestBody, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/users/1/transactions", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid user ID", func(t *testing.T) {
		handler, _ := setupTransactionHandler()
		router := setupGin()
		router.POST("/users/:userId/transactions", handler.Create)

		createReq := CreateTransactionRequest{
			Amount:      100.50,
			Type:        "expense",
			Description: "Test transaction",
			CategoryID:  1,
		}

		requestBody, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/users/invalid/transactions", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid user ID", response["error"])
	})

	t.Run("should return bad request for invalid JSON", func(t *testing.T) {
		handler, _ := setupTransactionHandler()
		router := setupGin()
		router.POST("/users/:userId/transactions", handler.Create)

		req := httptest.NewRequest("POST", "/users/1/transactions", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return bad request for invalid amount", func(t *testing.T) {
		handler, _ := setupTransactionHandler()
		router := setupGin()
		router.POST("/users/:userId/transactions", handler.Create)

		createReq := CreateTransactionRequest{
			Amount:      -10.00, // Invalid negative amount
			Type:        "expense",
			Description: "Test transaction",
			CategoryID:  1,
		}

		requestBody, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/users/1/transactions", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return bad request for invalid transaction type", func(t *testing.T) {
		handler, _ := setupTransactionHandler()
		router := setupGin()
		router.POST("/users/:userId/transactions", handler.Create)

		createReq := CreateTransactionRequest{
			Amount:      100.50,
			Type:        "invalid", // Invalid type
			Description: "Test transaction",
			CategoryID:  1,
		}

		requestBody, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/users/1/transactions", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return bad request for invalid date format", func(t *testing.T) {
		handler, _ := setupTransactionHandler()
		router := setupGin()
		router.POST("/users/:userId/transactions", handler.Create)

		createReq := CreateTransactionRequest{
			Amount:      100.50,
			Type:        "expense",
			Description: "Test transaction",
			CategoryID:  1,
			Date:        "invalid-date",
		}

		requestBody, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/users/1/transactions", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid date format. Use YYYY-MM-DD", response["error"])
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		handler, mockService := setupTransactionHandler()
		router := setupGin()
		router.POST("/users/:userId/transactions", handler.Create)

		createReq := CreateTransactionRequest{
			Amount:      100.50,
			Type:        "expense",
			Description: "Test transaction",
			CategoryID:  1,
		}

		mockService.On("Create", mock.AnythingOfType("*domain.Transaction")).Return(errors.New("database error"))

		requestBody, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/users/1/transactions", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Failed to create transaction", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestTransactionHandler_List(t *testing.T) {
	t.Run("should list transactions successfully", func(t *testing.T) {
		handler, mockService := setupTransactionHandler()
		router := setupGin()
		router.GET("/users/:userId/transactions", handler.List)

		expectedTransactions := []domain.Transaction{
			{
				ID:          1,
				UserID:      1,
				Amount:      100.50,
				Type:        "expense",
				Description: "Grocery shopping",
				CategoryID:  1,
				Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			},
			{
				ID:          2,
				UserID:      1,
				Amount:      2000.00,
				Type:        "income",
				Description: "Salary",
				CategoryID:  2,
				Date:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		}

		mockService.On("ListWithFilters", uint(1), (*string)(nil), (*uint)(nil), (*time.Time)(nil), (*time.Time)(nil), 100, 0).Return(expectedTransactions, nil)

		req := httptest.NewRequest("GET", "/users/1/transactions", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []domain.Transaction
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response, 2)
		assert.Equal(t, expectedTransactions[0].ID, response[0].ID)
		assert.Equal(t, expectedTransactions[1].ID, response[1].ID)
		mockService.AssertExpectations(t)
	})

	t.Run("should list transactions with filters", func(t *testing.T) {
		handler, mockService := setupTransactionHandler()
		router := setupGin()
		router.GET("/users/:userId/transactions", handler.List)

		expectedTransactions := []domain.Transaction{
			{
				ID:          1,
				UserID:      1,
				Amount:      100.50,
				Type:        "expense",
				Description: "Grocery shopping",
				CategoryID:  1,
				Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			},
		}

		expenseType := "expense"
		categoryID := uint(1)
		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

		mockService.On("ListWithFilters", uint(1), &expenseType, &categoryID, &startDate, &endDate, 50, 10).Return(expectedTransactions, nil)

		req := httptest.NewRequest("GET", "/users/1/transactions?type=expense&category_id=1&start_date=2024-01-01&end_date=2024-01-31&limit=50&offset=10", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []domain.Transaction
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response, 1)
		assert.Equal(t, expectedTransactions[0].ID, response[0].ID)
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid user ID", func(t *testing.T) {
		handler, _ := setupTransactionHandler()
		router := setupGin()
		router.GET("/users/:userId/transactions", handler.List)

		req := httptest.NewRequest("GET", "/users/invalid/transactions", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid user ID", response["error"])
	})

	t.Run("should return bad request for invalid category ID", func(t *testing.T) {
		handler, _ := setupTransactionHandler()
		router := setupGin()
		router.GET("/users/:userId/transactions", handler.List)

		req := httptest.NewRequest("GET", "/users/1/transactions?category_id=invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid category ID", response["error"])
	})

	t.Run("should return bad request for invalid date format", func(t *testing.T) {
		handler, _ := setupTransactionHandler()
		router := setupGin()
		router.GET("/users/:userId/transactions", handler.List)

		req := httptest.NewRequest("GET", "/users/1/transactions?start_date=invalid-date", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid start date format. Use YYYY-MM-DD", response["error"])
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		handler, mockService := setupTransactionHandler()
		router := setupGin()
		router.GET("/users/:userId/transactions", handler.List)

		mockService.On("ListWithFilters", uint(1), (*string)(nil), (*uint)(nil), (*time.Time)(nil), (*time.Time)(nil), 100, 0).Return([]domain.Transaction{}, errors.New("database error"))

		req := httptest.NewRequest("GET", "/users/1/transactions", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Failed to retrieve transactions", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestTransactionHandler_ExportCSV(t *testing.T) {
	t.Run("should export CSV successfully", func(t *testing.T) {
		handler, mockService := setupTransactionHandler()
		router := setupGin()
		router.GET("/users/:userId/transactions/export/csv", handler.ExportCSV)

		expectedTransactions := []domain.Transaction{
			{
				ID:          1,
				UserID:      1,
				Amount:      100.50,
				Type:        "expense",
				Description: "Grocery shopping",
				CategoryID:  1,
				Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				CreatedAt:   time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			},
		}

		mockService.On("ListWithFilters", uint(1), (*string)(nil), (*uint)(nil), (*time.Time)(nil), (*time.Time)(nil), 100, 0).Return(expectedTransactions, nil)

		req := httptest.NewRequest("GET", "/users/1/transactions/export/csv", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment; filename=")
		assert.Contains(t, w.Body.String(), "ID,Amount,Type,Description,Category ID,Date,Created At")
		assert.Contains(t, w.Body.String(), "1,100.50,expense,Grocery shopping,1,2024-01-15")
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid user ID", func(t *testing.T) {
		handler, _ := setupTransactionHandler()
		router := setupGin()
		router.GET("/users/:userId/transactions/export/csv", handler.ExportCSV)

		req := httptest.NewRequest("GET", "/users/invalid/transactions/export/csv", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid user ID", response["error"])
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		handler, mockService := setupTransactionHandler()
		router := setupGin()
		router.GET("/users/:userId/transactions/export/csv", handler.ExportCSV)

		mockService.On("ListWithFilters", uint(1), (*string)(nil), (*uint)(nil), (*time.Time)(nil), (*time.Time)(nil), 100, 0).Return([]domain.Transaction{}, errors.New("database error"))

		req := httptest.NewRequest("GET", "/users/1/transactions/export/csv", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Failed to retrieve transactions", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestTransactionHandler_ExportPDF(t *testing.T) {
	t.Run("should export PDF successfully", func(t *testing.T) {
		handler, mockService := setupTransactionHandler()
		router := setupGin()
		router.GET("/users/:userId/transactions/export/pdf", handler.ExportPDF)

		expectedTransactions := []domain.Transaction{
			{
				ID:          1,
				UserID:      1,
				Amount:      100.50,
				Type:        "expense",
				Description: "Grocery shopping",
				CategoryID:  1,
				Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			},
		}

		mockService.On("ListWithFilters", uint(1), (*string)(nil), (*uint)(nil), (*time.Time)(nil), (*time.Time)(nil), 100, 0).Return(expectedTransactions, nil)

		req := httptest.NewRequest("GET", "/users/1/transactions/export/pdf", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment; filename=")
		assert.Contains(t, w.Body.String(), "Transaction Report")
		assert.Contains(t, w.Body.String(), "100.50")
		assert.Contains(t, w.Body.String(), "Grocery shopping")
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid user ID", func(t *testing.T) {
		handler, _ := setupTransactionHandler()
		router := setupGin()
		router.GET("/users/:userId/transactions/export/pdf", handler.ExportPDF)

		req := httptest.NewRequest("GET", "/users/invalid/transactions/export/pdf", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid user ID", response["error"])
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		handler, mockService := setupTransactionHandler()
		router := setupGin()
		router.GET("/users/:userId/transactions/export/pdf", handler.ExportPDF)

		mockService.On("ListWithFilters", uint(1), (*string)(nil), (*uint)(nil), (*time.Time)(nil), (*time.Time)(nil), 100, 0).Return([]domain.Transaction{}, errors.New("database error"))

		req := httptest.NewRequest("GET", "/users/1/transactions/export/pdf", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Failed to retrieve transactions", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestTransactionHandler_GetByID(t *testing.T) {
	t.Run("should get transaction by ID successfully", func(t *testing.T) {
		handler, mockService := setupTransactionHandler()
		router := setupGin()
		router.GET("/transactions/:id", handler.GetByID)

		expectedTransaction := &domain.Transaction{
			ID:          1,
			UserID:      1,
			Amount:      100.50,
			Type:        "expense",
			Description: "Grocery shopping",
			CategoryID:  1,
			Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		}

		mockService.On("GetByID", uint(1)).Return(expectedTransaction, nil)

		req := httptest.NewRequest("GET", "/transactions/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response domain.Transaction
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, expectedTransaction.ID, response.ID)
		assert.Equal(t, expectedTransaction.Amount, response.Amount)
		mockService.AssertExpectations(t)
	})

	t.Run("should return not found when transaction doesn't exist", func(t *testing.T) {
		handler, mockService := setupTransactionHandler()
		router := setupGin()
		router.GET("/transactions/:id", handler.GetByID)

		mockService.On("GetByID", uint(999)).Return(nil, errors.New("transaction not found"))

		req := httptest.NewRequest("GET", "/transactions/999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Transaction not found", response["error"])
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid transaction ID", func(t *testing.T) {
		handler, _ := setupTransactionHandler()
		router := setupGin()
		router.GET("/transactions/:id", handler.GetByID)

		req := httptest.NewRequest("GET", "/transactions/invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid transaction ID", response["error"])
	})
}

func TestTransactionHandler_Update(t *testing.T) {
	t.Run("should update transaction successfully", func(t *testing.T) {
		handler, mockService := setupTransactionHandler()
		router := setupGin()
		router.PUT("/transactions/:id", handler.Update)

		existingTransaction := &domain.Transaction{
			ID:          1,
			UserID:      1,
			Amount:      100.50,
			Type:        "expense",
			Description: "Old description",
			CategoryID:  1,
		}

		updateReq := CreateTransactionRequest{
			Amount:      150.75,
			Type:        "expense",
			Description: "Updated description",
			CategoryID:  2,
		}

		mockService.On("GetByID", uint(1)).Return(existingTransaction, nil)
		mockService.On("Update", mock.MatchedBy(func(t *domain.Transaction) bool {
			return t.Amount == 150.75 && t.Description == "Updated description"
		})).Return(nil)

		requestBody, _ := json.Marshal(updateReq)
		req := httptest.NewRequest("PUT", "/transactions/1", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return not found when transaction doesn't exist", func(t *testing.T) {
		handler, mockService := setupTransactionHandler()
		router := setupGin()
		router.PUT("/transactions/:id", handler.Update)

		updateReq := CreateTransactionRequest{
			Amount:      150.75,
			Type:        "expense",
			Description: "Updated description",
			CategoryID:  2,
		}

		mockService.On("GetByID", uint(999)).Return(nil, errors.New("transaction not found"))

		requestBody, _ := json.Marshal(updateReq)
		req := httptest.NewRequest("PUT", "/transactions/999", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Transaction not found", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestTransactionHandler_Delete(t *testing.T) {
	t.Run("should delete transaction successfully", func(t *testing.T) {
		handler, mockService := setupTransactionHandler()
		router := setupGin()
		router.DELETE("/transactions/:id", handler.Delete)

		mockService.On("Delete", uint(1)).Return(nil)

		req := httptest.NewRequest("DELETE", "/transactions/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Transaction deleted successfully", response["message"])
		mockService.AssertExpectations(t)
	})

	t.Run("should return not found when transaction doesn't exist", func(t *testing.T) {
		handler, mockService := setupTransactionHandler()
		router := setupGin()
		router.DELETE("/transactions/:id", handler.Delete)

		mockService.On("Delete", uint(999)).Return(errors.New("transaction not found"))

		req := httptest.NewRequest("DELETE", "/transactions/999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Transaction not found", response["error"])
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid transaction ID", func(t *testing.T) {
		handler, _ := setupTransactionHandler()
		router := setupGin()
		router.DELETE("/transactions/:id", handler.Delete)

		req := httptest.NewRequest("DELETE", "/transactions/invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid transaction ID", response["error"])
	})
}