package timescale

import (
	"fmt"
	"time"
)

// MarketDataPoint represents a single price data point
type MarketDataPoint struct {
	Time      time.Time
	Symbol    string
	AssetType string
	Price     float64
	Volume    float64
	High24h   float64
	Low24h    float64
	Change24h float64
	MarketCap float64
	Source    string
}

// OHLCV represents candlestick data
type OHLCV struct {
	Time   time.Time
	Symbol string
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// MarketDataRepository handles time-series market data operations
type MarketDataRepository struct {
	db *TimescaleDB
}

// NewMarketDataRepository creates a new market data repository
func NewMarketDataRepository(db *TimescaleDB) *MarketDataRepository {
	return &MarketDataRepository{db: db}
}

// Insert inserts a market data point
func (r *MarketDataRepository) Insert(data *MarketDataPoint) error {
	if r.db == nil {
		return fmt.Errorf("timescaledb not initialized")
	}

	query := `
		INSERT INTO market_data (time, symbol, asset_type, price, volume, high_24h, low_24h, change_24h, market_cap, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.DB().Exec(query,
		data.Time,
		data.Symbol,
		data.AssetType,
		data.Price,
		data.Volume,
		data.High24h,
		data.Low24h,
		data.Change24h,
		data.MarketCap,
		data.Source,
	)

	return err
}

// InsertBatch inserts multiple data points in a single transaction
func (r *MarketDataRepository) InsertBatch(dataPoints []*MarketDataPoint) error {
	if r.db == nil {
		return fmt.Errorf("timescaledb not initialized")
	}

	tx, err := r.db.DB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO market_data (time, symbol, asset_type, price, volume, high_24h, low_24h, change_24h, market_cap, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, data := range dataPoints {
		if _, err := stmt.Exec(
			data.Time,
			data.Symbol,
			data.AssetType,
			data.Price,
			data.Volume,
			data.High24h,
			data.Low24h,
			data.Change24h,
			data.MarketCap,
			data.Source,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetLatestPrice gets the most recent price for a symbol
func (r *MarketDataRepository) GetLatestPrice(symbol string) (float64, error) {
	if r.db == nil {
		return 0, fmt.Errorf("timescaledb not initialized")
	}

	var price float64
	query := `
		SELECT price FROM market_data
		WHERE symbol = $1
		ORDER BY time DESC
		LIMIT 1
	`

	err := r.db.DB().QueryRow(query, symbol).Scan(&price)
	return price, err
}

// GetPriceHistory retrieves historical prices for a symbol
func (r *MarketDataRepository) GetPriceHistory(symbol string, start, end time.Time) ([]*MarketDataPoint, error) {
	if r.db == nil {
		return nil, fmt.Errorf("timescaledb not initialized")
	}

	query := `
		SELECT time, symbol, asset_type, price, volume, high_24h, low_24h, change_24h, market_cap, source
		FROM market_data
		WHERE symbol = $1 AND time >= $2 AND time <= $3
		ORDER BY time ASC
	`

	rows, err := r.db.DB().Query(query, symbol, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dataPoints []*MarketDataPoint
	for rows.Next() {
		var dp MarketDataPoint
		err := rows.Scan(
			&dp.Time,
			&dp.Symbol,
			&dp.AssetType,
			&dp.Price,
			&dp.Volume,
			&dp.High24h,
			&dp.Low24h,
			&dp.Change24h,
			&dp.MarketCap,
			&dp.Source,
		)
		if err != nil {
			return nil, err
		}
		dataPoints = append(dataPoints, &dp)
	}

	return dataPoints, rows.Err()
}

// GetOHLCV retrieves candlestick data for a symbol
func (r *MarketDataRepository) GetOHLCV(symbol string, interval string, start, end time.Time) ([]*OHLCV, error) {
	if r.db == nil {
		return nil, fmt.Errorf("timescaledb not initialized")
	}

	query := fmt.Sprintf(`
		SELECT
			time_bucket('%s', time) AS bucket,
			symbol,
			FIRST(price, time) AS open,
			MAX(price) AS high,
			MIN(price) AS low,
			LAST(price, time) AS close,
			SUM(volume) AS volume
		FROM market_data
		WHERE symbol = $1 AND time >= $2 AND time <= $3
		GROUP BY bucket, symbol
		ORDER BY bucket ASC
	`, interval)

	rows, err := r.db.DB().Query(query, symbol, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candles []*OHLCV
	for rows.Next() {
		var candle OHLCV
		err := rows.Scan(
			&candle.Time,
			&candle.Symbol,
			&candle.Open,
			&candle.High,
			&candle.Low,
			&candle.Close,
			&candle.Volume,
		)
		if err != nil {
			return nil, err
		}
		candles = append(candles, &candle)
	}

	return candles, rows.Err()
}

// GetRecentCandles retrieves the last N candles for a symbol
func (r *MarketDataRepository) GetRecentCandles(symbol string, interval string, count int) ([]*OHLCV, error) {
	if r.db == nil {
		return nil, fmt.Errorf("timescaledb not initialized")
	}

	query := fmt.Sprintf(`
		SELECT
			time_bucket('%s', time) AS bucket,
			symbol,
			FIRST(price, time) AS open,
			MAX(price) AS high,
			MIN(price) AS low,
			LAST(price, time) AS close,
			SUM(volume) AS volume
		FROM market_data
		WHERE symbol = $1
		GROUP BY bucket, symbol
		ORDER BY bucket DESC
		LIMIT $2
	`, interval)

	rows, err := r.db.DB().Query(query, symbol, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candles []*OHLCV
	for rows.Next() {
		var candle OHLCV
		err := rows.Scan(
			&candle.Time,
			&candle.Symbol,
			&candle.Open,
			&candle.High,
			&candle.Low,
			&candle.Close,
			&candle.Volume,
		)
		if err != nil {
			return nil, err
		}
		candles = append(candles, &candle)
	}

	// Reverse to get chronological order
	for i, j := 0, len(candles)-1; i < j; i, j = i+1, j-1 {
		candles[i], candles[j] = candles[j], candles[i]
	}

	return candles, rows.Err()
}

// GetPriceAtTime gets the price closest to a specific time
func (r *MarketDataRepository) GetPriceAtTime(symbol string, t time.Time) (float64, error) {
	if r.db == nil {
		return 0, fmt.Errorf("timescaledb not initialized")
	}

	var price float64
	query := `
		SELECT price FROM market_data
		WHERE symbol = $1 AND time <= $2
		ORDER BY time DESC
		LIMIT 1
	`

	err := r.db.DB().QueryRow(query, symbol, t).Scan(&price)
	return price, err
}

// GetAvgPrice gets the average price over a time period
func (r *MarketDataRepository) GetAvgPrice(symbol string, start, end time.Time) (float64, error) {
	if r.db == nil {
		return 0, fmt.Errorf("timescaledb not initialized")
	}

	var avgPrice float64
	query := `
		SELECT AVG(price) FROM market_data
		WHERE symbol = $1 AND time >= $2 AND time <= $3
	`

	err := r.db.DB().QueryRow(query, symbol, start, end).Scan(&avgPrice)
	return avgPrice, err
}

// GetVolatility calculates price volatility (standard deviation) over a period
func (r *MarketDataRepository) GetVolatility(symbol string, start, end time.Time) (float64, error) {
	if r.db == nil {
		return 0, fmt.Errorf("timescaledb not initialized")
	}

	var volatility float64
	query := `
		SELECT STDDEV(price) FROM market_data
		WHERE symbol = $1 AND time >= $2 AND time <= $3
	`

	err := r.db.DB().QueryRow(query, symbol, start, end).Scan(&volatility)
	return volatility, err
}
