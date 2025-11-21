package postgres

import (
	"time"

	"go-finance-advisor/internal/domain"

	"gorm.io/gorm"
)

// AllocationPlanRepository handles allocation plan operations
type AllocationPlanRepository struct {
	db *gorm.DB
}

// NewAllocationPlanRepository creates a new allocation plan repository
func NewAllocationPlanRepository(db *gorm.DB) *AllocationPlanRepository {
	return &AllocationPlanRepository{db: db}
}

// Create creates a new allocation plan with allocations
func (r *AllocationPlanRepository) Create(plan *domain.AllocationPlan) error {
	return r.db.Create(plan).Error
}

// GetByID retrieves a plan by ID with all allocations
func (r *AllocationPlanRepository) GetByID(id uint) (*domain.AllocationPlan, error) {
	var plan domain.AllocationPlan
	err := r.db.Preload("Allocations").First(&plan, id).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// GetUserPlans retrieves all plans for a user
func (r *AllocationPlanRepository) GetUserPlans(userID uint) ([]domain.AllocationPlan, error) {
	var plans []domain.AllocationPlan
	err := r.db.Where("user_id = ?", userID).
		Preload("Allocations").
		Order("created_at DESC").
		Find(&plans).Error
	return plans, err
}

// GetActivePlans retrieves active (non-expired, non-executed) plans for a user
func (r *AllocationPlanRepository) GetActivePlans(userID uint) ([]domain.AllocationPlan, error) {
	var plans []domain.AllocationPlan
	err := r.db.Where("user_id = ? AND is_executed = ? AND valid_until > ?", userID, false, time.Now()).
		Preload("Allocations").
		Order("created_at DESC").
		Find(&plans).Error
	return plans, err
}

// GetLatestPlan retrieves the most recent plan for a user
func (r *AllocationPlanRepository) GetLatestPlan(userID uint) (*domain.AllocationPlan, error) {
	var plan domain.AllocationPlan
	err := r.db.Where("user_id = ?", userID).
		Preload("Allocations").
		Order("created_at DESC").
		First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// MarkAsExecuted marks a plan as executed
func (r *AllocationPlanRepository) MarkAsExecuted(id uint) error {
	return r.db.Model(&domain.AllocationPlan{}).
		Where("id = ?", id).
		Update("is_executed", true).Error
}

// DeletePlan soft deletes a plan and its allocations
func (r *AllocationPlanRepository) DeletePlan(id uint) error {
	return r.db.Delete(&domain.AllocationPlan{}, id).Error
}

// GetPlansByDateRange retrieves plans created within a date range
func (r *AllocationPlanRepository) GetPlansByDateRange(userID uint, start, end time.Time) ([]domain.AllocationPlan, error) {
	var plans []domain.AllocationPlan
	err := r.db.Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, start, end).
		Preload("Allocations").
		Order("created_at DESC").
		Find(&plans).Error
	return plans, err
}

// GetPlanStats returns statistics about user's allocation plans
func (r *AllocationPlanRepository) GetPlanStats(userID uint) (map[string]interface{}, error) {
	type Stats struct {
		TotalPlans     int64
		ExecutedPlans  int64
		TotalBudget    float64
		AvgConfidence  float64
	}

	var stats Stats
	err := r.db.Model(&domain.AllocationPlan{}).
		Where("user_id = ?", userID).
		Select(`
			COUNT(*) as total_plans,
			SUM(CASE WHEN is_executed = true THEN 1 ELSE 0 END) as executed_plans,
			SUM(total_budget) as total_budget,
			AVG(confidence) as avg_confidence
		`).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_plans":    stats.TotalPlans,
		"executed_plans": stats.ExecutedPlans,
		"pending_plans":  stats.TotalPlans - stats.ExecutedPlans,
		"total_budget":   stats.TotalBudget,
		"avg_confidence": stats.AvgConfidence,
	}, nil
}
