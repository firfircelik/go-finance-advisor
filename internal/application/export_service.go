package application

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go-finance-advisor/internal/domain"
	"gorm.io/gorm"
)

type ExportService struct {
	DB *gorm.DB
}

func NewExportService(db *gorm.DB) *ExportService {
	return &ExportService{DB: db}
}

// ExportTransactions exports user transactions in the specified format
func (s *ExportService) ExportTransactions(userID uint, format domain.ExportFormat, startDate, endDate *time.Time) ([]byte, string, error) {
	// Get transactions
	var transactions []domain.Transaction
	query := s.DB.Preload("Category").Where("user_id = ?", userID)

	if startDate != nil && endDate != nil {
		query = query.Where("date BETWEEN ? AND ?", *startDate, *endDate)
	}

	err := query.Order("date DESC").Find(&transactions).Error
	if err != nil {
		return nil, "", err
	}

	switch format {
	case domain.ExportFormatCSV:
		return s.exportTransactionsCSV(transactions)
	case domain.ExportFormatJSON:
		return s.exportTransactionsJSON(transactions)
	default:
		return nil, "", fmt.Errorf("unsupported export format: %s", format)
	}
}

// ExportFinancialReport exports a financial report in the specified format
func (s *ExportService) ExportFinancialReport(userID uint, reportType string, year, month int, format domain.ExportFormat) ([]byte, string, error) {
	// Generate the report first
	reportsService := NewReportsService(s.DB)
	var report *domain.FinancialReport
	var err error

	switch reportType {
	case "monthly":
		report, err = reportsService.GenerateMonthlyReport(userID, year, month)
	case "yearly":
		report, err = reportsService.GenerateYearlyReport(userID, year)
	default:
		return nil, "", fmt.Errorf("unsupported report type: %s", reportType)
	}

	if err != nil {
		return nil, "", err
	}

	switch format {
	case domain.ExportFormatJSON:
		return s.exportReportJSON(report)
	case domain.ExportFormatCSV:
		return s.exportReportCSV(report)
	default:
		return nil, "", fmt.Errorf("unsupported export format: %s", format)
	}
}

// ExportBudgets exports user budgets in the specified format
func (s *ExportService) ExportBudgets(userID uint, format domain.ExportFormat) ([]byte, string, error) {
	// Get budgets
	var budgets []domain.Budget
	err := s.DB.Preload("Category").Where("user_id = ?", userID).Find(&budgets).Error
	if err != nil {
		return nil, "", err
	}

	switch format {
	case domain.ExportFormatCSV:
		return s.exportBudgetsCSV(budgets)
	case domain.ExportFormatJSON:
		return s.exportBudgetsJSON(budgets)
	default:
		return nil, "", fmt.Errorf("unsupported export format: %s", format)
	}
}

// ExportAllData exports all user financial data
func (s *ExportService) ExportAllData(userID uint, format domain.ExportFormat) ([]byte, string, error) {
	// Get all data
	var transactions []domain.Transaction
	var budgets []domain.Budget
	var categories []domain.Category

	// Get transactions
	err := s.DB.Preload("Category").Where("user_id = ?", userID).Find(&transactions).Error
	if err != nil {
		return nil, "", err
	}

	// Get budgets
	err = s.DB.Preload("Category").Where("user_id = ?", userID).Find(&budgets).Error
	if err != nil {
		return nil, "", err
	}

	// Get categories (all categories since they are global)
	err = s.DB.Find(&categories).Error
	if err != nil {
		return nil, "", err
	}

	exportData := domain.ExportData{
		Transactions: transactions,
		Budgets:      budgets,
		Categories:   categories,
		ExportedAt:   time.Now(),
		Format:       format,
	}

	switch format {
	case domain.ExportFormatJSON:
		return s.exportAllDataJSON(exportData)
	default:
		return nil, "", fmt.Errorf("unsupported export format for all data: %s", format)
	}
}

func (s *ExportService) exportTransactionsCSV(transactions []domain.Transaction) ([]byte, string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"ID", "Date", "Description", "Amount", "Type", "Category", "Created At"}
	if err := writer.Write(header); err != nil {
		return nil, "", err
	}

	// Write data
	for _, tx := range transactions {
		categoryName := ""
		if tx.Category.Name != "" {
			categoryName = tx.Category.Name
		}

		record := []string{
			strconv.FormatUint(uint64(tx.ID), 10),
			tx.Date.Format("2006-01-02"),
			tx.Description,
			strconv.FormatFloat(tx.Amount, 'f', 2, 64),
			tx.Type,
			categoryName,
			tx.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(record); err != nil {
			return nil, "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("transactions_%s.csv", time.Now().Format("2006-01-02"))
	return buf.Bytes(), filename, nil
}

func (s *ExportService) exportTransactionsJSON(transactions []domain.Transaction) ([]byte, string, error) {
	data, err := json.MarshalIndent(transactions, "", "  ")
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("transactions_%s.json", time.Now().Format("2006-01-02"))
	return data, filename, nil
}

func (s *ExportService) exportReportJSON(report *domain.FinancialReport) ([]byte, string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("financial_report_%s_%s.json", report.ReportType, report.StartDate.Format("2006-01"))
	return data, filename, nil
}

func (s *ExportService) exportReportCSV(report *domain.FinancialReport) ([]byte, string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write report summary
	summaryHeader := []string{"Metric", "Value"}
	if err := writer.Write(summaryHeader); err != nil {
		return nil, "", err
	}

	summaryData := [][]string{
		{"Report Type", report.ReportType},
		{"Period", fmt.Sprintf("%s to %s", report.StartDate.Format("2006-01-02"), report.EndDate.Format("2006-01-02"))},
		{"Total Income", strconv.FormatFloat(report.TotalIncome, 'f', 2, 64)},
		{"Total Expenses", strconv.FormatFloat(report.TotalExpenses, 'f', 2, 64)},
		{"Net Income", strconv.FormatFloat(report.NetIncome, 'f', 2, 64)},
		{"Savings Rate", strconv.FormatFloat(report.SavingsRate, 'f', 2, 64) + "%"},
		{"Transaction Count", strconv.Itoa(report.TransactionCount)},
	}

	for _, row := range summaryData {
		if err := writer.Write(row); err != nil {
			return nil, "", err
		}
	}

	// Add empty row
	writer.Write([]string{"", ""})

	// Write category breakdown
	categoryHeader := []string{"Category", "Type", "Amount", "Percentage", "Transaction Count"}
	if err := writer.Write(categoryHeader); err != nil {
		return nil, "", err
	}

	for _, category := range report.CategoryBreakdown {
		row := []string{
			category.CategoryName,
			"expense", // Default type since CategoryType doesn't exist
			strconv.FormatFloat(category.TotalAmount, 'f', 2, 64),
			strconv.FormatFloat(category.PercentageOfTotal, 'f', 2, 64) + "%",
			strconv.Itoa(category.TransactionCount),
		}
		if err := writer.Write(row); err != nil {
			return nil, "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("financial_report_%s_%s.csv", report.ReportType, report.StartDate.Format("2006-01"))
	return buf.Bytes(), filename, nil
}

func (s *ExportService) exportBudgetsCSV(budgets []domain.Budget) ([]byte, string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"ID", "Category", "Amount", "Period", "Start Date", "End Date", "Is Active", "Created At"}
	if err := writer.Write(header); err != nil {
		return nil, "", err
	}

	// Write data
	for _, budget := range budgets {
		categoryName := ""
		if budget.Category.Name != "" {
			categoryName = budget.Category.Name
		}

		record := []string{
			strconv.FormatUint(uint64(budget.ID), 10),
			categoryName,
			strconv.FormatFloat(budget.Amount, 'f', 2, 64),
			budget.Period,
			budget.StartDate.Format("2006-01-02"),
			budget.EndDate.Format("2006-01-02"),
			strconv.FormatBool(budget.IsActive),
			budget.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(record); err != nil {
			return nil, "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("budgets_%s.csv", time.Now().Format("2006-01-02"))
	return buf.Bytes(), filename, nil
}

func (s *ExportService) exportBudgetsJSON(budgets []domain.Budget) ([]byte, string, error) {
	data, err := json.MarshalIndent(budgets, "", "  ")
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("budgets_%s.json", time.Now().Format("2006-01-02"))
	return data, filename, nil
}

func (s *ExportService) exportAllDataJSON(exportData domain.ExportData) ([]byte, string, error) {
	data, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("financial_data_export_%s.json", time.Now().Format("2006-01-02"))
	return data, filename, nil
}
