package benchmarks

import (
	"testing"
	"time"

	"go-finance-advisor/internal/application"
	"go-finance-advisor/internal/domain"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupBenchDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to benchmark database: " + err.Error())
	}
	err = db.AutoMigrate(&domain.User{}, &domain.Transaction{}, &domain.Category{}, &domain.Recommendation{})
	if err != nil {
		panic("Failed to migrate benchmark database: " + err.Error())
	}
	return db
}

func BenchmarkGenerateAdvice(b *testing.B) {
	db := setupBenchDB()
	advisorService := &application.AdvisorService{DB: db}
	userService := &application.UserService{DB: db}
	txService := &application.TransactionService{DB: db}

	// Create test user
	user := domain.User{RiskTolerance: "moderate"}
	userService.Create(&user)

	// Create sample transactions
	for i := 0; i < 100; i++ {
		tx := domain.Transaction{
			UserID:      1,
			Type:        "income",
			Description: "Test income",
			Amount:      5000.0,
			CategoryID:  1,
			Date:        time.Now().AddDate(0, -1, 0),
		}
		txService.Create(&tx)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := advisorService.GenerateAdvice(&user)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCalculateMonthlySavings(b *testing.B) {
	db := setupBenchDB()
	advisorService := &application.AdvisorService{DB: db}
	txService := &application.TransactionService{DB: db}

	// Create sample transactions
	for i := 0; i < 1000; i++ {
		tx := domain.Transaction{
			UserID:      1,
			Type:        "income",
			Description: "Test income",
			Amount:      5000.0,
			CategoryID:  1,
			Date:        time.Now().AddDate(0, -1, 0),
		}
		txService.Create(&tx)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := advisorService.CalculateMonthlySavings(1)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTransactionCreate(b *testing.B) {
	db := setupBenchDB()
	txService := &application.TransactionService{DB: db}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx := domain.Transaction{
			UserID:      1,
			Type:        "income",
			Description: "Benchmark test",
			Amount:      1000.0,
			CategoryID:  1,
			Date:        time.Now(),
		}
		err := txService.Create(&tx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTransactionList(b *testing.B) {
	db := setupBenchDB()
	txService := &application.TransactionService{DB: db}

	// Pre-populate with transactions
	for i := 0; i < 1000; i++ {
		tx := domain.Transaction{
			UserID:      1,
			Type:        "income",
			Description: "Test transaction",
			Amount:      1000.0,
			CategoryID:  1,
			Date:        time.Now(),
		}
		txService.Create(&tx)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := txService.List(1)
		if err != nil {
			b.Fatal(err)
		}
	}
}
