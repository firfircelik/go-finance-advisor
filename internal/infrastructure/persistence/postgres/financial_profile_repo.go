package postgres

import (
	"fmt"

	"go-finance-advisor/internal/domain"

	"gorm.io/gorm"
)

// FinancialProfileRepository handles financial profile data operations
type FinancialProfileRepository struct {
	db *gorm.DB
}

// NewFinancialProfileRepository creates a new financial profile repository
func NewFinancialProfileRepository(db *gorm.DB) *FinancialProfileRepository {
	return &FinancialProfileRepository{db: db}
}

// Create creates a new financial profile
func (r *FinancialProfileRepository) Create(profile *domain.FinancialProfile) error {
	return r.db.Create(profile).Error
}

// GetByUserID retrieves a financial profile by user ID
func (r *FinancialProfileRepository) GetByUserID(userID uint) (*domain.FinancialProfile, error) {
	var profile domain.FinancialProfile
	err := r.db.Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// Update updates a financial profile
func (r *FinancialProfileRepository) Update(profile *domain.FinancialProfile) error {
	return r.db.Save(profile).Error
}

// Delete soft deletes a financial profile
func (r *FinancialProfileRepository) Delete(id uint) error {
	return r.db.Delete(&domain.FinancialProfile{}, id).Error
}

// GetOrCreate gets existing profile or creates a new one with defaults
func (r *FinancialProfileRepository) GetOrCreate(userID uint) (*domain.FinancialProfile, error) {
	profile, err := r.GetByUserID(userID)
	if err == nil {
		return profile, nil
	}

	if err == gorm.ErrRecordNotFound {
		profile = &domain.FinancialProfile{
			UserID:          userID,
			MonthlyIncome:   0,
			FixedExpenses:   0,
			InvestmentRatio: 0.1, // 10% default
			UseRatio:        true,
			RiskTolerance:   "moderate",
		}
		if err := r.Create(profile); err != nil {
			return nil, err
		}
		return profile, nil
	}

	return nil, err
}

// UpdateInvestmentSettings updates only investment-related fields
func (r *FinancialProfileRepository) UpdateInvestmentSettings(
	userID uint,
	monthlyIncome, fixedExpenses, investmentRatio, investmentAmount float64,
	useRatio bool,
) error {
	updates := map[string]interface{}{
		"monthly_income":    monthlyIncome,
		"fixed_expenses":    fixedExpenses,
		"investment_ratio":  investmentRatio,
		"investment_amount": investmentAmount,
		"use_ratio":         useRatio,
	}

	result := r.db.Model(&domain.FinancialProfile{}).
		Where("user_id = ?", userID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no financial profile found for user %d", userID)
	}

	return nil
}
