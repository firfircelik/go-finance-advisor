package indicators

import (
	"fmt"
	"math"
)

// PriceData represents a single price data point
type PriceData struct {
	Close  float64
	High   float64
	Low    float64
	Volume float64
}

// RSI calculates the Relative Strength Index
// RSI = 100 - (100 / (1 + RS))
// RS = Average Gain / Average Loss
func RSI(prices []float64, period int) ([]float64, error) {
	if len(prices) < period+1 {
		return nil, fmt.Errorf("insufficient data: need at least %d prices, got %d", period+1, len(prices))
	}

	if period <= 0 {
		return nil, fmt.Errorf("period must be positive, got %d", period)
	}

	rsi := make([]float64, len(prices))

	// Calculate price changes
	changes := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		changes[i-1] = prices[i] - prices[i-1]
	}

	// Calculate initial average gain and loss
	var avgGain, avgLoss float64
	for i := 0; i < period; i++ {
		if changes[i] > 0 {
			avgGain += changes[i]
		} else {
			avgLoss += math.Abs(changes[i])
		}
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// Calculate RSI for each period
	for i := period; i < len(prices); i++ {
		if i > period {
			// Smooth the averages using Wilder's smoothing method
			currentChange := changes[i-1]
			if currentChange > 0 {
				avgGain = (avgGain*float64(period-1) + currentChange) / float64(period)
				avgLoss = (avgLoss * float64(period-1)) / float64(period)
			} else {
				avgGain = (avgGain * float64(period-1)) / float64(period)
				avgLoss = (avgLoss*float64(period-1) + math.Abs(currentChange)) / float64(period)
			}
		}

		// Calculate RS and RSI
		var rs float64
		if avgLoss == 0 {
			rs = 100
		} else {
			rs = avgGain / avgLoss
		}

		rsi[i] = 100 - (100 / (1 + rs))
	}

	return rsi, nil
}

// SMA calculates Simple Moving Average
func SMA(prices []float64, period int) ([]float64, error) {
	if len(prices) < period {
		return nil, fmt.Errorf("insufficient data: need at least %d prices, got %d", period, len(prices))
	}

	if period <= 0 {
		return nil, fmt.Errorf("period must be positive, got %d", period)
	}

	sma := make([]float64, len(prices))

	for i := period - 1; i < len(prices); i++ {
		sum := 0.0
		for j := 0; j < period; j++ {
			sum += prices[i-j]
		}
		sma[i] = sum / float64(period)
	}

	return sma, nil
}

// EMA calculates Exponential Moving Average
// EMA = Price * K + EMA(previous) * (1 - K)
// K = 2 / (N + 1)
func EMA(prices []float64, period int) ([]float64, error) {
	if len(prices) < period {
		return nil, fmt.Errorf("insufficient data: need at least %d prices, got %d", period, len(prices))
	}

	if period <= 0 {
		return nil, fmt.Errorf("period must be positive, got %d", period)
	}

	ema := make([]float64, len(prices))
	multiplier := 2.0 / float64(period+1)

	// Start with SMA for the first value
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	ema[period-1] = sum / float64(period)

	// Calculate EMA for remaining values
	for i := period; i < len(prices); i++ {
		ema[i] = (prices[i]-ema[i-1])*multiplier + ema[i-1]
	}

	return ema, nil
}

// MACDResult contains MACD indicator values
type MACDResult struct {
	MACD      []float64
	Signal    []float64
	Histogram []float64
}

// MACD calculates Moving Average Convergence Divergence
// MACD Line = EMA(12) - EMA(26)
// Signal Line = EMA(9) of MACD Line
// Histogram = MACD Line - Signal Line
func MACD(prices []float64, fastPeriod, slowPeriod, signalPeriod int) (*MACDResult, error) {
	if len(prices) < slowPeriod {
		return nil, fmt.Errorf("insufficient data: need at least %d prices, got %d", slowPeriod, len(prices))
	}

	// Calculate fast and slow EMAs
	fastEMA, err := EMA(prices, fastPeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate fast EMA: %w", err)
	}

	slowEMA, err := EMA(prices, slowPeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate slow EMA: %w", err)
	}

	// Calculate MACD line
	macdLine := make([]float64, len(prices))
	for i := 0; i < len(prices); i++ {
		if fastEMA[i] != 0 && slowEMA[i] != 0 {
			macdLine[i] = fastEMA[i] - slowEMA[i]
		}
	}

	// Calculate signal line (EMA of MACD line)
	signalLine, err := EMA(macdLine, signalPeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate signal line: %w", err)
	}

	// Calculate histogram
	histogram := make([]float64, len(prices))
	for i := 0; i < len(prices); i++ {
		if macdLine[i] != 0 && signalLine[i] != 0 {
			histogram[i] = macdLine[i] - signalLine[i]
		}
	}

	return &MACDResult{
		MACD:      macdLine,
		Signal:    signalLine,
		Histogram: histogram,
	}, nil
}

// BollingerBandsResult contains Bollinger Bands values
type BollingerBandsResult struct {
	Upper  []float64
	Middle []float64
	Lower  []float64
}

// BollingerBands calculates Bollinger Bands
// Middle Band = SMA(N)
// Upper Band = SMA(N) + (K * Standard Deviation)
// Lower Band = SMA(N) - (K * Standard Deviation)
func BollingerBands(prices []float64, period int, stdDevMultiplier float64) (*BollingerBandsResult, error) {
	if len(prices) < period {
		return nil, fmt.Errorf("insufficient data: need at least %d prices, got %d", period, len(prices))
	}

	// Calculate middle band (SMA)
	middleBand, err := SMA(prices, period)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate SMA: %w", err)
	}

	upperBand := make([]float64, len(prices))
	lowerBand := make([]float64, len(prices))

	// Calculate standard deviation and bands
	for i := period - 1; i < len(prices); i++ {
		// Calculate standard deviation for the period
		sum := 0.0
		for j := 0; j < period; j++ {
			diff := prices[i-j] - middleBand[i]
			sum += diff * diff
		}
		stdDev := math.Sqrt(sum / float64(period))

		upperBand[i] = middleBand[i] + (stdDevMultiplier * stdDev)
		lowerBand[i] = middleBand[i] - (stdDevMultiplier * stdDev)
	}

	return &BollingerBandsResult{
		Upper:  upperBand,
		Middle: middleBand,
		Lower:  lowerBand,
	}, nil
}

// Stochastic calculates Stochastic Oscillator
// %K = 100 * (Close - Lowest Low) / (Highest High - Lowest Low)
// %D = SMA of %K
type StochasticResult struct {
	K []float64
	D []float64
}

func Stochastic(highs, lows, closes []float64, kPeriod, dPeriod int) (*StochasticResult, error) {
	if len(highs) != len(lows) || len(lows) != len(closes) {
		return nil, fmt.Errorf("price arrays must have equal length")
	}

	if len(closes) < kPeriod {
		return nil, fmt.Errorf("insufficient data: need at least %d prices, got %d", kPeriod, len(closes))
	}

	k := make([]float64, len(closes))

	// Calculate %K
	for i := kPeriod - 1; i < len(closes); i++ {
		highestHigh := highs[i]
		lowestLow := lows[i]

		// Find highest high and lowest low in the period
		for j := 0; j < kPeriod; j++ {
			if highs[i-j] > highestHigh {
				highestHigh = highs[i-j]
			}
			if lows[i-j] < lowestLow {
				lowestLow = lows[i-j]
			}
		}

		range_ := highestHigh - lowestLow
		if range_ == 0 {
			k[i] = 50 // Neutral value when no range
		} else {
			k[i] = 100 * (closes[i] - lowestLow) / range_
		}
	}

	// Calculate %D (SMA of %K)
	d, err := SMA(k, dPeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate %D: %w", err)
	}

	return &StochasticResult{
		K: k,
		D: d,
	}, nil
}

// ATR calculates Average True Range
func ATR(highs, lows, closes []float64, period int) ([]float64, error) {
	if len(highs) != len(lows) || len(lows) != len(closes) {
		return nil, fmt.Errorf("price arrays must have equal length")
	}

	if len(closes) < period+1 {
		return nil, fmt.Errorf("insufficient data: need at least %d prices, got %d", period+1, len(closes))
	}

	atr := make([]float64, len(closes))
	trueRanges := make([]float64, len(closes))

	// Calculate True Range for each period
	for i := 1; i < len(closes); i++ {
		highLow := highs[i] - lows[i]
		highClose := math.Abs(highs[i] - closes[i-1])
		lowClose := math.Abs(lows[i] - closes[i-1])

		trueRanges[i] = math.Max(highLow, math.Max(highClose, lowClose))
	}

	// Calculate initial ATR (simple average)
	sum := 0.0
	for i := 1; i <= period; i++ {
		sum += trueRanges[i]
	}
	atr[period] = sum / float64(period)

	// Calculate subsequent ATR values using smoothed average
	for i := period + 1; i < len(closes); i++ {
		atr[i] = (atr[i-1]*float64(period-1) + trueRanges[i]) / float64(period)
	}

	return atr, nil
}

// ADX calculates Average Directional Index
// Measures trend strength (not direction)
type ADXResult struct {
	ADX      []float64
	PlusDI   []float64
	MinusDI  []float64
}

func ADX(highs, lows, closes []float64, period int) (*ADXResult, error) {
	if len(highs) != len(lows) || len(lows) != len(closes) {
		return nil, fmt.Errorf("price arrays must have equal length")
	}

	if len(closes) < period*2 {
		return nil, fmt.Errorf("insufficient data: need at least %d prices, got %d", period*2, len(closes))
	}

	plusDM := make([]float64, len(closes))
	minusDM := make([]float64, len(closes))
	tr := make([]float64, len(closes))

	// Calculate +DM, -DM, and TR
	for i := 1; i < len(closes); i++ {
		highDiff := highs[i] - highs[i-1]
		lowDiff := lows[i-1] - lows[i]

		if highDiff > lowDiff && highDiff > 0 {
			plusDM[i] = highDiff
		}
		if lowDiff > highDiff && lowDiff > 0 {
			minusDM[i] = lowDiff
		}

		highLow := highs[i] - lows[i]
		highClose := math.Abs(highs[i] - closes[i-1])
		lowClose := math.Abs(lows[i] - closes[i-1])
		tr[i] = math.Max(highLow, math.Max(highClose, lowClose))
	}

	// Smooth +DM, -DM, and TR
	smoothPlusDM := make([]float64, len(closes))
	smoothMinusDM := make([]float64, len(closes))
	smoothTR := make([]float64, len(closes))

	// Initial sums
	for i := 1; i <= period; i++ {
		smoothPlusDM[period] += plusDM[i]
		smoothMinusDM[period] += minusDM[i]
		smoothTR[period] += tr[i]
	}

	// Smooth subsequent values
	for i := period + 1; i < len(closes); i++ {
		smoothPlusDM[i] = smoothPlusDM[i-1] - (smoothPlusDM[i-1] / float64(period)) + plusDM[i]
		smoothMinusDM[i] = smoothMinusDM[i-1] - (smoothMinusDM[i-1] / float64(period)) + minusDM[i]
		smoothTR[i] = smoothTR[i-1] - (smoothTR[i-1] / float64(period)) + tr[i]
	}

	// Calculate +DI and -DI
	plusDI := make([]float64, len(closes))
	minusDI := make([]float64, len(closes))

	for i := period; i < len(closes); i++ {
		if smoothTR[i] != 0 {
			plusDI[i] = 100 * smoothPlusDM[i] / smoothTR[i]
			minusDI[i] = 100 * smoothMinusDM[i] / smoothTR[i]
		}
	}

	// Calculate DX and ADX
	dx := make([]float64, len(closes))
	for i := period; i < len(closes); i++ {
		diDiff := math.Abs(plusDI[i] - minusDI[i])
		diSum := plusDI[i] + minusDI[i]

		if diSum != 0 {
			dx[i] = 100 * diDiff / diSum
		}
	}

	// Calculate ADX as EMA of DX
	adx, err := EMA(dx, period)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate ADX: %w", err)
	}

	return &ADXResult{
		ADX:     adx,
		PlusDI:  plusDI,
		MinusDI: minusDI,
	}, nil
}

// GetLatestValue returns the latest non-zero value from a slice
func GetLatestValue(values []float64) float64 {
	for i := len(values) - 1; i >= 0; i-- {
		if values[i] != 0 {
			return values[i]
		}
	}
	return 0
}

// InterpretRSI provides trading signal interpretation for RSI
func InterpretRSI(rsi float64) string {
	if rsi >= 70 {
		return "SELL" // Overbought
	} else if rsi <= 30 {
		return "BUY" // Oversold
	}
	return "NEUTRAL"
}

// InterpretMACD provides trading signal interpretation for MACD
func InterpretMACD(macd, signal float64) string {
	if macd > signal {
		return "BUY" // Bullish crossover
	} else if macd < signal {
		return "SELL" // Bearish crossover
	}
	return "NEUTRAL"
}

// InterpretBollingerBands provides trading signal interpretation for Bollinger Bands
func InterpretBollingerBands(price, upper, lower float64) string {
	if price >= upper {
		return "SELL" // Price at upper band (overbought)
	} else if price <= lower {
		return "BUY" // Price at lower band (oversold)
	}
	return "NEUTRAL"
}
