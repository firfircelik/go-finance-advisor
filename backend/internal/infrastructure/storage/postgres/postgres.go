package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

func New(dsn string) (*sql.DB, error) {
	var db *sql.DB
	var err error

	// Determine if it's SQLite or Postgres based on DSN
	if strings.HasPrefix(dsn, "postgres://") || strings.Contains(dsn, "@") {
		// Postgres connection
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			return nil, err
		}

		// Retry connection logic for Docker startup
		for i := 0; i < 10; i++ {
			if err := db.Ping(); err == nil {
				break
			}
			log.Printf("Waiting for database... (%d/10)", i+1)
			time.Sleep(1 * time.Second)
		}

		if err := db.Ping(); err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
	} else {
		// SQLite connection
		db, err = sql.Open("sqlite", dsn)
		if err != nil {
			return nil, err
		}
		log.Printf("Using SQLite database: %s", dsn)
	}

	if err := applyMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}
	return db, nil
}

func applyMigrations(db *sql.DB) error {
	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS schema_migrations(version INTEGER PRIMARY KEY)"); err != nil {
		return err
	}
	var cur int
	_ = db.QueryRow("SELECT COALESCE(MAX(version),0) FROM schema_migrations").Scan(&cur)
	type Mig struct {
		v   int
		sql string
	}
	// Adapted for PostgreSQL
	migs := []Mig{
		{1, "CREATE TABLE IF NOT EXISTS users(email TEXT PRIMARY KEY, password_hash TEXT, name TEXT)"},
		{2, "CREATE TABLE IF NOT EXISTS holdings(asset TEXT, price DOUBLE PRECISION, holdings TEXT, value DOUBLE PRECISION, change_pct DOUBLE PRECISION, icon TEXT, icon_color TEXT, category TEXT)"},
		{3, "CREATE TABLE IF NOT EXISTS performance(x DOUBLE PRECISION, y DOUBLE PRECISION)"},
		{4, "CREATE TABLE IF NOT EXISTS ticker(symbol TEXT PRIMARY KEY, price DOUBLE PRECISION, change DOUBLE PRECISION)"},
		{5, "CREATE TABLE IF NOT EXISTS liabilities(name TEXT, amount DOUBLE PRECISION)"},
		{6, "CREATE TABLE IF NOT EXISTS refresh_tokens(token TEXT PRIMARY KEY, email TEXT, expires_at BIGINT, revoked INTEGER)"},
		{7, "CREATE TABLE IF NOT EXISTS transactions(id TEXT PRIMARY KEY, email TEXT, date TEXT, description TEXT, amount DOUBLE PRECISION, category TEXT, type TEXT)"},
		{8, "CREATE TABLE IF NOT EXISTS budgets(category TEXT PRIMARY KEY, budgeted DOUBLE PRECISION, spent DOUBLE PRECISION)"},
		{9, "CREATE TABLE IF NOT EXISTS goals(id TEXT PRIMARY KEY, email TEXT, name TEXT, target_amount DOUBLE PRECISION, current_amount DOUBLE PRECISION, deadline TEXT)"},
		{10, "CREATE TABLE IF NOT EXISTS documents(id TEXT PRIMARY KEY, email TEXT, filename TEXT, upload_date TEXT, size BIGINT)"},
		// Performance indexes for user-based queries
		{11, "CREATE INDEX IF NOT EXISTS idx_transactions_email ON transactions(email)"},
		{12, "CREATE INDEX IF NOT EXISTS idx_goals_email ON goals(email)"},
		{13, "CREATE INDEX IF NOT EXISTS idx_refresh_tokens_email ON refresh_tokens(email)"},
		{14, "CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at)"},
		{15, "CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(date)"},
		{16, "CREATE INDEX IF NOT EXISTS idx_transactions_category ON transactions(category)"},
		// Make holdings user-scoped
		{17, "ALTER TABLE holdings ADD COLUMN IF NOT EXISTS email TEXT"},
		{18, "CREATE INDEX IF NOT EXISTS idx_holdings_email ON holdings(email)"},
		// Make performance user-scoped
		{19, "ALTER TABLE performance ADD COLUMN IF NOT EXISTS email TEXT"},
		{20, "CREATE INDEX IF NOT EXISTS idx_performance_email ON performance(email)"},
		// Make liabilities user-scoped
		{21, "ALTER TABLE liabilities ADD COLUMN IF NOT EXISTS email TEXT"},
		{22, "CREATE INDEX IF NOT EXISTS idx_liabilities_email ON liabilities(email)"},
		// Make budgets user-scoped
		{23, "ALTER TABLE budgets ADD COLUMN IF NOT EXISTS email TEXT"},
		{24, "CREATE INDEX IF NOT EXISTS idx_budgets_email ON budgets(email)"},
		// Update primary keys for user-scoped tables - Postgres approach differs from SQLite
		// We will create new tables and migrate data if needed, but for now assuming fresh start or compatible schema
		// Postgres supports composite primary keys directly
		{25, "DROP TABLE IF EXISTS holdings_new"}, // Cleanup from previous attempts
		{26, "CREATE TABLE IF NOT EXISTS holdings_new(email TEXT, asset TEXT, price DOUBLE PRECISION, holdings TEXT, value DOUBLE PRECISION, change_pct DOUBLE PRECISION, icon TEXT, icon_color TEXT, category TEXT, PRIMARY KEY(email, asset))"},
		{27, "INSERT INTO holdings_new(email, asset, price, holdings, value, change_pct, icon, icon_color, category) SELECT email, asset, price, holdings, value, change_pct, icon, icon_color, category FROM holdings ON CONFLICT DO NOTHING"},
		{28, "DROP TABLE holdings"},
		{29, "ALTER TABLE holdings_new RENAME TO holdings"},

		{30, "DROP TABLE IF EXISTS budgets_new"},
		{31, "CREATE TABLE IF NOT EXISTS budgets_new(email TEXT, category TEXT, budgeted DOUBLE PRECISION, spent DOUBLE PRECISION, PRIMARY KEY(email, category))"},
		{32, "INSERT INTO budgets_new(email, category, budgeted, spent) SELECT email, category, budgeted, spent FROM budgets ON CONFLICT DO NOTHING"},
		{33, "DROP TABLE budgets"},
		{34, "ALTER TABLE budgets_new RENAME TO budgets"},
	}
	for _, m := range migs {
		if m.v > cur {
			if _, err := db.Exec(m.sql); err != nil {
				return fmt.Errorf("migration %d failed: %w", m.v, err)
			}
			if _, err := db.Exec("INSERT INTO schema_migrations(version) VALUES($1)", m.v); err != nil {
				return fmt.Errorf("failed to record migration %d: %w", m.v, err)
			}
		}
	}
	return nil
}
