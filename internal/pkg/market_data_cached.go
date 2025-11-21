package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go-finance-advisor/internal/infrastructure/cache"

	"github.com/cenkalti/backoff/v4"
)

// CachedMarketService extends RealTimeMarketService with caching and retry logic
type CachedMarketService struct {
	*RealTimeMarketService
	cache *cache.RedisCache
}

// NewCachedMarketService creates a new market service with caching support
func NewCachedMarketService(redisCache *cache.RedisCache) *CachedMarketService {
	return &CachedMarketService{
		RealTimeMarketService: NewRealTimeMarketService(),
		cache:                 redisCache,
	}
}

// GetCryptoPricesWithCache fetches crypto prices with caching and retry logic
func (s *CachedMarketService) GetCryptoPricesWithCache(ctx context.Context) ([]CryptoPrice, error) {
	cacheKey := "market:crypto:prices"
	cacheTTL := 1 * time.Minute

	// Try to get from cache first
	if s.cache != nil && s.cache.IsEnabled() {
		var cachedPrices []CryptoPrice
		if err := s.cache.Get(ctx, cacheKey, &cachedPrices); err == nil {
			log.Println("Cache hit for crypto prices")
			return cachedPrices, nil
		}
		log.Println("Cache miss for crypto prices")
	}

	// If not in cache, fetch with retry logic
	var prices []CryptoPrice
	operation := func() error {
		var err error
		prices, err = s.fetchCryptoPricesFromAPI(ctx)
		return err
	}

	// Configure exponential backoff
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 30 * time.Second
	b.InitialInterval = 500 * time.Millisecond
	b.MaxInterval = 5 * time.Second

	if err := backoff.Retry(operation, backoff.WithContext(b, ctx)); err != nil {
		return nil, fmt.Errorf("failed to fetch crypto prices after retries: %w", err)
	}

	// Cache the result
	if s.cache != nil && s.cache.IsEnabled() {
		if err := s.cache.Set(ctx, cacheKey, prices, cacheTTL); err != nil {
			log.Printf("Failed to cache crypto prices: %v", err)
		}
	}

	return prices, nil
}

// GetStockPricesWithCache fetches stock prices with caching and retry logic
func (s *CachedMarketService) GetStockPricesWithCache(ctx context.Context, symbols []string) ([]StockPrice, error) {
	cacheKey := fmt.Sprintf("market:stocks:prices:%v", symbols)
	cacheTTL := 5 * time.Minute // Stocks update less frequently

	// Try to get from cache first
	if s.cache != nil && s.cache.IsEnabled() {
		var cachedPrices []StockPrice
		if err := s.cache.Get(ctx, cacheKey, &cachedPrices); err == nil {
			log.Println("Cache hit for stock prices")
			return cachedPrices, nil
		}
		log.Println("Cache miss for stock prices")
	}

	// If not in cache, fetch with retry logic
	var prices []StockPrice
	operation := func() error {
		var err error
		prices, err = s.fetchStockPricesFromAPI(ctx, symbols)
		return err
	}

	// Configure exponential backoff
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 30 * time.Second
	b.InitialInterval = 1 * time.Second
	b.MaxInterval = 10 * time.Second

	if err := backoff.Retry(operation, backoff.WithContext(b, ctx)); err != nil {
		return nil, fmt.Errorf("failed to fetch stock prices after retries: %w", err)
	}

	// Cache the result
	if s.cache != nil && s.cache.IsEnabled() {
		if err := s.cache.Set(ctx, cacheKey, prices, cacheTTL); err != nil {
			log.Printf("Failed to cache stock prices: %v", err)
		}
	}

	return prices, nil
}

// fetchCryptoPricesFromAPI fetches from CoinGecko API (internal method with timeout)
func (s *CachedMarketService) fetchCryptoPricesFromAPI(ctx context.Context) ([]CryptoPrice, error) {
	url := "https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=10&page=1&sparkline=false"

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch crypto data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var cryptos []CryptoPrice
	if err := json.NewDecoder(resp.Body).Decode(&cryptos); err != nil {
		return nil, fmt.Errorf("failed to parse crypto data: %w", err)
	}

	return cryptos, nil
}

// fetchStockPricesFromAPI fetches from Alpha Vantage API (internal method with timeout)
func (s *CachedMarketService) fetchStockPricesFromAPI(ctx context.Context, symbols []string) ([]StockPrice, error) {
	var stocks []StockPrice

	// Popular stocks to track
	defaultSymbols := []string{"AAPL", "GOOGL", "MSFT", "TSLA", "AMZN"}
	if len(symbols) == 0 {
		symbols = defaultSymbols
	}

	for _, symbol := range symbols {
		// Using Alpha Vantage API (demo key - replace with real key for production)
		url := fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=demo", symbol)

		req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
		if err != nil {
			continue // Skip failed requests
		}

		resp, err := s.client.Do(req)
		if err != nil {
			continue
		}

		var data map[string]map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		if quote, ok := data["Global Quote"]; ok {
			var price, change, changePct float64
			var volume int64

			if _, err := fmt.Sscanf(quote["05. price"], "%f", &price); err != nil {
				continue
			}
			if _, err := fmt.Sscanf(quote["09. change"], "%f", &change); err != nil {
				continue
			}
			if _, err := fmt.Sscanf(quote["10. change percent"], "%f%%", &changePct); err != nil {
				continue
			}
			if _, err := fmt.Sscanf(quote["06. volume"], "%d", &volume); err != nil {
				continue
			}

			stocks = append(stocks, StockPrice{
				Symbol:    symbol,
				Price:     price,
				Change:    change,
				ChangePct: changePct,
				Volume:    volume,
			})
		}

		// Rate limiting
		time.Sleep(200 * time.Millisecond)
	}

	if len(stocks) == 0 {
		return nil, fmt.Errorf("no stock data retrieved")
	}

	return stocks, nil
}

// AnalyzeMarketWithCache performs market analysis with caching
func (s *CachedMarketService) AnalyzeMarketWithCache(ctx context.Context) (*MarketAnalysis, error) {
	cacheKey := "market:analysis"
	cacheTTL := 2 * time.Minute

	// Try to get from cache first
	if s.cache != nil && s.cache.IsEnabled() {
		var cachedAnalysis MarketAnalysis
		if err := s.cache.Get(ctx, cacheKey, &cachedAnalysis); err == nil {
			log.Println("Cache hit for market analysis")
			return &cachedAnalysis, nil
		}
		log.Println("Cache miss for market analysis")
	}

	// Fetch fresh data
	cryptos, err := s.GetCryptoPricesWithCache(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get crypto prices: %w", err)
	}

	stocks, err := s.GetStockPricesWithCache(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock prices: %w", err)
	}

	// AI-powered market analysis
	marketTrend := s.calculateMarketTrend(cryptos, stocks)
	volatility := s.calculateVolatility(cryptos, stocks)
	recommendation := s.generateMarketRecommendation(marketTrend, volatility)

	// Advanced AI features
	sentimentScore := s.calculateSentimentScore(cryptos, stocks)
	confidenceLevel := s.calculateConfidenceLevel(cryptos, stocks)
	riskScore := s.calculateRiskScore(cryptos, stocks)
	predictedReturn := s.predictMarketReturn(cryptos, stocks, sentimentScore)

	analysis := &MarketAnalysis{
		Cryptos:         cryptos,
		Stocks:          stocks,
		MarketTrend:     marketTrend,
		Volatility:      volatility,
		Recommendation:  recommendation,
		SentimentScore:  sentimentScore,
		ConfidenceLevel: confidenceLevel,
		RiskScore:       riskScore,
		PredictedReturn: predictedReturn,
		LastUpdated:     time.Now(),
	}

	// Cache the result
	if s.cache != nil && s.cache.IsEnabled() {
		if err := s.cache.Set(ctx, cacheKey, analysis, cacheTTL); err != nil {
			log.Printf("Failed to cache market analysis: %v", err)
		}
	}

	return analysis, nil
}
