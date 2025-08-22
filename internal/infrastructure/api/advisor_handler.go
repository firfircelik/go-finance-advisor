package api

import (
	"net/http"
	"strconv"

	"go-finance-advisor/internal/application"
	"go-finance-advisor/internal/domain"
	"go-finance-advisor/internal/pkg"

	"github.com/gin-gonic/gin"
)

// AdvisorServiceInterface defines the contract for advisor service operations
type AdvisorServiceInterface interface {
	GenerateAdvice(user domain.User) (*application.InvestmentAdvice, error)
}

// MarketServiceInterface defines the contract for market service operations
type MarketServiceInterface interface {
	GetCryptoPrices() ([]pkg.CryptoPrice, error)
	GetStockPrices(symbols []string) ([]pkg.StockPrice, error)
	GetMarketData() (*pkg.MarketData, error)
	GetMarketSummary() (map[string]interface{}, error)
	AnalyzeMarket() (*pkg.MarketAnalysis, error)
	GenerateRecommendations(riskTolerance string, monthlyIncome float64, analysis *pkg.MarketAnalysis) []domain.Recommendation
	GenerateAdviceText(riskTolerance string, analysis *pkg.MarketAnalysis) string
	GeneratePersonalizedAdvice(user *domain.User, monthlyIncome float64) (*pkg.InvestmentRecommendation, error)
	PerformAIRiskAssessment(user *domain.User, monthlyIncome float64, goals []string) (*pkg.AIRiskAssessment, error)
}

type AdvisorHandler struct {
	Advisor       AdvisorServiceInterface
	Users         UserServiceInterface
	MarketService MarketServiceInterface
}

// NewAdvisorHandler creates a new advisor handler with real-time market service
func NewAdvisorHandler(advisor AdvisorServiceInterface, users UserServiceInterface, marketService MarketServiceInterface) *AdvisorHandler {
	return &AdvisorHandler{
		Advisor:       advisor,
		Users:         users,
		MarketService: marketService,
	}
}

// GetAdvice provides basic investment advice (legacy endpoint)
func (h *AdvisorHandler) GetAdvice(c *gin.Context) {
	userIDStr := c.Param("userId")
	uid, _ := strconv.Atoi(userIDStr)

	user, err := h.Users.GetByID(uint(uid))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	advice, err := h.Advisor.GenerateAdvice(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, advice)
}

// GetRealTimeAdvice provides personalized investment advice with real-time market data
func (h *AdvisorHandler) GetRealTimeAdvice(c *gin.Context) {
	userIDStr := c.Param("userId")
	uid, _ := strconv.Atoi(userIDStr)

	user, err := h.Users.GetByID(uint(uid))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get monthly income from query parameter or calculate from transactions
	monthlyIncomeStr := c.DefaultQuery("monthly_income", "5000")
	monthlyIncome, _ := strconv.ParseFloat(monthlyIncomeStr, 64)

	advice, err := h.MarketService.GeneratePersonalizedAdvice(&user, monthlyIncome)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, advice)
}

// GetMarketData provides real-time cryptocurrency and stock prices
func (h *AdvisorHandler) GetMarketData(c *gin.Context) {
	analysis, err := h.MarketService.AnalyzeMarket()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// GetCryptoPrices provides real-time cryptocurrency prices
func (h *AdvisorHandler) GetCryptoPrices(c *gin.Context) {
	cryptos, err := h.MarketService.GetCryptoPrices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cryptos": cryptos,
		"count":   len(cryptos),
		"updated": "real-time",
	})
}

// GetStockPrices provides real-time stock prices
func (h *AdvisorHandler) GetStockPrices(c *gin.Context) {
	// Get symbols from query parameter
	symbolsParam := c.Query("symbols")
	var symbols []string
	if symbolsParam != "" {
		symbols = []string{symbolsParam}
	}

	stocks, err := h.MarketService.GetStockPrices(symbols)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stocks":  stocks,
		"count":   len(stocks),
		"updated": "real-time",
	})
}

// GetMarketSummary provides a quick market overview
func (h *AdvisorHandler) GetMarketSummary(c *gin.Context) {
	summary, err := h.MarketService.GetMarketSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetPortfolioRecommendations provides AI-powered portfolio recommendations based on risk profile
func (h *AdvisorHandler) GetPortfolioRecommendations(c *gin.Context) {
	userIDStr := c.Param("userId")
	uid, _ := strconv.Atoi(userIDStr)

	user, err := h.Users.GetByID(uint(uid))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	monthlyIncomeStr := c.DefaultQuery("monthly_income", "5000")
	monthlyIncome, _ := strconv.ParseFloat(monthlyIncomeStr, 64)

	// Get AI-enhanced market analysis
	analysis, err := h.MarketService.AnalyzeMarket()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate AI-powered recommendations
	recommendations := h.MarketService.GenerateRecommendations(user.RiskTolerance, monthlyIncome, analysis)
	advice := h.MarketService.GenerateAdviceText(user.RiskTolerance, analysis)

	c.JSON(http.StatusOK, gin.H{
		"user_id":         user.ID,
		"risk_profile":    user.RiskTolerance,
		"recommendations": recommendations,
		"advice":          advice,
		"market_analysis": analysis,
		"ai_features": gin.H{
			"sentiment_score":  analysis.SentimentScore,
			"confidence_level": analysis.ConfidenceLevel,
			"risk_score":       analysis.RiskScore,
			"predicted_return": analysis.PredictedReturn,
		},
		"created_at": "real-time",
	})
}

// GetAIRiskAssessment provides comprehensive AI-driven risk analysis for a user
func (h *AdvisorHandler) GetAIRiskAssessment(c *gin.Context) {
	userIDStr := c.Param("userId")
	uid, _ := strconv.Atoi(userIDStr)

	user, err := h.Users.GetByID(uint(uid))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	monthlyIncomeStr := c.DefaultQuery("monthly_income", "5000")
	monthlyIncome, _ := strconv.ParseFloat(monthlyIncomeStr, 64)

	// Parse investment goals from query parameters
	goalsParam := c.Query("goals")
	var goals []string
	if goalsParam != "" {
		goals = []string{goalsParam}
	} else {
		goals = []string{"retirement", "wealth_building"}
	}

	// Perform AI risk assessment
	assessment, err := h.MarketService.PerformAIRiskAssessment(&user, monthlyIncome, goals)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, assessment)
}

// GetAIMarketPrediction provides AI-powered market predictions
func (h *AdvisorHandler) GetAIMarketPrediction(c *gin.Context) {
	// Get prediction timeframe from query parameter
	timeframeStr := c.DefaultQuery("timeframe", "30")
	timeframe, _ := strconv.Atoi(timeframeStr)

	// Get market analysis with AI predictions
	analysis, err := h.MarketService.AnalyzeMarket()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create prediction response
	prediction := gin.H{
		"timeframe_days":   timeframe,
		"predicted_return": analysis.PredictedReturn,
		"confidence_level": analysis.ConfidenceLevel,
		"risk_score":       analysis.RiskScore,
		"sentiment_score":  analysis.SentimentScore,
		"market_trend":     analysis.MarketTrend,
		"volatility":       analysis.Volatility,
		"recommendation":   analysis.Recommendation,
		"ai_insights": gin.H{
			"bullish_signals": analysis.SentimentScore > 0.6,
			"bearish_signals": analysis.SentimentScore < 0.4,
			"high_confidence": analysis.ConfidenceLevel > 0.7,
			"low_risk":        analysis.RiskScore < 0.3,
			"high_volatility": analysis.Volatility == "high",
		},
		"generated_at": analysis.LastUpdated,
	}

	c.JSON(http.StatusOK, prediction)
}

// GetAIPortfolioOptimization provides AI-optimized portfolio suggestions
func (h *AdvisorHandler) GetAIPortfolioOptimization(c *gin.Context) {
	userIDStr := c.Param("userId")
	uid, _ := strconv.Atoi(userIDStr)

	user, err := h.Users.GetByID(uint(uid))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	monthlyIncomeStr := c.DefaultQuery("monthly_income", "5000")
	monthlyIncome, _ := strconv.ParseFloat(monthlyIncomeStr, 64)

	// Get current portfolio value from query parameter
	currentValueStr := c.DefaultQuery("current_value", "0")
	currentValue, _ := strconv.ParseFloat(currentValueStr, 64)

	// Perform AI risk assessment for optimization
	assessment, err := h.MarketService.PerformAIRiskAssessment(&user, monthlyIncome, []string{"optimization"})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get market analysis for optimization
	analysis, err := h.MarketService.AnalyzeMarket()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate optimization suggestions
	optimization := gin.H{
		"user_id":                 user.ID,
		"current_portfolio_value": currentValue,
		"risk_assessment":         assessment,
		"optimized_allocation":    assessment.RecommendedAllocation,
		"rebalancing_needed":      currentValue > 0,
		"expected_return":         analysis.PredictedReturn,
		"optimization_score":      assessment.ConfidenceScore,
		"ai_recommendations": gin.H{
			"increase_stocks": assessment.RiskScore > 0.5 && analysis.SentimentScore > 0.6,
			"reduce_crypto":   analysis.Volatility == "high",
			"add_bonds":       assessment.RiskScore < 0.3,
			"hold_cash":       analysis.SentimentScore < 0.4,
		},
		"next_review_date": assessment.CreatedAt.AddDate(0, 1, 0), // Monthly review
		"generated_at":     assessment.CreatedAt,
	}

	c.JSON(http.StatusOK, optimization)
}
