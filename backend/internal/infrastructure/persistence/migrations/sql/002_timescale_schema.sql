-- GoFinVisor TimescaleDB Schema Migration
-- Creates time-series tables with hypertables, compression, and retention policies

-- Create TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- Market Data Table (Time-series)
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
);

-- Convert to hypertable with 1-day chunks
SELECT create_hypertable('market_data', 'time',
    if_not_exists => TRUE,
    chunk_time_interval => INTERVAL '1 day'
);

-- Create indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_market_data_symbol_time
    ON market_data (symbol, time DESC);

CREATE INDEX IF NOT EXISTS idx_market_data_asset_type_time
    ON market_data (asset_type, time DESC);

CREATE INDEX IF NOT EXISTS idx_market_data_source
    ON market_data (source, time DESC);

-- Enable compression on market_data
ALTER TABLE market_data SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol,asset_type',
    timescaledb.compress_orderby = 'time DESC'
);

-- Add compression policy (compress data older than 7 days)
SELECT add_compression_policy('market_data', INTERVAL '7 days', if_not_exists => TRUE);

-- Add retention policy (keep data for 365 days)
SELECT add_retention_policy('market_data', INTERVAL '365 days', if_not_exists => TRUE);

-- Technical Indicators Table (Time-series)
CREATE TABLE IF NOT EXISTS technical_indicators (
    time       TIMESTAMPTZ NOT NULL,
    symbol     TEXT NOT NULL,
    indicator  TEXT NOT NULL,
    value      DOUBLE PRECISION NOT NULL,
    period     INTEGER NOT NULL,
    signal     TEXT,
    metadata   TEXT
);

-- Convert to hypertable with 1-hour chunks (indicators calculated more frequently)
SELECT create_hypertable('technical_indicators', 'time',
    if_not_exists => TRUE,
    chunk_time_interval => INTERVAL '1 hour'
);

-- Create indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_technical_indicators_symbol_time
    ON technical_indicators (symbol, time DESC);

CREATE INDEX IF NOT EXISTS idx_technical_indicators_indicator_period
    ON technical_indicators (indicator, period, time DESC);

CREATE INDEX IF NOT EXISTS idx_technical_indicators_signal
    ON technical_indicators (signal, time DESC);

-- Enable compression on technical_indicators
ALTER TABLE technical_indicators SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol,indicator,period',
    timescaledb.compress_orderby = 'time DESC'
);

-- Add compression policy (compress data older than 3 days)
SELECT add_compression_policy('technical_indicators', INTERVAL '3 days', if_not_exists => TRUE);

-- Add retention policy (keep indicators for 90 days)
SELECT add_retention_policy('technical_indicators', INTERVAL '90 days', if_not_exists => TRUE);

-- Portfolio Performance History Table (Time-series)
CREATE TABLE IF NOT EXISTS portfolio_snapshots (
    time         TIMESTAMPTZ NOT NULL,
    portfolio_id INTEGER NOT NULL,
    user_id      INTEGER NOT NULL,
    total_value  DOUBLE PRECISION NOT NULL,
    daily_return DOUBLE PRECISION DEFAULT 0,
    total_return DOUBLE PRECISION DEFAULT 0,
    composition  JSONB
);

-- Convert to hypertable with 1-day chunks
SELECT create_hypertable('portfolio_snapshots', 'time',
    if_not_exists => TRUE,
    chunk_time_interval => INTERVAL '1 day'
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_portfolio_snapshots_portfolio_id
    ON portfolio_snapshots (portfolio_id, time DESC);

CREATE INDEX IF NOT EXISTS idx_portfolio_snapshots_user_id
    ON portfolio_snapshots (user_id, time DESC);

-- Enable compression
ALTER TABLE portfolio_snapshots SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'portfolio_id,user_id',
    timescaledb.compress_orderby = 'time DESC'
);

-- Add compression policy (compress data older than 30 days)
SELECT add_compression_policy('portfolio_snapshots', INTERVAL '30 days', if_not_exists => TRUE);

-- Add retention policy (keep performance history for 2 years)
SELECT add_retention_policy('portfolio_snapshots', INTERVAL '730 days', if_not_exists => TRUE);

-- Create continuous aggregates for common queries

-- 1-hour average prices
CREATE MATERIALIZED VIEW IF NOT EXISTS market_data_1h
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', time) AS bucket,
    symbol,
    asset_type,
    FIRST(price, time) AS open,
    MAX(price) AS high,
    MIN(price) AS low,
    LAST(price, time) AS close,
    AVG(price) AS avg_price,
    SUM(volume) AS volume
FROM market_data
GROUP BY bucket, symbol, asset_type
WITH NO DATA;

-- Add refresh policy for 1-hour aggregate
SELECT add_continuous_aggregate_policy('market_data_1h',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE
);

-- Daily average prices
CREATE MATERIALIZED VIEW IF NOT EXISTS market_data_1d
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', time) AS bucket,
    symbol,
    asset_type,
    FIRST(price, time) AS open,
    MAX(price) AS high,
    MIN(price) AS low,
    LAST(price, time) AS close,
    AVG(price) AS avg_price,
    SUM(volume) AS volume,
    STDDEV(price) AS volatility
FROM market_data
GROUP BY bucket, symbol, asset_type
WITH NO DATA;

-- Add refresh policy for daily aggregate
SELECT add_continuous_aggregate_policy('market_data_1d',
    start_offset => INTERVAL '3 days',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- Create helper functions

-- Function to get latest price
CREATE OR REPLACE FUNCTION get_latest_price(p_symbol TEXT)
RETURNS DOUBLE PRECISION AS $$
DECLARE
    latest_price DOUBLE PRECISION;
BEGIN
    SELECT price INTO latest_price
    FROM market_data
    WHERE symbol = p_symbol
    ORDER BY time DESC
    LIMIT 1;

    RETURN COALESCE(latest_price, 0);
END;
$$ LANGUAGE plpgsql;

-- Function to calculate price change percentage
CREATE OR REPLACE FUNCTION calculate_price_change(p_symbol TEXT, p_hours INTEGER)
RETURNS DOUBLE PRECISION AS $$
DECLARE
    current_price DOUBLE PRECISION;
    past_price DOUBLE PRECISION;
    change_percent DOUBLE PRECISION;
BEGIN
    -- Get current price
    SELECT price INTO current_price
    FROM market_data
    WHERE symbol = p_symbol
    ORDER BY time DESC
    LIMIT 1;

    -- Get price from N hours ago
    SELECT price INTO past_price
    FROM market_data
    WHERE symbol = p_symbol
      AND time <= NOW() - (p_hours || ' hours')::INTERVAL
    ORDER BY time DESC
    LIMIT 1;

    -- Calculate percentage change
    IF past_price > 0 THEN
        change_percent := ((current_price - past_price) / past_price) * 100;
    ELSE
        change_percent := 0;
    END IF;

    RETURN COALESCE(change_percent, 0);
END;
$$ LANGUAGE plpgsql;

-- Function to get indicator signal strength
CREATE OR REPLACE FUNCTION get_indicator_consensus(p_symbol TEXT, p_hours INTEGER DEFAULT 24)
RETURNS TABLE(buy_signals INTEGER, sell_signals INTEGER, hold_signals INTEGER) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COUNT(*) FILTER (WHERE signal = 'BUY') AS buy_signals,
        COUNT(*) FILTER (WHERE signal = 'SELL') AS sell_signals,
        COUNT(*) FILTER (WHERE signal = 'NEUTRAL') AS hold_signals
    FROM technical_indicators
    WHERE symbol = p_symbol
      AND time >= NOW() - (p_hours || ' hours')::INTERVAL;
END;
$$ LANGUAGE plpgsql;
