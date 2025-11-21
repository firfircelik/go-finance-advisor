package coordinator

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"go-finance-advisor/internal/config"
	"go-finance-advisor/internal/infrastructure/market_data/binance"
	"go-finance-advisor/internal/infrastructure/market_data/yahoo"
	"go-finance-advisor/internal/infrastructure/persistence/timescale"
)

// DataSource represents a market data source
type DataSource string

const (
	SourceBinance DataSource = "binance"
	SourceYahoo   DataSource = "yahoo"
)

// IngestionStatus represents the status of a data source
type IngestionStatus struct {
	Source      DataSource
	IsRunning   bool
	SymbolCount int
	LastUpdate  time.Time
	ErrorCount  int64
	LastError   string
}

// IngestionCoordinator manages multiple market data fetchers
type IngestionCoordinator struct {
	binanceFetcher *binance.CryptoFetcher
	yahooFetcher   *yahoo.StockFetcher
	marketDataRepo *timescale.MarketDataRepository
	indicatorRepo  *timescale.IndicatorRepository
	log            *slog.Logger
	cfg            *config.Config
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	mu             sync.RWMutex
	status         map[DataSource]*IngestionStatus
}

// NewIngestionCoordinator creates a new data ingestion coordinator
func NewIngestionCoordinator(
	cfg *config.Config,
	marketDataRepo *timescale.MarketDataRepository,
	indicatorRepo *timescale.IndicatorRepository,
	log *slog.Logger,
) *IngestionCoordinator {
	ctx, cancel := context.WithCancel(context.Background())

	coordinator := &IngestionCoordinator{
		marketDataRepo: marketDataRepo,
		indicatorRepo:  indicatorRepo,
		log:            log,
		cfg:            cfg,
		ctx:            ctx,
		cancel:         cancel,
		status: map[DataSource]*IngestionStatus{
			SourceBinance: {
				Source:    SourceBinance,
				IsRunning: false,
			},
			SourceYahoo: {
				Source:    SourceYahoo,
				IsRunning: false,
			},
		},
	}

	// Initialize Binance fetcher
	coordinator.binanceFetcher = binance.NewCryptoFetcher(
		&cfg.Binance,
		marketDataRepo,
		log.With("source", "binance"),
	)

	// Initialize Yahoo fetcher
	coordinator.yahooFetcher = yahoo.NewStockFetcher(
		&cfg.Yahoo,
		marketDataRepo,
		log.With("source", "yahoo"),
	)

	return coordinator
}

// Start starts all configured data fetchers
func (c *IngestionCoordinator) Start() error {
	c.log.Info("Starting data ingestion coordinator")

	var errs []error

	// Start Binance fetcher
	if len(c.cfg.Binance.WatchSymbols) > 0 {
		if err := c.startBinance(); err != nil {
			c.log.Error("Failed to start Binance fetcher", "error", err)
			errs = append(errs, fmt.Errorf("binance: %w", err))
		}
	} else {
		c.log.Info("Binance fetcher disabled (no symbols configured)")
	}

	// Start Yahoo fetcher
	if c.cfg.Yahoo.Enabled && len(c.cfg.Yahoo.WatchSymbols) > 0 {
		if err := c.startYahoo(); err != nil {
			c.log.Error("Failed to start Yahoo fetcher", "error", err)
			errs = append(errs, fmt.Errorf("yahoo: %w", err))
		}
	} else {
		c.log.Info("Yahoo fetcher disabled")
	}

	// Start monitoring routine
	c.wg.Add(1)
	go c.monitoringRoutine()

	if len(errs) == 2 {
		return fmt.Errorf("all data sources failed to start")
	}

	c.log.Info("Data ingestion coordinator started successfully")
	return nil
}

// startBinance starts the Binance fetcher
func (c *IngestionCoordinator) startBinance() error {
	c.log.Info("Starting Binance fetcher")

	if err := c.binanceFetcher.Start(); err != nil {
		c.updateStatus(SourceBinance, false, err.Error())
		return err
	}

	c.updateStatus(SourceBinance, true, "")
	c.log.Info("Binance fetcher started", "symbols", len(c.cfg.Binance.WatchSymbols))
	return nil
}

// startYahoo starts the Yahoo fetcher
func (c *IngestionCoordinator) startYahoo() error {
	c.log.Info("Starting Yahoo fetcher")

	if err := c.yahooFetcher.Start(); err != nil {
		c.updateStatus(SourceYahoo, false, err.Error())
		return err
	}

	c.updateStatus(SourceYahoo, true, "")
	c.log.Info("Yahoo fetcher started", "symbols", len(c.cfg.Yahoo.WatchSymbols))
	return nil
}

// monitoringRoutine monitors the health of data sources
func (c *IngestionCoordinator) monitoringRoutine() {
	defer c.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.checkHealth()
		}
	}
}

// checkHealth checks the health of all data sources
func (c *IngestionCoordinator) checkHealth() {
	// Check Binance
	if c.binanceFetcher != nil {
		isRunning := c.binanceFetcher.IsRunning()
		c.mu.Lock()
		if status, ok := c.status[SourceBinance]; ok {
			status.IsRunning = isRunning
			status.SymbolCount = len(c.binanceFetcher.GetSymbols())
			status.LastUpdate = time.Now()

			if !isRunning && status.LastError == "" {
				c.log.Warn("Binance fetcher is not running")
			}
		}
		c.mu.Unlock()
	}

	// Check Yahoo
	if c.yahooFetcher != nil {
		c.mu.Lock()
		if status, ok := c.status[SourceYahoo]; ok {
			status.IsRunning = true // Yahoo doesn't have connection status
			status.SymbolCount = len(c.yahooFetcher.GetSymbols())
			status.LastUpdate = time.Now()
		}
		c.mu.Unlock()
	}
}

// updateStatus updates the status of a data source
func (c *IngestionCoordinator) updateStatus(source DataSource, isRunning bool, errorMsg string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if status, ok := c.status[source]; ok {
		status.IsRunning = isRunning
		status.LastUpdate = time.Now()

		if errorMsg != "" {
			status.ErrorCount++
			status.LastError = errorMsg
		} else {
			status.LastError = ""
		}
	}
}

// Stop stops all data fetchers
func (c *IngestionCoordinator) Stop() error {
	c.log.Info("Stopping data ingestion coordinator")

	// Cancel context to stop monitoring
	c.cancel()

	var wg sync.WaitGroup

	// Stop Binance fetcher
	if c.binanceFetcher != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.binanceFetcher.Stop(); err != nil {
				c.log.Error("Failed to stop Binance fetcher", "error", err)
			}
			c.updateStatus(SourceBinance, false, "")
		}()
	}

	// Stop Yahoo fetcher
	if c.yahooFetcher != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.yahooFetcher.Stop(); err != nil {
				c.log.Error("Failed to stop Yahoo fetcher", "error", err)
			}
			c.updateStatus(SourceYahoo, false, "")
		}()
	}

	// Wait for all fetchers to stop
	wg.Wait()

	// Wait for monitoring routine to finish
	c.wg.Wait()

	c.log.Info("Data ingestion coordinator stopped")
	return nil
}

// GetStatus returns the current status of all data sources
func (c *IngestionCoordinator) GetStatus() map[DataSource]*IngestionStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Create a copy to avoid race conditions
	statusCopy := make(map[DataSource]*IngestionStatus)
	for source, status := range c.status {
		statusCopy[source] = &IngestionStatus{
			Source:      status.Source,
			IsRunning:   status.IsRunning,
			SymbolCount: status.SymbolCount,
			LastUpdate:  status.LastUpdate,
			ErrorCount:  status.ErrorCount,
			LastError:   status.LastError,
		}
	}

	return statusCopy
}

// AddSymbol adds a symbol to the appropriate data source
func (c *IngestionCoordinator) AddSymbol(symbol, assetType string) error {
	switch assetType {
	case "crypto":
		if c.binanceFetcher == nil {
			return fmt.Errorf("binance fetcher not initialized")
		}
		return c.binanceFetcher.AddSymbols([]string{symbol})

	case "stock":
		if c.yahooFetcher == nil {
			return fmt.Errorf("yahoo fetcher not initialized")
		}
		c.yahooFetcher.AddSymbols([]string{symbol})
		return nil

	default:
		return fmt.Errorf("unsupported asset type: %s", assetType)
	}
}

// RemoveSymbol removes a symbol from the appropriate data source
func (c *IngestionCoordinator) RemoveSymbol(symbol, assetType string) error {
	switch assetType {
	case "crypto":
		if c.binanceFetcher == nil {
			return fmt.Errorf("binance fetcher not initialized")
		}
		c.binanceFetcher.RemoveSymbols([]string{symbol})
		return nil

	case "stock":
		if c.yahooFetcher == nil {
			return fmt.Errorf("yahoo fetcher not initialized")
		}
		c.yahooFetcher.RemoveSymbols([]string{symbol})
		return nil

	default:
		return fmt.Errorf("unsupported asset type: %s", assetType)
	}
}

// FetchHistoricalStock fetches historical stock data
func (c *IngestionCoordinator) FetchHistoricalStock(symbol string, start, end time.Time) error {
	if c.yahooFetcher == nil {
		return fmt.Errorf("yahoo fetcher not initialized")
	}

	return c.yahooFetcher.FetchHistoricalAndStore(symbol, start, end)
}

// GetAllSymbols returns all symbols being tracked
func (c *IngestionCoordinator) GetAllSymbols() map[string][]string {
	result := make(map[string][]string)

	if c.binanceFetcher != nil {
		result["crypto"] = c.binanceFetcher.GetSymbols()
	}

	if c.yahooFetcher != nil {
		result["stock"] = c.yahooFetcher.GetSymbols()
	}

	return result
}

// IsHealthy returns true if at least one data source is running
func (c *IngestionCoordinator) IsHealthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, status := range c.status {
		if status.IsRunning {
			return true
		}
	}

	return false
}

// RestartSource restarts a specific data source
func (c *IngestionCoordinator) RestartSource(source DataSource) error {
	c.log.Info("Restarting data source", "source", source)

	switch source {
	case SourceBinance:
		if c.binanceFetcher != nil {
			if err := c.binanceFetcher.Stop(); err != nil {
				c.log.Error("Failed to stop Binance fetcher", "error", err)
			}
			time.Sleep(1 * time.Second)
			return c.startBinance()
		}
		return fmt.Errorf("binance fetcher not initialized")

	case SourceYahoo:
		if c.yahooFetcher != nil {
			if err := c.yahooFetcher.Stop(); err != nil {
				c.log.Error("Failed to stop Yahoo fetcher", "error", err)
			}
			time.Sleep(1 * time.Second)
			return c.startYahoo()
		}
		return fmt.Errorf("yahoo fetcher not initialized")

	default:
		return fmt.Errorf("unknown data source: %s", source)
	}
}
