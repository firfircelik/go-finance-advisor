package binance

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-finance-advisor/internal/domain"
)

type CryptoFetcher struct {
	client *http.Client
}

func NewCryptoFetcher() *CryptoFetcher {
	return &CryptoFetcher{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Binance Klines API returns array of arrays:
// [
//   [
//     1499040000000,      // Open time
//     "0.01634790",       // Open
//     "0.80000000",       // High
//     "0.01575800",       // Low
//     "0.01577100",       // Close
//     "148976.11427815",  // Volume
//     1499644799999,      // Close time
//     "2434.19055334",    // Quote asset volume
//     308,                // Number of trades
//     "1756.87402397",    // Taker buy base asset volume
//     "28.46694368",      // Taker buy quote asset volume
//     "17928899.62484339" // Ignore.
//   ]
// ]
func (f *CryptoFetcher) FetchHistoricalData(symbol string, start, end time.Time) ([]domain.MarketData, error) {
	// Binance API requires limits or intervals. For simplicity, we fetch 1d interval.
	// Note: Binance API has limits on how much data can be fetched in one go (default 500, max 1000).
	// For a full backfill, we might need pagination.
	// For this MVP, we will fetch the last 1000 days if the range allows, or just use the start/end params.
	
	startTime := start.UnixMilli()
	endTime := end.UnixMilli()
	
	// Ensure symbol is uppercase
	symbol = strings.ToUpper(symbol)

	url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=1d&startTime=%d&endTime=%d&limit=1000", symbol, startTime, endTime)

	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data for %s: %w", symbol, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("binance api error: %s - %s", resp.Status, string(body))
	}

	var rawData [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var marketData []domain.MarketData
	for _, kline := range rawData {
		if len(kline) < 6 {
			continue
		}

		// Parse timestamp
		tsFloat, ok := kline[0].(float64)
		if !ok {
			continue
		}
		date := time.UnixMilli(int64(tsFloat))

		// Parse prices (strings)
		open, _ := strconv.ParseFloat(kline[1].(string), 64)
		high, _ := strconv.ParseFloat(kline[2].(string), 64)
		low, _ := strconv.ParseFloat(kline[3].(string), 64)
		closePrice, _ := strconv.ParseFloat(kline[4].(string), 64)
		volume, _ := strconv.ParseFloat(kline[5].(string), 64)

		md := domain.MarketData{
			Symbol: symbol,
			Date:   date,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  closePrice,
			Volume: volume,
		}
		marketData = append(marketData, md)
	}

	return marketData, nil
}

type BinanceExchangeInfo struct {
	Symbols []struct {
		Symbol     string `json:"symbol"`
		Status     string `json:"status"`
		QuoteAsset string `json:"quoteAsset"`
	} `json:"symbols"`
}

func (f *CryptoFetcher) FetchAllUSDTPairs() ([]string, error) {
	url := "https://api.binance.com/api/v3/exchangeInfo"
	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exchange info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance api error: %s", resp.Status)
	}

	var info BinanceExchangeInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode exchange info: %w", err)
	}

	var pairs []string
	for _, s := range info.Symbols {
		if s.Status == "TRADING" && s.QuoteAsset == "USDT" {
			pairs = append(pairs, s.Symbol)
		}
	}

	return pairs, nil
}
