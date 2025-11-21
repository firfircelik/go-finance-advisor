package timescale

import (
	"fmt"
	"time"
)

// TechnicalIndicator represents a calculated technical indicator
type TechnicalIndicator struct {
	Time      time.Time
	Symbol    string
	Indicator string  // RSI, MACD, SMA, EMA, BB_UPPER, BB_LOWER, etc.
	Value     float64
	Period    int    // Period used for calculation (e.g., 14 for RSI-14)
	Signal    string // BUY, SELL, NEUTRAL
	Metadata  string // JSON metadata for complex indicators (e.g., MACD signal line)
}

// IndicatorRepository handles technical indicator storage and retrieval
type IndicatorRepository struct {
	db *TimescaleDB
}

// NewIndicatorRepository creates a new indicator repository
func NewIndicatorRepository(db *TimescaleDB) *IndicatorRepository {
	return &IndicatorRepository{db: db}
}

// Insert inserts a single technical indicator
func (r *IndicatorRepository) Insert(indicator *TechnicalIndicator) error {
	if r.db == nil {
		return fmt.Errorf("timescaledb not initialized")
	}

	query := `
		INSERT INTO technical_indicators (time, symbol, indicator, value, period, signal, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.DB().Exec(query,
		indicator.Time,
		indicator.Symbol,
		indicator.Indicator,
		indicator.Value,
		indicator.Period,
		indicator.Signal,
		indicator.Metadata,
	)

	return err
}

// InsertBatch inserts multiple indicators in a single transaction
func (r *IndicatorRepository) InsertBatch(indicators []*TechnicalIndicator) error {
	if r.db == nil {
		return fmt.Errorf("timescaledb not initialized")
	}

	tx, err := r.db.DB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO technical_indicators (time, symbol, indicator, value, period, signal, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, ind := range indicators {
		if _, err := stmt.Exec(
			ind.Time,
			ind.Symbol,
			ind.Indicator,
			ind.Value,
			ind.Period,
			ind.Signal,
			ind.Metadata,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetLatestIndicator gets the most recent value for a specific indicator
func (r *IndicatorRepository) GetLatestIndicator(symbol, indicator string, period int) (*TechnicalIndicator, error) {
	if r.db == nil {
		return nil, fmt.Errorf("timescaledb not initialized")
	}

	var ind TechnicalIndicator
	query := `
		SELECT time, symbol, indicator, value, period, signal, metadata
		FROM technical_indicators
		WHERE symbol = $1 AND indicator = $2 AND period = $3
		ORDER BY time DESC
		LIMIT 1
	`

	err := r.db.DB().QueryRow(query, symbol, indicator, period).Scan(
		&ind.Time,
		&ind.Symbol,
		&ind.Indicator,
		&ind.Value,
		&ind.Period,
		&ind.Signal,
		&ind.Metadata,
	)

	if err != nil {
		return nil, err
	}

	return &ind, nil
}

// GetLatestIndicators gets all latest indicators for a symbol
func (r *IndicatorRepository) GetLatestIndicators(symbol string) ([]*TechnicalIndicator, error) {
	if r.db == nil {
		return nil, fmt.Errorf("timescaledb not initialized")
	}

	query := `
		SELECT DISTINCT ON (indicator, period)
			time, symbol, indicator, value, period, signal, metadata
		FROM technical_indicators
		WHERE symbol = $1
		ORDER BY indicator, period, time DESC
	`

	rows, err := r.db.DB().Query(query, symbol)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indicators []*TechnicalIndicator
	for rows.Next() {
		var ind TechnicalIndicator
		err := rows.Scan(
			&ind.Time,
			&ind.Symbol,
			&ind.Indicator,
			&ind.Value,
			&ind.Period,
			&ind.Signal,
			&ind.Metadata,
		)
		if err != nil {
			return nil, err
		}
		indicators = append(indicators, &ind)
	}

	return indicators, rows.Err()
}

// GetIndicatorHistory retrieves historical indicator values
func (r *IndicatorRepository) GetIndicatorHistory(symbol, indicator string, period int, start, end time.Time) ([]*TechnicalIndicator, error) {
	if r.db == nil {
		return nil, fmt.Errorf("timescaledb not initialized")
	}

	query := `
		SELECT time, symbol, indicator, value, period, signal, metadata
		FROM technical_indicators
		WHERE symbol = $1 AND indicator = $2 AND period = $3
		  AND time >= $4 AND time <= $5
		ORDER BY time ASC
	`

	rows, err := r.db.DB().Query(query, symbol, indicator, period, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indicators []*TechnicalIndicator
	for rows.Next() {
		var ind TechnicalIndicator
		err := rows.Scan(
			&ind.Time,
			&ind.Symbol,
			&ind.Indicator,
			&ind.Value,
			&ind.Period,
			&ind.Signal,
			&ind.Metadata,
		)
		if err != nil {
			return nil, err
		}
		indicators = append(indicators, &ind)
	}

	return indicators, rows.Err()
}

// GetIndicatorsByTimeRange retrieves all indicators for a symbol in a time range
func (r *IndicatorRepository) GetIndicatorsByTimeRange(symbol string, start, end time.Time) ([]*TechnicalIndicator, error) {
	if r.db == nil {
		return nil, fmt.Errorf("timescaledb not initialized")
	}

	query := `
		SELECT time, symbol, indicator, value, period, signal, metadata
		FROM technical_indicators
		WHERE symbol = $1 AND time >= $2 AND time <= $3
		ORDER BY time ASC, indicator, period
	`

	rows, err := r.db.DB().Query(query, symbol, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indicators []*TechnicalIndicator
	for rows.Next() {
		var ind TechnicalIndicator
		err := rows.Scan(
			&ind.Time,
			&ind.Symbol,
			&ind.Indicator,
			&ind.Value,
			&ind.Period,
			&ind.Signal,
			&ind.Metadata,
		)
		if err != nil {
			return nil, err
		}
		indicators = append(indicators, &ind)
	}

	return indicators, rows.Err()
}

// GetSignalsByIndicator retrieves all indicators with a specific signal
func (r *IndicatorRepository) GetSignalsByIndicator(indicator, signal string, since time.Time) ([]*TechnicalIndicator, error) {
	if r.db == nil {
		return nil, fmt.Errorf("timescaledb not initialized")
	}

	query := `
		SELECT time, symbol, indicator, value, period, signal, metadata
		FROM technical_indicators
		WHERE indicator = $1 AND signal = $2 AND time >= $3
		ORDER BY time DESC
	`

	rows, err := r.db.DB().Query(query, indicator, signal, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indicators []*TechnicalIndicator
	for rows.Next() {
		var ind TechnicalIndicator
		err := rows.Scan(
			&ind.Time,
			&ind.Symbol,
			&ind.Indicator,
			&ind.Value,
			&ind.Period,
			&ind.Signal,
			&ind.Metadata,
		)
		if err != nil {
			return nil, err
		}
		indicators = append(indicators, &ind)
	}

	return indicators, rows.Err()
}

// GetRecentBuySignals retrieves recent BUY signals across all symbols
func (r *IndicatorRepository) GetRecentBuySignals(hours int) ([]*TechnicalIndicator, error) {
	if r.db == nil {
		return nil, fmt.Errorf("timescaledb not initialized")
	}

	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	query := `
		SELECT time, symbol, indicator, value, period, signal, metadata
		FROM technical_indicators
		WHERE signal = 'BUY' AND time >= $1
		ORDER BY time DESC, symbol
	`

	rows, err := r.db.DB().Query(query, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indicators []*TechnicalIndicator
	for rows.Next() {
		var ind TechnicalIndicator
		err := rows.Scan(
			&ind.Time,
			&ind.Symbol,
			&ind.Indicator,
			&ind.Value,
			&ind.Period,
			&ind.Signal,
			&ind.Metadata,
		)
		if err != nil {
			return nil, err
		}
		indicators = append(indicators, &ind)
	}

	return indicators, rows.Err()
}

// GetIndicatorStats calculates statistics for an indicator over a time period
func (r *IndicatorRepository) GetIndicatorStats(symbol, indicator string, period int, start, end time.Time) (map[string]float64, error) {
	if r.db == nil {
		return nil, fmt.Errorf("timescaledb not initialized")
	}

	query := `
		SELECT
			AVG(value) as avg_value,
			MIN(value) as min_value,
			MAX(value) as max_value,
			STDDEV(value) as std_dev,
			COUNT(*) as count
		FROM technical_indicators
		WHERE symbol = $1 AND indicator = $2 AND period = $3
		  AND time >= $4 AND time <= $5
	`

	var avgValue, minValue, maxValue, stdDev float64
	var count int64

	err := r.db.DB().QueryRow(query, symbol, indicator, period, start, end).Scan(
		&avgValue,
		&minValue,
		&maxValue,
		&stdDev,
		&count,
	)

	if err != nil {
		return nil, err
	}

	stats := map[string]float64{
		"avg":    avgValue,
		"min":    minValue,
		"max":    maxValue,
		"stddev": stdDev,
		"count":  float64(count),
	}

	return stats, nil
}

// DeleteOldIndicators removes indicators older than the specified date
func (r *IndicatorRepository) DeleteOldIndicators(before time.Time) (int64, error) {
	if r.db == nil {
		return 0, fmt.Errorf("timescaledb not initialized")
	}

	result, err := r.db.DB().Exec(`
		DELETE FROM technical_indicators
		WHERE time < $1
	`, before)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
