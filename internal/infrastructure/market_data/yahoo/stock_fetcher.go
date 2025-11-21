package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-finance-advisor/internal/config"
	"go-finance-advisor/internal/infrastructure/persistence/timescale"
)

const (
	yahooAPIEndpoint = "https://query1.finance.yahoo.com/v8/finance/chart/"
	userAgent        = "Mozilla/5.0 (compatible; GoFinVisor/1.0)"
	requestTimeout   = 10 * time.Second
)

// Quote represents a Yahoo Finance quote
type Quote struct {
	Symbol             string    `json:"symbol"`
	RegularMarketPrice float64   `json:"regularMarketPrice"`
	RegularMarketOpen  float64   `json:"regularMarketOpen"`
	RegularMarketHigh  float64   `json:"regularMarketDayHigh"`
	RegularMarketLow   float64   `json:"regularMarketDayLow"`
	RegularMarketVolume float64  `json:"regularMarketVolume"`
	MarketCap          float64   `json:"marketCap"`
	PriceChange        float64   `json:"regularMarketChange"`
	PriceChangePercent float64   `json:"regularMarketChangePercent"`
	Timestamp          time.Time `json:"-"`
}

// YahooResponse represents the API response structure
type YahooResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Symbol             string  `json:"symbol"`
				RegularMarketPrice float64 `json:"regularMarketPrice"`
				ChartPreviousClose float64 `json:"chartPreviousClose"`
				PreviousClose      float64 `json:"previousClose"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []int64   `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
	} `json:"chart"`
}

// StockFetcher fetches stock data from Yahoo Finance
type StockFetcher struct {
	httpClient     *http.Client
	marketDataRepo *timescale.MarketDataRepository
	log            *slog.Logger
	cfg            *config.YahooConfig
	symbols        []string
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	ticker         *time.Ticker
	mu             sync.RWMutex
}

// NewStockFetcher creates a new Yahoo Finance stock fetcher
func NewStockFetcher(
	cfg *config.YahooConfig,
	marketDataRepo *timescale.MarketDataRepository,
	log *slog.Logger,
) *StockFetcher {
	ctx, cancel := context.WithCancel(context.Background())

	return &StockFetcher{
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
		marketDataRepo: marketDataRepo,
		log:            log,
		cfg:            cfg,
		symbols:        cfg.WatchSymbols,
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Start starts the stock data fetcher
func (f *StockFetcher) Start() error {
	if !f.cfg.Enabled {
		f.log.Info("Yahoo Finance fetcher is disabled")
		return nil
	}

	if len(f.symbols) == 0 {
		return fmt.Errorf("no symbols configured for Yahoo Finance fetcher")
	}

	f.log.Info("Starting Yahoo Finance stock fetcher",
		"symbols", strings.Join(f.symbols, ", "),
		"interval", f.cfg.FetchInterval)

	// Fetch immediately on start
	f.fetchAll()

	// Start periodic fetching
	f.ticker = time.NewTicker(f.cfg.FetchInterval)
	f.wg.Add(1)
	go f.fetchRoutine()

	f.log.Info("Yahoo Finance stock fetcher started successfully")
	return nil
}

// fetchRoutine periodically fetches stock data
func (f *StockFetcher) fetchRoutine() {
	defer f.wg.Done()

	for {
		select {
		case <-f.ctx.Done():
			return
		case <-f.ticker.C:
			f.fetchAll()
		}
	}
}

// fetchAll fetches data for all symbols
func (f *StockFetcher) fetchAll() {
	f.mu.RLock()
	symbols := make([]string, len(f.symbols))
	copy(symbols, f.symbols)
	f.mu.RUnlock()

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // Limit concurrent requests

	for _, symbol := range symbols {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()

			semaphore <- struct{}{} // Acquire
			defer func() { <-semaphore }() // Release

			if err := f.fetchSymbol(sym); err != nil {
				f.log.Error("Failed to fetch symbol", "symbol", sym, "error", err)
			}
		}(symbol)
	}

	wg.Wait()
}

// fetchSymbol fetches data for a single symbol
func (f *StockFetcher) fetchSymbol(symbol string) error {
	quote, err := f.getQuote(symbol)
	if err != nil {
		return fmt.Errorf("failed to get quote for %s: %w", symbol, err)
	}

	// Create market data point
	dataPoint := &timescale.MarketDataPoint{
		Time:      quote.Timestamp,
		Symbol:    quote.Symbol,
		AssetType: "stock",
		Price:     quote.RegularMarketPrice,
		Volume:    quote.RegularMarketVolume,
		High24h:   quote.RegularMarketHigh,
		Low24h:    quote.RegularMarketLow,
		Change24h: quote.PriceChangePercent,
		MarketCap: quote.MarketCap,
		Source:    "yahoo",
	}

	// Insert into TimescaleDB
	if err := f.marketDataRepo.Insert(dataPoint); err != nil {
		return fmt.Errorf("failed to insert data: %w", err)
	}

	f.log.Debug("Fetched stock data",
		"symbol", symbol,
		"price", quote.RegularMarketPrice,
		"volume", quote.RegularMarketVolume,
		"change", quote.PriceChangePercent)

	return nil
}

// getQuote fetches a quote from Yahoo Finance API
func (f *StockFetcher) getQuote(symbol string) (*Quote, error) {
	// Build URL
	apiURL := fmt.Sprintf("%s%s", yahooAPIEndpoint, url.PathEscape(symbol))
	params := url.Values{}
	params.Set("interval", "1d")
	params.Set("range", "1d")
	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	// Create request
	req, err := http.NewRequestWithContext(f.ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)

	// Execute request
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	var yahooResp YahooResponse
	if err := json.NewDecoder(resp.Body).Decode(&yahooResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for API error
	if yahooResp.Chart.Error != nil {
		return nil, fmt.Errorf("yahoo api error: %s - %s",
			yahooResp.Chart.Error.Code,
			yahooResp.Chart.Error.Description)
	}

	// Validate response
	if len(yahooResp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data in response")
	}

	result := yahooResp.Chart.Result[0]

	// Extract latest price data
	quote := &Quote{
		Symbol:             result.Meta.Symbol,
		RegularMarketPrice: result.Meta.RegularMarketPrice,
		Timestamp:          time.Now(),
	}

	// Extract OHLCV data if available
	if len(result.Indicators.Quote) > 0 {
		quoteData := result.Indicators.Quote[0]
		lastIdx := len(quoteData.Close) - 1

		if lastIdx >= 0 {
			quote.RegularMarketHigh = getMaxFloat(quoteData.High)
			quote.RegularMarketLow = getMinFloat(quoteData.Low)
			quote.RegularMarketVolume = float64(sumInt64(quoteData.Volume))

			if len(quoteData.Open) > 0 {
				quote.RegularMarketOpen = quoteData.Open[0]
			}

			// Calculate price change
			if result.Meta.ChartPreviousClose > 0 {
				quote.PriceChange = quote.RegularMarketPrice - result.Meta.ChartPreviousClose
				quote.PriceChangePercent = (quote.PriceChange / result.Meta.ChartPreviousClose) * 100
			}
		}
	}

	return quote, nil
}

// getHistoricalData fetches historical data for a symbol
func (f *StockFetcher) getHistoricalData(symbol string, start, end time.Time) ([]*timescale.MarketDataPoint, error) {
	// Build URL
	apiURL := fmt.Sprintf("%s%s", yahooAPIEndpoint, url.PathEscape(symbol))
	params := url.Values{}
	params.Set("period1", strconv.FormatInt(start.Unix(), 10))
	params.Set("period2", strconv.FormatInt(end.Unix(), 10))
	params.Set("interval", "1d")
	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	// Create request
	req, err := http.NewRequestWithContext(f.ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)

	// Execute request
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	var yahooResp YahooResponse
	if err := json.NewDecoder(resp.Body).Decode(&yahooResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for error
	if yahooResp.Chart.Error != nil {
		return nil, fmt.Errorf("yahoo api error: %s", yahooResp.Chart.Error.Description)
	}

	// Validate response
	if len(yahooResp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data in response")
	}

	result := yahooResp.Chart.Result[0]

	// Build data points
	var dataPoints []*timescale.MarketDataPoint

	if len(result.Indicators.Quote) > 0 {
		quoteData := result.Indicators.Quote[0]

		for i := range result.Timestamp {
			if i >= len(quoteData.Close) {
				break
			}

			dataPoint := &timescale.MarketDataPoint{
				Time:      time.Unix(result.Timestamp[i], 0),
				Symbol:    symbol,
				AssetType: "stock",
				Price:     quoteData.Close[i],
				High24h:   quoteData.High[i],
				Low24h:    quoteData.Low[i],
				Source:    "yahoo",
			}

			if i < len(quoteData.Volume) {
				dataPoint.Volume = float64(quoteData.Volume[i])
			}

			dataPoints = append(dataPoints, dataPoint)
		}
	}

	return dataPoints, nil
}

// FetchHistoricalAndStore fetches historical data and stores in database
func (f *StockFetcher) FetchHistoricalAndStore(symbol string, start, end time.Time) error {
	f.log.Info("Fetching historical data", "symbol", symbol, "start", start, "end", end)

	dataPoints, err := f.getHistoricalData(symbol, start, end)
	if err != nil {
		return fmt.Errorf("failed to get historical data: %w", err)
	}

	if len(dataPoints) == 0 {
		f.log.Warn("No historical data found", "symbol", symbol)
		return nil
	}

	// Insert batch
	if err := f.marketDataRepo.InsertBatch(dataPoints); err != nil {
		return fmt.Errorf("failed to insert batch: %w", err)
	}

	f.log.Info("Historical data stored", "symbol", symbol, "count", len(dataPoints))
	return nil
}

// Stop stops the stock data fetcher
func (f *StockFetcher) Stop() error {
	f.log.Info("Stopping Yahoo Finance stock fetcher")

	if f.ticker != nil {
		f.ticker.Stop()
	}

	f.cancel()
	f.wg.Wait()

	f.log.Info("Yahoo Finance stock fetcher stopped")
	return nil
}

// AddSymbols adds new symbols to watch
func (f *StockFetcher) AddSymbols(symbols []string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.symbols = append(f.symbols, symbols...)
	f.log.Info("Added new symbols", "symbols", strings.Join(symbols, ", "))
}

// RemoveSymbols removes symbols from watch list
func (f *StockFetcher) RemoveSymbols(symbols []string) {
	f.mu.Lock()
	defer f.mu.Unlock()

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
func (f *StockFetcher) GetSymbols() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	symbols := make([]string, len(f.symbols))
	copy(symbols, f.symbols)
	return symbols
}

// Helper functions

func getMaxFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

func getMinFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values {
		if v > 0 && v < min {
			min = v
		}
	}
	return min
}

func sumInt64(values []int64) int64 {
	sum := int64(0)
	for _, v := range values {
		sum += v
	}
	return sum
}
