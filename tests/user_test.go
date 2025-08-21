package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-finance-advisor/internal/application"
	"go-finance-advisor/internal/domain"
	"go-finance-advisor/internal/infrastructure/api"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic("Failed to connect to test database: " + err.Error())
	}
	if err := db.AutoMigrate(&domain.User{}, &domain.Transaction{}); err != nil {
		panic("Failed to migrate test database: " + err.Error())
	}
	return db
}

func TestCreateUser(t *testing.T) {
	db := setupTestDB()
	userService := &application.UserService{DB: db}
	userHandler := api.NewUserHandler(userService)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/users", userHandler.Create)

	user := domain.User{
		Email:         "test@example.com",
		FirstName:     "Test",
		LastName:      "User",
		RiskTolerance: "moderate",
	}
	jsonData, _ := json.Marshal(user)

	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	var response domain.User
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "moderate", response.RiskTolerance)
	assert.NotZero(t, response.ID)
}

func TestGetUser(t *testing.T) {
	db := setupTestDB()
	userService := &application.UserService{DB: db}
	userHandler := api.NewUserHandler(userService)

	// Create a user first
	user := domain.User{
		Email:         "test2@example.com",
		FirstName:     "Test",
		LastName:      "User2",
		RiskTolerance: "aggressive",
	}
	userService.Create(&user)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/users/:userId", userHandler.Get)

	req := httptest.NewRequest("GET", "/users/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.User
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "aggressive", response.RiskTolerance)
	assert.Equal(t, uint(1), response.ID)
}

func TestUpdateUserRisk(t *testing.T) {
	db := setupTestDB()
	userService := &application.UserService{DB: db}
	userHandler := &api.UserHandler{Service: userService}

	// Create a user first
	user := domain.User{
		Email:         "test3@example.com",
		FirstName:     "Test",
		LastName:      "User3",
		RiskTolerance: "moderate",
	}
	userService.Create(&user)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PUT("/users/:userId/risk", userHandler.UpdateRisk)

	update := map[string]string{"risk_tolerance": "conservative"}
	jsonData, _ := json.Marshal(update)

	req := httptest.NewRequest("PUT", "/users/1/risk", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.User
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "conservative", response.RiskTolerance)
}
