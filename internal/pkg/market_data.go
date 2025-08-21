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

	// 1. Bitcoin Price (CoinGecko)
	btcResp, err := client.Get("https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bitcoin data: %v", err)
	}
	defer btcResp.Body.Close()

	var btcData map[string]map[string]float64
	if err := json.NewDecoder(btcResp.Body).Decode(&btcData); err != nil {
		return nil, fmt.Errorf("failed to parse bitcoin data: %v", err)
	}

	// 2. S&P 500 Price (Alpha Vantage)
	// Using demo key, real usage requires API key
	sp500Resp, err := client.Get("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=SPX&apikey=demo")
	if err != nil || sp500Resp.StatusCode != http.StatusOK {
		// Fallback: Sabit değerler
		return &MarketData{
			BitcoinPrice: btcData["bitcoin"]["usd"],
			SP500Price:   4500.0,
		}, nil
	}
	defer sp500Resp.Body.Close()

	var sp500Data map[string]map[string]string
	if err := json.NewDecoder(sp500Resp.Body).Decode(&sp500Data); err != nil {
		return &MarketData{
			BitcoinPrice: btcData["bitcoin"]["usd"],
			SP500Price:   4500.0,
		}, nil
	}

	// Alpha Vantage'tan gerçek veri
	sp500Price := 4500.0
	if priceStr, ok := sp500Data["Global Quote"]["05. price"]; ok {
		fmt.Sscanf(priceStr, "%f", &sp500Price)
	}

	return &MarketData{
		BitcoinPrice: btcData["bitcoin"]["usd"],
		SP500Price:   sp500Price,
	}, nil
}
