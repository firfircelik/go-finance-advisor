package application

import (
	"errors"
	"time"

	"go-finance-advisor/internal/domain"

	"gorm.io/gorm"
)

type BudgetService struct {
	DB *gorm.DB
}

func NewBudgetService(db *gorm.DB) *BudgetService {
	return &BudgetService{DB: db}
}

// CreateBudget creates a new budget for a user
func (s *BudgetService) CreateBudget(budget *domain.Budget) error {
	// Check if budget already exists for this user, category, and period
	var existingBudget domain.Budget
	err := s.DB.Where("user_id = ? AND category_id = ? AND start_date <= ? AND end_date >= ? AND is_active = ?",
		budget.UserID, budget.CategoryID, budget.EndDate, budget.StartDate, true).First(&existingBudget).Error

	if err == nil {
		return errors.New("budget already exists for this category and period")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Set default values
	if budget.Period == "" {
		budget.Period = "monthly"
	}

	if budget.StartDate.IsZero() {
		budget.StartDate = time.Now().Truncate(24 * time.Hour)
	}

	if budget.EndDate.IsZero() {
		// Set end date based on period
		switch budget.Period {
		case "weekly":
			budget.EndDate = budget.StartDate.AddDate(0, 0, 7)
		case "monthly":
			budget.EndDate = budget.StartDate.AddDate(0, 1, 0)
		case "yearly":
			budget.EndDate = budget.StartDate.AddDate(1, 0, 0)
		default:
			budget.EndDate = budget.StartDate.AddDate(0, 1, 0) // Default to monthly
		}
	}

	budget.Remaining = budget.Amount
	budget.IsActive = true

	return s.DB.Create(budget).Error
}

// UpdateBudget updates an existing budget
func (s *BudgetService) UpdateBudget(budgetID uint, updates *domain.Budget) error {
	var budget domain.Budget
	err := s.DB.First(&budget, budgetID).Error
	if err != nil {
		return err
	}

	// Update allowed fields
	budget.Amount = updates.Amount
	budget.Period = updates.Period
	budget.StartDate = updates.StartDate
	budget.EndDate = updates.EndDate
	budget.IsActive = updates.IsActive

	// Recalculate remaining amount
	budget.CalculateRemaining()

	return s.DB.Save(&budget).Error
}

// GetBudgetsByUser retrieves all budgets for a user
func (s *BudgetService) GetBudgetsByUser(userID uint) ([]domain.Budget, error) {
	var budgets []domain.Budget
	err := s.DB.Preload("Category").Where("user_id = ?", userID).Order("created_at DESC").Find(&budgets).Error
	return budgets, err
}

// GetActiveBudgetsByUser retrieves active budgets for a user
func (s *BudgetService) GetActiveBudgetsByUser(userID uint) ([]domain.Budget, error) {
	var budgets []domain.Budget
	err := s.DB.Preload("Category").Where("user_id = ? AND is_active = ? AND end_date >= ?",
		userID, true, time.Now()).Order("created_at DESC").Find(&budgets).Error
	return budgets, err
}

// GetBudgetByID retrieves a specific budget
func (s *BudgetService) GetBudgetByID(budgetID uint) (*domain.Budget, error) {
	var budget domain.Budget
	err := s.DB.Preload("Category").First(&budget, budgetID).Error
	if err != nil {
		return nil, err
	}
	return &budget, nil
}

// DeleteBudget deletes a budget
func (s *BudgetService) DeleteBudget(budgetID uint) error {
	return s.DB.Delete(&domain.Budget{}, budgetID).Error
}

// UpdateBudgetSpending updates the spent amount for budgets when a transaction is added
func (s *BudgetService) UpdateBudgetSpending(userID, categoryID uint, amount float64, transactionDate time.Time) error {
	// Find active budgets that cover this transaction date
	var budgets []domain.Budget
	err := s.DB.Where("user_id = ? AND category_id = ? AND is_active = ? AND start_date <= ? AND end_date >= ?",
		userID, categoryID, true, transactionDate, transactionDate).Find(&budgets).Error

	if err != nil {
		return err
	}

	// Update spent amount for each matching budget
	for _, budget := range budgets {
		budget.Spent += amount
		budget.CalculateRemaining()
		s.DB.Save(&budget)
	}

	return nil
}

// GetBudgetSummary calculates budget summary for a user
func (s *BudgetService) GetBudgetSummary(userID uint) (*domain.BudgetSummary, error) {
	var budgets []domain.Budget
	err := s.DB.Where("user_id = ? AND is_active = ? AND end_date >= ?",
		userID, true, time.Now()).Find(&budgets).Error

	if err != nil {
		return nil, err
	}

	totalBudget := 0.0
	totalSpent := 0.0
	overBudgetCount := 0

	for _, budget := range budgets {
		totalBudget += budget.Amount
		totalSpent += budget.Spent

		if budget.Spent > budget.Amount {
			overBudgetCount++
		}
	}

	totalRemaining := totalBudget - totalSpent
	percentageUsed := 0.0
	if totalBudget > 0 {
		percentageUsed = (totalSpent / totalBudget) * 100
	}

	// Determine budget status
	budgetStatus := "on_track"
	if percentageUsed >= 100 {
		budgetStatus = "over_budget"
	} else if percentageUsed >= 80 {
		budgetStatus = "warning"
	}

	return &domain.BudgetSummary{
		TotalBudget:    totalBudget,
		TotalSpent:     totalSpent,
		TotalRemaining: totalRemaining,
		PercentageUsed: percentageUsed,
		BudgetStatus:   budgetStatus,
	}, nil
}

// GetBudgetsByCategory retrieves budgets for a specific category
func (s *BudgetService) GetBudgetsByCategory(userID, categoryID uint) ([]domain.Budget, error) {
	var budgets []domain.Budget
	err := s.DB.Preload("Category").Where("user_id = ? AND category_id = ?",
		userID, categoryID).Order("created_at DESC").Find(&budgets).Error
	return budgets, err
}

// GetBudgetsByPeriod retrieves budgets for a specific period
func (s *BudgetService) GetBudgetsByPeriod(userID uint, startDate, endDate time.Time) ([]domain.Budget, error) {
	var budgets []domain.Budget
	err := s.DB.Preload("Category").Where("user_id = ? AND start_date <= ? AND end_date >= ?",
		userID, endDate, startDate).Order("created_at DESC").Find(&budgets).Error
	return budgets, err
}

// RefreshBudgetSpending recalculates spent amounts for all budgets
func (s *BudgetService) RefreshBudgetSpending(userID uint) error {
	var budgets []domain.Budget
	err := s.DB.Where("user_id = ? AND is_active = ?", userID, true).Find(&budgets).Error
	if err != nil {
		return err
	}

	for _, budget := range budgets {
		// Calculate actual spent amount from transactions
		var totalSpent float64
		err := s.DB.Model(&domain.Transaction{}).Where(
			"user_id = ? AND category_id = ? AND date BETWEEN ? AND ?",
			budget.UserID, budget.CategoryID, budget.StartDate, budget.EndDate,
		).Select("COALESCE(SUM(amount), 0)").Scan(&totalSpent).Error

		if err != nil {
			continue
		}

		budget.Spent = totalSpent
		budget.CalculateRemaining()
		s.DB.Save(&budget)
	}

	return nil
}
