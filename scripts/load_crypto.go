package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

const (
	dbURL = "postgres://user:password@localhost:5433/finance?sslmode=disable"
)

type CryptoData struct {
	Prices [][]float64 `json:"prices"`
}

func main() {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS market_data (
			symbol VARCHAR(20),
			date DATE,
			close_price DECIMAL(15,2),
			volume BIGINT DEFAULT 0,
			created_at TIMESTAMP DEFAULT NOW(),
			PRIMARY KEY (symbol, date)
		)
	`)
	if err != nil {
		log.Fatal("Create table error:", err)
	}

	// Top cryptocurrencies
	cryptos := []struct {
		id     string
		symbol string
	}{
		{"bitcoin", "BTC"},
		{"ethereum", "ETH"},
		{"tether", "USDT"},
		{"binancecoin", "BNB"},
		{"solana", "SOL"},
		{"ripple", "XRP"},
		{"usd-coin", "USDC"},
		{"cardano", "ADA"},
		{"dogecoin", "DOGE"},
		{"tron", "TRX"},
	}

	fmt.Println("📊 Loading cryptocurrency data...")
	for i, crypto := range cryptos {
		fmt.Printf("[%d/%d] Loading %s (%s)...\n", i+1, len(cryptos), crypto.symbol, crypto.id)

		// CoinGecko API - free, no key required
		url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s/market_chart?vs_currency=usd&days=365", crypto.id)

		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error fetching %s: %v", crypto.symbol, err)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var data CryptoData
		if err := json.Unmarshal(body, &data); err != nil {
			log.Printf("Error parsing %s: %v", crypto.symbol, err)
			continue
		}

		// Insert data
		count := 0
		for _, price := range data.Prices {
			timestamp := int64(price[0]) / 1000
			date := time.Unix(timestamp, 0).Format("2006-01-02")
			closePrice := price[1]

			_, err := db.Exec(`
				INSERT INTO market_data (symbol, date, close_price)
				VALUES ($1, $2, $3)
				ON CONFLICT (symbol, date) DO UPDATE 
				SET close_price = EXCLUDED.close_price
			`, crypto.symbol, date, closePrice)

			if err == nil {
				count++
			}
		}

		fmt.Printf("  ✓ Loaded %d records for %s\n", count, crypto.symbol)

		// Rate limiting - CoinGecko allows ~10-50 calls/min
		if i < len(cryptos)-1 {
			time.Sleep(2 * time.Second)
		}
	}

	// Check results
	var total int
	db.QueryRow("SELECT COUNT(*) FROM market_data").Scan(&total)
	fmt.Printf("\n✅ Complete! Total records: %d\n", total)
}
