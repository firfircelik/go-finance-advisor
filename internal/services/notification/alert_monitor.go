package notification

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"go-finance-advisor/internal/domain"
	"go-finance-advisor/internal/infrastructure/persistence/postgres"
	"go-finance-advisor/internal/infrastructure/persistence/timescale"
	ws "go-finance-advisor/internal/infrastructure/websocket"
)

// AlertMonitor monitors price alerts and triggers notifications
type AlertMonitor struct {
	priceAlertRepo *postgres.PriceAlertRepository
	marketDataRepo *timescale.MarketDataRepository
	emailService   *EmailService
	wsHub          *ws.Hub
	log            *slog.Logger
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	checkInterval  time.Duration
}

// NewAlertMonitor creates a new alert monitor
func NewAlertMonitor(
	priceAlertRepo *postgres.PriceAlertRepository,
	marketDataRepo *timescale.MarketDataRepository,
	emailService   *EmailService,
	wsHub          *ws.Hub,
	log *slog.Logger,
) *AlertMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &AlertMonitor{
		priceAlertRepo: priceAlertRepo,
		marketDataRepo: marketDataRepo,
		emailService:   emailService,
		wsHub:          wsHub,
		log:            log,
		ctx:            ctx,
		cancel:         cancel,
		checkInterval:  30 * time.Second, // Check every 30 seconds
	}
}

// Start starts the alert monitoring service
func (m *AlertMonitor) Start() error {
	m.log.Info("Starting alert monitor", "interval", m.checkInterval)

	// Start monitoring routine
	m.wg.Add(1)
	go m.monitoringRoutine()

	m.log.Info("Alert monitor started successfully")
	return nil
}

// monitoringRoutine periodically checks alerts
func (m *AlertMonitor) monitoringRoutine() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	// Run immediately on start
	m.checkAllAlerts()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkAllAlerts()
		}
	}
}

// checkAllAlerts checks all active price alerts
func (m *AlertMonitor) checkAllAlerts() {
	alerts, err := m.priceAlertRepo.GetActiveAlerts()
	if err != nil {
		m.log.Error("Failed to get active alerts", "error", err)
		return
	}

	if len(alerts) == 0 {
		return
	}

	m.log.Debug("Checking alerts", "count", len(alerts))

	// Process alerts concurrently
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // Limit concurrent checks

	for _, alert := range alerts {
		wg.Add(1)
		go func(a domain.PriceAlert) {
			defer wg.Done()

			semaphore <- struct{}{} // Acquire
			defer func() { <-semaphore }() // Release

			if err := m.checkAlert(&a); err != nil {
				m.log.Error("Failed to check alert",
					"alert_id", a.ID,
					"symbol", a.Symbol,
					"error", err)
			}
		}(alert)
	}

	wg.Wait()
}

// checkAlert checks a single price alert
func (m *AlertMonitor) checkAlert(alert *domain.PriceAlert) error {
	// Get current price
	currentPrice, err := m.marketDataRepo.GetLatestPrice(alert.Symbol)
	if err != nil {
		return fmt.Errorf("failed to get price: %w", err)
	}

	// Check if alert condition is met
	triggered := false

	switch alert.Condition {
	case "above":
		triggered = currentPrice >= alert.TargetPrice
	case "below":
		triggered = currentPrice <= alert.TargetPrice
	case "equals":
		// Allow 0.1% tolerance for equality
		tolerance := alert.TargetPrice * 0.001
		triggered = (currentPrice >= alert.TargetPrice-tolerance) &&
			(currentPrice <= alert.TargetPrice+tolerance)
	default:
		return fmt.Errorf("unknown condition: %s", alert.Condition)
	}

	if !triggered {
		return nil
	}

	// Alert triggered!
	m.log.Info("Alert triggered",
		"alert_id", alert.ID,
		"symbol", alert.Symbol,
		"target", alert.TargetPrice,
		"current", currentPrice)

	// Update alert
	alert.CurrentPrice = currentPrice
	alert.IsTriggered = true
	alert.TriggeredAt = time.Now()

	if err := m.priceAlertRepo.TriggerAlert(alert.ID, currentPrice); err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}

	// Send notifications
	if err := m.sendNotifications(alert); err != nil {
		m.log.Error("Failed to send notifications",
			"alert_id", alert.ID,
			"error", err)
	}

	return nil
}

// sendNotifications sends alert notifications via multiple channels
func (m *AlertMonitor) sendNotifications(alert *domain.PriceAlert) error {
	// Prepare notification data
	notificationData := map[string]interface{}{
		"Symbol":       alert.Symbol,
		"AlertType":    alert.AlertType,
		"TargetPrice":  alert.TargetPrice,
		"CurrentPrice": alert.CurrentPrice,
		"Condition":    alert.Condition,
		"Message":      alert.Message,
	}

	// Send WebSocket notification
	if m.wsHub != nil {
		m.wsHub.BroadcastToUser(alert.UserID, ws.MessageTypeAlert, notificationData)
		m.log.Debug("WebSocket notification sent", "user_id", alert.UserID)
	}

	// Send email notification (if email service is enabled)
	if m.emailService != nil && m.emailService.IsEnabled() {
		// TODO: Get user email from user repository
		// For now, we'll just mark notification as sent
		if err := m.priceAlertRepo.MarkNotificationSent(alert.ID); err != nil {
			m.log.Error("Failed to mark notification as sent", "error", err)
		}
	}

	return nil
}

// Stop stops the alert monitoring service
func (m *AlertMonitor) Stop() error {
	m.log.Info("Stopping alert monitor")

	m.cancel()
	m.wg.Wait()

	m.log.Info("Alert monitor stopped")
	return nil
}

// CheckAlertNow immediately checks a specific alert
func (m *AlertMonitor) CheckAlertNow(alertID uint) error {
	alert, err := m.priceAlertRepo.GetByID(alertID)
	if err != nil {
		return fmt.Errorf("failed to get alert: %w", err)
	}

	if !alert.IsActive {
		return fmt.Errorf("alert is not active")
	}

	return m.checkAlert(alert)
}

// MonitorSymbol starts monitoring a new symbol
func (m *AlertMonitor) MonitorSymbol(symbol string) error {
	m.log.Info("Adding symbol to monitoring", "symbol", symbol)

	// Get all alerts for this symbol
	alerts, err := m.priceAlertRepo.GetBySymbol(symbol)
	if err != nil {
		return fmt.Errorf("failed to get alerts for symbol: %w", err)
	}

	m.log.Info("Found alerts for symbol", "symbol", symbol, "count", len(alerts))

	// Check each alert immediately
	for _, alert := range alerts {
		if alert.IsActive && !alert.IsTriggered {
			if err := m.checkAlert(&alert); err != nil {
				m.log.Error("Failed to check alert", "alert_id", alert.ID, "error", err)
			}
		}
	}

	return nil
}

// GetAlertStats returns statistics about alert monitoring
func (m *AlertMonitor) GetAlertStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get total active alerts
	activeAlerts, err := m.priceAlertRepo.GetActiveAlerts()
	if err != nil {
		return nil, fmt.Errorf("failed to get active alerts: %w", err)
	}

	stats["active_alerts"] = len(activeAlerts)

	// Get triggered alerts count (last 24 hours)
	// TODO: Add method to repository to get triggered alerts count

	return stats, nil
}
