package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-finance-advisor/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBudgetService is a mock implementation of BudgetService
type MockBudgetService struct {
	mock.Mock
}

func (m *MockBudgetService) CreateBudget(budget *domain.Budget) error {
	args := m.Called(budget)
	return args.Error(0)
}

func (m *MockBudgetService) GetBudgetsByUser(userID uint) ([]domain.Budget, error) {
	args := m.Called(userID)
	return args.Get(0).([]domain.Budget), args.Error(1)
}

func (m *MockBudgetService) GetBudgetByID(budgetID uint) (*domain.Budget, error) {
	args := m.Called(budgetID)
	return args.Get(0).(*domain.Budget), args.Error(1)
}

func (m *MockBudgetService) UpdateBudget(budgetID uint, budget *domain.Budget) error {
	args := m.Called(budgetID, budget)
	return args.Error(0)
}

func (m *MockBudgetService) DeleteBudget(budgetID uint) error {
	args := m.Called(budgetID)
	return args.Error(0)
}

func (m *MockBudgetService) GetBudgetSummary(userID uint) (*domain.BudgetSummary, error) {
	args := m.Called(userID)
	return args.Get(0).(*domain.BudgetSummary), args.Error(1)
}

func setupBudgetHandler() (*BudgetHandler, *MockBudgetService) {
	mockService := &MockBudgetService{}
	handler := &BudgetHandler{
		Service: mockService,
	}
	return handler, mockService
}

func TestBudgetHandler_CreateBudget(t *testing.T) {
	t.Run("should create budget successfully", func(t *testing.T) {
		handler, mockService := setupBudgetHandler()
		router := setupGin()
		router.POST("/users/:userId/budgets", handler.CreateBudget)

		reqBody := CreateBudgetRequest{
			CategoryID: 1,
			Amount:     500.00,
			Period:     "monthly",
			StartDate:  "2024-01-01",
		}

		mockService.On("CreateBudget", mock.AnythingOfType("*domain.Budget")).Return(nil)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/users/1/budgets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response domain.Budget
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, uint(1), response.UserID)
		assert.Equal(t, uint(1), response.CategoryID)
		assert.Equal(t, 500.00, response.Amount)
		assert.Equal(t, "monthly", response.Period)
		assert.True(t, response.IsActive)
		mockService.AssertExpectations(t)
	})

	t.Run("should create budget with calculated end date", func(t *testing.T) {
		handler, mockService := setupBudgetHandler()
		router := setupGin()
		router.POST("/users/:userId/budgets", handler.CreateBudget)

		reqBody := CreateBudgetRequest{
			CategoryID: 1,
			Amount:     1000.00,
			Period:     "weekly",
			StartDate:  "2024-01-01",
		}

		mockService.On("CreateBudget", mock.AnythingOfType("*domain.Budget")).Return(nil)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/users/1/budgets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response domain.Budget
		json.Unmarshal(w.Body.Bytes(), &response)
		expectedEndDate := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, expectedEndDate, response.EndDate)
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid user ID", func(t *testing.T) {
		handler, _ := setupBudgetHandler()
		router := setupGin()
		router.POST("/users/:userId/budgets", handler.CreateBudget)

		reqBody := CreateBudgetRequest{
			CategoryID: 1,
			Amount:     500.00,
			Period:     "monthly",
			StartDate:  "2024-01-01",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/users/invalid/budgets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid user ID", response["error"])
	})

	t.Run("should return bad request for invalid JSON", func(t *testing.T) {
		handler, _ := setupBudgetHandler()
		router := setupGin()
		router.POST("/users/:userId/budgets", handler.CreateBudget)

		req := httptest.NewRequest("POST", "/users/1/budgets", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return bad request for invalid start date", func(t *testing.T) {
		handler, _ := setupBudgetHandler()
		router := setupGin()
		router.POST("/users/:userId/budgets", handler.CreateBudget)

		reqBody := CreateBudgetRequest{
			CategoryID: 1,
			Amount:     500.00,
			Period:     "monthly",
			StartDate:  "invalid-date",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/users/1/budgets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid start date format. Use YYYY-MM-DD", response["error"])
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		handler, mockService := setupBudgetHandler()
		router := setupGin()
		router.POST("/users/:userId/budgets", handler.CreateBudget)

		reqBody := CreateBudgetRequest{
			CategoryID: 1,
			Amount:     500.00,
			Period:     "monthly",
			StartDate:  "2024-01-01",
		}

		mockService.On("CreateBudget", mock.AnythingOfType("*domain.Budget")).Return(errors.New("database error"))

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/users/1/budgets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Failed to create budget", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestBudgetHandler_GetBudgets(t *testing.T) {
	t.Run("should get budgets successfully", func(t *testing.T) {
		handler, mockService := setupBudgetHandler()
		router := setupGin()
		router.GET("/users/:userId/budgets", handler.GetBudgets)

		expectedBudgets := []domain.Budget{
			{
				ID:         1,
				UserID:     1,
				CategoryID: 1,
				Amount:     500.00,
				Period:     "monthly",
				IsActive:   true,
			},
			{
				ID:         2,
				UserID:     1,
				CategoryID: 2,
				Amount:     200.00,
				Period:     "weekly",
				IsActive:   true,
			},
		}

		mockService.On("GetBudgetsByUser", uint(1)).Return(expectedBudgets, nil)

		req := httptest.NewRequest("GET", "/users/1/budgets", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []domain.Budget
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response, 2)
		assert.Equal(t, expectedBudgets[0].ID, response[0].ID)
		assert.Equal(t, expectedBudgets[1].ID, response[1].ID)
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid user ID", func(t *testing.T) {
		handler, _ := setupBudgetHandler()
		router := setupGin()
		router.GET("/users/:userId/budgets", handler.GetBudgets)

		req := httptest.NewRequest("GET", "/users/invalid/budgets", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid user ID", response["error"])
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		handler, mockService := setupBudgetHandler()
		router := setupGin()
		router.GET("/users/:userId/budgets", handler.GetBudgets)

		mockService.On("GetBudgetsByUser", uint(1)).Return([]domain.Budget{}, errors.New("database error"))

		req := httptest.NewRequest("GET", "/users/1/budgets", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Failed to retrieve budgets", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestBudgetHandler_GetBudget(t *testing.T) {
	t.Run("should get budget successfully", func(t *testing.T) {
		handler, mockService := setupBudgetHandler()
		router := setupGin()
		router.GET("/users/:userId/budgets/:budgetId", handler.GetBudget)

		expectedBudget := &domain.Budget{
			ID:         1,
			UserID:     1,
			CategoryID: 1,
			Amount:     500.00,
			Period:     "monthly",
			IsActive:   true,
		}

		mockService.On("GetBudgetByID", uint(1)).Return(expectedBudget, nil)

		req := httptest.NewRequest("GET", "/users/1/budgets/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response domain.Budget
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, expectedBudget.ID, response.ID)
		assert.Equal(t, expectedBudget.UserID, response.UserID)
		mockService.AssertExpectations(t)
	})

	t.Run("should return forbidden for different user", func(t *testing.T) {
		handler, mockService := setupBudgetHandler()
		router := setupGin()
		router.GET("/users/:userId/budgets/:budgetId", handler.GetBudget)

		expectedBudget := &domain.Budget{
			ID:         1,
			UserID:     2, // Different user
			CategoryID: 1,
			Amount:     500.00,
			Period:     "monthly",
			IsActive:   true,
		}

		mockService.On("GetBudgetByID", uint(1)).Return(expectedBudget, nil)

		req := httptest.NewRequest("GET", "/users/1/budgets/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Access denied", response["error"])
		mockService.AssertExpectations(t)
	})

	t.Run("should return not found for non-existent budget", func(t *testing.T) {
		handler, mockService := setupBudgetHandler()
		router := setupGin()
		router.GET("/users/:userId/budgets/:budgetId", handler.GetBudget)

		mockService.On("GetBudgetByID", uint(999)).Return((*domain.Budget)(nil), errors.New("not found"))

		req := httptest.NewRequest("GET", "/users/1/budgets/999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Budget not found", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestBudgetHandler_UpdateBudget(t *testing.T) {
	t.Run("should update budget successfully", func(t *testing.T) {
		handler, mockService := setupBudgetHandler()
		router := setupGin()
		router.PUT("/users/:userId/budgets/:budgetId", handler.UpdateBudget)

		existingBudget := &domain.Budget{
			ID:         1,
			UserID:     1,
			CategoryID: 1,
			Amount:     500.00,
			Period:     "monthly",
			IsActive:   true,
		}

		updatedBudget := &domain.Budget{
			ID:         1,
			UserID:     1,
			CategoryID: 1,
			Amount:     600.00,
			Period:     "monthly",
			IsActive:   false,
		}

		newAmount := 600.00
		newActive := false
		reqBody := UpdateBudgetRequest{
			Amount:   &newAmount,
			IsActive: &newActive,
		}

		mockService.On("GetBudgetByID", uint(1)).Return(existingBudget, nil).Once()
		mockService.On("UpdateBudget", uint(1), mock.AnythingOfType("*domain.Budget")).Return(nil)
		mockService.On("GetBudgetByID", uint(1)).Return(updatedBudget, nil).Once()

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PUT", "/users/1/budgets/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response domain.Budget
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, 600.00, response.Amount)
		assert.False(t, response.IsActive)
		mockService.AssertExpectations(t)
	})

	t.Run("should return forbidden for different user", func(t *testing.T) {
		handler, mockService := setupBudgetHandler()
		router := setupGin()
		router.PUT("/users/:userId/budgets/:budgetId", handler.UpdateBudget)

		existingBudget := &domain.Budget{
			ID:         1,
			UserID:     2, // Different user
			CategoryID: 1,
			Amount:     500.00,
			Period:     "monthly",
			IsActive:   true,
		}

		newAmount := 600.00
		reqBody := UpdateBudgetRequest{
			Amount: &newAmount,
		}

		mockService.On("GetBudgetByID", uint(1)).Return(existingBudget, nil)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PUT", "/users/1/budgets/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Access denied", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestBudgetHandler_DeleteBudget(t *testing.T) {
	t.Run("should delete budget successfully", func(t *testing.T) {
		handler, mockService := setupBudgetHandler()
		router := setupGin()
		router.DELETE("/users/:userId/budgets/:budgetId", handler.DeleteBudget)

		existingBudget := &domain.Budget{
			ID:         1,
			UserID:     1,
			CategoryID: 1,
			Amount:     500.00,
			Period:     "monthly",
			IsActive:   true,
		}

		mockService.On("GetBudgetByID", uint(1)).Return(existingBudget, nil)
		mockService.On("DeleteBudget", uint(1)).Return(nil)

		req := httptest.NewRequest("DELETE", "/users/1/budgets/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Budget deleted successfully", response["message"])
		mockService.AssertExpectations(t)
	})

	t.Run("should return forbidden for different user", func(t *testing.T) {
		handler, mockService := setupBudgetHandler()
		router := setupGin()
		router.DELETE("/users/:userId/budgets/:budgetId", handler.DeleteBudget)

		existingBudget := &domain.Budget{
			ID:         1,
			UserID:     2, // Different user
			CategoryID: 1,
			Amount:     500.00,
			Period:     "monthly",
			IsActive:   true,
		}

		mockService.On("GetBudgetByID", uint(1)).Return(existingBudget, nil)

		req := httptest.NewRequest("DELETE", "/users/1/budgets/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Access denied", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestBudgetHandler_GetBudgetSummary(t *testing.T) {
	t.Run("should get budget summary successfully", func(t *testing.T) {
		handler, mockService := setupBudgetHandler()
		router := setupGin()
		router.GET("/users/:userId/budgets/summary", handler.GetBudgetSummary)

		expectedSummary := map[string]interface{}{
			"total_budgets":    3,
			"active_budgets":   2,
			"total_allocated": 1500.00,
			"total_spent":     800.00,
			"remaining":       700.00,
		}

		mockService.On("GetBudgetSummary", uint(1)).Return(expectedSummary, nil)

		req := httptest.NewRequest("GET", "/users/1/budgets/summary", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(3), response["total_budgets"])
		assert.Equal(t, float64(2), response["active_budgets"])
		assert.Equal(t, 1500.00, response["total_allocated"])
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid user ID", func(t *testing.T) {
		handler, _ := setupBudgetHandler()
		router := setupGin()
		router.GET("/users/:userId/budgets/summary", handler.GetBudgetSummary)

		req := httptest.NewRequest("GET", "/users/invalid/budgets/summary", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid user ID", response["error"])
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		handler, mockService := setupBudgetHandler()
		router := setupGin()
		router.GET("/users/:userId/budgets/summary", handler.GetBudgetSummary)

		mockService.On("GetBudgetSummary", uint(1)).Return(map[string]interface{}{}, errors.New("database error"))

		req := httptest.NewRequest("GET", "/users/1/budgets/summary", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Failed to generate budget summary", response["error"])
		mockService.AssertExpectations(t)
	})
}