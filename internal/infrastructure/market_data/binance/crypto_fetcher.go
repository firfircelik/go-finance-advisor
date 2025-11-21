package binance

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-finance-advisor/internal/config"
	"go-finance-advisor/internal/infrastructure/persistence/timescale"
)

// CryptoFetcher fetches cryptocurrency data from Binance WebSocket
type CryptoFetcher struct {
	wsClient      *WebSocketClient
	marketDataRepo *timescale.MarketDataRepository
	log           *slog.Logger
	cfg           *config.BinanceConfig
	symbols       []string
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	dataBuffer    chan *timescale.MarketDataPoint
	bufferSize    int
}

// NewCryptoFetcher creates a new cryptocurrency data fetcher
func NewCryptoFetcher(
	cfg *config.BinanceConfig,
	marketDataRepo *timescale.MarketDataRepository,
	log *slog.Logger,
) *CryptoFetcher {
	ctx, cancel := context.WithCancel(context.Background())

	fetcher := &CryptoFetcher{
		wsClient:       NewWebSocketClient(log),
		marketDataRepo: marketDataRepo,
		log:            log,
		cfg:            cfg,
		symbols:        cfg.WatchSymbols,
		ctx:            ctx,
		cancel:         cancel,
		dataBuffer:     make(chan *timescale.MarketDataPoint, 1000),
		bufferSize:     100,
	}

	// Set up data handlers
	fetcher.wsClient.OnTicker(fetcher.handleTicker)
	fetcher.wsClient.OnTrade(fetcher.handleTrade)
	fetcher.wsClient.OnKline(fetcher.handleKline)

	return fetcher
}

// Start starts the cryptocurrency data fetcher
func (f *CryptoFetcher) Start() error {
	if len(f.symbols) == 0 {
		return fmt.Errorf("no symbols configured for Binance fetcher")
	}

	f.log.Info("Starting Binance crypto fetcher", "symbols", strings.Join(f.symbols, ", "))

	// Connect to WebSocket
	if err := f.wsClient.Connect(); err != nil {
		return fmt.Errorf("failed to connect to Binance WebSocket: %w", err)
	}

	// Subscribe to ticker streams for all symbols
	if err := f.wsClient.SubscribeTicker(f.symbols); err != nil {
		return fmt.Errorf("failed to subscribe to tickers: %w", err)
	}

	// Optionally subscribe to trade streams for more granular data
	if f.cfg.EnableTrades {
		if err := f.wsClient.SubscribeTrades(f.symbols); err != nil {
			f.log.Error("Failed to subscribe to trades", "error", err)
		}
	}

	// Optionally subscribe to klines for candlestick data
	if f.cfg.EnableKlines {
		interval := f.cfg.KlineInterval
		if interval == "" {
			interval = "1m" // Default to 1-minute candles
		}
		if err := f.wsClient.SubscribeKlines(f.symbols, interval); err != nil {
			f.log.Error("Failed to subscribe to klines", "error", err)
		}
	}

	// Start batch writer
	f.wg.Add(1)
	go f.batchWriter()

	f.log.Info("Binance crypto fetcher started successfully")
	return nil
}

// handleTicker processes ticker data from WebSocket
func (f *CryptoFetcher) handleTicker(data interface{}) {
	ticker, ok := data.(TickerData)
	if !ok {
		f.log.Error("Invalid ticker data type")
		return
	}

	// Parse ticker data
	price, err := strconv.ParseFloat(ticker.LastPrice, 64)
	if err != nil {
		f.log.Error("Failed to parse price", "symbol", ticker.Symbol, "error", err)
		return
	}

	volume, err := strconv.ParseFloat(ticker.Volume, 64)
	if err != nil {
		f.log.Error("Failed to parse volume", "symbol", ticker.Symbol, "error", err)
		return
	}

	high, err := strconv.ParseFloat(ticker.HighPrice, 64)
	if err != nil {
		f.log.Error("Failed to parse high price", "symbol", ticker.Symbol, "error", err)
		return
	}

	low, err := strconv.ParseFloat(ticker.LowPrice, 64)
	if err != nil {
		f.log.Error("Failed to parse low price", "symbol", ticker.Symbol, "error", err)
		return
	}

	changePercent, err := strconv.ParseFloat(ticker.PriceChangePercent, 64)
	if err != nil {
		f.log.Error("Failed to parse change percent", "symbol", ticker.Symbol, "error", err)
		return
	}

	// Create market data point
	dataPoint := &timescale.MarketDataPoint{
		Time:      ticker.Timestamp,
		Symbol:    ticker.Symbol,
		AssetType: "crypto",
		Price:     price,
		Volume:    volume,
		High24h:   high,
		Low24h:    low,
		Change24h: changePercent,
		Source:    "binance",
	}

	// Send to buffer
	select {
	case f.dataBuffer <- dataPoint:
	case <-f.ctx.Done():
		return
	default:
		f.log.Warn("Data buffer full, dropping data point", "symbol", ticker.Symbol)
	}
}

// handleTrade processes trade data from WebSocket
func (f *CryptoFetcher) handleTrade(data interface{}) {
	trade, ok := data.(TradeData)
	if !ok {
		f.log.Error("Invalid trade data type")
		return
	}

	// Create market data point from trade
	dataPoint := &timescale.MarketDataPoint{
		Time:      time.UnixMilli(trade.TradeTime),
		Symbol:    trade.Symbol,
		AssetType: "crypto",
		Price:     trade.Price,
		Volume:    trade.Quantity,
		Source:    "binance_trade",
	}

	// Send to buffer
	select {
	case f.dataBuffer <- dataPoint:
	case <-f.ctx.Done():
		return
	default:
		f.log.Warn("Data buffer full, dropping trade data", "symbol", trade.Symbol)
	}
}

// handleKline processes kline/candlestick data from WebSocket
func (f *CryptoFetcher) handleKline(data interface{}) {
	kline, ok := data.(KlineData)
	if !ok {
		f.log.Error("Invalid kline data type")
		return
	}

	// Only process closed candles to avoid duplicates
	if !kline.Kline.IsClosed {
		return
	}

	// Create market data point from kline
	dataPoint := &timescale.MarketDataPoint{
		Time:      time.UnixMilli(kline.Kline.StartTime),
		Symbol:    kline.Symbol,
		AssetType: "crypto",
		Price:     kline.Kline.Close,
		Volume:    kline.Kline.Volume,
		High24h:   kline.Kline.High,
		Low24h:    kline.Kline.Low,
		Source:    fmt.Sprintf("binance_kline_%s", kline.Kline.Interval),
	}

	// Send to buffer
	select {
	case f.dataBuffer <- dataPoint:
	case <-f.ctx.Done():
		return
	default:
		f.log.Warn("Data buffer full, dropping kline data", "symbol", kline.Symbol)
	}
}

// batchWriter batches data points and writes them to TimescaleDB
func (f *CryptoFetcher) batchWriter() {
	defer f.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var batch []*timescale.MarketDataPoint

	flush := func() {
		if len(batch) == 0 {
			return
		}

		if err := f.marketDataRepo.InsertBatch(batch); err != nil {
			f.log.Error("Failed to insert batch", "count", len(batch), "error", err)
		} else {
			f.log.Debug("Batch inserted successfully", "count", len(batch))
		}

		batch = batch[:0] // Clear batch
	}

	for {
		select {
		case <-f.ctx.Done():
			// Flush remaining data
			flush()
			return

		case dataPoint := <-f.dataBuffer:
			batch = append(batch, dataPoint)

			// Flush if batch is full
			if len(batch) >= f.bufferSize {
				flush()
			}

		case <-ticker.C:
			// Periodic flush
			flush()
		}
	}
}

// Stop stops the cryptocurrency data fetcher
func (f *CryptoFetcher) Stop() error {
	f.log.Info("Stopping Binance crypto fetcher")

	// Cancel context to stop all goroutines
	f.cancel()

	// Close WebSocket connection
	if err := f.wsClient.Close(); err != nil {
		f.log.Error("Failed to close WebSocket", "error", err)
	}

	// Wait for batch writer to finish
	f.wg.Wait()

	f.log.Info("Binance crypto fetcher stopped")
	return nil
}

// AddSymbols adds new symbols to watch
func (f *CryptoFetcher) AddSymbols(symbols []string) error {
	f.symbols = append(f.symbols, symbols...)

	if f.wsClient.IsConnected() {
		if err := f.wsClient.SubscribeTicker(symbols); err != nil {
			return fmt.Errorf("failed to subscribe to new symbols: %w", err)
		}
	}

	f.log.Info("Added new symbols", "symbols", strings.Join(symbols, ", "))
	return nil
}

// RemoveSymbols removes symbols from watch list
func (f *CryptoFetcher) RemoveSymbols(symbols []string) {
	removeMap := make(map[string]bool)
	for _, s := range symbols {
		removeMap[s] = true
	}

	newSymbols := make([]string, 0)
	for _, s := range f.symbols {
		if !removeMap[s] {
			newSymbols = append(newSymbols, s)
		}
	}

	f.symbols = newSymbols
	f.log.Info("Removed symbols", "symbols", strings.Join(symbols, ", "))
}

// GetSymbols returns the list of watched symbols
func (f *CryptoFetcher) GetSymbols() []string {
	return f.symbols
}

// IsRunning returns true if the fetcher is connected and running
func (f *CryptoFetcher) IsRunning() bool {
	return f.wsClient.IsConnected()
}
