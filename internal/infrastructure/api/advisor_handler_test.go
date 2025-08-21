package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-finance-advisor/internal/application"
	"go-finance-advisor/internal/domain"
	"go-finance-advisor/internal/pkg"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAdvisorService is a mock implementation of AdvisorService
type MockAdvisorService struct {
	mock.Mock
}

func (m *MockAdvisorService) GenerateAdvice(user domain.User) (*application.InvestmentAdvice, error) {
	args := m.Called(user)
	return args.Get(0).(*application.InvestmentAdvice), args.Error(1)
}

// MockUserService is a mock implementation of UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetByID(userID uint) (domain.User, error) {
	args := m.Called(userID)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUserService) Create(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserService) Update(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserService) Delete(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockUserService) GetByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) Register(email, password, firstName, lastName string) (*domain.User, error) {
	args := m.Called(email, password, firstName, lastName)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) Login(email, password string) (*domain.User, error) {
	args := m.Called(email, password)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) ValidateCredentials(email, password string) (*domain.User, error) {
	args := m.Called(email, password)
	return args.Get(0).(*domain.User), args.Error(1)
}

// MockRealTimeMarketService is a mock implementation of RealTimeMarketService
type MockRealTimeMarketService struct {
	mock.Mock
}

func (m *MockRealTimeMarketService) GeneratePersonalizedAdvice(user *domain.User, monthlyIncome float64) (*pkg.InvestmentRecommendation, error) {
	args := m.Called(user, monthlyIncome)
	return args.Get(0).(*pkg.InvestmentRecommendation), args.Error(1)
}

func (m *MockRealTimeMarketService) AnalyzeMarket() (*pkg.MarketAnalysis, error) {
	args := m.Called()
	return args.Get(0).(*pkg.MarketAnalysis), args.Error(1)
}

func (m *MockRealTimeMarketService) GetCryptoPrices() ([]pkg.CryptoPrice, error) {
	args := m.Called()
	return args.Get(0).([]pkg.CryptoPrice), args.Error(1)
}

func (m *MockRealTimeMarketService) GetStockPrices(symbols []string) ([]pkg.StockPrice, error) {
	args := m.Called(symbols)
	return args.Get(0).([]pkg.StockPrice), args.Error(1)
}

func (m *MockRealTimeMarketService) GetMarketData() (*pkg.MarketData, error) {
	args := m.Called()
	return args.Get(0).(*pkg.MarketData), args.Error(1)
}

func (m *MockRealTimeMarketService) GetMarketSummary() (map[string]interface{}, error) {
	args := m.Called()
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockRealTimeMarketService) GenerateRecommendations(riskTolerance string, monthlyIncome float64, analysis *pkg.MarketAnalysis) []pkg.InvestmentRecommendation {
	args := m.Called(riskTolerance, monthlyIncome, analysis)
	return args.Get(0).([]pkg.InvestmentRecommendation)
}

func (m *MockRealTimeMarketService) GenerateAdviceText(riskTolerance string, analysis *pkg.MarketAnalysis) string {
	args := m.Called(riskTolerance, analysis)
	return args.String(0)
}

func (m *MockRealTimeMarketService) PerformAIRiskAssessment(user *domain.User, monthlyIncome float64, goals []string) (*pkg.AIRiskAssessment, error) {
	args := m.Called(user, monthlyIncome, goals)
	return args.Get(0).(*pkg.AIRiskAssessment), args.Error(1)
}

func setupAdvisorHandler() (*AdvisorHandler, *MockAdvisorService, *MockUserService, *MockRealTimeMarketService) {
	mockAdvisorService := &MockAdvisorService{}
	mockUserService := &MockUserService{}
	mockMarketService := &MockRealTimeMarketService{}

	handler := &AdvisorHandler{
		Advisor:       mockAdvisorService,
		Users:         mockUserService,
		MarketService: mockMarketService,
	}

	return handler, mockAdvisorService, mockUserService, mockMarketService
}

func TestAdvisorHandler_GetAdvice(t *testing.T) {
	t.Run("should get advice successfully", func(t *testing.T) {
		handler, mockAdvisorService, mockUserService, _ := setupAdvisorHandler()
		router := setupGin()
		router.GET("/advice/:userId", handler.GetAdvice)

		user := &domain.User{
			ID:            1,
			Email:         "test@example.com",
			RiskTolerance: "moderate",
		}

		advice := &application.InvestmentAdvice{
			MonthlySavings: 1000.0,
			Risk:           "moderate",
			Recommendations: []application.Recommendation{
				{Asset: "SPY", Amount: 500.0, Percent: 50},
				{Asset: "BTC", Amount: 500.0, Percent: 50},
			},
		}

		mockUserService.On("GetByID", uint(1)).Return(*user, nil)
		mockAdvisorService.On("GenerateAdvice", *user).Return(advice, nil)

		req := httptest.NewRequest("GET", "/advice/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response application.InvestmentAdvice
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, 1000.0, response.MonthlySavings)
		assert.Equal(t, "moderate", response.Risk)
		mockUserService.AssertExpectations(t)
		mockAdvisorService.AssertExpectations(t)
	})

	t.Run("should return not found for non-existent user", func(t *testing.T) {
		handler, _, mockUserService, _ := setupAdvisorHandler()
		router := setupGin()
		router.GET("/advice/:userId", handler.GetAdvice)

		mockUserService.On("GetByID", uint(999)).Return(domain.User{}, errors.New("user not found"))

		req := httptest.NewRequest("GET", "/advice/999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "user not found", response["error"])
		mockUserService.AssertExpectations(t)
	})

	t.Run("should return internal server error when advisor service fails", func(t *testing.T) {
		handler, mockAdvisorService, mockUserService, _ := setupAdvisorHandler()
		router := setupGin()
		router.GET("/advice/:userId", handler.GetAdvice)

		user := &domain.User{
			ID:            1,
			Email:         "test@example.com",
			RiskTolerance: "moderate",
		}

		mockUserService.On("GetByID", uint(1)).Return(*user, nil)
		mockAdvisorService.On("GenerateAdvice", *user).Return((*application.InvestmentAdvice)(nil), errors.New("advisor service error"))

		req := httptest.NewRequest("GET", "/advice/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "advisor service error", response["error"])
		mockUserService.AssertExpectations(t)
		mockAdvisorService.AssertExpectations(t)
	})
}

func TestAdvisorHandler_GetRealTimeAdvice(t *testing.T) {
	t.Run("should get real-time advice successfully", func(t *testing.T) {
		handler, _, mockUserService, mockMarketService := setupAdvisorHandler()
		router := setupGin()
		router.GET("/advice/realtime/:userId", handler.GetRealTimeAdvice)

		user := &domain.User{
			ID:            1,
			Email:         "test@example.com",
			RiskTolerance: "moderate",
		}

		advice := &pkg.InvestmentRecommendation{
			UserID:      1,
			RiskProfile: "moderate",
			Recommendations: []domain.Recommendation{
				{Symbol: "SPY", Action: "buy", Reason: "Diversify portfolio", Confidence: 80.0, RiskLevel: "moderate", Timeframe: "long", IsActive: true},
				{Symbol: "BTC", Action: "buy", Reason: "Consider bonds", Confidence: 60.0, RiskLevel: "moderate", Timeframe: "medium", IsActive: true},
			},
			Advice:    "Diversify portfolio, Consider bonds",
			CreatedAt: time.Now(),
		}

		mockUserService.On("GetByID", uint(1)).Return(*user, nil)
        mockMarketService.On("GeneratePersonalizedAdvice", user, 5000.0).Return(advice, nil)

		req := httptest.NewRequest("GET", "/advice/realtime/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response pkg.InvestmentRecommendation
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, uint(1), response.UserID)
		assert.Equal(t, "moderate", response.RiskProfile)
		assert.NotEmpty(t, response.Recommendations)
		assert.NotEmpty(t, response.Advice)
		mockUserService.AssertExpectations(t)
		mockMarketService.AssertExpectations(t)
	})

	t.Run("should get real-time advice with custom monthly income", func(t *testing.T) {
		handler, _, mockUserService, mockMarketService := setupAdvisorHandler()
		router := setupGin()
		router.GET("/advice/realtime/:userId", handler.GetRealTimeAdvice)

		user := &domain.User{
			ID:            1,
			Email:         "test@example.com",
			RiskTolerance: "aggressive",
		}

		advice := &pkg.InvestmentRecommendation{
			UserID:      1,
			RiskProfile: "aggressive",
			Recommendations: []domain.Recommendation{
				{
					Type:         "stock",
					Symbol:       "QQQ",
					Action:       "buy",
					Reason:       "Growth stocks for aggressive portfolio",
					Confidence:   80.0,
					CurrentPrice: 350.0,
					RiskLevel:    "high",
					Timeframe:    "medium",
					IsActive:     true,
				},
			},
			Advice:    "Invest in growth stocks, Consider crypto",
			CreatedAt: time.Now(),
		}

		mockUserService.On("GetByID", uint(1)).Return(*user, nil)
		mockMarketService.On("GeneratePersonalizedAdvice", user, 8000.0).Return(advice, nil)

		req := httptest.NewRequest("GET", "/advice/realtime/1?monthly_income=8000", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response pkg.InvestmentRecommendation
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "aggressive", response.RiskProfile)
		mockUserService.AssertExpectations(t)
		mockMarketService.AssertExpectations(t)
	})

	t.Run("should return not found for non-existent user", func(t *testing.T) {
		handler, _, mockUserService, _ := setupAdvisorHandler()
		router := setupGin()
		router.GET("/advice/realtime/:userId", handler.GetRealTimeAdvice)

		mockUserService.On("GetByID", uint(999)).Return(domain.User{}, errors.New("user not found"))

		req := httptest.NewRequest("GET", "/advice/realtime/999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "User not found", response["error"])
		mockUserService.AssertExpectations(t)
	})
}

func TestAdvisorHandler_GetMarketData(t *testing.T) {
	t.Run("should get market data successfully", func(t *testing.T) {
		handler, _, _, mockMarketService := setupAdvisorHandler()
		router := setupGin()
		router.GET("/market/data", handler.GetMarketData)

		analysis := &pkg.MarketAnalysis{
			MarketTrend:      "bullish",
			Volatility:       "moderate",
			SentimentScore:   0.75,
			ConfidenceLevel:  0.85,
			RiskScore:        0.4,
			PredictedReturn:  0.08,
			Recommendation:   "buy",
			LastUpdated:      time.Now(),
		}

		mockMarketService.On("AnalyzeMarket").Return(analysis, nil)

		req := httptest.NewRequest("GET", "/market/data", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response pkg.MarketAnalysis
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "bullish", response.MarketTrend)
		assert.Equal(t, 0.75, response.SentimentScore)
		mockMarketService.AssertExpectations(t)
	})

	t.Run("should return internal server error when market service fails", func(t *testing.T) {
		handler, _, _, mockMarketService := setupAdvisorHandler()
		router := setupGin()
		router.GET("/market/data", handler.GetMarketData)

		mockMarketService.On("AnalyzeMarket").Return((*pkg.MarketAnalysis)(nil), errors.New("market service error"))

		req := httptest.NewRequest("GET", "/market/data", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "market service error", response["error"])
		mockMarketService.AssertExpectations(t)
	})
}

func TestAdvisorHandler_GetCryptoPrices(t *testing.T) {
	t.Run("should get crypto prices successfully", func(t *testing.T) {
		handler, _, _, mockMarketService := setupAdvisorHandler()
		router := setupGin()
		router.GET("/market/crypto", handler.GetCryptoPrices)

		cryptos := []pkg.CryptoPrice{
			{
				Symbol:    "BTC",
				Name:      "Bitcoin",
				Price:     45000.0,
				Change24h: 2.5,
				Volume24h: 1000000000,
				MarketCap: 850000000000,
			},
			{
				Symbol:    "ETH",
				Name:      "Ethereum",
				Price:     3000.0,
				Change24h: 1.8,
				Volume24h: 500000000,
				MarketCap: 360000000000,
			},
		}

		mockMarketService.On("GetCryptoPrices").Return(cryptos, nil)

		req := httptest.NewRequest("GET", "/market/crypto", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(2), response["count"])
		assert.Equal(t, "real-time", response["updated"])
		cryptosList := response["cryptos"].([]interface{})
		assert.Len(t, cryptosList, 2)
		mockMarketService.AssertExpectations(t)
	})

	t.Run("should return internal server error when crypto service fails", func(t *testing.T) {
		handler, _, _, mockMarketService := setupAdvisorHandler()
		router := setupGin()
		router.GET("/market/crypto", handler.GetCryptoPrices)

		mockMarketService.On("GetCryptoPrices").Return([]pkg.CryptoPrice{}, errors.New("crypto service error"))

		req := httptest.NewRequest("GET", "/market/crypto", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "crypto service error", response["error"])
		mockMarketService.AssertExpectations(t)
	})
}

func TestAdvisorHandler_GetStockPrices(t *testing.T) {
	t.Run("should get stock prices successfully", func(t *testing.T) {
		handler, _, _, mockMarketService := setupAdvisorHandler()
		router := setupGin()
		router.GET("/market/stocks", handler.GetStockPrices)

		stocks := []pkg.StockPrice{
			{
				Symbol:    "AAPL",
				Price:     150.0,
				Change:    2.5,
				ChangePct: 1.7,
				Volume:    50000000,
			},
		}

		mockMarketService.On("GetStockPrices", []string(nil)).Return(stocks, nil)

		req := httptest.NewRequest("GET", "/market/stocks", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(1), response["count"])
		assert.Equal(t, "real-time", response["updated"])
		mockMarketService.AssertExpectations(t)
	})

	t.Run("should get stock prices with specific symbols", func(t *testing.T) {
		handler, _, _, mockMarketService := setupAdvisorHandler()
		router := setupGin()
		router.GET("/market/stocks", handler.GetStockPrices)

		stocks := []pkg.StockPrice{
			{
				Symbol:    "AAPL",
				Price:     150.0,
				Change:    2.5,
				ChangePct: 1.7,
				Volume:    50000000,
			},
		}

		mockMarketService.On("GetStockPrices", []string{"AAPL"}).Return(stocks, nil)

		req := httptest.NewRequest("GET", "/market/stocks?symbols=AAPL", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(1), response["count"])
		mockMarketService.AssertExpectations(t)
	})
}

func TestAdvisorHandler_GetMarketSummary(t *testing.T) {
	t.Run("should get market summary successfully", func(t *testing.T) {
		handler, _, _, mockMarketService := setupAdvisorHandler()
		router := setupGin()
		router.GET("/market/summary", handler.GetMarketSummary)

		summary := map[string]interface{}{
			"overall_trend":     "bullish",
			"market_cap":        50000000000000,
			"volume_24h":        2000000000000,
			"dominance_index":   0.45,
			"fear_greed_index":  75,
			"top_gainers":       []string{"AAPL", "TSLA"},
			"top_losers":        []string{"META", "NFLX"},
		}

		mockMarketService.On("GetMarketSummary").Return(summary, nil)

		req := httptest.NewRequest("GET", "/market/summary", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "bullish", response["overall_trend"])
		assert.Equal(t, float64(75), response["fear_greed_index"])
		mockMarketService.AssertExpectations(t)
	})

	t.Run("should return internal server error when market service fails", func(t *testing.T) {
		handler, _, _, mockMarketService := setupAdvisorHandler()
		router := setupGin()
		router.GET("/market/summary", handler.GetMarketSummary)

		mockMarketService.On("GetMarketSummary").Return((map[string]interface{})(nil), errors.New("market service error"))

		req := httptest.NewRequest("GET", "/market/summary", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "market service error", response["error"])
		mockMarketService.AssertExpectations(t)
	})
}

func TestAdvisorHandler_GetPortfolioRecommendations(t *testing.T) {
	t.Run("should get portfolio recommendations successfully", func(t *testing.T) {
		handler, _, mockUserService, mockMarketService := setupAdvisorHandler()
		router := setupGin()
		router.GET("/portfolio/recommendations/:userId", handler.GetPortfolioRecommendations)

		user := &domain.User{
			ID:            1,
			Email:         "test@example.com",
			RiskTolerance: "moderate",
		}

		analysis := &pkg.MarketAnalysis{
			MarketTrend:      "bullish",
			SentimentScore:   0.75,
			ConfidenceLevel:  0.85,
			RiskScore:        0.4,
			PredictedReturn:  0.08,
			Volatility:       "moderate",
			Recommendation:   "buy",
			LastUpdated:      time.Now(),
		}

		recommendations := []pkg.InvestmentRecommendation{
			{
				UserID:      1,
				RiskProfile: "moderate",
				Recommendations: []domain.Recommendation{
					{
						Type:         "stocks",
						Symbol:       "SPY",
						Action:       "buy",
						Reason:       "Market is bullish",
						Confidence:   80.0,
						CurrentPrice: 400.0,
						RiskLevel:    "moderate",
						Timeframe:    "medium",
						IsActive:     true,
					},
				},
				Advice:    "Invest in balanced portfolio",
				CreatedAt: time.Now(),
			},
		}

		adviceText := "Based on current market conditions, consider a balanced approach."

		mockUserService.On("GetByID", uint(1)).Return(*user, nil)
		mockMarketService.On("AnalyzeMarket").Return(analysis, nil)
		mockMarketService.On("GenerateRecommendations", "moderate", 5000.0, analysis).Return(recommendations)
		mockMarketService.On("GenerateAdviceText", "moderate", analysis).Return(adviceText)

		req := httptest.NewRequest("GET", "/portfolio/recommendations/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(1), response["user_id"])
		assert.Equal(t, "moderate", response["risk_profile"])
		assert.NotNil(t, response["recommendations"])
		assert.NotNil(t, response["ai_features"])
		mockUserService.AssertExpectations(t)
		mockMarketService.AssertExpectations(t)
	})

	t.Run("should return not found for non-existent user", func(t *testing.T) {
		handler, _, mockUserService, _ := setupAdvisorHandler()
		router := setupGin()
		router.GET("/portfolio/recommendations/:userId", handler.GetPortfolioRecommendations)

		mockUserService.On("GetByID", uint(999)).Return(domain.User{}, errors.New("user not found"))

		req := httptest.NewRequest("GET", "/portfolio/recommendations/999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "User not found", response["error"])
		mockUserService.AssertExpectations(t)
	})
}

func TestAdvisorHandler_GetAIRiskAssessment(t *testing.T) {
	t.Run("should get AI risk assessment successfully", func(t *testing.T) {
		handler, _, mockUserService, mockMarketService := setupAdvisorHandler()
		router := setupGin()
		router.GET("/ai/risk-assessment/:userId", handler.GetAIRiskAssessment)

		user := &domain.User{
			ID:            1,
			Email:         "test@example.com",
			RiskTolerance: "moderate",
		}

		assessment := &pkg.AIRiskAssessment{
			UserID:                1,
			RiskScore:             0.6,
			RiskCategory:          "moderate",
			ConfidenceScore:       0.85,
			RecommendedAllocation: map[string]float64{"stocks": 60, "bonds": 30, "cash": 10},
			FactorsAnalyzed:       []string{"market volatility", "inflation risk"},
			CreatedAt:             time.Now(),
		}

		mockUserService.On("GetByID", uint(1)).Return(*user, nil)
        mockMarketService.On("PerformAIRiskAssessment", user, 5000.0, []string{"retirement", "wealth_building"}).Return(assessment, nil)

		req := httptest.NewRequest("GET", "/ai/risk-assessment/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response pkg.AIRiskAssessment
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, uint(1), response.UserID)
		assert.Equal(t, "moderate", response.RiskCategory)
		mockUserService.AssertExpectations(t)
		mockMarketService.AssertExpectations(t)
	})

	t.Run("should get AI risk assessment with custom goals", func(t *testing.T) {
		handler, _, mockUserService, mockMarketService := setupAdvisorHandler()
		router := setupGin()
		router.GET("/ai/risk-assessment/:userId", handler.GetAIRiskAssessment)

		user := &domain.User{
			ID:            1,
			Email:         "test@example.com",
			RiskTolerance: "aggressive",
		}

		assessment := &pkg.AIRiskAssessment{
			UserID:                1,
			RiskScore:             0.8,
			RiskCategory:          "aggressive",
			ConfidenceScore:       0.75,
			RecommendedAllocation: map[string]float64{"stocks": 80, "crypto": 15, "cash": 5},
			FactorsAnalyzed:       []string{"high volatility", "market timing risk"},
			CreatedAt:             time.Now(),
		}

		mockUserService.On("GetByID", uint(1)).Return(*user, nil)
		mockMarketService.On("PerformAIRiskAssessment", user, 8000.0, []string{"growth"}).Return(assessment, nil)

		req := httptest.NewRequest("GET", "/ai/risk-assessment/1?monthly_income=8000&goals=growth", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response pkg.AIRiskAssessment
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "aggressive", response.RiskCategory)
		mockUserService.AssertExpectations(t)
		mockMarketService.AssertExpectations(t)
	})
}

func TestAdvisorHandler_GetAIMarketPrediction(t *testing.T) {
	t.Run("should get AI market prediction successfully", func(t *testing.T) {
		handler, _, _, mockMarketService := setupAdvisorHandler()
		router := setupGin()
		router.GET("/ai/market-prediction", handler.GetAIMarketPrediction)

		analysis := &pkg.MarketAnalysis{
			MarketTrend:      "bullish",
			Volatility:       "moderate",
			SentimentScore:   0.75,
			ConfidenceLevel:  0.85,
			RiskScore:        0.4,
			PredictedReturn:  0.08,
			Recommendation:   "buy",
			LastUpdated:      time.Now(),
		}

		mockMarketService.On("AnalyzeMarket").Return(analysis, nil)

		req := httptest.NewRequest("GET", "/ai/market-prediction", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(30), response["timeframe_days"])
		assert.Equal(t, 0.08, response["predicted_return"])
		assert.Equal(t, "bullish", response["market_trend"])
		assert.NotNil(t, response["ai_insights"])
		mockMarketService.AssertExpectations(t)
	})

	t.Run("should get AI market prediction with custom timeframe", func(t *testing.T) {
		handler, _, _, mockMarketService := setupAdvisorHandler()
		router := setupGin()
		router.GET("/ai/market-prediction", handler.GetAIMarketPrediction)

		analysis := &pkg.MarketAnalysis{
			MarketTrend:      "bearish",
			Volatility:       "high",
			SentimentScore:   0.25,
			ConfidenceLevel:  0.90,
			RiskScore:        0.8,
			PredictedReturn:  -0.05,
			Recommendation:   "sell",
			LastUpdated:      time.Now(),
		}

		mockMarketService.On("AnalyzeMarket").Return(analysis, nil)

		req := httptest.NewRequest("GET", "/ai/market-prediction?timeframe=60", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(60), response["timeframe_days"])
		assert.Equal(t, "bearish", response["market_trend"])
		assert.Equal(t, "high", response["volatility"])
		mockMarketService.AssertExpectations(t)
	})
}

func TestAdvisorHandler_GetAIPortfolioOptimization(t *testing.T) {
	t.Run("should get AI portfolio optimization successfully", func(t *testing.T) {
		handler, _, mockUserService, mockMarketService := setupAdvisorHandler()
		router := setupGin()
		router.GET("/ai/portfolio-optimization/:userId", handler.GetAIPortfolioOptimization)

		user := &domain.User{
			ID:            1,
			Email:         "test@example.com",
			RiskTolerance: "moderate",
		}

		assessment := &pkg.AIRiskAssessment{
			UserID:                1,
			RiskScore:             0.6,
			RiskCategory:          "moderate",
			ConfidenceScore:       0.85,
			RecommendedAllocation: map[string]float64{"stocks": 60, "bonds": 30, "cash": 10},
			CreatedAt:             time.Now(),
		}

		analysis := &pkg.MarketAnalysis{
			MarketTrend:      "bullish",
			Volatility:       "moderate",
			SentimentScore:   0.75,
			PredictedReturn:  0.08,
			LastUpdated:      time.Now(),
		}

		mockUserService.On("GetByID", uint(1)).Return(*user, nil)
        mockMarketService.On("PerformAIRiskAssessment", user, 5000.0, []string{"optimization"}).Return(assessment, nil)
        mockMarketService.On("AnalyzeMarket").Return(analysis, nil)

		req := httptest.NewRequest("GET", "/ai/portfolio-optimization/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(1), response["user_id"])
		assert.Equal(t, float64(0), response["current_portfolio_value"])
		assert.NotNil(t, response["risk_assessment"])
		assert.NotNil(t, response["ai_recommendations"])
		mockUserService.AssertExpectations(t)
		mockMarketService.AssertExpectations(t)
	})

	t.Run("should get AI portfolio optimization with current portfolio value", func(t *testing.T) {
		handler, _, mockUserService, mockMarketService := setupAdvisorHandler()
		router := setupGin()
		router.GET("/ai/portfolio-optimization/:userId", handler.GetAIPortfolioOptimization)

		user := &domain.User{
			ID:            1,
			Email:         "test@example.com",
			RiskTolerance: "aggressive",
		}

		assessment := &pkg.AIRiskAssessment{
			UserID:                1,
			RiskScore:             0.8,
			RiskCategory:          "aggressive",
			ConfidenceScore:       0.75,
			RecommendedAllocation: map[string]float64{"stocks": 80, "crypto": 15, "cash": 5},
			CreatedAt:             time.Now(),
		}

		analysis := &pkg.MarketAnalysis{
			MarketTrend:      "bullish",
			Volatility:       "high",
			SentimentScore:   0.85,
			PredictedReturn:  0.12,
			LastUpdated:      time.Now(),
		}

		mockUserService.On("GetByID", uint(1)).Return(*user, nil)
        mockMarketService.On("PerformAIRiskAssessment", user, 8000.0, []string{"optimization"}).Return(assessment, nil)
        mockMarketService.On("AnalyzeMarket").Return(analysis, nil)

		req := httptest.NewRequest("GET", "/ai/portfolio-optimization/1?monthly_income=8000&current_value=100000", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(100000), response["current_portfolio_value"])
		assert.Equal(t, true, response["rebalancing_needed"])
		mockUserService.AssertExpectations(t)
		mockMarketService.AssertExpectations(t)
	})
}