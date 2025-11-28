-- GoFinVisor Initial Schema Migration
-- Creates all PostgreSQL tables for the application

-- Financial Profiles Table
CREATE TABLE IF NOT EXISTS financial_profiles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE,
    monthly_income DOUBLE PRECISION DEFAULT 0,
    fixed_expenses DOUBLE PRECISION DEFAULT 0,
    investment_ratio DOUBLE PRECISION DEFAULT 0.1,
    investment_amount DOUBLE PRECISION DEFAULT 0,
    use_ratio BOOLEAN DEFAULT true,
    risk_tolerance VARCHAR(20) DEFAULT 'moderate',
    investment_goals TEXT,
    time_horizon INTEGER DEFAULT 12,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_financial_profiles_user_id ON financial_profiles(user_id);
CREATE INDEX idx_financial_profiles_risk_tolerance ON financial_profiles(risk_tolerance);

-- Portfolios Table
CREATE TABLE IF NOT EXISTS portfolios (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    total_value DOUBLE PRECISION DEFAULT 0,
    currency VARCHAR(10) DEFAULT 'USD',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_portfolios_user_id ON portfolios(user_id);
CREATE INDEX idx_portfolios_name ON portfolios(name);

-- Holdings Table
CREATE TABLE IF NOT EXISTS holdings (
    id SERIAL PRIMARY KEY,
    portfolio_id INTEGER NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    asset_type VARCHAR(20) NOT NULL,
    quantity DOUBLE PRECISION NOT NULL,
    purchase_price DOUBLE PRECISION NOT NULL,
    current_price DOUBLE PRECISION DEFAULT 0,
    total_value DOUBLE PRECISION DEFAULT 0,
    purchase_date TIMESTAMP WITH TIME ZONE,
    unrealized_pnl DOUBLE PRECISION DEFAULT 0,
    unrealized_pnl_percent DOUBLE PRECISION DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id) ON DELETE CASCADE
);

CREATE INDEX idx_holdings_portfolio_id ON holdings(portfolio_id);
CREATE INDEX idx_holdings_symbol ON holdings(symbol);
CREATE INDEX idx_holdings_asset_type ON holdings(asset_type);
CREATE UNIQUE INDEX idx_holdings_portfolio_symbol ON holdings(portfolio_id, symbol);

-- Signals Table
CREATE TABLE IF NOT EXISTS signals (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    asset_type VARCHAR(20) NOT NULL,
    action VARCHAR(10) NOT NULL,
    confidence DOUBLE PRECISION NOT NULL,
    reasoning TEXT,
    indicators JSONB,
    entry_price DOUBLE PRECISION,
    target_price DOUBLE PRECISION,
    stop_loss DOUBLE PRECISION,
    valid_until TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_signals_symbol ON signals(symbol);
CREATE INDEX idx_signals_action ON signals(action);
CREATE INDEX idx_signals_confidence ON signals(confidence DESC);
CREATE INDEX idx_signals_is_active ON signals(is_active);
CREATE INDEX idx_signals_valid_until ON signals(valid_until);
CREATE INDEX idx_signals_created_at ON signals(created_at DESC);

-- Allocation Plans Table
CREATE TABLE IF NOT EXISTS allocation_plans (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    total_budget DOUBLE PRECISION NOT NULL,
    risk_level VARCHAR(20) NOT NULL,
    strategy VARCHAR(50),
    is_executed BOOLEAN DEFAULT false,
    executed_at TIMESTAMP WITH TIME ZONE,
    performance_score DOUBLE PRECISION DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_allocation_plans_user_id ON allocation_plans(user_id);
CREATE INDEX idx_allocation_plans_is_executed ON allocation_plans(is_executed);
CREATE INDEX idx_allocation_plans_created_at ON allocation_plans(created_at DESC);

-- Allocations Table (part of allocation plan)
CREATE TABLE IF NOT EXISTS allocations (
    id SERIAL PRIMARY KEY,
    plan_id INTEGER NOT NULL,
    signal_id INTEGER,
    symbol VARCHAR(20) NOT NULL,
    asset_type VARCHAR(20) NOT NULL,
    allocated_amount DOUBLE PRECISION NOT NULL,
    percentage DOUBLE PRECISION NOT NULL,
    expected_return DOUBLE PRECISION,
    reasoning TEXT,
    priority INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (plan_id) REFERENCES allocation_plans(id) ON DELETE CASCADE,
    FOREIGN KEY (signal_id) REFERENCES signals(id) ON DELETE SET NULL
);

CREATE INDEX idx_allocations_plan_id ON allocations(plan_id);
CREATE INDEX idx_allocations_signal_id ON allocations(signal_id);
CREATE INDEX idx_allocations_symbol ON allocations(symbol);

-- Price Alerts Table
CREATE TABLE IF NOT EXISTS price_alerts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    asset_type VARCHAR(20) NOT NULL,
    alert_type VARCHAR(20) NOT NULL,
    target_price DOUBLE PRECISION NOT NULL,
    current_price DOUBLE PRECISION,
    condition VARCHAR(20) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    is_triggered BOOLEAN DEFAULT false,
    triggered_at TIMESTAMP WITH TIME ZONE,
    notification_sent BOOLEAN DEFAULT false,
    message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_price_alerts_user_id ON price_alerts(user_id);
CREATE INDEX idx_price_alerts_symbol ON price_alerts(symbol);
CREATE INDEX idx_price_alerts_is_active ON price_alerts(is_active);
CREATE INDEX idx_price_alerts_is_triggered ON price_alerts(is_triggered);

-- Notifications Table
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(200) NOT NULL,
    message TEXT NOT NULL,
    severity VARCHAR(20) DEFAULT 'info',
    is_read BOOLEAN DEFAULT false,
    related_entity_type VARCHAR(50),
    related_entity_id INTEGER,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    read_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_is_read ON notifications(is_read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX idx_notifications_type ON notifications(type);

-- Add trigger to update updated_at timestamp automatically
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_financial_profiles_updated_at BEFORE UPDATE ON financial_profiles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_portfolios_updated_at BEFORE UPDATE ON portfolios FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_holdings_updated_at BEFORE UPDATE ON holdings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_signals_updated_at BEFORE UPDATE ON signals FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_allocation_plans_updated_at BEFORE UPDATE ON allocation_plans FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_price_alerts_updated_at BEFORE UPDATE ON price_alerts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
