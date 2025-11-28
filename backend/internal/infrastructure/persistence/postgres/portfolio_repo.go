package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"go-finance-advisor/internal/domain"
)

type PortfolioRepository struct {
	db *sql.DB
}

func NewPortfolioRepository(db *sql.DB) *PortfolioRepository {
	return &PortfolioRepository{db: db}
}

func (r *PortfolioRepository) SaveMarketData(data []domain.MarketData) error {
	if len(data) == 0 {
		return nil
	}

	// Batch insert
	// INSERT INTO market_data (symbol, date, open, high, low, close, volume) VALUES ...
	// ON CONFLICT (symbol, date) DO UPDATE SET ...

	query := "INSERT INTO market_data (symbol, date, open, high, low, close, volume) VALUES "
	values := []interface{}{}
	placeholders := []string{}

	for i, md := range data {
		offset := i * 7
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)", 
			offset+1, offset+2, offset+3, offset+4, offset+5, offset+6, offset+7))
		
		values = append(values, md.Symbol, md.Date, md.Open, md.High, md.Low, md.Close, md.Volume)
	}

	query += strings.Join(placeholders, ",")
	query += ` ON CONFLICT (symbol, date) DO UPDATE SET 
		open = EXCLUDED.open,
		high = EXCLUDED.high,
		low = EXCLUDED.low,
		close = EXCLUDED.close,
		volume = EXCLUDED.volume;`

	_, err := r.db.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("failed to batch insert market data: %w", err)
	}

	return nil
}

func (r *PortfolioRepository) EnsureInstrument(symbol string, instrumentType domain.InstrumentType, name string) error {
	query := `
		INSERT INTO financial_instruments (symbol, type, name)
		VALUES ($1, $2, $3)
		ON CONFLICT (symbol) DO NOTHING;
	`
	_, err := r.db.Exec(query, symbol, instrumentType, name)
	if err != nil {
		return fmt.Errorf("failed to ensure instrument %s: %w", symbol, err)
	}
	return nil
}
