package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"go-finance-advisor/internal/domain"
	"go-finance-advisor/internal/infrastructure/persistence/postgres"
	"go-finance-advisor/internal/infrastructure/persistence/timescale"
	"go-finance-advisor/internal/services/analysis/indicators"
)

// SignalEngine generates trading signals based on technical analysis
type SignalEngine struct {
	marketDataRepo *timescale.MarketDataRepository
	indicatorRepo  *timescale.IndicatorRepository
	signalRepo     *postgres.SignalRepository
	log            *slog.Logger
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	symbols        []string
	mu             sync.RWMutex
}

// IndicatorValues holds calculated indicator values
type IndicatorValues struct {
	RSI14         float64
	RSI7          float64
	SMA20         float64
	SMA50         float64
	SMA200        float64
	EMA12         float64
	EMA26         float64
	MACD          float64
	MACDSignal    float64
	MACDHistogram float64
	BBUpper       float64
	BBMiddle      float64
	BBLower       float64
	ADX           float64
}

// SignalScore holds the trading signal score
type SignalScore struct {
	Action     string
	Confidence float64
	Indicators map[string]string
	Reasoning  []string
}

// NewSignalEngine creates a new signal generation engine
func NewSignalEngine(
	marketDataRepo *timescale.MarketDataRepository,
	indicatorRepo *timescale.IndicatorRepository,
	signalRepo *postgres.SignalRepository,
	log *slog.Logger,
) *SignalEngine {
	ctx, cancel := context.WithCancel(context.Background())

	return &SignalEngine{
		marketDataRepo: marketDataRepo,
		indicatorRepo:  indicatorRepo,
		signalRepo:     signalRepo,
		log:            log,
		ctx:            ctx,
		cancel:         cancel,
		symbols:        make([]string, 0),
	}
}

// Start starts the signal generation engine
func (e *SignalEngine) Start(symbols []string, interval time.Duration) error {
	e.mu.Lock()
	e.symbols = symbols
	e.mu.Unlock()

	e.log.Info("Starting signal generation engine",
		"symbols", len(symbols),
		"interval", interval)

	// Generate signals immediately on start
	e.generateAllSignals()

	// Start periodic signal generation
	e.wg.Add(1)
	go e.generationRoutine(interval)

	e.log.Info("Signal generation engine started successfully")
	return nil
}

// generationRoutine periodically generates signals
func (e *SignalEngine) generationRoutine(interval time.Duration) {
	defer e.wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.generateAllSignals()
		}
	}
}

// generateAllSignals generates signals for all symbols
func (e *SignalEngine) generateAllSignals() {
	e.mu.RLock()
	symbols := make([]string, len(e.symbols))
	copy(symbols, e.symbols)
	e.mu.RUnlock()

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // Limit concurrent processing

	for _, symbol := range symbols {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()

			semaphore <- struct{}{} // Acquire
			defer func() { <-semaphore }() // Release

			if err := e.GenerateSignal(sym); err != nil {
				e.log.Error("Failed to generate signal",
					"symbol", sym,
					"error", err)
			}
		}(symbol)
	}

	wg.Wait()
}

// GenerateSignal generates a trading signal for a symbol
func (e *SignalEngine) GenerateSignal(symbol string) error {
	e.log.Debug("Generating signal", "symbol", symbol)

	// Get historical price data (100 periods for reliable indicators)
	end := time.Now()
	start := end.Add(-100 * 24 * time.Hour) // ~100 days

	dataPoints, err := e.marketDataRepo.GetPriceHistory(symbol, start, end)
	if err != nil {
		return fmt.Errorf("failed to get price history: %w", err)
	}

	if len(dataPoints) < 50 {
		e.log.Warn("Insufficient data for signal generation", "symbol", symbol, "count", len(dataPoints))
		return nil
	}

	// Extract price arrays
	closes := make([]float64, len(dataPoints))
	highs := make([]float64, len(dataPoints))
	lows := make([]float64, len(dataPoints))

	for i, dp := range dataPoints {
		closes[i] = dp.Price
		highs[i] = dp.High24h
		lows[i] = dp.Low24h
	}

	// Calculate indicators
	indicatorValues, err := e.calculateIndicators(closes, highs, lows)
	if err != nil {
		return fmt.Errorf("failed to calculate indicators: %w", err)
	}

	// Store indicators in TimescaleDB
	if err := e.storeIndicators(symbol, indicatorValues); err != nil {
		e.log.Error("Failed to store indicators", "symbol", symbol, "error", err)
	}

	// Generate trading signal
	signalScore := e.analyzeIndicators(indicatorValues, closes[len(closes)-1])

	// Determine asset type
	assetType := "stock"
	if len(symbol) > 3 && symbol[len(symbol)-4:] == "USDT" {
		assetType = "crypto"
	}

	// Create signal
	signal := e.createSignal(symbol, assetType, dataPoints[len(dataPoints)-1].Price, signalScore, indicatorValues)

	// Store signal in PostgreSQL
	if err := e.signalRepo.Create(signal); err != nil {
		return fmt.Errorf("failed to create signal: %w", err)
	}

	e.log.Info("Signal generated",
		"symbol", symbol,
		"action", signal.Action,
		"confidence", signal.Confidence)

	return nil
}

// calculateIndicators calculates all technical indicators
func (e *SignalEngine) calculateIndicators(closes, highs, lows []float64) (*IndicatorValues, error) {
	values := &IndicatorValues{}

	// RSI
	rsi14, err := indicators.RSI(closes, 14)
	if err == nil {
		values.RSI14 = indicators.GetLatestValue(rsi14)
	}

	rsi7, err := indicators.RSI(closes, 7)
	if err == nil {
		values.RSI7 = indicators.GetLatestValue(rsi7)
	}

	// Moving Averages
	sma20, err := indicators.SMA(closes, 20)
	if err == nil {
		values.SMA20 = indicators.GetLatestValue(sma20)
	}

	sma50, err := indicators.SMA(closes, 50)
	if err == nil {
		values.SMA50 = indicators.GetLatestValue(sma50)
	}

	sma200, err := indicators.SMA(closes, 200)
	if err == nil {
		values.SMA200 = indicators.GetLatestValue(sma200)
	}

	// MACD
	macdResult, err := indicators.MACD(closes, 12, 26, 9)
	if err == nil {
		values.MACD = indicators.GetLatestValue(macdResult.MACD)
		values.MACDSignal = indicators.GetLatestValue(macdResult.Signal)
		values.MACDHistogram = indicators.GetLatestValue(macdResult.Histogram)
		values.EMA12 = indicators.GetLatestValue(macdResult.MACD) + indicators.GetLatestValue(macdResult.Signal)
	}

	// Bollinger Bands
	bbResult, err := indicators.BollingerBands(closes, 20, 2.0)
	if err == nil {
		values.BBUpper = indicators.GetLatestValue(bbResult.Upper)
		values.BBMiddle = indicators.GetLatestValue(bbResult.Middle)
		values.BBLower = indicators.GetLatestValue(bbResult.Lower)
	}

	// ADX (trend strength)
	adxResult, err := indicators.ADX(highs, lows, closes, 14)
	if err == nil {
		values.ADX = indicators.GetLatestValue(adxResult.ADX)
	}

	return values, nil
}

// analyzeIndicators analyzes indicators and generates signal score
func (e *SignalEngine) analyzeIndicators(ind *IndicatorValues, currentPrice float64) *SignalScore {
	score := &SignalScore{
		Indicators: make(map[string]string),
		Reasoning:  make([]string, 0),
	}

	buySignals := 0
	sellSignals := 0
	totalSignals := 0

	// RSI Analysis
	if ind.RSI14 > 0 {
		totalSignals++
		rsiSignal := indicators.InterpretRSI(ind.RSI14)
		score.Indicators["RSI14"] = rsiSignal

		if rsiSignal == "BUY" {
			buySignals++
			score.Reasoning = append(score.Reasoning, fmt.Sprintf("RSI(14) is oversold at %.2f", ind.RSI14))
		} else if rsiSignal == "SELL" {
			sellSignals++
			score.Reasoning = append(score.Reasoning, fmt.Sprintf("RSI(14) is overbought at %.2f", ind.RSI14))
		}
	}

	// MACD Analysis
	if ind.MACD != 0 && ind.MACDSignal != 0 {
		totalSignals++
		macdSignal := indicators.InterpretMACD(ind.MACD, ind.MACDSignal)
		score.Indicators["MACD"] = macdSignal

		if macdSignal == "BUY" {
			buySignals++
			score.Reasoning = append(score.Reasoning, "MACD shows bullish crossover")
		} else if macdSignal == "SELL" {
			sellSignals++
			score.Reasoning = append(score.Reasoning, "MACD shows bearish crossover")
		}
	}

	// Bollinger Bands Analysis
	if ind.BBUpper > 0 && ind.BBLower > 0 {
		totalSignals++
		bbSignal := indicators.InterpretBollingerBands(currentPrice, ind.BBUpper, ind.BBLower)
		score.Indicators["BB"] = bbSignal

		if bbSignal == "BUY" {
			buySignals++
			score.Reasoning = append(score.Reasoning, "Price near lower Bollinger Band")
		} else if bbSignal == "SELL" {
			sellSignals++
			score.Reasoning = append(score.Reasoning, "Price near upper Bollinger Band")
		}
	}

	// Moving Average Crossover
	if ind.SMA20 > 0 && ind.SMA50 > 0 {
		totalSignals++
		if ind.SMA20 > ind.SMA50 {
			buySignals++
			score.Indicators["MA_Cross"] = "BUY"
			score.Reasoning = append(score.Reasoning, "SMA(20) above SMA(50) - bullish trend")
		} else {
			sellSignals++
			score.Indicators["MA_Cross"] = "SELL"
			score.Reasoning = append(score.Reasoning, "SMA(20) below SMA(50) - bearish trend")
		}
	}

	// Price vs SMA200 (long-term trend)
	if ind.SMA200 > 0 {
		totalSignals++
		if currentPrice > ind.SMA200 {
			buySignals++
			score.Indicators["LT_Trend"] = "BUY"
			score.Reasoning = append(score.Reasoning, "Price above SMA(200) - long-term bullish")
		} else {
			sellSignals++
			score.Indicators["LT_Trend"] = "SELL"
			score.Reasoning = append(score.Reasoning, "Price below SMA(200) - long-term bearish")
		}
	}

	// Determine action and confidence
	if totalSignals == 0 {
		score.Action = "HOLD"
		score.Confidence = 0
		return score
	}

	buyRatio := float64(buySignals) / float64(totalSignals)
	sellRatio := float64(sellSignals) / float64(totalSignals)

	// Strong consensus required for BUY/SELL
	if buyRatio >= 0.6 {
		score.Action = "BUY"
		score.Confidence = buyRatio * 100
	} else if sellRatio >= 0.6 {
		score.Action = "SELL"
		score.Confidence = sellRatio * 100
	} else {
		score.Action = "HOLD"
		score.Confidence = 50.0
		score.Reasoning = append(score.Reasoning, "Mixed signals - no clear consensus")
	}

	// Adjust confidence based on trend strength (ADX)
	if ind.ADX > 25 {
		score.Confidence += 10 // Strong trend increases confidence
		score.Reasoning = append(score.Reasoning, fmt.Sprintf("Strong trend detected (ADX: %.2f)", ind.ADX))
	} else if ind.ADX < 20 {
		score.Confidence -= 10 // Weak trend decreases confidence
		score.Reasoning = append(score.Reasoning, "Weak trend - market consolidation")
	}

	// Cap confidence at 95%
	if score.Confidence > 95 {
		score.Confidence = 95
	} else if score.Confidence < 5 {
		score.Confidence = 5
	}

	return score
}

// createSignal creates a domain signal from analysis
func (e *SignalEngine) createSignal(symbol, assetType string, currentPrice float64, score *SignalScore, ind *IndicatorValues) *domain.Signal {
	// Serialize indicators to JSON
	indicatorsJSON, _ := json.Marshal(score.Indicators)

	// Calculate entry, target, and stop loss
	entryPrice := currentPrice
	var targetPrice, stopLoss float64

	if score.Action == "BUY" {
		targetPrice = currentPrice * 1.05  // 5% profit target
		stopLoss = currentPrice * 0.97     // 3% stop loss
	} else if score.Action == "SELL" {
		targetPrice = currentPrice * 0.95  // 5% profit target (short)
		stopLoss = currentPrice * 1.03     // 3% stop loss (short)
	}

	// Build reasoning text
	reasoning := fmt.Sprintf("%s\n\nIndicators:\n", score.Action)
	for _, reason := range score.Reasoning {
		reasoning += fmt.Sprintf("- %s\n", reason)
	}

	return &domain.Signal{
		Symbol:      symbol,
		AssetType:   assetType,
		Action:      score.Action,
		Confidence:  score.Confidence,
		Reasoning:   reasoning,
		Indicators:  string(indicatorsJSON),
		EntryPrice:  entryPrice,
		TargetPrice: targetPrice,
		StopLoss:    stopLoss,
		ValidUntil:  time.Now().Add(24 * time.Hour),
		IsActive:    true,
	}
}

// storeIndicators stores calculated indicators in TimescaleDB
func (e *SignalEngine) storeIndicators(symbol string, ind *IndicatorValues) error {
	now := time.Now()
	indicators := []*timescale.TechnicalIndicator{
		{Time: now, Symbol: symbol, Indicator: "RSI", Value: ind.RSI14, Period: 14, Signal: indicators.InterpretRSI(ind.RSI14)},
		{Time: now, Symbol: symbol, Indicator: "SMA", Value: ind.SMA20, Period: 20, Signal: "NEUTRAL"},
		{Time: now, Symbol: symbol, Indicator: "SMA", Value: ind.SMA50, Period: 50, Signal: "NEUTRAL"},
		{Time: now, Symbol: symbol, Indicator: "MACD", Value: ind.MACD, Period: 12, Signal: indicators.InterpretMACD(ind.MACD, ind.MACDSignal)},
		{Time: now, Symbol: symbol, Indicator: "BB_UPPER", Value: ind.BBUpper, Period: 20, Signal: "NEUTRAL"},
		{Time: now, Symbol: symbol, Indicator: "BB_LOWER", Value: ind.BBLower, Period: 20, Signal: "NEUTRAL"},
		{Time: now, Symbol: symbol, Indicator: "ADX", Value: ind.ADX, Period: 14, Signal: "NEUTRAL"},
	}

	return e.indicatorRepo.InsertBatch(indicators)
}

// Stop stops the signal generation engine
func (e *SignalEngine) Stop() error {
	e.log.Info("Stopping signal generation engine")

	e.cancel()
	e.wg.Wait()

	// Deactivate expired signals
	if err := e.signalRepo.DeactivateExpiredSignals(); err != nil {
		e.log.Error("Failed to deactivate expired signals", "error", err)
	}

	e.log.Info("Signal generation engine stopped")
	return nil
}

// AddSymbols adds symbols to watch
func (e *SignalEngine) AddSymbols(symbols []string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.symbols = append(e.symbols, symbols...)
	e.log.Info("Added symbols to signal engine", "count", len(symbols))
}

// RemoveSymbols removes symbols from watch list
func (e *SignalEngine) RemoveSymbols(symbols []string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	removeMap := make(map[string]bool)
	for _, s := range symbols {
		removeMap[s] = true
	}

	newSymbols := make([]string, 0)
	for _, s := range e.symbols {
		if !removeMap[s] {
			newSymbols = append(newSymbols, s)
		}
	}

	e.symbols = newSymbols
	e.log.Info("Removed symbols from signal engine", "count", len(symbols))
}
