package pkg

import (
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

	// Default fallback values
	bitcoinPrice := 45000.0
	sp500Price := 4500.0

	// 1. Bitcoin Price (CoinGecko)
	btcResp, err := client.Get("https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd")
	if err == nil && btcResp.StatusCode == http.StatusOK {
		defer btcResp.Body.Close()
		var btcData map[string]map[string]float64
		if err := json.NewDecoder(btcResp.Body).Decode(&btcData); err == nil {
			if bitcoin, ok := btcData["bitcoin"]; ok {
				if price, ok := bitcoin["usd"]; ok {
					bitcoinPrice = price
				}
			}
		}
	}

	// 2. S&P 500 Price (Alpha Vantage)
	// Using demo key, real usage requires API key
	sp500Resp, err := client.Get("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=SPX&apikey=demo")
	if err == nil && sp500Resp.StatusCode == http.StatusOK {
		defer sp500Resp.Body.Close()
		var sp500Data map[string]map[string]string
		if err := json.NewDecoder(sp500Resp.Body).Decode(&sp500Data); err == nil {
			if globalQuote, ok := sp500Data["Global Quote"]; ok {
				if priceStr, ok := globalQuote["05. price"]; ok {
					fmt.Sscanf(priceStr, "%f", &sp500Price)
				}
			}
		}
	}

	return &MarketData{
		BitcoinPrice: bitcoinPrice,
		SP500Price:   sp500Price,
	}, nil
}
