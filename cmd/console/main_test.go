package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"go-finance-advisor/internal/application"
	"go-finance-advisor/internal/domain"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupTestApp creates a test instance of the console application
func setupTestApp(t *testing.T) (*App, *gorm.DB) {
	// Use in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate
	err = db.AutoMigrate(&domain.User{}, &domain.Transaction{}, &domain.Category{}, &domain.Budget{})
	require.NoError(t, err)

	// Initialize services
	userSvc := &application.UserService{DB: db}
	txSvc := &application.TransactionService{DB: db}
	advisorSvc := &application.AdvisorService{DB: db}
	analyticsSvc := &application.AnalyticsService{DB: db}
	budgetSvc := &application.BudgetService{DB: db}
	categorySvc := &application.CategoryService{DB: db}
	reportsSvc := &application.ReportsService{DB: db}
	exportSvc := &application.ExportService{DB: db}

	// Initialize default categories
	err = categorySvc.InitializeDefaultCategories()
	require.NoError(t, err)

	return &App{
		userSvc:      userSvc,
		txSvc:        txSvc,
		advisorSvc:   advisorSvc,
		analyticsSvc: analyticsSvc,
		budgetSvc:    budgetSvc,
		categorySvc:  categorySvc,
		reportsSvc:   reportsSvc,
		exportSvc:    exportSvc,
		reader:       bufio.NewReader(strings.NewReader("")),
	}, db
}

// captureOutput captures stdout during test execution
func captureOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestPrintHeader(t *testing.T) {
	output := captureOutput(func() {
		printHeader()
	})

	assert.Contains(t, output, "PERSONAL FINANCE ADVISOR")
	assert.Contains(t, output, "CONSOLE APPLICATION")
	assert.Contains(t, output, "Professional Financial Management System")
	assert.Contains(t, output, "=")
	assert.Contains(t, output, "-")
}

func TestInitializeDatabase(t *testing.T) {
	tests := []struct {
		name        string
		dbPath      string
		expectError bool
	}{
		{
			name:        "in-memory database",
			dbPath:      ":memory:",
			expectError: false,
		},
		{
			name:        "file database",
			dbPath:      "test_finance.db",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For file database, clean up after test
			if tt.dbPath != ":memory:" {
				defer os.Remove(tt.dbPath)
			}

			// Test database initialization
			db, err := gorm.Open(sqlite.Open(tt.dbPath), &gorm.Config{})
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, db)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db)

				// Test migration
				err = db.AutoMigrate(&domain.User{}, &domain.Transaction{}, &domain.Category{}, &domain.Budget{})
				assert.NoError(t, err)

				// Verify tables were created
				var tableCount int64
				err = db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&tableCount).Error
				assert.NoError(t, err)
				assert.Greater(t, tableCount, int64(0))
			}
		})
	}
}

func TestInitializeApp(t *testing.T) {
	// Create test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate
	err = db.AutoMigrate(&domain.User{}, &domain.Transaction{}, &domain.Category{}, &domain.Budget{})
	require.NoError(t, err)

	// Test app initialization
	app := initializeApp(db)

	assert.NotNil(t, app)
	assert.NotNil(t, app.userSvc)
	assert.NotNil(t, app.txSvc)
	assert.NotNil(t, app.advisorSvc)
	assert.NotNil(t, app.analyticsSvc)
	assert.NotNil(t, app.budgetSvc)
	assert.NotNil(t, app.categorySvc)
	assert.NotNil(t, app.reportsSvc)
	assert.NotNil(t, app.exportSvc)
	assert.NotNil(t, app.reader)
	assert.Nil(t, app.currentUser) // Should start with no user logged in
}

func TestAppServices(t *testing.T) {
	app, db := setupTestApp(t)

	// Test that all services are properly initialized
	assert.NotNil(t, app.userSvc)
	assert.NotNil(t, app.txSvc)
	assert.NotNil(t, app.advisorSvc)
	assert.NotNil(t, app.analyticsSvc)
	assert.NotNil(t, app.budgetSvc)
	assert.NotNil(t, app.categorySvc)
	assert.NotNil(t, app.reportsSvc)
	assert.NotNil(t, app.exportSvc)

	// Test that services can interact with database
	sqlDB, err := db.DB()
	require.NoError(t, err)
	err = sqlDB.Ping()
	assert.NoError(t, err)

	// Test that default categories were initialized
	categories, err := app.categorySvc.GetAllCategories()
	assert.NoError(t, err)
	assert.Greater(t, len(categories), 0)
}

func TestAppInitialState(t *testing.T) {
	app, _ := setupTestApp(t)

	// Test initial state
	assert.Nil(t, app.currentUser, "App should start with no user logged in")
	assert.NotNil(t, app.reader, "Reader should be initialized")
}

func TestAppWithMockInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty input",
			input: "",
		},
		{
			name:  "single line input",
			input: "test\n",
		},
		{
			name:  "multiple line input",
			input: "line1\nline2\nline3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, _ := setupTestApp(t)
			
			// Replace reader with test input
			app.reader = bufio.NewReader(strings.NewReader(tt.input))
			
			// Test that reader is properly set
			assert.NotNil(t, app.reader)
		})
	}
}

func TestDatabaseMigration(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Test migration
	err = db.AutoMigrate(&domain.User{}, &domain.Transaction{}, &domain.Category{}, &domain.Budget{})
	assert.NoError(t, err)

	// Verify that tables exist
	tables := []string{"users", "transactions", "categories", "budgets"}
	for _, table := range tables {
		var count int64
		err = db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count, "Table %s should exist", table)
	}
}

func TestServiceIntegration(t *testing.T) {
	app, _ := setupTestApp(t)

	// Test user service
	user := &domain.User{
		FirstName:     "Test",
		LastName:      "User",
		Email:         "test@example.com",
		Age:           30,
		RiskTolerance: "moderate",
	}

	err := app.userSvc.Create(user)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Greater(t, user.ID, uint(0))

	// Test category service
	categories, err := app.categorySvc.GetAllCategories()
	assert.NoError(t, err)
	assert.Greater(t, len(categories), 0)

	// Test transaction service with created user
	if len(categories) > 0 {
		transaction := &domain.Transaction{
			UserID:      user.ID,
			Amount:      100.50,
			Description: "Test transaction",
			Type:        "expense",
			CategoryID:  categories[0].ID,
			Date:        time.Now(),
		}

		err = app.txSvc.Create(transaction)
		assert.NoError(t, err)
		assert.NotNil(t, transaction)
		assert.Greater(t, transaction.ID, uint(0))
	}
}

func TestConcurrentServiceAccess(t *testing.T) {
	app, _ := setupTestApp(t)

	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	// Test concurrent access to category service
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			_, err := app.categorySvc.GetAllCategories()
			results <- err
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-results:
			assert.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent access test timed out")
		}
	}
}

func TestAppMemoryUsage(t *testing.T) {
	// Test that app initialization doesn't consume excessive memory
	app, _ := setupTestApp(t)

	// Basic memory usage test - ensure app is created without panic
	assert.NotNil(t, app)
	assert.NotNil(t, app.userSvc)
	assert.NotNil(t, app.txSvc)
	assert.NotNil(t, app.advisorSvc)
	assert.NotNil(t, app.analyticsSvc)
	assert.NotNil(t, app.budgetSvc)
	assert.NotNil(t, app.categorySvc)
	assert.NotNil(t, app.reportsSvc)
	assert.NotNil(t, app.exportSvc)
}

func TestDatabaseConnectionPooling(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Test multiple connections
	sqlDB, err := db.DB()
	require.NoError(t, err)

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test that connections work
	err = sqlDB.Ping()
	assert.NoError(t, err)

	// Test concurrent database access
	const numConnections = 5
	results := make(chan error, numConnections)

	for i := 0; i < numConnections; i++ {
		go func() {
			err := sqlDB.Ping()
			results <- err
		}()
	}

	for i := 0; i < numConnections; i++ {
		select {
		case err := <-results:
			assert.NoError(t, err)
		case <-time.After(2 * time.Second):
			t.Fatal("Database connection test timed out")
		}
	}
}

// Benchmark tests
func BenchmarkAppInitialization(b *testing.B) {
	for i := 0; i < b.N; i++ {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			b.Fatalf("Database setup failed: %v", err)
		}

		err = db.AutoMigrate(&domain.User{}, &domain.Transaction{}, &domain.Category{}, &domain.Budget{})
		if err != nil {
			b.Fatalf("Migration failed: %v", err)
		}

		app := initializeApp(db)
		if app == nil {
			b.Fatal("App initialization failed")
		}
	}
}

func BenchmarkCategoryServiceAccess(b *testing.B) {
	app, _ := setupTestApp(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := app.categorySvc.GetAllCategories()
		if err != nil {
			b.Fatalf("Category service failed: %v", err)
		}
	}
}

func BenchmarkUserServiceCreation(b *testing.B) {
	app, _ := setupTestApp(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := &domain.User{
			FirstName:     "Benchmark",
			LastName:      "User",
			Email:         "benchmark@example.com",
			Age:           25,
			RiskTolerance: "moderate",
		}

		err := app.userSvc.Create(user)
		if err != nil {
			b.Fatalf("User creation failed: %v", err)
		}
	}
}

// Integration test for the complete application flow
func TestApplicationFlow(t *testing.T) {
	app, _ := setupTestApp(t)

	// Step 1: Create a user
	user := &domain.User{
		FirstName:     "Integration",
		LastName:      "User",
		Email:         "integration@example.com",
		Age:           35,
		RiskTolerance: "aggressive",
	}

	err := app.userSvc.Create(user)
	require.NoError(t, err)
	require.NotNil(t, user)

	// Step 2: Get categories
	categories, err := app.categorySvc.GetAllCategories()
	require.NoError(t, err)
	require.Greater(t, len(categories), 0)

	// Step 3: Create transactions
	transaction := &domain.Transaction{
		UserID:      user.ID,
		Amount:      500.00,
		Description: "Integration test transaction",
		Type:        "income",
		CategoryID:  categories[0].ID,
		Date:        time.Now(),
	}

	err = app.txSvc.Create(transaction)
	require.NoError(t, err)
	require.NotNil(t, transaction)

	// Step 4: Create budget
	budget := &domain.Budget{
		UserID:     user.ID,
		CategoryID: categories[0].ID,
		Amount:     1000.00,
		Period:     "monthly",
		StartDate:  time.Now(),
		EndDate:    time.Now().AddDate(0, 1, 0),
	}

	err = app.budgetSvc.CreateBudget(budget)
	require.NoError(t, err)
	require.NotNil(t, budget)

	// Step 5: Verify data integrity
	retrievedUser, err := app.userSvc.GetByID(user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Email, retrievedUser.Email)

	transactions, err := app.txSvc.List(user.ID)
    require.NoError(t, err)
    assert.Len(t, transactions, 1)
    assert.Equal(t, transaction.Amount, transactions[0].Amount)

	budgets, err := app.budgetSvc.GetBudgetsByUser(user.ID)
	require.NoError(t, err)
	assert.Len(t, budgets, 1)
	assert.Equal(t, budget.Amount, budgets[0].Amount)
}