package postgres

import (
	"time"

	"go-finance-advisor/internal/domain"

	"gorm.io/gorm"
)

// PriceAlertRepository handles price alert operations
type PriceAlertRepository struct {
	db *gorm.DB
}

// NewPriceAlertRepository creates a new price alert repository
func NewPriceAlertRepository(db *gorm.DB) *PriceAlertRepository {
	return &PriceAlertRepository{db: db}
}

// Create creates a new price alert
func (r *PriceAlertRepository) Create(alert *domain.PriceAlert) error {
	return r.db.Create(alert).Error
}

// GetByID retrieves an alert by ID
func (r *PriceAlertRepository) GetByID(id uint) (*domain.PriceAlert, error) {
	var alert domain.PriceAlert
	err := r.db.First(&alert, id).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

// GetUserAlerts retrieves all alerts for a user
func (r *PriceAlertRepository) GetUserAlerts(userID uint) ([]domain.PriceAlert, error) {
	var alerts []domain.PriceAlert
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&alerts).Error
	return alerts, err
}

// GetActiveAlerts retrieves all non-triggered alerts
func (r *PriceAlertRepository) GetActiveAlerts() ([]domain.PriceAlert, error) {
	var alerts []domain.PriceAlert
	err := r.db.Where("is_triggered = ?", false).
		Find(&alerts).Error
	return alerts, err
}

// GetActiveAlertsForSymbol retrieves active alerts for a specific symbol
func (r *PriceAlertRepository) GetActiveAlertsForSymbol(symbol string) ([]domain.PriceAlert, error) {
	var alerts []domain.PriceAlert
	err := r.db.Where("symbol = ? AND is_triggered = ?", symbol, false).
		Find(&alerts).Error
	return alerts, err
}

// MarkAsTriggered marks an alert as triggered
func (r *PriceAlertRepository) MarkAsTriggered(id uint) error {
	now := time.Now()
	return r.db.Model(&domain.PriceAlert{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_triggered": true,
			"triggered_at": &now,
		}).Error
}

// Delete deletes an alert
func (r *PriceAlertRepository) Delete(id uint) error {
	return r.db.Delete(&domain.PriceAlert{}, id).Error
}

// DeleteUserAlert deletes an alert if it belongs to the user
func (r *PriceAlertRepository) DeleteUserAlert(id, userID uint) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).
		Delete(&domain.PriceAlert{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// CheckAndTriggerAlerts checks current price against alerts and triggers them
func (r *PriceAlertRepository) CheckAndTriggerAlerts(symbol string, currentPrice float64) ([]domain.PriceAlert, error) {
	var triggered []domain.PriceAlert

	// Get all active alerts for this symbol
	alerts, err := r.GetActiveAlertsForSymbol(symbol)
	if err != nil {
		return nil, err
	}

	for _, alert := range alerts {
		shouldTrigger := false

		switch alert.Condition {
		case "above":
			if currentPrice >= alert.TargetPrice {
				shouldTrigger = true
			}
		case "below":
			if currentPrice <= alert.TargetPrice {
				shouldTrigger = true
			}
		}

		if shouldTrigger {
			if err := r.MarkAsTriggered(alert.ID); err == nil {
				alert.IsTriggered = true
				now := time.Now()
				alert.TriggeredAt = &now
				triggered = append(triggered, alert)
			}
		}
	}

	return triggered, nil
}

// GetAlertStats returns statistics about user's alerts
func (r *PriceAlertRepository) GetAlertStats(userID uint) (map[string]int64, error) {
	type Stats struct {
		Total     int64
		Active    int64
		Triggered int64
	}

	var stats Stats
	err := r.db.Model(&domain.PriceAlert{}).
		Where("user_id = ?", userID).
		Select(`
			COUNT(*) as total,
			SUM(CASE WHEN is_triggered = false THEN 1 ELSE 0 END) as active,
			SUM(CASE WHEN is_triggered = true THEN 1 ELSE 0 END) as triggered
		`).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	return map[string]int64{
		"total":     stats.Total,
		"active":    stats.Active,
		"triggered": stats.Triggered,
	}, nil
}
