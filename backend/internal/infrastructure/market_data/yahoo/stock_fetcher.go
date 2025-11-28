package yahoo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go-finance-advisor/internal/domain"
)

type StockFetcher struct {
	client *http.Client
}

func NewStockFetcher() *StockFetcher {
	return &StockFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Yahoo Finance Chart API Response Structure
type yahooChartResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Currency             string  `json:"currency"`
				Symbol               string  `json:"symbol"`
				ExchangeName         string  `json:"exchangeName"`
				InstrumentType       string  `json:"instrumentType"`
				FirstTradeDate       int64   `json:"firstTradeDate"`
				RegularMarketTime    int64   `json:"regularMarketTime"`
				Gmtoffset            int     `json:"gmtoffset"`
				Timezone             string  `json:"timezone"`
				ExchangeTimezoneName string  `json:"exchangeTimezoneName"`
				RegularMarketPrice   float64 `json:"regularMarketPrice"`
				ChartPreviousClose   float64 `json:"chartPreviousClose"`
				PriceHint            int     `json:"priceHint"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []*float64 `json:"open"`
					Low    []*float64 `json:"low"`
					High   []*float64 `json:"high"`
					Close  []*float64 `json:"close"`
					Volume []*int64   `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error interface{} `json:"error"`
	} `json:"chart"`
}

func (f *StockFetcher) FetchHistoricalData(symbol string, startDate, endDate time.Time) ([]domain.MarketData, error) {
	period1 := startDate.Unix()
	period2 := endDate.Unix()

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?period1=%d&period2=%d&interval=1d", symbol, period1, period2)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Yahoo Finance requires User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("yahoo api returned status: %d, body: %s", resp.StatusCode, string(body))
	}

	var chartResp yahooChartResponse
	if err := json.NewDecoder(resp.Body).Decode(&chartResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(chartResp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data found for symbol: %s", symbol)
	}

	result := chartResp.Chart.Result[0]
	timestamps := result.Timestamp
	quote := result.Indicators.Quote[0]

	var marketData []domain.MarketData

	for i, ts := range timestamps {
		// Skip if any data point is missing (can happen with Yahoo)
		if quote.Open[i] == nil || quote.High[i] == nil || quote.Low[i] == nil || quote.Close[i] == nil || quote.Volume[i] == nil {
			continue
		}

		md := domain.MarketData{
			Symbol: symbol,
			Date:   time.Unix(ts, 0),
			Open:   *quote.Open[i],
			High:   *quote.High[i],
			Low:    *quote.Low[i],
			Close:  *quote.Close[i],
			Volume: float64(*quote.Volume[i]),
		}
		marketData = append(marketData, md)
	}

	return marketData, nil
}
