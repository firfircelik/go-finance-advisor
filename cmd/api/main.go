package main

import (
	"log"

	"go-finance-advisor/internal/application"
	"go-finance-advisor/internal/domain"
	"go-finance-advisor/internal/infrastructure/api"
	"go-finance-advisor/internal/infrastructure/middleware"
	"go-finance-advisor/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Database setup
	db, err := gorm.Open(sqlite.Open("finance.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate
	db.AutoMigrate(&domain.User{}, &domain.Transaction{}, &domain.Category{}, &domain.Budget{}, &domain.Recommendation{})

	userSvc := &application.UserService{DB: db}
	txSvc := &application.TransactionService{DB: db}
	advisorSvc := &application.AdvisorService{DB: db}
	analyticsSvc := &application.AnalyticsService{DB: db}
	budgetSvc := &application.BudgetService{DB: db}
	categorySvc := &application.CategoryService{DB: db}
	reportsSvc := application.NewReportsService(db)
	exportSvc := application.NewExportService(db)
	marketSvc := &pkg.RealTimeMarketService{}

	userHandler := &api.UserHandler{Service: userSvc}
	txHandler := &api.TransactionHandler{Service: txSvc}
	advisorHandler := api.NewAdvisorHandler(advisorSvc, userSvc, marketSvc)
	analyticsHandler := &api.AnalyticsHandler{Service: analyticsSvc}
	budgetHandler := &api.BudgetHandler{Service: budgetSvc}
	categoryHandler := &api.CategoryHandler{Service: categorySvc}
	reportsHandler := &api.ReportsHandler{Service: reportsSvc}
	exportHandler := api.NewExportHandler(exportSvc)

	r := gin.Default()

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

	// Static files
	r.Static("/web", "./web")
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/web/")
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

	log.Println("Server listening on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
