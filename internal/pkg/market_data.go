// Package pkg provides utility functions for market data and financial calculations.
package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// MarketData represents current market prices
type MarketData struct {
	BitcoinPrice float64 `json:"bitcoin"`
	SP500Price   float64 `json:"sp500"`
}

// FetchMarketData fetches real-time market data from APIs
func FetchMarketData() (*MarketData, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	ctx := context.Background()

	// Default fallback values
	bitcoinPrice := 45000.0
	sp500Price := 4500.0

	// 1. Bitcoin Price (CoinGecko)
	btcURL := "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd"
	btcReq, _ := http.NewRequestWithContext(ctx, "GET", btcURL, http.NoBody)
	btcResp, err := client.Do(btcReq)
	if err == nil && btcResp.StatusCode == http.StatusOK {
		defer func() {
			if closeErr := btcResp.Body.Close(); closeErr != nil {
				// Log error if needed, but don't fail the function
			}
		}()
		var btcData map[string]map[string]float64
		if decodeErr := json.NewDecoder(btcResp.Body).Decode(&btcData); decodeErr == nil {
			if bitcoin, ok := btcData["bitcoin"]; ok {
				if price, ok := bitcoin["usd"]; ok {
					bitcoinPrice = price
				}
			}
		}
	}

	// 2. S&P 500 Price (Alpha Vantage)
	// Using demo key, real usage requires API key
	sp500URL := "https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=SPX&apikey=demo"
	sp500Req, _ := http.NewRequestWithContext(ctx, "GET", sp500URL, http.NoBody)
	sp500Resp, err := client.Do(sp500Req)
	if err == nil && sp500Resp.StatusCode == http.StatusOK {
		defer func() {
			if closeErr := sp500Resp.Body.Close(); closeErr != nil {
				// Log error if needed, but don't fail the function
			}
		}()
		var sp500Data map[string]map[string]string
		if err := json.NewDecoder(sp500Resp.Body).Decode(&sp500Data); err == nil {
			if globalQuote, ok := sp500Data["Global Quote"]; ok {
				if priceStr, ok := globalQuote["05. price"]; ok {
					if _, scanErr := fmt.Sscanf(priceStr, "%f", &sp500Price); scanErr != nil {
						// Log error if needed, but don't fail the function
					}
				}
			}
		}
	}

	return &MarketData{
		BitcoinPrice: bitcoinPrice,
		SP500Price:   sp500Price,
	}, nil
}
