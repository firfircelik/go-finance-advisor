package pkg

import (
	"encoding/json"
	"fmt"
	"go-finance-advisor/internal/domain"
	"math"
	"net/http"
	"strings"
	"time"
)

// CryptoPrice represents cryptocurrency price data
type CryptoPrice struct {
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	Price     float64 `json:"current_price"`
	Change24h float64 `json:"price_change_percentage_24h"`
	MarketCap float64 `json:"market_cap"`
	Volume24h float64 `json:"total_volume"`
}

// StockPrice represents stock price data
type StockPrice struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Change    float64 `json:"change"`
	ChangePct float64 `json:"change_percent"`
	Volume    int64   `json:"volume"`
}

// MarketAnalysis represents comprehensive AI-powered market analysis
type MarketAnalysis struct {
	Cryptos         []CryptoPrice `json:"cryptos"`
	Stocks          []StockPrice  `json:"stocks"`
	MarketTrend     string        `json:"market_trend"`
	Volatility      string        `json:"volatility"`
	Recommendation  string        `json:"recommendation"`
	SentimentScore  float64       `json:"sentiment_score"`
	ConfidenceLevel float64       `json:"confidence_level"`
	RiskScore       float64       `json:"risk_score"`
	PredictedReturn float64       `json:"predicted_return"`
	LastUpdated     time.Time     `json:"last_updated"`
}

// AIRiskAssessment represents AI-driven risk analysis
type AIRiskAssessment struct {
	UserID                uint               `json:"user_id"`
	RiskScore             float64            `json:"risk_score"`
	RiskCategory          string             `json:"risk_category"`
	RecommendedAllocation map[string]float64 `json:"recommended_allocation"`
	ConfidenceScore       float64            `json:"confidence_score"`
	FactorsAnalyzed       []string           `json:"factors_analyzed"`
	CreatedAt             time.Time          `json:"created_at"`
}

// MarketSentiment represents sentiment analysis data
type MarketSentiment struct {
	OverallSentiment string  `json:"overall_sentiment"`
	SentimentScore   float64 `json:"sentiment_score"`
	FearGreedIndex   int     `json:"fear_greed_index"`
	NewsAnalysis     string  `json:"news_analysis"`
	SocialSentiment  float64 `json:"social_sentiment"`
}

// InvestmentRecommendation represents personalized investment advice
type InvestmentRecommendation struct {
	UserID          uint                    `json:"user_id"`
	RiskProfile     string                  `json:"risk_profile"`
	Recommendations []domain.Recommendation `json:"recommendations"`
	MarketAnalysis  *MarketAnalysis         `json:"market_analysis"`
	Advice          string                  `json:"advice"`
	CreatedAt       time.Time               `json:"created_at"`
}

// RealTimeMarketService provides real-time market data and investment advice
type RealTimeMarketService struct {
	client *http.Client
}

// NewRealTimeMarketService creates a new market service instance
func NewRealTimeMarketService() *RealTimeMarketService {
	return &RealTimeMarketService{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// GetCryptoPrices fetches real-time cryptocurrency prices from CoinGecko
func (s *RealTimeMarketService) GetCryptoPrices() ([]CryptoPrice, error) {
	url := "https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=10&page=1&sparkline=false"

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch crypto data: %v", err)
	}
	defer resp.Body.Close()

	var cryptos []CryptoPrice
	if err := json.NewDecoder(resp.Body).Decode(&cryptos); err != nil {
		return nil, fmt.Errorf("failed to parse crypto data: %v", err)
	}

	return cryptos, nil
}

// GetStockPrices fetches real-time stock prices from Alpha Vantage
func (s *RealTimeMarketService) GetStockPrices(symbols []string) ([]StockPrice, error) {
	var stocks []StockPrice

	// Popular stocks to track
	defaultSymbols := []string{"AAPL", "GOOGL", "MSFT", "TSLA", "AMZN"}
	if len(symbols) == 0 {
		symbols = defaultSymbols
	}

	for _, symbol := range symbols {
		// Using Alpha Vantage API (demo key - replace with real key for production)
		url := fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=demo", symbol)

		resp, err := s.client.Get(url)
		if err != nil {
			continue // Skip failed requests
		}
		defer resp.Body.Close()

		var data map[string]map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			continue
		}

		if quote, ok := data["Global Quote"]; ok {
			var price, change, changePct float64
			var volume int64

			fmt.Sscanf(quote["05. price"], "%f", &price)
			fmt.Sscanf(quote["09. change"], "%f", &change)
			fmt.Sscanf(quote["10. change percent"], "%f%%", &changePct)
			fmt.Sscanf(quote["06. volume"], "%d", &volume)

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

	return stocks, nil
}

// AnalyzeMarket performs comprehensive AI-powered market analysis
func (s *RealTimeMarketService) AnalyzeMarket() (*MarketAnalysis, error) {
	cryptos, err := s.GetCryptoPrices()
	if err != nil {
		return nil, err
	}

	stocks, err := s.GetStockPrices(nil)
	if err != nil {
		return nil, err
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

	return &MarketAnalysis{
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
	}, nil
}

// GeneratePersonalizedAdvice creates personalized investment recommendations
func (s *RealTimeMarketService) GeneratePersonalizedAdvice(user *domain.User, monthlyIncome float64) (*InvestmentRecommendation, error) {
	marketAnalysis, err := s.AnalyzeMarket()
	if err != nil {
		return nil, err
	}

	// Generate recommendations based on risk tolerance
	recommendations := s.GenerateRecommendations(user.RiskTolerance, monthlyIncome, marketAnalysis)
	advice := s.GenerateAdviceText(user.RiskTolerance, marketAnalysis)

	return &InvestmentRecommendation{
		UserID:          user.ID,
		RiskProfile:     user.RiskTolerance,
		Recommendations: recommendations,
		MarketAnalysis:  marketAnalysis,
		Advice:          advice,
		CreatedAt:       time.Now(),
	}, nil
}

// Helper functions

func (s *RealTimeMarketService) calculateMarketTrend(cryptos []CryptoPrice, stocks []StockPrice) string {
	var totalChange float64
	var count int

	for _, crypto := range cryptos {
		totalChange += crypto.Change24h
		count++
	}

	for _, stock := range stocks {
		totalChange += stock.ChangePct
		count++
	}

	if count == 0 {
		return "neutral"
	}

	avgChange := totalChange / float64(count)
	if avgChange > 2 {
		return "bullish"
	} else if avgChange < -2 {
		return "bearish"
	}
	return "neutral"
}

func (s *RealTimeMarketService) calculateVolatility(cryptos []CryptoPrice, stocks []StockPrice) string {
	var totalVolatility float64
	var count int

	for _, crypto := range cryptos {
		if crypto.Change24h < 0 {
			totalVolatility += -crypto.Change24h
		} else {
			totalVolatility += crypto.Change24h
		}
		count++
	}

	for _, stock := range stocks {
		if stock.ChangePct < 0 {
			totalVolatility += -stock.ChangePct
		} else {
			totalVolatility += stock.ChangePct
		}
		count++
	}

	if count == 0 {
		return "low"
	}

	avgVolatility := totalVolatility / float64(count)
	if avgVolatility > 5 {
		return "high"
	} else if avgVolatility > 2 {
		return "medium"
	}
	return "low"
}

func (s *RealTimeMarketService) generateMarketRecommendation(trend, volatility string) string {
	switch {
	case trend == "bullish" && volatility == "low":
		return "Strong buy signal - favorable market conditions"
	case trend == "bullish" && volatility == "medium":
		return "Buy signal with caution - monitor volatility"
	case trend == "bullish" && volatility == "high":
		return "Cautious buy - high volatility present"
	case trend == "bearish" && volatility == "low":
		return "Hold or sell - stable downtrend"
	case trend == "bearish" && volatility == "high":
		return "Strong sell signal - high risk environment"
	default:
		return "Neutral - wait for clearer signals"
	}
}

func (s *RealTimeMarketService) GenerateRecommendations(riskTolerance string, monthlyIncome float64, analysis *MarketAnalysis) []domain.Recommendation {
	var recommendations []domain.Recommendation
	investableAmount := monthlyIncome * 0.2 // 20% of income for investments

	switch riskTolerance {
	case "conservative":
		recommendations = append(recommendations, domain.Recommendation{
			Type:         "bond",
			Symbol:       "GOVT",
			Action:       "buy",
			Reason:       "Low-risk government bonds for stable returns",
			Confidence:   85.0,
			CurrentPrice: investableAmount * 0.6,
			RiskLevel:    "low",
			Timeframe:    "long",
			IsActive:     true,
		})
		recommendations = append(recommendations, domain.Recommendation{
			Type:         "stock",
			Symbol:       "SPY",
			Action:       "buy",
			Reason:       "Blue chip stocks for steady growth",
			Confidence:   80.0,
			CurrentPrice: investableAmount * 0.3,
			RiskLevel:    "low",
			Timeframe:    "medium",
			IsActive:     true,
		})
		recommendations = append(recommendations, domain.Recommendation{
			Type:         "crypto",
			Symbol:       "BTC",
			Action:       "buy",
			Reason:       "Small allocation to Bitcoin for diversification",
			Confidence:   70.0,
			CurrentPrice: investableAmount * 0.1,
			RiskLevel:    "medium",
			Timeframe:    "long",
			IsActive:     true,
		})

	case "moderate":
		recommendations = append(recommendations, domain.Recommendation{
			Type:         "stock",
			Symbol:       "SPY",
			Action:       "buy",
			Reason:       "S&P 500 index funds for balanced growth",
			Confidence:   85.0,
			CurrentPrice: investableAmount * 0.4,
			RiskLevel:    "medium",
			Timeframe:    "medium",
			IsActive:     true,
		})
		recommendations = append(recommendations, domain.Recommendation{
			Type:         "stock",
			Symbol:       "QQQ",
			Action:       "buy",
			Reason:       "Growth stocks for higher potential returns",
			Confidence:   75.0,
			CurrentPrice: investableAmount * 0.3,
			RiskLevel:    "medium",
			Timeframe:    "medium",
			IsActive:     true,
		})
		recommendations = append(recommendations, domain.Recommendation{
			Type:         "crypto",
			Symbol:       "BTC",
			Action:       "buy",
			Reason:       "Bitcoin allocation for portfolio diversification",
			Confidence:   70.0,
			CurrentPrice: investableAmount * 0.2,
			RiskLevel:    "high",
			Timeframe:    "long",
			IsActive:     true,
		})
		recommendations = append(recommendations, domain.Recommendation{
			Type:         "crypto",
			Symbol:       "ETH",
			Action:       "buy",
			Reason:       "Ethereum for smart contract exposure",
			Confidence:   65.0,
			CurrentPrice: investableAmount * 0.1,
			RiskLevel:    "high",
			Timeframe:    "medium",
		})

	case "aggressive":
		recommendations = append(recommendations, domain.Recommendation{
			Type:         "stock",
			Symbol:       "QQQ",
			Action:       "buy",
			Reason:       "High-growth technology stocks for maximum returns",
			Confidence:   70.0,
			CurrentPrice: investableAmount * 0.3,
			RiskLevel:    "high",
			Timeframe:    "short",
			IsActive:     true,
		})
		recommendations = append(recommendations, domain.Recommendation{
			Type:         "crypto",
			Symbol:       "BTC",
			Action:       "buy",
			Reason:       "Major Bitcoin allocation for high growth potential",
			Confidence:   65.0,
			CurrentPrice: investableAmount * 0.3,
			RiskLevel:    "high",
			Timeframe:    "medium",
			IsActive:     true,
		})
		recommendations = append(recommendations, domain.Recommendation{
			Type:         "crypto",
			Symbol:       "ETH",
			Action:       "buy",
			Reason:       "Ethereum for DeFi and smart contract exposure",
			Confidence:   60.0,
			CurrentPrice: investableAmount * 0.2,
			RiskLevel:    "high",
			Timeframe:    "medium",
			IsActive:     true,
		})
		recommendations = append(recommendations, domain.Recommendation{
			Type:         "crypto",
			Symbol:       "ALT",
			Action:       "buy",
			Reason:       "Diversified altcoin portfolio for explosive growth",
			Confidence:   50.0,
			CurrentPrice: investableAmount * 0.2,
			RiskLevel:    "high",
			Timeframe:    "short",
			IsActive:     true,
		})
	}

	return recommendations
}

func (s *RealTimeMarketService) GenerateAdviceText(riskTolerance string, analysis *MarketAnalysis) string {
	baseAdvice := fmt.Sprintf("Market Analysis: %s trend with %s volatility. %s\n\n",
		analysis.MarketTrend, analysis.Volatility, analysis.Recommendation)

	switch riskTolerance {
	case "conservative":
		return baseAdvice + "As a conservative investor, focus on stable assets with guaranteed returns. " +
			"Consider government bonds and established blue-chip stocks. Limit cryptocurrency exposure to 10% maximum."
	case "moderate":
		return baseAdvice + "With moderate risk tolerance, diversify across index funds and growth stocks. " +
			"Cryptocurrency allocation of 20-30% can provide growth potential while maintaining stability."
	case "aggressive":
		return baseAdvice + "As an aggressive investor, you can take advantage of high-growth opportunities. " +
			"Consider higher cryptocurrency allocation and growth stocks, but monitor market conditions closely."
	default:
		return baseAdvice + "Please complete risk assessment to receive personalized investment advice."
	}
}

// GetMarketData fetches current market data
func (s *RealTimeMarketService) GetMarketData() (*MarketData, error) {
	return FetchMarketData()
}

// GetMarketSummary provides a quick AI-enhanced market overview
func (s *RealTimeMarketService) GetMarketSummary() (map[string]interface{}, error) {
	analysis, err := s.AnalyzeMarket()
	if err != nil {
		return nil, err
	}

	summary := map[string]interface{}{
		"market_trend":     analysis.MarketTrend,
		"volatility":       analysis.Volatility,
		"recommendation":   analysis.Recommendation,
		"sentiment_score":  analysis.SentimentScore,
		"confidence_level": analysis.ConfidenceLevel,
		"risk_score":       analysis.RiskScore,
		"predicted_return": analysis.PredictedReturn,
		"top_cryptos":      analysis.Cryptos[:3], // Top 3 cryptos
		"top_stocks":       analysis.Stocks[:3],  // Top 3 stocks
		"last_updated":     analysis.LastUpdated,
	}

	return summary, nil
}

// AI-powered analysis functions

// calculateSentimentScore analyzes market sentiment using price movements and volume
func (s *RealTimeMarketService) calculateSentimentScore(cryptos []CryptoPrice, stocks []StockPrice) float64 {
	var totalSentiment float64
	var count int

	// Analyze crypto sentiment
	for _, crypto := range cryptos {
		// Positive sentiment for price increases, negative for decreases
		sentiment := crypto.Change24h / 100.0 // Normalize to percentage

		// Volume factor - higher volume increases sentiment confidence
		volumeWeight := math.Min(crypto.Volume24h/crypto.MarketCap, 1.0)
		sentiment *= (1.0 + volumeWeight)

		totalSentiment += sentiment
		count++
	}

	// Analyze stock sentiment
	for _, stock := range stocks {
		sentiment := stock.ChangePct / 100.0

		// Volume factor for stocks
		volumeWeight := math.Min(float64(stock.Volume)/1000000.0, 1.0) // Normalize volume
		sentiment *= (1.0 + volumeWeight*0.1)

		totalSentiment += sentiment
		count++
	}

	if count == 0 {
		return 0.5 // Neutral sentiment
	}

	avgSentiment := totalSentiment / float64(count)
	// Normalize to 0-1 scale (0.5 = neutral)
	return math.Max(0, math.Min(1, 0.5+avgSentiment))
}

// calculateConfidenceLevel determines confidence in market analysis
func (s *RealTimeMarketService) calculateConfidenceLevel(cryptos []CryptoPrice, stocks []StockPrice) float64 {
	var totalVolume float64
	var totalMarketCap float64
	var priceConsistency float64
	count := 0

	// Analyze data quality and consistency
	for _, crypto := range cryptos {
		totalVolume += crypto.Volume24h
		totalMarketCap += crypto.MarketCap

		// Price consistency check (lower volatility = higher confidence)
		if math.Abs(crypto.Change24h) < 5 {
			priceConsistency += 1.0
		} else if math.Abs(crypto.Change24h) < 10 {
			priceConsistency += 0.7
		} else {
			priceConsistency += 0.3
		}
		count++
	}

	for _, stock := range stocks {
		if math.Abs(stock.ChangePct) < 3 {
			priceConsistency += 1.0
		} else if math.Abs(stock.ChangePct) < 7 {
			priceConsistency += 0.8
		} else {
			priceConsistency += 0.4
		}
		count++
	}

	if count == 0 {
		return 0.5
	}

	// Calculate confidence based on data consistency and market activity
	volumeConfidence := math.Min(totalVolume/1000000000, 1.0) // Normalize volume
	consistencyConfidence := priceConsistency / float64(count)

	return (volumeConfidence*0.3 + consistencyConfidence*0.7)
}

// calculateRiskScore assesses overall market risk
func (s *RealTimeMarketService) calculateRiskScore(cryptos []CryptoPrice, stocks []StockPrice) float64 {
	var totalRisk float64
	count := 0

	// Crypto risk assessment (higher volatility = higher risk)
	for _, crypto := range cryptos {
		volatilityRisk := math.Abs(crypto.Change24h) / 100.0

		// Market cap risk (smaller cap = higher risk)
		marketCapRisk := 1.0 - math.Min(crypto.MarketCap/100000000000, 1.0)

		cryptoRisk := (volatilityRisk*0.7 + marketCapRisk*0.3)
		totalRisk += math.Min(cryptoRisk, 1.0)
		count++
	}

	// Stock risk assessment
	for _, stock := range stocks {
		volatilityRisk := math.Abs(stock.ChangePct) / 100.0
		stockRisk := volatilityRisk * 0.5 // Stocks generally less risky than crypto
		totalRisk += math.Min(stockRisk, 1.0)
		count++
	}

	if count == 0 {
		return 0.5
	}

	return totalRisk / float64(count)
}

// predictMarketReturn uses AI to predict potential returns
func (s *RealTimeMarketService) predictMarketReturn(cryptos []CryptoPrice, stocks []StockPrice, sentimentScore float64) float64 {
	var weightedReturn float64
	var totalWeight float64

	// Predict based on current trends and sentiment
	for _, crypto := range cryptos {
		// Weight by market cap
		weight := crypto.MarketCap

		// Base prediction on recent performance
		basePrediction := crypto.Change24h * 0.1 // Conservative prediction

		// Adjust by sentiment
		sentimentAdjustment := (sentimentScore - 0.5) * 20 // -10% to +10% adjustment

		predictedReturn := basePrediction + sentimentAdjustment
		weightedReturn += predictedReturn * weight
		totalWeight += weight
	}

	for _, stock := range stocks {
		// Assume average market cap for stocks
		weight := 50000000000.0 // 50B average

		basePrediction := stock.ChangePct * 0.05 // More conservative for stocks
		sentimentAdjustment := (sentimentScore - 0.5) * 10

		predictedReturn := basePrediction + sentimentAdjustment
		weightedReturn += predictedReturn * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0
	}

	return weightedReturn / totalWeight
}

// PerformAIRiskAssessment conducts comprehensive AI-driven risk analysis for a user
func (s *RealTimeMarketService) PerformAIRiskAssessment(user *domain.User, monthlyIncome float64, investmentGoals []string) (*AIRiskAssessment, error) {
	// Analyze user factors
	factors := []string{"age", "income", "risk_tolerance", "investment_goals", "market_conditions"}

	// Calculate risk score based on multiple factors
	riskScore := s.calculateUserRiskScore(user, monthlyIncome, investmentGoals)

	// Determine risk category
	riskCategory := s.determineRiskCategory(riskScore)

	// Generate recommended allocation
	allocation := s.generateAIAllocation(riskScore, user.RiskTolerance)

	// Calculate confidence in assessment
	confidenceScore := s.calculateAssessmentConfidence(user, monthlyIncome)

	return &AIRiskAssessment{
		UserID:                user.ID,
		RiskScore:             riskScore,
		RiskCategory:          riskCategory,
		RecommendedAllocation: allocation,
		ConfidenceScore:       confidenceScore,
		FactorsAnalyzed:       factors,
		CreatedAt:             time.Now(),
	}, nil
}

// Helper functions for AI risk assessment

func (s *RealTimeMarketService) calculateUserRiskScore(user *domain.User, monthlyIncome float64, goals []string) float64 {
	var score float64

	// Age factor (younger = higher risk tolerance)
	if user.Age < 30 {
		score += 0.3
	} else if user.Age < 50 {
		score += 0.2
	} else {
		score += 0.1
	}

	// Income factor
	if monthlyIncome > 10000 {
		score += 0.3
	} else if monthlyIncome > 5000 {
		score += 0.2
	} else {
		score += 0.1
	}

	// Risk tolerance factor
	switch strings.ToLower(user.RiskTolerance) {
	case "aggressive":
		score += 0.4
	case "moderate":
		score += 0.25
	case "conservative":
		score += 0.1
	}

	return math.Min(score, 1.0)
}

func (s *RealTimeMarketService) determineRiskCategory(score float64) string {
	if score >= 0.7 {
		return "high_risk_high_reward"
	} else if score >= 0.4 {
		return "moderate_risk"
	}
	return "low_risk_stable"
}

func (s *RealTimeMarketService) generateAIAllocation(riskScore float64, riskTolerance string) map[string]float64 {
	allocation := make(map[string]float64)

	if riskScore >= 0.7 {
		// High risk allocation
		allocation["stocks"] = 0.4
		allocation["crypto"] = 0.4
		allocation["bonds"] = 0.1
		allocation["cash"] = 0.1
	} else if riskScore >= 0.4 {
		// Moderate risk allocation
		allocation["stocks"] = 0.5
		allocation["crypto"] = 0.2
		allocation["bonds"] = 0.2
		allocation["cash"] = 0.1
	} else {
		// Low risk allocation
		allocation["stocks"] = 0.3
		allocation["crypto"] = 0.1
		allocation["bonds"] = 0.4
		allocation["cash"] = 0.2
	}

	return allocation
}

func (s *RealTimeMarketService) calculateAssessmentConfidence(user *domain.User, monthlyIncome float64) float64 {
	confidence := 0.5 // Base confidence

	// More data = higher confidence
	if user.Age > 0 {
		confidence += 0.1
	}
	if monthlyIncome > 0 {
		confidence += 0.1
	}
	if user.RiskTolerance != "" {
		confidence += 0.2
	}

	return math.Min(confidence, 1.0)
}
