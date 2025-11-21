package postgres

import (
	"time"

	"go-finance-advisor/internal/domain"

	"gorm.io/gorm"
)

// SignalRepository handles trading signal operations
type SignalRepository struct {
	db *gorm.DB
}

// NewSignalRepository creates a new signal repository
func NewSignalRepository(db *gorm.DB) *SignalRepository {
	return &SignalRepository{db: db}
}

// Create creates a new signal
func (r *SignalRepository) Create(signal *domain.Signal) error {
	return r.db.Create(signal).Error
}

// GetByID retrieves a signal by ID
func (r *SignalRepository) GetByID(id uint) (*domain.Signal, error) {
	var signal domain.Signal
	err := r.db.First(&signal, id).Error
	if err != nil {
		return nil, err
	}
	return &signal, nil
}

// GetActiveSignals retrieves all active signals
func (r *SignalRepository) GetActiveSignals() ([]domain.Signal, error) {
	var signals []domain.Signal
	err := r.db.Where("is_active = ? AND valid_until > ?", true, time.Now()).
		Order("confidence DESC").
		Find(&signals).Error
	return signals, err
}

// GetActiveSignalsBySymbol retrieves active signals for a specific symbol
func (r *SignalRepository) GetActiveSignalsBySymbol(symbol string) ([]domain.Signal, error) {
	var signals []domain.Signal
	err := r.db.Where("symbol = ? AND is_active = ? AND valid_until > ?", symbol, true, time.Now()).
		Order("created_at DESC").
		Find(&signals).Error
	return signals, err
}

// GetSignalsByAssetType retrieves active signals filtered by asset type
func (r *SignalRepository) GetSignalsByAssetType(assetType string) ([]domain.Signal, error) {
	var signals []domain.Signal
	err := r.db.Where("asset_type = ? AND is_active = ? AND valid_until > ?", assetType, true, time.Now()).
		Order("confidence DESC").
		Find(&signals).Error
	return signals, err
}

// GetBuySignals retrieves all active BUY signals
func (r *SignalRepository) GetBuySignals() ([]domain.Signal, error) {
	var signals []domain.Signal
	err := r.db.Where("action = ? AND is_active = ? AND valid_until > ?", "BUY", true, time.Now()).
		Order("confidence DESC").
		Find(&signals).Error
	return signals, err
}

// DeactivateSignal marks a signal as inactive
func (r *SignalRepository) DeactivateSignal(id uint) error {
	return r.db.Model(&domain.Signal{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

// DeactivateExpiredSignals deactivates all expired signals
func (r *SignalRepository) DeactivateExpiredSignals() error {
	return r.db.Model(&domain.Signal{}).
		Where("is_active = ? AND valid_until <= ?", true, time.Now()).
		Update("is_active", false).Error
}

// GetRecentSignals retrieves signals from the last N days
func (r *SignalRepository) GetRecentSignals(days int) ([]domain.Signal, error) {
	var signals []domain.Signal
	since := time.Now().AddDate(0, 0, -days)
	err := r.db.Where("created_at >= ?", since).
		Order("created_at DESC").
		Find(&signals).Error
	return signals, err
}

// CountActiveSignalsByAction counts signals by action type
func (r *SignalRepository) CountActiveSignalsByAction() (map[string]int64, error) {
	type ActionCount struct {
		Action string
		Count  int64
	}

	var results []ActionCount
	err := r.db.Model(&domain.Signal{}).
		Select("action, COUNT(*) as count").
		Where("is_active = ? AND valid_until > ?", true, time.Now()).
		Group("action").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, result := range results {
		counts[result.Action] = result.Count
	}

	return counts, nil
}
