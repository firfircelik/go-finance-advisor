package application

import (
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-finance-advisor/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupExportTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&domain.User{}, &domain.Transaction{}, &domain.Budget{}, &domain.Category{}, &domain.FinancialReport{})
	return db
}

func createExportTestData(db *gorm.DB) (uint, uint) {
	// Create test user
	user := domain.User{
		FirstName: "Test",
		LastName:  "User",
		Email:     "test@example.com",
	}
	db.Create(&user)

	// Create test categories
	foodCategory := domain.Category{
		Name: "Food",
		Type: "expense",
	}
	db.Create(&foodCategory)

	salaryCategory := domain.Category{
		Name: "Salary",
		Type: "income",
	}
	db.Create(&salaryCategory)

	// Create test transactions
	transactions := []domain.Transaction{
		{
			UserID:      user.ID,
			CategoryID:  foodCategory.ID,
			Amount:      50.00,
			Type:        "expense",
			Description: "Grocery shopping",
			Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			UserID:      user.ID,
			CategoryID:  salaryCategory.ID,
			Amount:      3000.00,
			Type:        "income",
			Description: "Monthly salary",
			Date:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			UserID:      user.ID,
			CategoryID:  foodCategory.ID,
			Amount:      25.50,
			Type:        "expense",
			Description: "Restaurant dinner",
			Date:        time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tx := range transactions {
		db.Create(&tx)
	}

	// Create test budgets
	budgets := []domain.Budget{
		{
			UserID:     user.ID,
			CategoryID: foodCategory.ID,
			Amount:     200.00,
			Period:     "monthly",
			StartDate:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			IsActive:   true,
		},
		{
			UserID:     user.ID,
			CategoryID: salaryCategory.ID,
			Amount:     500.00,
			Period:     "monthly",
			StartDate:  time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
			IsActive:   false,
		},
	}
	for _, budget := range budgets {
		db.Create(&budget)
	}

	return user.ID, foodCategory.ID
}

func TestExportService_ExportTransactions(t *testing.T) {
	db := setupExportTestDB()
	service := NewExportService(db)
	userID, _ := createExportTestData(db)

	t.Run("export transactions as CSV", func(t *testing.T) {
		data, filename, err := service.ExportTransactions(userID, domain.ExportFormatCSV, nil, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, data)
		assert.Contains(t, filename, "transactions_")
		assert.Contains(t, filename, ".csv")

		// Parse CSV and verify content
		reader := csv.NewReader(strings.NewReader(string(data)))
		records, err := reader.ReadAll()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(records), 2) // Header + at least 1 data row

		// Check header
		header := records[0]
		assert.Contains(t, header, "ID")
		assert.Contains(t, header, "Amount")
		assert.Contains(t, header, "Type")
		assert.Contains(t, header, "Description")
	})

	t.Run("export transactions as JSON", func(t *testing.T) {
		data, filename, err := service.ExportTransactions(userID, domain.ExportFormatJSON, nil, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, data)
		assert.Contains(t, filename, "transactions_")
		assert.Contains(t, filename, ".json")

		// Parse JSON and verify content
		var transactions []domain.Transaction
		err = json.Unmarshal(data, &transactions)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(transactions), 1)
		assert.Equal(t, userID, transactions[0].UserID)
	})

	t.Run("export transactions with date range", func(t *testing.T) {
		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

		data, filename, err := service.ExportTransactions(userID, domain.ExportFormatJSON, &startDate, &endDate)
		require.NoError(t, err)
		assert.NotEmpty(t, data)
		assert.Contains(t, filename, "transactions_")

		// Parse JSON and verify date filtering
		var transactions []domain.Transaction
		err = json.Unmarshal(data, &transactions)
		require.NoError(t, err)

		// Should only include January transactions
		for _, tx := range transactions {
			assert.True(t, tx.Date.After(startDate.Add(-time.Second)) && tx.Date.Before(endDate.Add(time.Second)))
		}
	})

	t.Run("export transactions with unsupported format", func(t *testing.T) {
		_, _, err := service.ExportTransactions(userID, domain.ExportFormatPDF, nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported export format")
	})

	t.Run("export transactions for non-existent user", func(t *testing.T) {
		data, filename, err := service.ExportTransactions(99999, domain.ExportFormatJSON, nil, nil)
		require.NoError(t, err) // Should not error, just return empty data
		assert.NotEmpty(t, filename)

		var transactions []domain.Transaction
		err = json.Unmarshal(data, &transactions)
		require.NoError(t, err)
		assert.Empty(t, transactions)
	})
}

func TestExportService_ExportBudgets(t *testing.T) {
	db := setupExportTestDB()
	service := NewExportService(db)
	userID, _ := createExportTestData(db)

	t.Run("export budgets as CSV", func(t *testing.T) {
		data, filename, err := service.ExportBudgets(userID, domain.ExportFormatCSV)
		require.NoError(t, err)
		assert.NotEmpty(t, data)
		assert.Contains(t, filename, "budgets_")
		assert.Contains(t, filename, ".csv")

		// Parse CSV and verify content
		reader := csv.NewReader(strings.NewReader(string(data)))
		records, err := reader.ReadAll()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(records), 2) // Header + at least 1 data row

		// Check header
		header := records[0]
		assert.Contains(t, header, "ID")
		assert.Contains(t, header, "Amount")
		assert.Contains(t, header, "Period")
		assert.Contains(t, header, "Is Active")
	})

	t.Run("export budgets as JSON", func(t *testing.T) {
		data, filename, err := service.ExportBudgets(userID, domain.ExportFormatJSON)
		require.NoError(t, err)
		assert.NotEmpty(t, data)
		assert.Contains(t, filename, "budgets_")
		assert.Contains(t, filename, ".json")

		// Parse JSON and verify content
		var budgets []domain.Budget
		err = json.Unmarshal(data, &budgets)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(budgets), 1)
		assert.Equal(t, userID, budgets[0].UserID)
	})

	t.Run("export budgets with unsupported format", func(t *testing.T) {
		_, _, err := service.ExportBudgets(userID, domain.ExportFormatPDF)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported export format")
	})
}

func TestExportService_ExportAllData(t *testing.T) {
	db := setupExportTestDB()
	service := NewExportService(db)
	userID, _ := createExportTestData(db)

	t.Run("export all data as JSON", func(t *testing.T) {
		data, filename, err := service.ExportAllData(userID, domain.ExportFormatJSON)
		require.NoError(t, err)
		assert.NotEmpty(t, data)
		assert.Contains(t, filename, "financial_data_export_")
		assert.Contains(t, filename, ".json")

		// Parse JSON and verify content
		var exportData domain.ExportData
		err = json.Unmarshal(data, &exportData)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(exportData.Transactions), 1)
		assert.GreaterOrEqual(t, len(exportData.Budgets), 1)
		assert.GreaterOrEqual(t, len(exportData.Categories), 0)
		assert.Equal(t, domain.ExportFormatJSON, exportData.Format)
		assert.False(t, exportData.ExportedAt.IsZero())
	})

	t.Run("export all data with unsupported format", func(t *testing.T) {
		_, _, err := service.ExportAllData(userID, domain.ExportFormatCSV)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported export format for all data")
	})
}

func TestExportService_ExportFinancialReport(t *testing.T) {
	db := setupExportTestDB()
	service := NewExportService(db)
	userID, _ := createExportTestData(db)

	// Create ReportsService to generate actual reports
	reportsService := NewReportsService(db)

	t.Run("export monthly report as JSON", func(t *testing.T) {
		// Generate the report first
		report, err := reportsService.GenerateMonthlyReport(userID, 2024, 1)
		require.NoError(t, err)
		require.NotNil(t, report)

		data, filename, err := service.ExportFinancialReport(userID, "monthly", 2024, 1, domain.ExportFormatJSON)
		require.NoError(t, err)
		assert.NotEmpty(t, data)
		assert.Contains(t, filename, "financial_report_monthly_")
		assert.Contains(t, filename, ".json")

		// Parse JSON and verify content
		var reportData domain.FinancialReport
		err = json.Unmarshal(data, &reportData)
		require.NoError(t, err)
		assert.Equal(t, userID, reportData.UserID)
		assert.Equal(t, "monthly", reportData.ReportType)
		assert.Equal(t, 2024, reportData.StartDate.Year())
		assert.Equal(t, time.January, reportData.StartDate.Month())
	})

	t.Run("export yearly report as JSON", func(t *testing.T) {
		// Generate the report first
		report, err := reportsService.GenerateYearlyReport(userID, 2024)
		require.NoError(t, err)
		require.NotNil(t, report)

		data, filename, err := service.ExportFinancialReport(userID, "yearly", 2024, 0, domain.ExportFormatJSON)
		require.NoError(t, err)
		assert.NotEmpty(t, data)
		assert.Contains(t, filename, "financial_report_yearly_")
		assert.Contains(t, filename, ".json")

		// Parse JSON and verify content
		var reportData domain.FinancialReport
		err = json.Unmarshal(data, &reportData)
		require.NoError(t, err)
		assert.Equal(t, userID, reportData.UserID)
		assert.Equal(t, "yearly", reportData.ReportType)
		assert.Equal(t, 2024, reportData.StartDate.Year())
	})

	t.Run("export report as CSV", func(t *testing.T) {
		// Generate the report first
		report, err := reportsService.GenerateMonthlyReport(userID, 2024, 1)
		require.NoError(t, err)
		require.NotNil(t, report)

		data, filename, err := service.ExportFinancialReport(userID, "monthly", 2024, 1, domain.ExportFormatCSV)
		require.NoError(t, err)
		assert.NotEmpty(t, data)
		assert.Contains(t, filename, "financial_report_monthly_")
		assert.Contains(t, filename, ".csv")

		// Verify CSV content is not empty
		assert.Greater(t, len(data), 0)
	})

	t.Run("export report with unsupported type", func(t *testing.T) {
		_, _, err := service.ExportFinancialReport(userID, "invalid", 2024, 1, domain.ExportFormatJSON)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported report type")
	})

	t.Run("export report with unsupported format", func(t *testing.T) {
		_, _, err := service.ExportFinancialReport(userID, "monthly", 2024, 1, domain.ExportFormatPDF)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported export format")
	})
}

func TestExportService_Integration(t *testing.T) {
	db := setupExportTestDB()
	service := NewExportService(db)
	userID, _ := createExportTestData(db)

	t.Run("export data maintains referential integrity", func(t *testing.T) {
		// Export transactions
		txData, _, err := service.ExportTransactions(userID, domain.ExportFormatJSON, nil, nil)
		require.NoError(t, err)

		var transactions []domain.Transaction
		err = json.Unmarshal(txData, &transactions)
		require.NoError(t, err)

		// Export budgets
		budgetData, _, err := service.ExportBudgets(userID, domain.ExportFormatJSON)
		require.NoError(t, err)

		var budgets []domain.Budget
		err = json.Unmarshal(budgetData, &budgets)
		require.NoError(t, err)

		// Verify relationships
		for _, tx := range transactions {
			assert.Equal(t, userID, tx.UserID)
			assert.NotZero(t, tx.CategoryID)
		}

		for _, budget := range budgets {
			assert.Equal(t, userID, budget.UserID)
			assert.NotZero(t, budget.CategoryID)
		}
	})

	t.Run("export handles empty data gracefully", func(t *testing.T) {
		// Create new user with no data
		emptyUser := domain.User{
			FirstName: "Empty",
			LastName:  "User",
			Email:     "empty@example.com",
		}
		db.Create(&emptyUser)

		// Export transactions for empty user
		data, filename, err := service.ExportTransactions(emptyUser.ID, domain.ExportFormatJSON, nil, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, filename)

		var transactions []domain.Transaction
		err = json.Unmarshal(data, &transactions)
		require.NoError(t, err)
		assert.Empty(t, transactions)

		// Export budgets for empty user
		data, filename, err = service.ExportBudgets(emptyUser.ID, domain.ExportFormatJSON)
		require.NoError(t, err)
		assert.NotEmpty(t, filename)

		var budgets []domain.Budget
		err = json.Unmarshal(data, &budgets)
		require.NoError(t, err)
		assert.Empty(t, budgets)
	})

	t.Run("export filename generation is consistent", func(t *testing.T) {
		_, filename1, err := service.ExportTransactions(userID, domain.ExportFormatCSV, nil, nil)
		require.NoError(t, err)

		_, filename2, err := service.ExportTransactions(userID, domain.ExportFormatJSON, nil, nil)
		require.NoError(t, err)

		// Both should contain date and appropriate extension
		assert.Contains(t, filename1, time.Now().Format("2006-01-02"))
		assert.Contains(t, filename2, time.Now().Format("2006-01-02"))
		assert.Contains(t, filename1, ".csv")
		assert.Contains(t, filename2, ".json")
	})
}
