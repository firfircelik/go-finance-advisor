package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"go-finance-advisor/internal/application"
	"go-finance-advisor/internal/config"
	"go-finance-advisor/internal/domain"
	"go-finance-advisor/internal/infrastructure/api"
	"go-finance-advisor/internal/infrastructure/middleware"
	"go-finance-advisor/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

// Build information (set via ldflags)
var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
	startTime time.Time
)

func main() {
	// Parse command line flags
	versionFlag := flag.Bool("version", false, "Show version information")
	healthFlag := flag.Bool("health", false, "Perform health check")
	flag.Parse()

	// Handle version flag
	if *versionFlag {
		fmt.Printf("Finance Advisor API\n")
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Build Time: %s\n", buildTime)
		fmt.Printf("Git Commit: %s\n", gitCommit)
		os.Exit(0)
	}

	// Handle health check flag
	if *healthFlag {
		fmt.Println("OK")
		os.Exit(0)
	}

	// Load environment variables from .env file (if exists)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := initLogger(cfg)
	logger.Info("Starting application",
		"version", version,
		"environment", cfg.App.Environment,
		"port", cfg.Server.Port,
	)

	// Set JWT secret in middleware
	middleware.SetJWTSecret(cfg.JWT.Secret)

	// Initialize database
	db, cleanup := initDatabase(cfg, logger)
	defer cleanup()

	// Auto migrate database schema
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Transaction{},
		&domain.Category{},
		&domain.Budget{},
		&domain.Recommendation{},
	); err != nil {
		logger.Error("Failed to migrate database", "error", err)
		log.Fatal(err)
	}

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

	// Initialize Gin router
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	// Add middleware
	r.Use(gin.Recovery())
	r.Use(loggerMiddleware(logger))
	r.Use(corsMiddleware(cfg))

	// Static files
	r.Static("/web", "./web")
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/web/")
	})

	// Health check routes (public)
	r.GET("/health", healthCheckHandler(db, logger))
	r.GET("/metrics", metricsHandler())

	// API routes
	v1 := r.Group("/api/v1")
	{
		// Health check routes
		v1.GET("/health", healthCheckHandler(db, logger))
		v1.GET("/metrics", metricsHandler())

		// Public routes
		v1.POST("/users", userHandler.Create)
		v1.POST("/auth/register", registerHandler(userHandler, cfg, logger))
		v1.POST("/auth/login", loginHandler(userHandler, cfg, logger))

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

	// Create HTTP server with timeouts
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server listening", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			log.Fatal(err)
		}
	}()

	// Record start time
	startTime = time.Now()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		log.Fatal(err)
	}

	logger.Info("Server exited gracefully")
}

// initLogger creates and configures the structured logger
func initLogger(cfg *config.Config) *slog.Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: parseLogLevel(cfg.Logging.Level),
	}

	if cfg.Logging.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// parseLogLevel converts string log level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// initDatabase initializes the database connection with proper configuration
func initDatabase(cfg *config.Config, logger *slog.Logger) (*gorm.DB, func()) {
	var db *gorm.DB
	var err error

	switch cfg.Database.Driver {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.Database.Path), &gorm.Config{})
	default:
		logger.Error("Unsupported database driver", "driver", cfg.Database.Driver)
		log.Fatalf("Unsupported database driver: %s", cfg.Database.Driver)
	}

	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		log.Fatal("Failed to connect to database:", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("Failed to get database instance", "error", err)
		log.Fatal(err)
	}

	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	logger.Info("Database connected successfully",
		"driver", cfg.Database.Driver,
		"max_idle_conns", cfg.Database.MaxIdleConns,
		"max_open_conns", cfg.Database.MaxOpenConns,
	)

	// Return cleanup function
	cleanup := func() {
		if err := sqlDB.Close(); err != nil {
			logger.Error("Error closing database", "error", err)
		} else {
			logger.Info("Database connection closed")
		}
	}

	return db, cleanup
}

// loggerMiddleware adds structured logging to Gin
func loggerMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("HTTP request",
			"method", method,
			"path", path,
			"status", statusCode,
			"latency", latency,
			"ip", clientIP,
		)
	}
}

// corsMiddleware handles CORS with configured origins
func corsMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range cfg.Server.AllowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				c.Header("Access-Control-Allow-Origin", allowedOrigin)
				break
			}
		}

		// If in development and origin is empty, allow
		if !allowed && cfg.IsDevelopment() && origin == "" {
			allowed = true
			c.Header("Access-Control-Allow-Origin", "*")
		}

		if allowed {
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// healthCheckHandler returns a comprehensive health check
func healthCheckHandler(db *gorm.DB, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check database connection
		sqlDB, err := db.DB()
		dbStatus := "connected"
		if err != nil {
			dbStatus = "error"
			logger.Error("Database health check failed", "error", err)
		} else if err := sqlDB.Ping(); err != nil {
			dbStatus = "disconnected"
			logger.Error("Database ping failed", "error", err)
		}

		status := "healthy"
		httpStatus := http.StatusOK

		if dbStatus != "connected" {
			status = "unhealthy"
			httpStatus = http.StatusServiceUnavailable
		}

		c.JSON(httpStatus, gin.H{
			"status":    status,
			"service":   "go-finance-advisor",
			"version":   version,
			"timestamp": time.Now().Unix(),
			"uptime":    time.Since(startTime).String(),
			"database":  dbStatus,
		})
	}
}

// metricsHandler returns system metrics
func metricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		c.JSON(http.StatusOK, gin.H{
			"uptime":           time.Since(startTime).String(),
			"goroutines":       runtime.NumGoroutine(),
			"memory_alloc_mb":  m.Alloc / 1024 / 1024,
			"memory_total_mb":  m.TotalAlloc / 1024 / 1024,
			"memory_sys_mb":    m.Sys / 1024 / 1024,
			"gc_runs":          m.NumGC,
			"last_gc_time":     time.Unix(0, int64(m.LastGC)).Format(time.RFC3339),
		})
	}
}

// registerHandler wraps the user handler Register with JWT expiration from config
func registerHandler(handler *api.UserHandler, cfg *config.Config, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		type RegisterRequest struct {
			Email     string `json:"email" binding:"required,email"`
			Password  string `json:"password" binding:"required,min=8"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		}

		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := handler.Service.Register(req.Email, req.Password, req.FirstName, req.LastName)
		if err != nil {
			if err.Error() == "user already exists" {
				c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
				return
			}
			logger.Error("Registration failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration failed"})
			return
		}

		// Generate JWT token with configured expiration
		token, err := middleware.GenerateToken(user.ID, cfg.JWT.Expiration)
		if err != nil {
			logger.Error("Token generation failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
			return
		}

		logger.Info("User registered successfully", "user_id", user.ID, "email", user.Email)

		c.JSON(http.StatusCreated, gin.H{
			"user": gin.H{
				"id":         user.ID,
				"email":      user.Email,
				"first_name": user.FirstName,
				"last_name":  user.LastName,
			},
			"token": token,
		})
	}
}

// loginHandler wraps the user handler Login with JWT expiration from config
func loginHandler(handler *api.UserHandler, cfg *config.Config, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		type LoginRequest struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := handler.Service.Login(req.Email, req.Password)
		if err != nil {
			logger.Warn("Login attempt failed", "email", req.Email)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// Generate JWT token with configured expiration
		token, err := middleware.GenerateToken(user.ID, cfg.JWT.Expiration)
		if err != nil {
			logger.Error("Token generation failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
			return
		}

		logger.Info("User logged in successfully", "user_id", user.ID, "email", user.Email)

		c.JSON(http.StatusOK, gin.H{
			"user": gin.H{
				"id":         user.ID,
				"email":      user.Email,
				"first_name": user.FirstName,
				"last_name":  user.LastName,
			},
			"token": token,
		})
	}
}
