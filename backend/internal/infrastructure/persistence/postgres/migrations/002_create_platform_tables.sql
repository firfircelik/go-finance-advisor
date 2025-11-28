CREATE TABLE IF NOT EXISTS users (
    email VARCHAR(255) PRIMARY KEY,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    token VARCHAR(512) PRIMARY KEY,
    email VARCHAR(255) REFERENCES users(email),
    expires_at BIGINT NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS holdings (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) REFERENCES users(email),
    asset VARCHAR(255) NOT NULL,
    price DECIMAL(20, 8),
    holdings VARCHAR(255), -- Stored as string to handle various formats or just display
    value DECIMAL(20, 8),
    change_pct DECIMAL(10, 2),
    icon VARCHAR(50),
    icon_color VARCHAR(20),
    category VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transactions (
    id VARCHAR(255) PRIMARY KEY, -- Using string ID as per handlers
    email VARCHAR(255) REFERENCES users(email),
    date VARCHAR(50), -- Storing as string for simplicity as per handlers
    description VARCHAR(255),
    amount DECIMAL(20, 8),
    category VARCHAR(50),
    type VARCHAR(20), -- 'income' or 'expense'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS budgets (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) REFERENCES users(email),
    category VARCHAR(50),
    budgeted DECIMAL(20, 8),
    spent DECIMAL(20, 8),
    UNIQUE(email, category)
);

CREATE TABLE IF NOT EXISTS goals (
    id VARCHAR(255) PRIMARY KEY,
    email VARCHAR(255) REFERENCES users(email),
    name VARCHAR(255),
    target_amount DECIMAL(20, 8),
    current_amount DECIMAL(20, 8),
    deadline VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS performance (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) REFERENCES users(email),
    x VARCHAR(50), -- Date/Label
    y DECIMAL(20, 8), -- Value
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS liabilities (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) REFERENCES users(email),
    name VARCHAR(255),
    amount DECIMAL(20, 8),
    interest_rate DECIMAL(5, 2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Ticker table for the scrolling ticker
CREATE TABLE IF NOT EXISTS ticker (
    symbol VARCHAR(20) PRIMARY KEY,
    price DECIMAL(20, 8),
    change DECIMAL(10, 2)
);
