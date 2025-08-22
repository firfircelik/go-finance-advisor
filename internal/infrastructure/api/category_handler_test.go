package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-finance-advisor/internal/application"
	"go-finance-advisor/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCategoryService is a mock implementation of CategoryService
type MockCategoryService struct {
	mock.Mock
}

func (m *MockCategoryService) InitializeDefaultCategories() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCategoryService) GetAllCategories() ([]domain.Category, error) {
	args := m.Called()
	return args.Get(0).([]domain.Category), args.Error(1)
}

func (m *MockCategoryService) GetCategoryByID(categoryID uint) (*domain.Category, error) {
	args := m.Called(categoryID)
	return args.Get(0).(*domain.Category), args.Error(1)
}

func (m *MockCategoryService) CreateCategory(category *domain.Category) error {
	args := m.Called(category)
	return args.Error(0)
}

func (m *MockCategoryService) UpdateCategory(categoryID uint, category *domain.Category) error {
	args := m.Called(categoryID, category)
	return args.Error(0)
}

func (m *MockCategoryService) DeleteCategory(categoryID uint) error {
	args := m.Called(categoryID)
	return args.Error(0)
}

func (m *MockCategoryService) GetCategoryUsageStats(userID uint) ([]application.CategoryUsageStats, error) {
	args := m.Called(userID)
	return args.Get(0).([]application.CategoryUsageStats), args.Error(1)
}

func (m *MockCategoryService) GetCategoriesByType(categoryType string) ([]domain.Category, error) {
	args := m.Called(categoryType)
	return args.Get(0).([]domain.Category), args.Error(1)
}

func setupCategoryHandler() (*CategoryHandler, *MockCategoryService) {
	mockService := &MockCategoryService{}
	handler := &CategoryHandler{
		Service: mockService,
	}
	return handler, mockService
}

func TestCategoryHandler_InitializeDefaultCategories(t *testing.T) {
	t.Run("should initialize default categories successfully", func(t *testing.T) {
		handler, mockService := setupCategoryHandler()
		router := setupGin()
		router.POST("/categories/initialize", handler.InitializeDefaultCategories)

		mockService.On("InitializeDefaultCategories").Return(nil)

		req := httptest.NewRequest("POST", "/categories/initialize", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Default categories initialized successfully", response["message"])
		mockService.AssertExpectations(t)
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		handler, mockService := setupCategoryHandler()
		router := setupGin()
		router.POST("/categories/initialize", handler.InitializeDefaultCategories)

		mockService.On("InitializeDefaultCategories").Return(errors.New("database error"))

		req := httptest.NewRequest("POST", "/categories/initialize", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Failed to initialize default categories", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestCategoryHandler_GetCategories(t *testing.T) {
	t.Run("should get categories successfully", func(t *testing.T) {
		handler, mockService := setupCategoryHandler()
		router := setupGin()
		router.GET("/categories", handler.GetCategories)

		expectedCategories := []domain.Category{
			{
				ID:          1,
				Name:        "Food",
				Type:        "expense",
				Description: "Food and dining",
				Icon:        "🍽️",
				Color:       "#FF6B6B",
				IsDefault:   true,
			},
			{
				ID:          2,
				Name:        "Salary",
				Type:        "income",
				Description: "Monthly salary",
				Icon:        "💰",
				Color:       "#4ECDC4",
				IsDefault:   true,
			},
		}

		mockService.On("GetAllCategories").Return(expectedCategories, nil)

		req := httptest.NewRequest("GET", "/categories", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []domain.Category
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response, 2)
		assert.Equal(t, "Food", response[0].Name)
		assert.Equal(t, "Salary", response[1].Name)
		mockService.AssertExpectations(t)
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		handler, mockService := setupCategoryHandler()
		router := setupGin()
		router.GET("/categories", handler.GetCategories)

		mockService.On("GetAllCategories").Return([]domain.Category{}, errors.New("database error"))

		req := httptest.NewRequest("GET", "/categories", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Failed to retrieve categories", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestCategoryHandler_GetCategory(t *testing.T) {
	t.Run("should get category successfully", func(t *testing.T) {
		handler, mockService := setupCategoryHandler()
		router := setupGin()
		router.GET("/categories/:categoryId", handler.GetCategory)

		expectedCategory := &domain.Category{
			ID:          1,
			Name:        "Food",
			Type:        "expense",
			Description: "Food and dining",
			Icon:        "🍽️",
			Color:       "#FF6B6B",
			IsDefault:   true,
		}

		mockService.On("GetCategoryByID", uint(1)).Return(expectedCategory, nil)

		req := httptest.NewRequest("GET", "/categories/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response domain.Category
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Food", response.Name)
		assert.Equal(t, "expense", response.Type)
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid category ID", func(t *testing.T) {
		handler, _ := setupCategoryHandler()
		router := setupGin()
		router.GET("/categories/:categoryId", handler.GetCategory)

		req := httptest.NewRequest("GET", "/categories/invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "invalid category ID", response["error"])
	})

	t.Run("should return not found for non-existent category", func(t *testing.T) {
		handler, mockService := setupCategoryHandler()
		router := setupGin()
		router.GET("/categories/:categoryId", handler.GetCategory)

		mockService.On("GetCategoryByID", uint(999)).Return((*domain.Category)(nil), errors.New("not found"))

		req := httptest.NewRequest("GET", "/categories/999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Category not found", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestCategoryHandler_CreateCategory(t *testing.T) {
	t.Run("should create category successfully", func(t *testing.T) {
		handler, mockService := setupCategoryHandler()
		router := setupGin()
		router.POST("/categories", handler.CreateCategory)

		reqBody := CreateCategoryRequest{
			Name:        "Custom Food",
			Type:        "expense",
			Description: "Custom food category",
			Icon:        "🍕",
			Color:       "#FF5722",
		}

		mockService.On("CreateCategory", mock.AnythingOfType("*domain.Category")).Return(nil)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response domain.Category
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Custom Food", response.Name)
		assert.Equal(t, "expense", response.Type)
		assert.False(t, response.IsDefault)
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid JSON", func(t *testing.T) {
		handler, _ := setupCategoryHandler()
		router := setupGin()
		router.POST("/categories", handler.CreateCategory)

		req := httptest.NewRequest("POST", "/categories", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return bad request for invalid type", func(t *testing.T) {
		handler, _ := setupCategoryHandler()
		router := setupGin()
		router.POST("/categories", handler.CreateCategory)

		reqBody := CreateCategoryRequest{
			Name: "Test Category",
			Type: "invalid", // Invalid type
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		handler, mockService := setupCategoryHandler()
		router := setupGin()
		router.POST("/categories", handler.CreateCategory)

		reqBody := CreateCategoryRequest{
			Name: "Test Category",
			Type: "expense",
		}

		mockService.On("CreateCategory", mock.AnythingOfType("*domain.Category")).Return(errors.New("database error"))

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Failed to create category", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestCategoryHandler_UpdateCategory(t *testing.T) {
	t.Run("should update category successfully", func(t *testing.T) {
		handler, mockService := setupCategoryHandler()
		router := setupGin()
		router.PUT("/categories/:categoryId", handler.UpdateCategory)

		existingCategory := &domain.Category{
			ID:          1,
			Name:        "Food",
			Type:        "expense",
			Description: "Food and dining",
			Icon:        "🍽️",
			Color:       "#FF6B6B",
			IsDefault:   false, // Custom category
		}

		updatedCategory := &domain.Category{
			ID:          1,
			Name:        "Updated Food",
			Type:        "expense",
			Description: "Updated food category",
			Icon:        "🍕",
			Color:       "#FF5722",
			IsDefault:   false,
		}

		newName := "Updated Food"
		newDescription := "Updated food category"
		newIcon := "🍕"
		newColor := "#FF5722"
		reqBody := UpdateCategoryRequest{
			Name:        &newName,
			Description: &newDescription,
			Icon:        &newIcon,
			Color:       &newColor,
		}

		mockService.On("GetCategoryByID", uint(1)).Return(existingCategory, nil).Once()
		mockService.On("UpdateCategory", uint(1), mock.AnythingOfType("*domain.Category")).Return(nil)
		mockService.On("GetCategoryByID", uint(1)).Return(updatedCategory, nil).Once()

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PUT", "/categories/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response domain.Category
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Updated Food", response.Name)
		assert.Equal(t, "Updated food category", response.Description)
		mockService.AssertExpectations(t)
	})

	t.Run("should return forbidden for default category", func(t *testing.T) {
		handler, mockService := setupCategoryHandler()
		router := setupGin()
		router.PUT("/categories/:categoryId", handler.UpdateCategory)

		defaultCategory := &domain.Category{
			ID:        1,
			Name:      "Food",
			Type:      "expense",
			IsDefault: true, // Default category
		}

		newName := "Updated Food"
		reqBody := UpdateCategoryRequest{
			Name: &newName,
		}

		mockService.On("GetCategoryByID", uint(1)).Return(defaultCategory, nil)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PUT", "/categories/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Default categories cannot be modified", response["error"])
		mockService.AssertExpectations(t)
	})

	t.Run("should return not found for non-existent category", func(t *testing.T) {
		handler, mockService := setupCategoryHandler()
		router := setupGin()
		router.PUT("/categories/:categoryId", handler.UpdateCategory)

		newName := "Updated Food"
		reqBody := UpdateCategoryRequest{
			Name: &newName,
		}

		mockService.On("GetCategoryByID", uint(999)).Return((*domain.Category)(nil), errors.New("not found"))

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PUT", "/categories/999", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Category not found", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestCategoryHandler_DeleteCategory(t *testing.T) {
	t.Run("should delete category successfully", func(t *testing.T) {
		handler, mockService := setupCategoryHandler()
		router := setupGin()
		router.DELETE("/categories/:categoryId", handler.DeleteCategory)

		customCategory := &domain.Category{
			ID:        1,
			Name:      "Custom Food",
			Type:      "expense",
			IsDefault: false, // Custom category
		}

		mockService.On("GetCategoryByID", uint(1)).Return(customCategory, nil)
		mockService.On("DeleteCategory", uint(1)).Return(nil)

		req := httptest.NewRequest("DELETE", "/categories/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Category deleted successfully", response["message"])
		mockService.AssertExpectations(t)
	})

	t.Run("should return forbidden for default category", func(t *testing.T) {
		handler, mockService := setupCategoryHandler()
		router := setupGin()
		router.DELETE("/categories/:categoryId", handler.DeleteCategory)

		defaultCategory := &domain.Category{
			ID:        1,
			Name:      "Food",
			Type:      "expense",
			IsDefault: true, // Default category
		}

		mockService.On("GetCategoryByID", uint(1)).Return(defaultCategory, nil)

		req := httptest.NewRequest("DELETE", "/categories/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Default categories cannot be deleted", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestCategoryHandler_GetCategoryUsage(t *testing.T) {
	t.Run("should get category usage successfully", func(t *testing.T) {
		handler, mockService := setupCategoryHandler()
		router := setupGin()
		router.GET("/categories/usage", handler.GetCategoryUsage)

		expectedStats := []application.CategoryUsageStats{
			{
				CategoryID:       1,
				CategoryName:     "Food",
				CategoryType:     "expense",
				TransactionCount: 5,
				TotalAmount:      250.0,
				AverageAmount:    50.0,
			},
			{
				CategoryID:       2,
				CategoryName:     "Salary",
				CategoryType:     "income",
				TransactionCount: 2,
				TotalAmount:      4000.0,
				AverageAmount:    2000.0,
			},
		}

		mockService.On("GetCategoryUsageStats", uint(0)).Return(expectedStats, nil)

		req := httptest.NewRequest("GET", "/categories/usage", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []application.CategoryUsageStats
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response, 2)
		assert.Equal(t, "Food", response[0].CategoryName)
		assert.Equal(t, "expense", response[0].CategoryType)
		assert.Equal(t, 5, response[0].TransactionCount)
		mockService.AssertExpectations(t)
	})

	t.Run("should get category usage for specific user", func(t *testing.T) {
		handler, mockService := setupCategoryHandler()
		router := setupGin()
		router.GET("/categories/usage", handler.GetCategoryUsage)

		expectedStats := []application.CategoryUsageStats{
			{
				CategoryID:       1,
				CategoryName:     "Transport",
				CategoryType:     "expense",
				TransactionCount: 3,
				TotalAmount:      150.0,
				AverageAmount:    50.0,
			},
		}

		mockService.On("GetCategoryUsageStats", uint(1)).Return(expectedStats, nil)

		req := httptest.NewRequest("GET", "/categories/usage?user_id=1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []application.CategoryUsageStats
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response, 1)
		assert.Equal(t, "Transport", response[0].CategoryName)
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid user ID", func(t *testing.T) {
		handler, _ := setupCategoryHandler()
		router := setupGin()
		router.GET("/categories/usage", handler.GetCategoryUsage)

		req := httptest.NewRequest("GET", "/categories/usage?user_id=invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "invalid user ID", response["error"])
	})
}

func TestCategoryHandler_GetIncomeCategories(t *testing.T) {
	t.Run("should get income categories successfully", func(t *testing.T) {
		handler, mockService := setupCategoryHandler()
		router := setupGin()
		router.GET("/categories/income", handler.GetIncomeCategories)

		expectedCategories := []domain.Category{
			{
				ID:          1,
				Name:        "Salary",
				Type:        "income",
				Description: "Monthly salary",
				Icon:        "💰",
				Color:       "#4ECDC4",
				IsDefault:   true,
			},
			{
				ID:          2,
				Name:        "Freelance",
				Type:        "income",
				Description: "Freelance work",
				Icon:        "💻",
				Color:       "#45B7D1",
				IsDefault:   true,
			},
		}

		mockService.On("GetCategoriesByType", "income").Return(expectedCategories, nil)

		req := httptest.NewRequest("GET", "/categories/income", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []domain.Category
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response, 2)
		assert.Equal(t, "income", response[0].Type)
		assert.Equal(t, "income", response[1].Type)
		mockService.AssertExpectations(t)
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		handler, mockService := setupCategoryHandler()
		router := setupGin()
		router.GET("/categories/income", handler.GetIncomeCategories)

		mockService.On("GetCategoriesByType", "income").Return([]domain.Category{}, errors.New("database error"))

		req := httptest.NewRequest("GET", "/categories/income", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Failed to retrieve income categories", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestCategoryHandler_GetExpenseCategories(t *testing.T) {
	t.Run("should get expense categories successfully", func(t *testing.T) {
		handler, mockService := setupCategoryHandler()
		router := setupGin()
		router.GET("/categories/expense", handler.GetExpenseCategories)

		expectedCategories := []domain.Category{
			{
				ID:          1,
				Name:        "Food",
				Type:        "expense",
				Description: "Food and dining",
				Icon:        "🍽️",
				Color:       "#FF6B6B",
				IsDefault:   true,
			},
			{
				ID:          2,
				Name:        "Transportation",
				Type:        "expense",
				Description: "Transport costs",
				Icon:        "🚗",
				Color:       "#FFA726",
				IsDefault:   true,
			},
		}

		mockService.On("GetCategoriesByType", "expense").Return(expectedCategories, nil)

		req := httptest.NewRequest("GET", "/categories/expense", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []domain.Category
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response, 2)
		assert.Equal(t, "expense", response[0].Type)
		assert.Equal(t, "expense", response[1].Type)
		mockService.AssertExpectations(t)
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		handler, mockService := setupCategoryHandler()
		router := setupGin()
		router.GET("/categories/expense", handler.GetExpenseCategories)

		mockService.On("GetCategoriesByType", "expense").Return([]domain.Category{}, errors.New("database error"))

		req := httptest.NewRequest("GET", "/categories/expense", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Failed to retrieve expense categories", response["error"])
		mockService.AssertExpectations(t)
	})
}
