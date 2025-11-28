package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"

	"go-finance-advisor/internal/domain"
	"go-finance-advisor/internal/infrastructure/market_data/binance"
	"go-finance-advisor/internal/infrastructure/market_data/yahoo"
	"go-finance-advisor/internal/infrastructure/persistence/postgres"
)

func main() {
	log.Println("Starting Market Data Backfill...")

	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5433/finance?sslmode=disable" // Default to localhost port mapped in docker-compose
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize components
	repo := postgres.NewPortfolioRepository(db)
	stockFetcher := yahoo.NewStockFetcher()
	cryptoFetcher := binance.NewCryptoFetcher()

	// Initialize Ticker Fetcher
	tickerFetcher := NewTickerFetcher()

	// Fetch Stock Tickers
	log.Println("Fetching BIST (Turkish) tickers...")
	bistTickers, err := tickerFetcher.GetBISTTickers()
	if err != nil {
		log.Printf("Error fetching BIST tickers: %v", err)
		// Fallback to a few if failed
		bistTickers = []string{"THYAO.IS", "GARAN.IS", "AKBNK.IS"}
	}
	log.Printf("Found %d BIST tickers.", len(bistTickers))

	log.Println("Fetching US (NASDAQ/NYSE) tickers...")
	usTickers, err := tickerFetcher.GetUSTickers()
	if err != nil {
		log.Printf("Error fetching US tickers: %v", err)
		// Fallback
		usTickers = []string{"AAPL", "MSFT", "GOOGL", "SPY"}
	}
	log.Printf("Found %d US tickers.", len(usTickers))

	// Combine lists
	// For MVP, we might want to limit the total number to avoid running for days
	// But user asked for "all". We will prioritize BIST then US.
	stocks := append(bistTickers, usTickers...)

	// Limit to top 2000 for now to be realistic about execution time in this session?
	// Or just let it run. Let's let it run but maybe shuffle or prioritize?
	// We'll just run it.

	endDate := time.Now()
	startDate := endDate.AddDate(-5, 0, 0) // Last 5 years

	// Backfill Stocks
	log.Printf("Starting Stock Backfill for %d symbols...", len(stocks))
	for i, symbol := range stocks {
		log.Printf("[%d/%d] Backfilling stock: %s", i+1, len(stocks), symbol)

		if err := repo.EnsureInstrument(symbol, domain.InstrumentTypeStock, symbol); err != nil {
			log.Printf("Error ensuring instrument %s: %v", symbol, err)
			continue
		}

		data, err := stockFetcher.FetchHistoricalData(symbol, startDate, endDate)
		if err != nil {
			log.Printf("Error fetching data for %s: %v", symbol, err)
			continue
		}

		if err := repo.SaveMarketData(data); err != nil {
			log.Printf("Error saving data for %s: %v", symbol, err)
			continue
		}

		log.Printf("Successfully backfilled %d records for %s", len(data), symbol)

		// Rate limiting delay between stocks
		// Yahoo is sensitive, so we wait a bit.
		time.Sleep(2 * time.Second)
	}

	// Backfill Crypto
	log.Println("Fetching all active USDT trading pairs from Binance...")
	cryptos, err := cryptoFetcher.FetchAllUSDTPairs()
	if err != nil {
		log.Fatalf("Failed to fetch crypto pairs: %v", err)
	}
	log.Printf("Found %d crypto pairs to backfill.", len(cryptos))

	for i, symbol := range cryptos {
		log.Printf("[%d/%d] Backfilling crypto: %s", i+1, len(cryptos), symbol)

		if err := repo.EnsureInstrument(symbol, domain.InstrumentTypeCrypto, symbol); err != nil {
			log.Printf("Error ensuring instrument %s: %v", symbol, err)
			continue
		}

		data, err := cryptoFetcher.FetchHistoricalData(symbol, startDate, endDate)
		if err != nil {
			log.Printf("Error fetching data for %s: %v", symbol, err)
			continue
		}

		if err := repo.SaveMarketData(data); err != nil {
			log.Printf("Error saving data for %s: %v", symbol, err)
			continue
		}

		log.Printf("Successfully backfilled %d records for %s", len(data), symbol)

		// Binance is more generous, but let's be polite
		time.Sleep(200 * time.Millisecond)
	}

	log.Println("Backfill completed successfully!")
}
