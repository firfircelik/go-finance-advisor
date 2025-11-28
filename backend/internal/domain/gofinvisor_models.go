package domain

import (
	"time"
)

type InstrumentType string

const (
	InstrumentTypeStock  InstrumentType = "STOCK"
	InstrumentTypeCrypto InstrumentType = "CRYPTO"
	InstrumentTypeForex  InstrumentType = "FOREX"
)

type FinancialInstrument struct {
	ID        int64          `json:"id"`
	Symbol    string         `json:"symbol"`
	Type      InstrumentType `json:"type"`
	Name      string         `json:"name"`
	CreatedAt time.Time      `json:"created_at"`
}

type MarketData struct {
	ID           int64     `json:"id"`
	InstrumentID int64     `json:"instrument_id"`
	Symbol       string    `json:"symbol"` // Denormalized for easier querying
	Date         time.Time `json:"date"`
	Open         float64   `json:"open"`
	High         float64   `json:"high"`
	Low          float64   `json:"low"`
	Close        float64   `json:"close"`
	Volume       float64   `json:"volume"`
	CreatedAt    time.Time `json:"created_at"`
}
