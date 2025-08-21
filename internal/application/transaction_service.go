package application

import (
	"time"

	"go-finance-advisor/internal/domain"

	"gorm.io/gorm"
)

type TransactionService struct {
	DB *gorm.DB
}

// Create creates a new transaction
func (s *TransactionService) Create(transaction *domain.Transaction) error {
	return s.DB.Create(transaction).Error
}

// List returns all transactions for a user
func (s *TransactionService) List(userID uint) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	err := s.DB.Preload("Category").Where("user_id = ?", userID).Order("date DESC").Find(&transactions).Error
	return transactions, err
}

// ListWithFilters returns transactions with filtering options
func (s *TransactionService) ListWithFilters(userID uint, transactionType *string, categoryID *uint, startDate *time.Time, endDate *time.Time, limit int, offset int) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	query := s.DB.Preload("Category").Where("user_id = ?", userID)

	// Apply filters
	if transactionType != nil && *transactionType != "" {
		query = query.Where("type = ?", *transactionType)
	}

	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}

	if startDate != nil {
		query = query.Where("date >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("date <= ?", *endDate)
	}

	// Apply pagination and ordering
	err := query.Order("date DESC").Limit(limit).Offset(offset).Find(&transactions).Error
	return transactions, err
}

// GetByID returns a transaction by ID
func (s *TransactionService) GetByID(id uint) (*domain.Transaction, error) {
	var transaction domain.Transaction
	err := s.DB.Preload("Category").First(&transaction, id).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// Update updates an existing transaction
func (s *TransactionService) Update(transaction *domain.Transaction) error {
	return s.DB.Save(transaction).Error
}

// Delete deletes a transaction
func (s *TransactionService) Delete(id uint) error {
	return s.DB.Delete(&domain.Transaction{}, id).Error
}

// GetTransactionsByDateRange returns transactions within a date range
func (s *TransactionService) GetTransactionsByDateRange(userID uint, startDate, endDate time.Time) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	err := s.DB.Preload("Category").Where("user_id = ? AND date >= ? AND date <= ?", userID, startDate, endDate).Order("date DESC").Find(&transactions).Error
	return transactions, err
}

// GetTransactionsByCategory returns transactions for a specific category
func (s *TransactionService) GetTransactionsByCategory(userID uint, categoryID uint, startDate, endDate time.Time) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	err := s.DB.Preload("Category").Where("user_id = ? AND category_id = ? AND date >= ? AND date <= ?", userID, categoryID, startDate, endDate).Order("date DESC").Find(&transactions).Error
	return transactions, err
}

// GetTransactionsByType returns transactions by type (income/expense)
func (s *TransactionService) GetTransactionsByType(userID uint, transactionType string, startDate, endDate time.Time) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	err := s.DB.Preload("Category").Where("user_id = ? AND type = ? AND date >= ? AND date <= ?", userID, transactionType, startDate, endDate).Order("date DESC").Find(&transactions).Error
	return transactions, err
}

// GetTotalByType returns the total amount for a transaction type within a date range
func (s *TransactionService) GetTotalByType(userID uint, transactionType string, startDate, endDate time.Time) (float64, error) {
	var total float64
	err := s.DB.Model(&domain.Transaction{}).Where("user_id = ? AND type = ? AND date >= ? AND date <= ?", userID, transactionType, startDate, endDate).Select("COALESCE(SUM(amount), 0)").Scan(&total).Error
	return total, err
}

// GetTotalByCategory returns the total amount for a category within a date range
func (s *TransactionService) GetTotalByCategory(userID uint, categoryID uint, startDate, endDate time.Time) (float64, error) {
	var total float64
	err := s.DB.Model(&domain.Transaction{}).Where("user_id = ? AND category_id = ? AND date >= ? AND date <= ?", userID, categoryID, startDate, endDate).Select("COALESCE(SUM(amount), 0)").Scan(&total).Error
	return total, err
}

// GetCategoryTotals returns total amounts grouped by category
func (s *TransactionService) GetCategoryTotals(userID uint, transactionType string, startDate, endDate time.Time) (map[uint]float64, error) {
	type CategoryTotal struct {
		CategoryID uint
		Total      float64
	}

	var results []CategoryTotal
	err := s.DB.Model(&domain.Transaction{}).
		Select("category_id, COALESCE(SUM(amount), 0) as total").
		Where("user_id = ? AND type = ? AND date >= ? AND date <= ?", userID, transactionType, startDate, endDate).
		Group("category_id").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	totals := make(map[uint]float64)
	for _, result := range results {
		totals[result.CategoryID] = result.Total
	}

	return totals, nil
}

// GetMonthlyTotals returns monthly totals for a transaction type
func (s *TransactionService) GetMonthlyTotals(userID uint, transactionType string, startDate, endDate time.Time) (map[string]float64, error) {
	type MonthlyTotal struct {
		Month string
		Total float64
	}

	var results []MonthlyTotal
	err := s.DB.Model(&domain.Transaction{}).
		Select("strftime('%Y-%m', date) as month, COALESCE(SUM(amount), 0) as total").
		Where("user_id = ? AND type = ? AND date >= ? AND date <= ?", userID, transactionType, startDate, endDate).
		Group("strftime('%Y-%m', date)").
		Order("month").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	totals := make(map[string]float64)
	for _, result := range results {
		totals[result.Month] = result.Total
	}

	return totals, nil
}

// GetDailyAverages calculates daily averages for income and expenses
func (s *TransactionService) GetDailyAverages(userID uint, startDate, endDate time.Time) (map[string]float64, error) {
	days := int(endDate.Sub(startDate).Hours()/24) + 1
	if days <= 0 {
		days = 1
	}

	incomeTotal, err := s.GetTotalByType(userID, "income", startDate, endDate)
	if err != nil {
		return nil, err
	}

	expenseTotal, err := s.GetTotalByType(userID, "expense", startDate, endDate)
	if err != nil {
		return nil, err
	}

	averages := map[string]float64{
		"daily_income":  incomeTotal / float64(days),
		"daily_expense": expenseTotal / float64(days),
		"daily_net":     (incomeTotal - expenseTotal) / float64(days),
	}

	return averages, nil
}