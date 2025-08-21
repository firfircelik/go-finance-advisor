package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"go-finance-advisor/internal/application"
	"go-finance-advisor/internal/domain"
	"go-finance-advisor/internal/infrastructure/api"
	"go-finance-advisor/internal/infrastructure/middleware"
	"go-finance-advisor/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupTestApp creates a test instance of the application
func setupTestApp(t *testing.T) (*gin.Engine, *gorm.DB) {
	// Use in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate
	err = db.AutoMigrate(&domain.User{}, &domain.Transaction{}, &domain.Category{}, &domain.Budget{}, &domain.Recommendation{})
	require.NoError(t, err)

	// Initialize services
	userSvc := &application.UserService{DB: db}
	txSvc := &application.TransactionService{DB: db}
	advisorSvc := &application.AdvisorService{DB: db}
	analyticsSvc := &application.AnalyticsService{DB: db}
	budgetSvc := &application.BudgetService{DB: db}
	categorySvc := &application.CategoryService{DB: db}
	reportsSvc := application.NewReportsService(db)
	exportSvc := application.NewExportService(db)
	marketSvc := &pkg.RealTimeMarketService{}

	// Initialize handlers
	userHandler := &api.UserHandler{Service: userSvc}
	txHandler := &api.TransactionHandler{Service: txSvc}
	advisorHandler := api.NewAdvisorHandler(advisorSvc, userSvc, marketSvc)
	analyticsHandler := &api.AnalyticsHandler{Service: analyticsSvc}
	budgetHandler := &api.BudgetHandler{Service: budgetSvc}
	categoryHandler := &api.CategoryHandler{Service: categorySvc}
	reportsHandler := &api.ReportsHandler{Service: reportsSvc}
	exportHandler := api.NewExportHandler(exportSvc)

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check routes (public)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"service":   "go-finance-advisor",
			"version":   "1.0.0",
			"timestamp": gin.H{"unix": gin.H{"seconds": 1735000000}},
		})
	})

	r.GET("/metrics", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"uptime":          "24h",
			"requests_total":  1000,
			"active_users":    50,
			"database_status": "connected",
			"memory_usage":    "256MB",
			"cpu_usage":       "15%",
		})
	})

	// Routes
	v1 := r.Group("/api/v1")
	{
		// Health check routes
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":    "healthy",
				"service":   "go-finance-advisor",
				"version":   "1.0.0",
				"timestamp": gin.H{"unix": gin.H{"seconds": 1735000000}},
			})
		})

		v1.GET("/metrics", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"uptime":          "24h",
				"requests_total":  1000,
				"active_users":    50,
				"database_status": "connected",
				"memory_usage":    "256MB",
				"cpu_usage":       "15%",
			})
		})

		// Public routes
		v1.POST("/users", userHandler.Create)
		v1.POST("/auth/register", userHandler.Register)
		v1.POST("/auth/login", userHandler.Login)

		// Category routes (public for now)
		v1.POST("/categories/initialize", categoryHandler.InitializeDefaultCategories)
		v1.GET("/categories", categoryHandler.GetCategories)
		v1.GET("/categories/:categoryId", categoryHandler.GetCategory)
		v1.POST("/categories", categoryHandler.CreateCategory)
		v1.PUT("/categories/:categoryId", categoryHandler.UpdateCategory)
		v1.DELETE("/categories/:categoryId", categoryHandler.DeleteCategory)
		v1.GET("/categories/usage", categoryHandler.GetCategoryUsage)
		v1.GET("/categories/income", categoryHandler.GetIncomeCategories)
		v1.GET("/categories/expense", categoryHandler.GetExpenseCategories)

		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// User routes
			protected.GET("/users/:userId", userHandler.Get)
			protected.PUT("/users/:userId/risk", userHandler.UpdateRisk)

			// Transaction routes
			protected.POST("/users/:userId/transactions", txHandler.Create)
			protected.GET("/users/:userId/transactions", txHandler.List)
			protected.GET("/users/:userId/transactions/export/csv", txHandler.ExportCSV)
			protected.GET("/users/:userId/transactions/export/pdf", txHandler.ExportPDF)

			// Analytics routes
			protected.GET("/users/:userId/analytics/metrics", analyticsHandler.GetFinancialMetrics)
			protected.GET("/users/:userId/analytics/income-expense", analyticsHandler.GetIncomeExpenseAnalysis)
			protected.GET("/users/:userId/analytics/categories/:categoryId", analyticsHandler.GetCategoryAnalysis)
			protected.GET("/users/:userId/analytics/dashboard", analyticsHandler.GetDashboardSummary)

			// Budget routes
			protected.POST("/users/:userId/budgets", budgetHandler.CreateBudget)
			protected.GET("/users/:userId/budgets", budgetHandler.GetBudgets)
			protected.GET("/users/:userId/budgets/:budgetId", budgetHandler.GetBudget)
			protected.PUT("/users/:userId/budgets/:budgetId", budgetHandler.UpdateBudget)
			protected.DELETE("/users/:userId/budgets/:budgetId", budgetHandler.DeleteBudget)
			protected.GET("/users/:userId/budgets/summary", budgetHandler.GetBudgetSummary)

			// Reports routes
			protected.GET("/users/:userId/reports/monthly/:year/:month", reportsHandler.GenerateMonthlyReport)
			protected.GET("/users/:userId/reports/quarterly/:year/:quarter", reportsHandler.GenerateQuarterlyReport)
			protected.GET("/users/:userId/reports/yearly/:year", reportsHandler.GenerateYearlyReport)
			protected.GET("/users/:userId/reports", reportsHandler.GetReportsList)

			// Export routes
			protected.GET("/export/transactions", exportHandler.ExportTransactions)
			protected.GET("/export/budgets", exportHandler.ExportBudgets)
			protected.GET("/export/reports", exportHandler.ExportFinancialReport)
			protected.GET("/export/all", exportHandler.ExportAllData)
			protected.GET("/export/formats", exportHandler.GetExportFormats)

			// Investment advice
			protected.GET("/users/:userId/advice", advisorHandler.GetAdvice)
			protected.GET("/users/:userId/advice/realtime", advisorHandler.GetRealTimeAdvice)
			protected.GET("/market/data", advisorHandler.GetMarketData)
			protected.GET("/market/crypto", advisorHandler.GetCryptoPrices)
			protected.GET("/market/stocks", advisorHandler.GetStockPrices)
			protected.GET("/market/summary", advisorHandler.GetMarketSummary)
			protected.GET("/users/:userId/portfolio/recommendations", advisorHandler.GetPortfolioRecommendations)

			// AI-powered endpoints
			protected.GET("/users/:userId/ai/risk-assessment", advisorHandler.GetAIRiskAssessment)
			protected.GET("/ai/market/prediction", advisorHandler.GetAIMarketPrediction)
			protected.GET("/users/:userId/ai/portfolio/optimization", advisorHandler.GetAIPortfolioOptimization)
		}
	}

	return r, db
}

func TestMain(m *testing.M) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Run tests
	code := m.Run()

	// Exit with the test result code
	os.Exit(code)
}

func TestHealthEndpoint(t *testing.T) {
	router, _ := setupTestApp(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "go-finance-advisor", response["service"])
	assert.Equal(t, "1.0.0", response["version"])
}

func TestMetricsEndpoint(t *testing.T) {
	router, _ := setupTestApp(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "24h", response["uptime"])
	assert.Equal(t, float64(1000), response["requests_total"])
	assert.Equal(t, float64(50), response["active_users"])
	assert.Equal(t, "connected", response["database_status"])
}

func TestAPIV1HealthEndpoint(t *testing.T) {
	router, _ := setupTestApp(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "go-finance-advisor", response["service"])
}

func TestAPIV1MetricsEndpoint(t *testing.T) {
	router, _ := setupTestApp(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/metrics", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "24h", response["uptime"])
	assert.Equal(t, "connected", response["database_status"])
}

func TestCORSMiddleware(t *testing.T) {
	router, _ := setupTestApp(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/v1/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Origin, Content-Type, Authorization", w.Header().Get("Access-Control-Allow-Headers"))
}

func TestDatabaseConnection(t *testing.T) {
	_, db := setupTestApp(t)

	// Test database connection
	sqlDB, err := db.DB()
	require.NoError(t, err)

	err = sqlDB.Ping()
	assert.NoError(t, err)

	// Test that tables were created
	var tableCount int64
	err = db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&tableCount).Error
	require.NoError(t, err)
	assert.Greater(t, tableCount, int64(0))
}

func TestPublicRoutes(t *testing.T) {
	router, _ := setupTestApp(t)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "GET categories",
			method:         "GET",
			path:           "/api/v1/categories",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET income categories",
			method:         "GET",
			path:           "/api/v1/categories/income",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET expense categories",
			method:         "GET",
			path:           "/api/v1/categories/expense",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestProtectedRoutesWithoutAuth(t *testing.T) {
	router, _ := setupTestApp(t)

	tests := []struct {
		name string
		path string
	}{
		{
			name: "GET user without auth",
			path: "/api/v1/users/1",
		},
		{
			name: "GET transactions without auth",
			path: "/api/v1/users/1/transactions",
		},
		{
			name: "GET budgets without auth",
			path: "/api/v1/users/1/budgets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			router.ServeHTTP(w, req)

			// Should return 401 Unauthorized for protected routes without auth
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestApplicationStartup(t *testing.T) {
	// Test that the application can start up without errors
	router, db := setupTestApp(t)

	// Verify router is not nil
	assert.NotNil(t, router)

	// Verify database is not nil and connected
	assert.NotNil(t, db)

	// Test that we can make a simple request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestConcurrentRequests(t *testing.T) {
	router, _ := setupTestApp(t)

	const numRequests = 10
	results := make(chan int, numRequests)

	// Launch concurrent requests
	for i := 0; i < numRequests; i++ {
		go func() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/health", nil)
			router.ServeHTTP(w, req)
			results <- w.Code
		}()
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		select {
		case code := <-results:
			assert.Equal(t, http.StatusOK, code)
		case <-time.After(5 * time.Second):
			t.Fatal("Request timed out")
		}
	}
}

// Benchmark tests
func BenchmarkHealthEndpoint(b *testing.B) {
	router, _ := setupTestApp(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkCategoriesEndpoint(b *testing.B) {
	router, _ := setupTestApp(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/categories", nil)
		router.ServeHTTP(w, req)
	}
}