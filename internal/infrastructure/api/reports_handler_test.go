package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go-finance-advisor/internal/domain"
)

// MockReportsService is a mock implementation of ReportsService
type MockReportsService struct {
	mock.Mock
}

func (m *MockReportsService) GenerateMonthlyReport(userID uint, year, month int) (*domain.FinancialReport, error) {
	args := m.Called(userID, year, month)
	return args.Get(0).(*domain.FinancialReport), args.Error(1)
}

func (m *MockReportsService) GenerateQuarterlyReport(userID uint, year, quarter int) (*domain.FinancialReport, error) {
	args := m.Called(userID, year, quarter)
	return args.Get(0).(*domain.FinancialReport), args.Error(1)
}

func (m *MockReportsService) GenerateYearlyReport(userID uint, year int) (*domain.FinancialReport, error) {
	args := m.Called(userID, year)
	return args.Get(0).(*domain.FinancialReport), args.Error(1)
}

func (m *MockReportsService) GenerateCustomReport(userID uint, startDate, endDate time.Time) (*domain.FinancialReport, error) {
	args := m.Called(userID, startDate, endDate)
	return args.Get(0).(*domain.FinancialReport), args.Error(1)
}

func setupReportsHandler() (*ReportsHandler, *MockReportsService) {
	mockService := new(MockReportsService)
	handler := &ReportsHandler{Service: mockService}
	return handler, mockService
}

func TestReportsHandler_GenerateMonthlyReport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful monthly report generation", func(t *testing.T) {
		handler, mockService := setupReportsHandler()
		expectedReport := &domain.FinancialReport{
			ID:            1,
			UserID:        1,
			ReportType:    "monthly",
			TotalIncome:   5000.00,
			TotalExpenses: 3000.00,
			NetIncome:     2000.00,
		}

		mockService.On("GenerateMonthlyReport", uint(1), 2023, 6).Return(expectedReport, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "1"},
			{Key: "year", Value: "2023"},
			{Key: "month", Value: "6"},
		}

		handler.GenerateMonthlyReport(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response domain.FinancialReport
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedReport.ID, response.ID)
		assert.Equal(t, expectedReport.UserID, response.UserID)
		assert.Equal(t, expectedReport.ReportType, response.ReportType)
		assert.Equal(t, expectedReport.ReportType, response.ReportType)
		assert.Equal(t, expectedReport.TotalIncome, response.TotalIncome)
		assert.Equal(t, expectedReport.TotalExpenses, response.TotalExpenses)
		assert.Equal(t, expectedReport.NetIncome, response.NetIncome)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid user ID", func(t *testing.T) {
		handler, _ := setupReportsHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "invalid"},
			{Key: "year", Value: "2023"},
			{Key: "month", Value: "6"},
		}

		handler.GenerateMonthlyReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid user ID", response["error"])
	})

	t.Run("invalid year", func(t *testing.T) {
		handler, _ := setupReportsHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "1"},
			{Key: "year", Value: "invalid"},
			{Key: "month", Value: "6"},
		}

		handler.GenerateMonthlyReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid year", response["error"])
	})

	t.Run("invalid month", func(t *testing.T) {
		handler, _ := setupReportsHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "1"},
			{Key: "year", Value: "2023"},
			{Key: "month", Value: "13"},
		}

		handler.GenerateMonthlyReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid month", response["error"])
	})

	t.Run("service error", func(t *testing.T) {
		handler, mockService := setupReportsHandler()
		mockService.On("GenerateMonthlyReport", uint(1), 2023, 6).Return((*domain.FinancialReport)(nil), errors.New("service error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "1"},
			{Key: "year", Value: "2023"},
			{Key: "month", Value: "6"},
		}

		handler.GenerateMonthlyReport(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to generate monthly report", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestReportsHandler_GenerateQuarterlyReport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful quarterly report generation", func(t *testing.T) {
		handler, mockService := setupReportsHandler()
		expectedReport := &domain.FinancialReport{
			ID:            1,
			UserID:        1,
			ReportType:    "quarterly",
			TotalIncome:   15000.00,
			TotalExpenses: 9000.00,
			NetIncome:     6000.00,
		}

		mockService.On("GenerateQuarterlyReport", uint(1), 2023, 2).Return(expectedReport, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "1"},
			{Key: "year", Value: "2023"},
			{Key: "quarter", Value: "2"},
		}

		handler.GenerateQuarterlyReport(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response domain.FinancialReport
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedReport.ID, response.ID)
		assert.Equal(t, expectedReport.UserID, response.UserID)
		assert.Equal(t, expectedReport.ReportType, response.ReportType)
		assert.Equal(t, expectedReport.ReportType, response.ReportType)
		assert.Equal(t, expectedReport.TotalIncome, response.TotalIncome)
		assert.Equal(t, expectedReport.TotalExpenses, response.TotalExpenses)
		assert.Equal(t, expectedReport.NetIncome, response.NetIncome)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid user ID", func(t *testing.T) {
		handler, _ := setupReportsHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "invalid"},
			{Key: "year", Value: "2023"},
			{Key: "quarter", Value: "2"},
		}

		handler.GenerateQuarterlyReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid user ID", response["error"])
	})

	t.Run("invalid year", func(t *testing.T) {
		handler, _ := setupReportsHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "1"},
			{Key: "year", Value: "invalid"},
			{Key: "quarter", Value: "2"},
		}

		handler.GenerateQuarterlyReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid year", response["error"])
	})

	t.Run("invalid quarter", func(t *testing.T) {
		handler, _ := setupReportsHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "1"},
			{Key: "year", Value: "2023"},
			{Key: "quarter", Value: "5"},
		}

		handler.GenerateQuarterlyReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid quarter", response["error"])
	})

	t.Run("service error", func(t *testing.T) {
		handler, mockService := setupReportsHandler()
		mockService.On("GenerateQuarterlyReport", uint(1), 2023, 2).Return((*domain.FinancialReport)(nil), errors.New("service error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "1"},
			{Key: "year", Value: "2023"},
			{Key: "quarter", Value: "2"},
		}

		handler.GenerateQuarterlyReport(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to generate quarterly report", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestReportsHandler_GenerateYearlyReport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful yearly report generation", func(t *testing.T) {
		handler, mockService := setupReportsHandler()
		expectedReport := &domain.FinancialReport{
			ID:            1,
			UserID:        1,
			ReportType:    "yearly",
			TotalIncome:   60000.00,
			TotalExpenses: 36000.00,
			NetIncome:     24000.00,
		}

		mockService.On("GenerateYearlyReport", uint(1), 2023).Return(expectedReport, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "1"},
			{Key: "year", Value: "2023"},
		}

		handler.GenerateYearlyReport(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response domain.FinancialReport
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedReport.ID, response.ID)
		assert.Equal(t, expectedReport.UserID, response.UserID)
		assert.Equal(t, expectedReport.ReportType, response.ReportType)
		assert.Equal(t, expectedReport.TotalIncome, response.TotalIncome)
		assert.Equal(t, expectedReport.TotalExpenses, response.TotalExpenses)
		assert.Equal(t, expectedReport.NetIncome, response.NetIncome)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid user ID", func(t *testing.T) {
		handler, _ := setupReportsHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "invalid"},
			{Key: "year", Value: "2023"},
		}

		handler.GenerateYearlyReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid user ID", response["error"])
	})

	t.Run("invalid year", func(t *testing.T) {
		handler, _ := setupReportsHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "1"},
			{Key: "year", Value: "invalid"},
		}

		handler.GenerateYearlyReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid year", response["error"])
	})

	t.Run("service error", func(t *testing.T) {
		handler, mockService := setupReportsHandler()
		mockService.On("GenerateYearlyReport", uint(1), 2023).Return((*domain.FinancialReport)(nil), errors.New("service error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "1"},
			{Key: "year", Value: "2023"},
		}

		handler.GenerateYearlyReport(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to generate yearly report", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestReportsHandler_GenerateCustomReport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful custom report generation", func(t *testing.T) {
		handler, mockService := setupReportsHandler()
		startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2023, 6, 30, 0, 0, 0, 0, time.UTC)
		expectedReport := &domain.FinancialReport{
			ID:            1,
			UserID:        1,
			ReportType:    "custom",
			TotalIncome:   30000.00,
			TotalExpenses: 18000.00,
			NetIncome:     12000.00,
		}

		mockService.On("GenerateCustomReport", uint(1), startDate, endDate).Return(expectedReport, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "1"},
		}
		c.Request = httptest.NewRequest("GET", "/reports/custom/1?start_date=2023-01-01&end_date=2023-06-30", nil)

		handler.GenerateCustomReport(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response domain.FinancialReport
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedReport.ID, response.ID)
		assert.Equal(t, expectedReport.UserID, response.UserID)
		assert.Equal(t, expectedReport.ReportType, response.ReportType)
		assert.Equal(t, expectedReport.ReportType, response.ReportType)
		assert.Equal(t, expectedReport.TotalIncome, response.TotalIncome)
		assert.Equal(t, expectedReport.TotalExpenses, response.TotalExpenses)
		assert.Equal(t, expectedReport.NetIncome, response.NetIncome)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid user ID", func(t *testing.T) {
		handler, _ := setupReportsHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "invalid"},
		}
		c.Request = httptest.NewRequest("GET", "/reports/custom/invalid?start_date=2023-01-01&end_date=2023-06-30", nil)

		handler.GenerateCustomReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid user ID", response["error"])
	})

	t.Run("missing start_date", func(t *testing.T) {
		handler, _ := setupReportsHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "1"},
		}
		c.Request = httptest.NewRequest("GET", "/reports/custom/1?end_date=2023-06-30", nil)

		handler.GenerateCustomReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "start_date and end_date are required", response["error"])
	})

	t.Run("missing end_date", func(t *testing.T) {
		handler, _ := setupReportsHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "1"},
		}
		c.Request = httptest.NewRequest("GET", "/reports/custom/1?start_date=2023-01-01", nil)

		handler.GenerateCustomReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "start_date and end_date are required", response["error"])
	})

	t.Run("invalid start_date format", func(t *testing.T) {
		handler, _ := setupReportsHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "1"},
		}
		c.Request = httptest.NewRequest("GET", "/reports/custom/1?start_date=invalid&end_date=2023-06-30", nil)

		handler.GenerateCustomReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid start_date format. Use YYYY-MM-DD", response["error"])
	})

	t.Run("invalid end_date format", func(t *testing.T) {
		handler, _ := setupReportsHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "1"},
		}
		c.Request = httptest.NewRequest("GET", "/reports/custom/1?start_date=2023-01-01&end_date=invalid", nil)

		handler.GenerateCustomReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid end_date format. Use YYYY-MM-DD", response["error"])
	})

	t.Run("start_date after end_date", func(t *testing.T) {
		handler, _ := setupReportsHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "1"},
		}
		c.Request = httptest.NewRequest("GET", "/reports/custom/1?start_date=2023-06-30&end_date=2023-01-01", nil)

		handler.GenerateCustomReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "start_date must be before end_date", response["error"])
	})

	t.Run("service error", func(t *testing.T) {
		handler, mockService := setupReportsHandler()
		startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2023, 6, 30, 0, 0, 0, 0, time.UTC)
		mockService.On("GenerateCustomReport", uint(1), startDate, endDate).Return((*domain.FinancialReport)(nil), errors.New("service error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "1"},
		}
		c.Request = httptest.NewRequest("GET", "/reports/custom/1?start_date=2023-01-01&end_date=2023-06-30", nil)

		handler.GenerateCustomReport(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to generate custom report", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestReportsHandler_GetReportsList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful reports list retrieval", func(t *testing.T) {
		handler, _ := setupReportsHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "1"},
		}

		handler.GetReportsList(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "reports")
		assert.Contains(t, response, "message")
		assert.Equal(t, "Reports are generated on-demand. Use the generate endpoints to create new reports.", response["message"])
		reports := response["reports"].([]interface{})
		assert.Len(t, reports, 0)
	})

	t.Run("invalid user ID", func(t *testing.T) {
		handler, _ := setupReportsHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "userId", Value: "invalid"},
		}

		handler.GetReportsList(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid user ID", response["error"])
	})
}
