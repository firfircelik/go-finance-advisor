package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-finance-advisor/internal/application"
	"go-finance-advisor/internal/domain"
	"go-finance-advisor/internal/infrastructure/api"
	"go-finance-advisor/internal/infrastructure/persistence"
	"go-finance-advisor/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAdvice(t *testing.T) {
	db := setupTestDB()
	userRepo := persistence.NewUserRepository(db)
	transactionRepo := persistence.NewTransactionRepository(db)
	marketService := &pkg.RealTimeMarketService{}
	userService := &application.UserService{DB: db}
	advisorService := &application.AdvisorService{DB: db}
	advisorHandler := api.NewAdvisorHandler(advisorService, userService, marketService)

	// Create test user
	user := &domain.User{
		ID:        1,
		FirstName: "Test",
		LastName:  "User",
		Email:     "test@example.com",
	}
	err := userRepo.Create(user)
	assert.NoError(t, err)

	// Create test category
	category := &domain.Category{
		ID:   1,
		Name: "Food",
	}
	db.Create(category)

	// Create test transactions
	transactions := []*domain.Transaction{
		{
			ID:          1,
			UserID:      1,
			Amount:      -50.0,
			Description: "Grocery shopping",
			CategoryID:  1,
			Date:        time.Now(),
		},
		{
			ID:          2,
			UserID:      1,
			Amount:      -30.0,
			Description: "Restaurant",
			CategoryID:  1,
			Date:        time.Now().AddDate(0, 0, -1),
		},
	}

	for _, transaction := range transactions {
		createErr := transactionRepo.Create(transaction)
		assert.NoError(t, createErr)
	}

	// Test advice generation
	req := httptest.NewRequest("GET", "/users/1/advice", http.NoBody)
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/users/:userId/advice", advisorHandler.GetAdvice)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "recommendations")
}

func TestGenerateAdviceConservative(t *testing.T) {
	db := setupTestDB()
	userService := &application.UserService{DB: db}
	advisorService := &application.AdvisorService{DB: db}
	marketService := &pkg.RealTimeMarketService{}
	advisorHandler := api.NewAdvisorHandler(advisorService, userService, marketService)

	// Create a conservative user
	user := domain.User{RiskTolerance: "conservative"}
	userService.Create(&user)

	// Create a transaction
	tx := domain.Transaction{
		UserID:      1,
		Type:        "income",
		Description: "Salary",
		Amount:      3000.0,
		CategoryID:  1,
		Date:        time.Now().AddDate(0, -1, 0),
	}
	txService := &application.TransactionService{DB: db}
	txService.Create(&tx)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/users/:userId/advice", advisorHandler.GetAdvice)

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/users/1/advice", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response application.InvestmentAdvice
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "conservative", response.Risk)
	assert.Len(t, response.Recommendations, 2)

	// For conservative risk: 70% SPY, 30% BTC
	for _, rec := range response.Recommendations {
		if rec.Asset == "SPY" {
			assert.Equal(t, 70.0, rec.Percent)
		} else if rec.Asset == "BTC" {
			assert.Equal(t, 30.0, rec.Percent)
		}
	}
}

// MockMarketService implements the MarketService interface for testing
type MockMarketService struct{}

func (m *MockMarketService) GetCryptoPrices() ([]pkg.CryptoPrice, error) {
	return []pkg.CryptoPrice{
		{Symbol: "BTC", Name: "Bitcoin", Price: 50000.0, Change24h: 2.5},
		{Symbol: "ETH", Name: "Ethereum", Price: 3000.0, Change24h: 1.8},
	}, nil
}

func (m *MockMarketService) GetStockPrices(symbols []string) ([]pkg.StockPrice, error) {
	return []pkg.StockPrice{
		{Symbol: "SPY", Price: 400.0, Change: 5.0, ChangePct: 1.25},
		{Symbol: "AAPL", Price: 150.0, Change: 2.0, ChangePct: 1.35},
	}, nil
}

func (m *MockMarketService) AnalyzeMarket() (*pkg.MarketAnalysis, error) {
	cryptos, _ := m.GetCryptoPrices()
	stocks, _ := m.GetStockPrices([]string{})
	return &pkg.MarketAnalysis{
		Cryptos:        cryptos,
		Stocks:         stocks,
		MarketTrend:    "bullish",
		Volatility:     "medium",
		Recommendation: "conservative",
		LastUpdated:    time.Now(),
	}, nil
}

func (m *MockMarketService) GeneratePersonalizedAdvice(user *domain.User, monthlyIncome float64) (*pkg.InvestmentRecommendation, error) {
	return &pkg.InvestmentRecommendation{
		UserID:      user.ID,
		RiskProfile: "conservative",
		Recommendations: []domain.Recommendation{
			{Symbol: "SPY", Action: "buy", Reason: "Stable growth", Confidence: 70.0, RiskLevel: "low", Timeframe: "long", IsActive: true},
			{Symbol: "BTC", Action: "buy", Reason: "Diversification", Confidence: 30.0, RiskLevel: "medium", Timeframe: "long", IsActive: true},
		},
		Advice:    "Conservative portfolio for stable growth",
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockMarketService) GetMarketSummary() (map[string]interface{}, error) {
	return map[string]interface{}{
		"market_trend":   "bullish",
		"volatility":     "medium",
		"recommendation": "conservative",
	}, nil
}
