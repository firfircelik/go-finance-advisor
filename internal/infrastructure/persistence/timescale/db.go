package timescale

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"go-finance-advisor/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
)

// TimescaleDB client for time-series data
type TimescaleDB struct {
	db  *sql.DB
	log *slog.Logger
}

// NewTimescaleDB creates a new TimescaleDB connection
func NewTimescaleDB(cfg *config.TimescaleDBConfig, log *slog.Logger) (*TimescaleDB, error) {
	if !cfg.Enabled {
		log.Info("TimescaleDB is disabled")
		return nil, nil
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to TimescaleDB: %w", err)
	}

	// Configure connection pool
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(1 * time.Hour)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping TimescaleDB: %w", err)
	}

	log.Info("TimescaleDB connected successfully",
		"host", cfg.Host,
		"database", cfg.Name,
	)

	tsdb := &TimescaleDB{
		db:  db,
		log: log,
	}

	// Initialize schema
	if err := tsdb.initSchema(cfg); err != nil {
		return nil, err
	}

	return tsdb, nil
}

// initSchema creates necessary tables and hypertables
func (ts *TimescaleDB) initSchema(cfg *config.TimescaleDBConfig) error {
	// Create TimescaleDB extension if not exists
	if _, err := ts.db.Exec("CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE"); err != nil {
		return fmt.Errorf("failed to create timescaledb extension: %w", err)
	}

	// Create market_data table
	if _, err := ts.db.Exec(`
		CREATE TABLE IF NOT EXISTS market_data (
			time        TIMESTAMPTZ NOT NULL,
			symbol      TEXT NOT NULL,
			asset_type  TEXT NOT NULL,
			price       DOUBLE PRECISION NOT NULL,
			volume      DOUBLE PRECISION DEFAULT 0,
			high_24h    DOUBLE PRECISION DEFAULT 0,
			low_24h     DOUBLE PRECISION DEFAULT 0,
			change_24h  DOUBLE PRECISION DEFAULT 0,
			market_cap  DOUBLE PRECISION DEFAULT 0,
			source      TEXT
		)
	`); err != nil {
		return fmt.Errorf("failed to create market_data table: %w", err)
	}

	// Convert to hypertable (if not already)
	if _, err := ts.db.Exec(`
		SELECT create_hypertable('market_data', 'time',
			if_not_exists => TRUE,
			chunk_time_interval => INTERVAL '` + cfg.ChunkInterval + `'
		)
	`); err != nil {
		ts.log.Warn("hypertable creation warning", "error", err)
	}

	// Create index on symbol and time
	if _, err := ts.db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_market_data_symbol_time
		ON market_data (symbol, time DESC)
	`); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	// Create technical_indicators table
	if _, err := ts.db.Exec(`
		CREATE TABLE IF NOT EXISTS technical_indicators (
			time       TIMESTAMPTZ NOT NULL,
			symbol     TEXT NOT NULL,
			indicator  TEXT NOT NULL,
			value      DOUBLE PRECISION NOT NULL,
			period     INTEGER NOT NULL,
			signal     TEXT
		)
	`); err != nil {
		return fmt.Errorf("failed to create technical_indicators table: %w", err)
	}

	// Convert to hypertable
	if _, err := ts.db.Exec(`
		SELECT create_hypertable('technical_indicators', 'time',
			if_not_exists => TRUE,
			chunk_time_interval => INTERVAL '1 hour'
		)
	`); err != nil {
		ts.log.Warn("hypertable creation warning", "error", err)
	}

	// Set up retention policy
	if _, err := ts.db.Exec(fmt.Sprintf(`
		SELECT add_retention_policy('market_data', INTERVAL '%d days', if_not_exists => TRUE)
	`, cfg.RetentionDays)); err != nil {
		ts.log.Warn("retention policy warning", "error", err)
	}

	// Enable compression on old data (7 days)
	if _, err := ts.db.Exec(`
		ALTER TABLE market_data SET (
			timescaledb.compress,
			timescaledb.compress_segmentby = 'symbol,asset_type'
		)
	`); err != nil {
		ts.log.Warn("compression setup warning", "error", err)
	}

	if _, err := ts.db.Exec(`
		SELECT add_compression_policy('market_data', INTERVAL '7 days', if_not_exists => TRUE)
	`); err != nil {
		ts.log.Warn("compression policy warning", "error", err)
	}

	ts.log.Info("TimescaleDB schema initialized successfully")
	return nil
}

// Close closes the database connection
func (ts *TimescaleDB) Close() error {
	if ts.db != nil {
		return ts.db.Close()
	}
	return nil
}

// DB returns the underlying *sql.DB
func (ts *TimescaleDB) DB() *sql.DB {
	return ts.db
}
