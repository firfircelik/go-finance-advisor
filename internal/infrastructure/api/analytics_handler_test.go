package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-finance-advisor/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAnalyticsService is a mock implementation of AnalyticsService
type MockAnalyticsService struct {
	mock.Mock
}

func (m *MockAnalyticsService) GetFinancialMetrics(userID uint, period string, startDate, endDate time.Time) (*domain.FinancialMetrics, error) {
	args := m.Called(userID, period, startDate, endDate)
	return args.Get(0).(*domain.FinancialMetrics), args.Error(1)
}

func (m *MockAnalyticsService) GetIncomeExpenseAnalysis(userID uint, period string, startDate, endDate time.Time) (*domain.IncomeExpenseAnalysis, error) {
	args := m.Called(userID, period, startDate, endDate)
	return args.Get(0).(*domain.IncomeExpenseAnalysis), args.Error(1)
}

func (m *MockAnalyticsService) GetCategoryAnalysis(userID, categoryID uint, startDate, endDate time.Time) (*domain.CategoryMetrics, error) {
	args := m.Called(userID, categoryID, startDate, endDate)
	return args.Get(0).(*domain.CategoryMetrics), args.Error(1)
}

func (m *MockAnalyticsService) GetDashboardSummary(userID uint, period string) (*domain.DashboardSummary, error) {
	args := m.Called(userID, period)
	return args.Get(0).(*domain.DashboardSummary), args.Error(1)
}

func setupAnalyticsHandler() (*AnalyticsHandler, *MockAnalyticsService) {
	mockService := &MockAnalyticsService{}
	handler := &AnalyticsHandler{
		Service: mockService,
	}
	return handler, mockService
}

func TestAnalyticsHandler_GetFinancialMetrics(t *testing.T) {
	t.Run("should get financial metrics successfully with default period", func(t *testing.T) {
		handler, mockService := setupAnalyticsHandler()
		router := setupGin()
		router.GET("/analytics/metrics/:userId", handler.GetFinancialMetrics)

		metrics := &domain.FinancialMetrics{
			UserID:        1,
			Period:        "monthly",
			TotalIncome:   5000.0,
			TotalExpenses: 3500.0,
			NetIncome:     1500.0,
			SavingsRate:   0.30,
			ExpenseRatio:  0.70,
			// BudgetUtilization field doesn't exist, removing
			StartDate: time.Now().AddDate(0, -1, 0),
			EndDate:   time.Now(),
		}

		mockService.On("GetFinancialMetrics", uint(1), "monthly", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(metrics, nil)

		req := httptest.NewRequest("GET", "/analytics/metrics/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response domain.FinancialMetrics
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, uint(1), response.UserID)
		assert.Equal(t, "monthly", response.Period)
		assert.Equal(t, 5000.0, response.TotalIncome)
		assert.Equal(t, 3500.0, response.TotalExpenses)
		mockService.AssertExpectations(t)
	})

	t.Run("should get financial metrics with custom period and dates", func(t *testing.T) {
		handler, mockService := setupAnalyticsHandler()
		router := setupGin()
		router.GET("/analytics/metrics/:userId", handler.GetFinancialMetrics)

		metrics := &domain.FinancialMetrics{
			UserID:        1,
			Period:        "weekly",
			TotalIncome:   1200.0,
			TotalExpenses: 800.0,
			NetIncome:     400.0,
			SavingsRate:   0.33,
			ExpenseRatio:  0.67,
			// BudgetUtilization field doesn't exist, removing
			StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		}

		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
		mockService.On("GetFinancialMetrics", uint(1), "weekly", startDate, endDate).Return(metrics, nil)

		req := httptest.NewRequest("GET", "/analytics/metrics/1?period=weekly&start_date=2024-01-01&end_date=2024-01-31", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response domain.FinancialMetrics
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "weekly", response.Period)
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid user ID", func(t *testing.T) {
		handler, _ := setupAnalyticsHandler()
		router := setupGin()
		router.GET("/analytics/metrics/:userId", handler.GetFinancialMetrics)

		req := httptest.NewRequest("GET", "/analytics/metrics/invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "invalid user ID", response["error"])
	})

	t.Run("should return bad request for invalid start date", func(t *testing.T) {
		handler, _ := setupAnalyticsHandler()
		router := setupGin()
		router.GET("/analytics/metrics/:userId", handler.GetFinancialMetrics)

		req := httptest.NewRequest("GET", "/analytics/metrics/1?start_date=invalid-date", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid start date format. Use YYYY-MM-DD", response["error"])
	})

	t.Run("should return bad request for invalid end date", func(t *testing.T) {
		handler, _ := setupAnalyticsHandler()
		router := setupGin()
		router.GET("/analytics/metrics/:userId", handler.GetFinancialMetrics)

		req := httptest.NewRequest("GET", "/analytics/metrics/1?end_date=invalid-date", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid end date format. Use YYYY-MM-DD", response["error"])
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		handler, mockService := setupAnalyticsHandler()
		router := setupGin()
		router.GET("/analytics/metrics/:userId", handler.GetFinancialMetrics)

		mockService.On("GetFinancialMetrics", uint(1), "monthly", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return((*domain.FinancialMetrics)(nil), errors.New("service error"))

		req := httptest.NewRequest("GET", "/analytics/metrics/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Failed to calculate financial metrics", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestAnalyticsHandler_GetIncomeExpenseAnalysis(t *testing.T) {
	t.Run("should get income expense analysis successfully", func(t *testing.T) {
		handler, mockService := setupAnalyticsHandler()
		router := setupGin()
		router.GET("/analytics/income-expense/:userId", handler.GetIncomeExpenseAnalysis)

		analysis := &domain.IncomeExpenseAnalysis{
			UserID:    1,
			Period:    "monthly",
			StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
			IncomeBreakdown: []domain.CategoryMetrics{
				{CategoryName: "salary", TotalAmount: 4500.0},
				{CategoryName: "freelance", TotalAmount: 500.0},
			},
			ExpenseBreakdown: []domain.CategoryMetrics{
				{CategoryName: "food", TotalAmount: 800.0},
				{CategoryName: "transport", TotalAmount: 300.0},
			},
			TopIncomeCategories: []domain.CategoryMetrics{
				{CategoryName: "salary", TotalAmount: 4500.0},
			},
			TopExpenseCategories: []domain.CategoryMetrics{
				{CategoryName: "food", TotalAmount: 800.0},
			},
			DailyAverages: domain.DailyAverages{
				DailyIncome:    161.29,
				DailyExpenses:  112.90,
				DailyNetIncome: 48.39,
			},
			WeeklyTrends: []domain.WeeklyTrend{
				{WeekNumber: 1, Income: 5000.0, Expenses: 3500.0, NetIncome: 1500.0},
			},
		}

		mockService.On("GetIncomeExpenseAnalysis", uint(1), "monthly", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(analysis, nil)

		req := httptest.NewRequest("GET", "/analytics/income-expense/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response domain.IncomeExpenseAnalysis
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, uint(1), response.UserID)
		assert.Equal(t, "monthly", response.Period)
		assert.NotEmpty(t, response.IncomeBreakdown)
		assert.NotEmpty(t, response.ExpenseBreakdown)
		mockService.AssertExpectations(t)
	})

	t.Run("should get income expense analysis with custom parameters", func(t *testing.T) {
		handler, mockService := setupAnalyticsHandler()
		router := setupGin()
		router.GET("/analytics/income-expense/:userId", handler.GetIncomeExpenseAnalysis)

		analysis := &domain.IncomeExpenseAnalysis{
			UserID:    1,
			Period:    "weekly",
			StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC),
			IncomeBreakdown: []domain.CategoryMetrics{
				{CategoryName: "salary", TotalAmount: 1000.0},
				{CategoryName: "bonus", TotalAmount: 200.0},
			},
			ExpenseBreakdown: []domain.CategoryMetrics{
				{CategoryName: "food", TotalAmount: 200.0},
				{CategoryName: "transport", TotalAmount: 100.0},
			},
			TopIncomeCategories: []domain.CategoryMetrics{
				{CategoryName: "salary", TotalAmount: 1000.0},
			},
			TopExpenseCategories: []domain.CategoryMetrics{
				{CategoryName: "food", TotalAmount: 200.0},
			},
			DailyAverages: domain.DailyAverages{
				DailyIncome:    171.43,
				DailyExpenses:  114.29,
				DailyNetIncome: 57.14,
			},
			WeeklyTrends: []domain.WeeklyTrend{
				{WeekNumber: 1, Income: 1200.0, Expenses: 800.0, NetIncome: 400.0},
			},
		}

		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)
		mockService.On("GetIncomeExpenseAnalysis", uint(1), "weekly", startDate, endDate).Return(analysis, nil)

		req := httptest.NewRequest("GET", "/analytics/income-expense/1?period=weekly&start_date=2024-01-01&end_date=2024-01-07", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response domain.IncomeExpenseAnalysis
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "weekly", response.Period)
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid user ID", func(t *testing.T) {
		handler, _ := setupAnalyticsHandler()
		router := setupGin()
		router.GET("/analytics/income-expense/:userId", handler.GetIncomeExpenseAnalysis)

		req := httptest.NewRequest("GET", "/analytics/income-expense/invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "invalid user ID", response["error"])
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		handler, mockService := setupAnalyticsHandler()
		router := setupGin()
		router.GET("/analytics/income-expense/:userId", handler.GetIncomeExpenseAnalysis)

		mockService.On("GetIncomeExpenseAnalysis", uint(1), "monthly", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return((*domain.IncomeExpenseAnalysis)(nil), errors.New("service error"))

		req := httptest.NewRequest("GET", "/analytics/income-expense/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Failed to generate income-expense analysis", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestAnalyticsHandler_GetCategoryAnalysis(t *testing.T) {
	t.Run("should get category analysis successfully", func(t *testing.T) {
		handler, mockService := setupAnalyticsHandler()
		router := setupGin()
		router.GET("/analytics/category/:userId/:categoryId", handler.GetCategoryAnalysis)

		analysis := &domain.CategoryMetrics{
			CategoryID:        5,
			CategoryName:      "Food",
			TotalAmount:       800.0,
			TransactionCount:  25,
			AverageAmount:     32.0,
			PercentageOfTotal: 22.8,
			Trend:             "stable",
		}

		mockService.On("GetCategoryAnalysis", uint(1), uint(5), mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(analysis, nil)

		req := httptest.NewRequest("GET", "/analytics/category/1/5", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response domain.CategoryMetrics
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, uint(5), response.CategoryID)
		assert.Equal(t, "Food", response.CategoryName)
		assert.Equal(t, 800.0, response.TotalAmount)
		assert.Equal(t, 25, response.TransactionCount)
		mockService.AssertExpectations(t)
	})

	t.Run("should get category analysis with custom date range", func(t *testing.T) {
		handler, mockService := setupAnalyticsHandler()
		router := setupGin()
		router.GET("/analytics/category/:userId/:categoryId", handler.GetCategoryAnalysis)

		analysis := &domain.CategoryMetrics{
			CategoryID:        3,
			CategoryName:      "Transport",
			TotalAmount:       300.0,
			TransactionCount:  10,
			AverageAmount:     30.0,
			PercentageOfTotal: 8.5,
			Trend:             "stable",
		}

		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		mockService.On("GetCategoryAnalysis", uint(1), uint(3), startDate, endDate).Return(analysis, nil)

		req := httptest.NewRequest("GET", "/analytics/category/1/3?start_date=2024-01-01&end_date=2024-01-15", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response domain.CategoryMetrics
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Transport", response.CategoryName)
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid user ID", func(t *testing.T) {
		handler, _ := setupAnalyticsHandler()
		router := setupGin()
		router.GET("/analytics/category/:userId/:categoryId", handler.GetCategoryAnalysis)

		req := httptest.NewRequest("GET", "/analytics/category/invalid/5", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "invalid user ID", response["error"])
	})

	t.Run("should return bad request for invalid category ID", func(t *testing.T) {
		handler, _ := setupAnalyticsHandler()
		router := setupGin()
		router.GET("/analytics/category/:userId/:categoryId", handler.GetCategoryAnalysis)

		req := httptest.NewRequest("GET", "/analytics/category/1/invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "invalid category ID", response["error"])
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		handler, mockService := setupAnalyticsHandler()
		router := setupGin()
		router.GET("/analytics/category/:userId/:categoryId", handler.GetCategoryAnalysis)

		mockService.On("GetCategoryAnalysis", uint(1), uint(5), mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return((*domain.CategoryMetrics)(nil), errors.New("service error"))

		req := httptest.NewRequest("GET", "/analytics/category/1/5", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Failed to generate category analysis", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestAnalyticsHandler_GetDashboardSummary(t *testing.T) {
	t.Run("should get dashboard summary successfully", func(t *testing.T) {
		handler, mockService := setupAnalyticsHandler()
		router := setupGin()
		router.GET("/analytics/dashboard/:userId", handler.GetDashboardSummary)

		dashboard := &domain.DashboardSummary{
			UserID:          1,
			Period:          "month",
			StartDate:       time.Now().AddDate(0, -1, 0),
			EndDate:         time.Now(),
			TotalBalance:    1500.0,
			MonthlyIncome:   5000.0,
			MonthlyExpenses: 3500.0,
			MonthlySavings:  1500.0,
			SavingsRate:     30.0,
			TopExpenseCategories: []domain.CategoryMetrics{
				{CategoryName: "Food", TotalAmount: 800.0},
				{CategoryName: "Transport", TotalAmount: 300.0},
			},
			RecentTransactions: []domain.Transaction{{ID: 1, Amount: 150.0, Description: "Grocery shopping"}},
			BudgetAlerts:       []domain.BudgetAlert{{BudgetID: 1, CategoryName: "Rent", BudgetAmount: 1200.0, SpentAmount: 1000.0, PercentageUsed: 83.3, AlertLevel: "warning", DaysRemaining: 5}},
			FinancialGoals:     []domain.FinancialGoal{},
			QuickStats:         domain.QuickStats{TotalTransactions: 25, AverageTransaction: 140.0, LargestExpense: 500.0, MostUsedCategory: "Food", DaysUntilNextBudget: 15, CashFlowTrend: "positive"},
		}

		mockService.On("GetDashboardSummary", uint(1), "month").Return(dashboard, nil)

		req := httptest.NewRequest("GET", "/analytics/dashboard/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response domain.DashboardSummary
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, uint(1), response.UserID)
		assert.Equal(t, "month", response.Period)
		assert.Equal(t, 5000.0, response.MonthlyIncome)
		assert.Equal(t, 3500.0, response.MonthlyExpenses)
		assert.Equal(t, 1500.0, response.MonthlySavings)
		mockService.AssertExpectations(t)
	})

	t.Run("should get dashboard summary with custom period", func(t *testing.T) {
		handler, mockService := setupAnalyticsHandler()
		router := setupGin()
		router.GET("/analytics/dashboard/:userId", handler.GetDashboardSummary)

		dashboard := &domain.DashboardSummary{
			UserID:          1,
			Period:          "week",
			StartDate:       time.Now().AddDate(0, 0, -7),
			EndDate:         time.Now(),
			TotalBalance:    400.0,
			MonthlyIncome:   1200.0,
			MonthlyExpenses: 800.0,
			MonthlySavings:  400.0,
			SavingsRate:     33.0,
			TopExpenseCategories: []domain.CategoryMetrics{
				{CategoryName: "Food", TotalAmount: 200.0},
			},
			RecentTransactions: []domain.Transaction{{ID: 2, Amount: 50.0, Description: "Coffee shop"}},
			BudgetAlerts:       []domain.BudgetAlert{},
			FinancialGoals:     []domain.FinancialGoal{},
			QuickStats:         domain.QuickStats{TotalTransactions: 10, AverageTransaction: 80.0, LargestExpense: 200.0, MostUsedCategory: "Food", DaysUntilNextBudget: 20, CashFlowTrend: "stable"},
		}

		mockService.On("GetDashboardSummary", uint(1), "week").Return(dashboard, nil)

		req := httptest.NewRequest("GET", "/analytics/dashboard/1?period=week", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response domain.DashboardSummary
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "week", response.Period)
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid user ID", func(t *testing.T) {
		handler, _ := setupAnalyticsHandler()
		router := setupGin()
		router.GET("/analytics/dashboard/:userId", handler.GetDashboardSummary)

		req := httptest.NewRequest("GET", "/analytics/dashboard/invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "invalid user ID", response["error"])
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		handler, mockService := setupAnalyticsHandler()
		router := setupGin()
		router.GET("/analytics/dashboard/:userId", handler.GetDashboardSummary)

		mockService.On("GetDashboardSummary", uint(1), "month").Return((*domain.DashboardSummary)(nil), errors.New("service error"))

		req := httptest.NewRequest("GET", "/analytics/dashboard/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Failed to generate dashboard summary", response["error"])
		mockService.AssertExpectations(t)
	})
}
