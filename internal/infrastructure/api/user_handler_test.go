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

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Note: MockUserService is defined in advisor_handler_test.go to avoid duplication

func setupUserHandler() (*UserHandler, *MockUserService) {
	mockService := new(MockUserService)
	handler := &UserHandler{Service: mockService}
	return handler, mockService
}

func setupGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestNewUserHandler(t *testing.T) {
	t.Run("should create new user handler", func(t *testing.T) {
		mockService := &MockUserService{}
		handler := NewUserHandler(mockService)

		assert.NotNil(t, handler)
		assert.Equal(t, mockService, handler.Service)
	})
}

func TestUserHandler_Create(t *testing.T) {
	t.Run("should create user successfully", func(t *testing.T) {
		handler, mockService := setupUserHandler()
		router := setupGin()
		router.POST("/users", handler.Create)

		user := domain.User{
			Email:         "test@example.com",
			Password:      "password123",
			FirstName:     "John",
			LastName:      "Doe",
			RiskTolerance: "moderate",
		}

		mockService.On("Create", mock.AnythingOfType("*domain.User")).Return(nil)

		requestBody, _ := json.Marshal(user)
		req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should set default risk tolerance", func(t *testing.T) {
		handler, mockService := setupUserHandler()
		router := setupGin()
		router.POST("/users", handler.Create)

		user := domain.User{
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "John",
			LastName:  "Doe",
			// RiskTolerance not set
		}

		mockService.On("Create", mock.MatchedBy(func(u *domain.User) bool {
			return u.RiskTolerance == "moderate"
		})).Return(nil)

		requestBody, _ := json.Marshal(user)
		req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid JSON", func(t *testing.T) {
		handler, _ := setupUserHandler()
		router := setupGin()
		router.POST("/users", handler.Create)

		req := httptest.NewRequest("POST", "/users", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response["error"], "invalid character")
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		handler, mockService := setupUserHandler()
		router := setupGin()
		router.POST("/users", handler.Create)

		user := domain.User{
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "John",
			LastName:  "Doe",
		}

		mockService.On("Create", mock.AnythingOfType("*domain.User")).Return(errors.New("database error"))

		requestBody, _ := json.Marshal(user)
		req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "db error", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_Get(t *testing.T) {
	t.Run("should get user successfully", func(t *testing.T) {
		handler, mockService := setupUserHandler()
		router := setupGin()
		router.GET("/users/:userId", handler.Get)

		expectedUser := domain.User{
			ID:        1,
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.On("GetByID", uint(1)).Return(expectedUser, nil)

		req := httptest.NewRequest("GET", "/users/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response domain.User
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, expectedUser.ID, response.ID)
		assert.Equal(t, expectedUser.Email, response.Email)
		mockService.AssertExpectations(t)
	})

	t.Run("should return not found when user doesn't exist", func(t *testing.T) {
		handler, mockService := setupUserHandler()
		router := setupGin()
		router.GET("/users/:userId", handler.Get)

		mockService.On("GetByID", uint(999)).Return(domain.User{}, errors.New("user not found"))

		req := httptest.NewRequest("GET", "/users/999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "user not found", response["error"])
		mockService.AssertExpectations(t)
	})

	t.Run("should handle invalid user ID", func(t *testing.T) {
		handler, mockService := setupUserHandler()
		router := setupGin()
		router.GET("/users/:userId", handler.Get)

		// Note: Gin will convert "invalid" to 0, so this tests the zero ID case
		mockService.On("GetByID", uint(0)).Return(domain.User{}, errors.New("user not found"))

		req := httptest.NewRequest("GET", "/users/invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// The handler will still call GetByID with 0, which should return not found
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_Register(t *testing.T) {
	t.Run("should register user successfully", func(t *testing.T) {
		handler, mockService := setupUserHandler()
		router := setupGin()
		router.POST("/register", handler.Register)

		registerReq := RegisterRequest{
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "John",
			LastName:  "Doe",
		}

		expectedUser := domain.User{
			ID:        1,
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
		}

		mockService.On("Register", "test@example.com", "password123", "John", "Doe").Return(&expectedUser, nil)

		requestBody, _ := json.Marshal(registerReq)
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "user")
		assert.Contains(t, response, "token")
		mockService.AssertExpectations(t)
	})

	t.Run("should return conflict when user already exists", func(t *testing.T) {
		handler, mockService := setupUserHandler()
		router := setupGin()
		router.POST("/register", handler.Register)

		registerReq := RegisterRequest{
			Email:     "existing@example.com",
			Password:  "password123",
			FirstName: "John",
			LastName:  "Doe",
		}

		mockService.On("Register", "existing@example.com", "password123", "John", "Doe").Return((*domain.User)(nil), errors.New("user already exists"))

		requestBody, _ := json.Marshal(registerReq)
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "User already exists", response["error"])
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid email", func(t *testing.T) {
		handler, _ := setupUserHandler()
		router := setupGin()
		router.POST("/register", handler.Register)

		registerReq := RegisterRequest{
			Email:     "invalid-email",
			Password:  "password123",
			FirstName: "John",
			LastName:  "Doe",
		}

		requestBody, _ := json.Marshal(registerReq)
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response["error"], "email")
	})

	t.Run("should return bad request for short password", func(t *testing.T) {
		handler, _ := setupUserHandler()
		router := setupGin()
		router.POST("/register", handler.Register)

		registerReq := RegisterRequest{
			Email:     "test@example.com",
			Password:  "short",
			FirstName: "John",
			LastName:  "Doe",
		}

		requestBody, _ := json.Marshal(registerReq)
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response["error"], "min")
	})
}

func TestUserHandler_Login(t *testing.T) {
	t.Run("should login user successfully", func(t *testing.T) {
		handler, mockService := setupUserHandler()
		router := setupGin()
		router.POST("/login", handler.Login)

		loginReq := LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		expectedUser := domain.User{
			ID:        1,
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
		}

		mockService.On("Login", "test@example.com", "password123").Return(&expectedUser, nil)

		requestBody, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "user")
		assert.Contains(t, response, "token")
		mockService.AssertExpectations(t)
	})

	t.Run("should return unauthorized for invalid credentials", func(t *testing.T) {
		handler, mockService := setupUserHandler()
		router := setupGin()
		router.POST("/login", handler.Login)

		loginReq := LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		mockService.On("Login", "test@example.com", "wrongpassword").Return((*domain.User)(nil), errors.New("invalid credentials"))

		requestBody, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid credentials", response["error"])
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid JSON", func(t *testing.T) {
		handler, _ := setupUserHandler()
		router := setupGin()
		router.POST("/login", handler.Login)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUserHandler_UpdateRisk(t *testing.T) {
	t.Run("should update risk tolerance successfully", func(t *testing.T) {
		handler, mockService := setupUserHandler()
		router := setupGin()
		router.PUT("/users/:userId/risk", handler.UpdateRisk)

		existingUser := domain.User{
			ID:            1,
			Email:         "test@example.com",
			RiskTolerance: "moderate",
		}

		updatedUser := existingUser
		updatedUser.RiskTolerance = "aggressive"

		riskUpdateReq := RiskUpdateRequest{
			RiskTolerance: "aggressive",
		}

		mockService.On("GetByID", uint(1)).Return(existingUser, nil)
		mockService.On("Update", mock.MatchedBy(func(u *domain.User) bool {
			return u.RiskTolerance == "aggressive"
		})).Return(nil)

		requestBody, _ := json.Marshal(riskUpdateReq)
		req := httptest.NewRequest("PUT", "/users/1/risk", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid risk tolerance", func(t *testing.T) {
		handler, _ := setupUserHandler()
		router := setupGin()
		router.PUT("/users/:userId/risk", handler.UpdateRisk)

		riskUpdateReq := RiskUpdateRequest{
			RiskTolerance: "invalid",
		}

		requestBody, _ := json.Marshal(riskUpdateReq)
		req := httptest.NewRequest("PUT", "/users/1/risk", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "invalid risk tolerance", response["error"])
	})

	t.Run("should return not found when user doesn't exist", func(t *testing.T) {
		handler, mockService := setupUserHandler()
		router := setupGin()
		router.PUT("/users/:userId/risk", handler.UpdateRisk)

		riskUpdateReq := RiskUpdateRequest{
			RiskTolerance: "conservative",
		}

		mockService.On("GetByID", uint(999)).Return(domain.User{}, errors.New("user not found"))

		requestBody, _ := json.Marshal(riskUpdateReq)
		req := httptest.NewRequest("PUT", "/users/999/risk", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "user not found", response["error"])
		mockService.AssertExpectations(t)
	})

	t.Run("should return internal server error when update fails", func(t *testing.T) {
		handler, mockService := setupUserHandler()
		router := setupGin()
		router.PUT("/users/:userId/risk", handler.UpdateRisk)

		existingUser := domain.User{
			ID:            1,
			Email:         "test@example.com",
			RiskTolerance: "moderate",
		}

		riskUpdateReq := RiskUpdateRequest{
			RiskTolerance: "conservative",
		}

		mockService.On("GetByID", uint(1)).Return(existingUser, nil)
		mockService.On("Update", mock.AnythingOfType("*domain.User")).Return(errors.New("database error"))

		requestBody, _ := json.Marshal(riskUpdateReq)
		req := httptest.NewRequest("PUT", "/users/1/risk", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "update failed", response["error"])
		mockService.AssertExpectations(t)
	})
}
