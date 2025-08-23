package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-finance-advisor/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockExportService is a mock implementation of ExportService
type MockExportService struct {
	mock.Mock
}

func (m *MockExportService) ExportTransactions(
	userID uint, format domain.ExportFormat, startDate, endDate *time.Time,
) (data []byte, filename string, err error) {
	args := m.Called(userID, format, startDate, endDate)
	return args.Get(0).([]byte), args.String(1), args.Error(2)
}

func (m *MockExportService) ExportBudgets(userID uint, format domain.ExportFormat) (data []byte, filename string, err error) {
	args := m.Called(userID, format)
	return args.Get(0).([]byte), args.String(1), args.Error(2)
}

func (m *MockExportService) ExportFinancialReport(
	userID uint, reportType string, year, month int, format domain.ExportFormat,
) (data []byte, filename string, err error) {
	args := m.Called(userID, reportType, year, month, format)
	return args.Get(0).([]byte), args.String(1), args.Error(2)
}

func (m *MockExportService) ExportAllData(userID uint, format domain.ExportFormat) (data []byte, filename string, err error) {
	args := m.Called(userID, format)
	return args.Get(0).([]byte), args.String(1), args.Error(2)
}

func setupExportHandler() (*ExportHandler, *MockExportService) {
	mockService := new(MockExportService)
	handler := &ExportHandler{Service: mockService}
	return handler, mockService
}

func TestExportHandler_ExportTransactions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful export with default format", func(t *testing.T) {
		handler, mockService := setupExportHandler()
		expectedData := []byte("transaction,amount\nTest,100.00")
		expectedFilename := "transactions.csv"

		mockService.On("ExportTransactions", uint(1), domain.ExportFormatCSV, (*time.Time)(nil), (*time.Time)(nil)).
			Return(expectedData, expectedFilename, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Request = httptest.NewRequest("GET", "/export/transactions", http.NoBody)

		handler.ExportTransactions(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), "transactions.csv")
		assert.Equal(t, expectedData, w.Body.Bytes())
		mockService.AssertExpectations(t)
	})

	t.Run("successful export with JSON format and date range", func(t *testing.T) {
		handler, mockService := setupExportHandler()
		expectedData := []byte(`{"transactions":[{"id":1,"amount":100.00}]}`)
		expectedFilename := "transactions.json"
		startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)

		mockService.On("ExportTransactions", uint(1), domain.ExportFormatJSON, &startDate, &endDate).Return(expectedData, expectedFilename, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Request = httptest.NewRequest("GET", "/export/transactions?format=json&start_date=2023-01-01&end_date=2023-12-31", http.NoBody)

		handler.ExportTransactions(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), "transactions.json")
		assert.Equal(t, expectedData, w.Body.Bytes())
		mockService.AssertExpectations(t)
	})

	t.Run("user not authenticated", func(t *testing.T) {
		handler, _ := setupExportHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/export/transactions", http.NoBody)

		handler.ExportTransactions(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "User not authenticated", response["error"])
	})

	t.Run("invalid export format", func(t *testing.T) {
		handler, _ := setupExportHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Request = httptest.NewRequest("GET", "/export/transactions?format=invalid", http.NoBody)

		handler.ExportTransactions(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid export format", response["error"])
	})

	t.Run("service error", func(t *testing.T) {
		handler, mockService := setupExportHandler()
		mockService.On("ExportTransactions", uint(1), domain.ExportFormatCSV, (*time.Time)(nil), (*time.Time)(nil)).
			Return([]byte{}, "", errors.New("export failed"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Request = httptest.NewRequest("GET", "/export/transactions", http.NoBody)

		handler.ExportTransactions(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "export failed", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestExportHandler_ExportBudgets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful export with default format", func(t *testing.T) {
		handler, mockService := setupExportHandler()
		expectedData := []byte("budget,amount\nFood,500.00")
		expectedFilename := "budgets.csv"

		mockService.On("ExportBudgets", uint(1), domain.ExportFormatCSV).Return(expectedData, expectedFilename, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Request = httptest.NewRequest("GET", "/export/budgets", http.NoBody)

		handler.ExportBudgets(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), "budgets.csv")
		assert.Equal(t, expectedData, w.Body.Bytes())
		mockService.AssertExpectations(t)
	})

	t.Run("successful export with JSON format", func(t *testing.T) {
		handler, mockService := setupExportHandler()
		expectedData := []byte(`{"budgets":[{"id":1,"amount":500.00}]}`)
		expectedFilename := "budgets.json"

		mockService.On("ExportBudgets", uint(1), domain.ExportFormatJSON).Return(expectedData, expectedFilename, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Request = httptest.NewRequest("GET", "/export/budgets?format=json", http.NoBody)

		handler.ExportBudgets(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), "budgets.json")
		assert.Equal(t, expectedData, w.Body.Bytes())
		mockService.AssertExpectations(t)
	})

	t.Run("user not authenticated", func(t *testing.T) {
		handler, _ := setupExportHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/export/budgets", http.NoBody)

		handler.ExportBudgets(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "User not authenticated", response["error"])
	})

	t.Run("invalid export format", func(t *testing.T) {
		handler, _ := setupExportHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Request = httptest.NewRequest("GET", "/export/budgets?format=xml", http.NoBody)

		handler.ExportBudgets(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid export format", response["error"])
	})

	t.Run("service error", func(t *testing.T) {
		handler, mockService := setupExportHandler()
		mockService.On("ExportBudgets", uint(1), domain.ExportFormatCSV).Return([]byte{}, "", errors.New("export failed"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Request = httptest.NewRequest("GET", "/export/budgets", http.NoBody)

		handler.ExportBudgets(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "export failed", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestExportHandler_ExportFinancialReport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful monthly report export", func(t *testing.T) {
		handler, mockService := setupExportHandler()
		expectedData := []byte(`{"report":{"type":"monthly","year":2023,"month":6}}`)
		expectedFilename := "monthly_report_2023_06.json"

		mockService.On("ExportFinancialReport", uint(1), "monthly", 2023, 6, domain.ExportFormatJSON).Return(expectedData, expectedFilename, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Request = httptest.NewRequest("GET", "/export/reports?type=monthly&year=2023&month=6&format=json", http.NoBody)

		handler.ExportFinancialReport(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), "monthly_report_2023_06.json")
		assert.Equal(t, expectedData, w.Body.Bytes())
		mockService.AssertExpectations(t)
	})

	t.Run("successful yearly report export", func(t *testing.T) {
		handler, mockService := setupExportHandler()
		expectedData := []byte(`{"report":{"type":"yearly","year":2023}}`)
		expectedFilename := "yearly_report_2023.json"

		mockService.On("ExportFinancialReport", uint(1), "yearly", 2023, 0, domain.ExportFormatJSON).Return(expectedData, expectedFilename, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Request = httptest.NewRequest("GET", "/export/reports?type=yearly&year=2023", http.NoBody)

		handler.ExportFinancialReport(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), "yearly_report_2023.json")
		assert.Equal(t, expectedData, w.Body.Bytes())
		mockService.AssertExpectations(t)
	})

	t.Run("user not authenticated", func(t *testing.T) {
		handler, _ := setupExportHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/export/reports?type=monthly&year=2023&month=6", http.NoBody)

		handler.ExportFinancialReport(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "User not authenticated", response["error"])
	})

	t.Run("missing report type", func(t *testing.T) {
		handler, _ := setupExportHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Request = httptest.NewRequest("GET", "/export/reports?year=2023", http.NoBody)

		handler.ExportFinancialReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Report type is required", response["error"])
	})

	t.Run("missing year", func(t *testing.T) {
		handler, _ := setupExportHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Request = httptest.NewRequest("GET", "/export/reports?type=monthly", http.NoBody)

		handler.ExportFinancialReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Year is required", response["error"])
	})

	t.Run("invalid year", func(t *testing.T) {
		handler, _ := setupExportHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Request = httptest.NewRequest("GET", "/export/reports?type=monthly&year=invalid", http.NoBody)

		handler.ExportFinancialReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid year", response["error"])
	})

	t.Run("invalid month", func(t *testing.T) {
		handler, _ := setupExportHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Request = httptest.NewRequest("GET", "/export/reports?type=monthly&year=2023&month=13", http.NoBody)

		handler.ExportFinancialReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid month", response["error"])
	})

	t.Run("missing month for monthly report", func(t *testing.T) {
		handler, _ := setupExportHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Request = httptest.NewRequest("GET", "/export/reports?type=monthly&year=2023", http.NoBody)

		handler.ExportFinancialReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Month is required for monthly reports", response["error"])
	})

	t.Run("invalid export format", func(t *testing.T) {
		handler, _ := setupExportHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Request = httptest.NewRequest("GET", "/export/reports?type=monthly&year=2023&month=6&format=xml", http.NoBody)

		handler.ExportFinancialReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid export format", response["error"])
	})

	t.Run("service error", func(t *testing.T) {
		handler, mockService := setupExportHandler()
		mockService.On("ExportFinancialReport", uint(1), "monthly", 2023, 6, domain.ExportFormatJSON).
			Return([]byte{}, "", errors.New("export failed"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Request = httptest.NewRequest("GET", "/export/reports?type=monthly&year=2023&month=6", http.NoBody)

		handler.ExportFinancialReport(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "export failed", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestExportHandler_ExportAllData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful export with default format", func(t *testing.T) {
		handler, mockService := setupExportHandler()
		expectedData := []byte(`{"user_data":{"transactions":[],"budgets":[],"categories":[]}}`)
		expectedFilename := "all_data.json"

		mockService.On("ExportAllData", uint(1), domain.ExportFormatJSON).Return(expectedData, expectedFilename, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Request = httptest.NewRequest("GET", "/export/all", http.NoBody)

		handler.ExportAllData(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), "all_data.json")
		assert.Equal(t, expectedData, w.Body.Bytes())
		mockService.AssertExpectations(t)
	})

	t.Run("successful export with explicit JSON format", func(t *testing.T) {
		handler, mockService := setupExportHandler()
		expectedData := []byte(`{"user_data":{"transactions":[],"budgets":[],"categories":[]}}`)
		expectedFilename := "all_data.json"

		mockService.On("ExportAllData", uint(1), domain.ExportFormatJSON).Return(expectedData, expectedFilename, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Request = httptest.NewRequest("GET", "/export/all?format=json", http.NoBody)

		handler.ExportAllData(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), "all_data.json")
		assert.Equal(t, expectedData, w.Body.Bytes())
		mockService.AssertExpectations(t)
	})

	t.Run("user not authenticated", func(t *testing.T) {
		handler, _ := setupExportHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/export/all", http.NoBody)

		handler.ExportAllData(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "User not authenticated", response["error"])
	})

	t.Run("unsupported format", func(t *testing.T) {
		handler, _ := setupExportHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Request = httptest.NewRequest("GET", "/export/all?format=csv", http.NoBody)

		handler.ExportAllData(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Only JSON format is supported for all data export", response["error"])
	})

	t.Run("service error", func(t *testing.T) {
		handler, mockService := setupExportHandler()
		mockService.On("ExportAllData", uint(1), domain.ExportFormatJSON).Return([]byte{}, "", errors.New("export failed"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Request = httptest.NewRequest("GET", "/export/all", http.NoBody)

		handler.ExportAllData(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "export failed", response["error"])
		mockService.AssertExpectations(t)
	})
}

func TestExportHandler_GetExportFormats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful retrieval", func(t *testing.T) {
		handler, _ := setupExportHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/export/formats", http.NoBody)

		handler.GetExportFormats(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check formats
		formats, ok := response["formats"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, formats, 2)

		// Check data types
		dataTypes, ok := response["data_types"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, dataTypes, 4)

		// Verify CSV format
		csvFormat := formats[0].(map[string]interface{})
		assert.Equal(t, "csv", csvFormat["value"])
		assert.Equal(t, "CSV", csvFormat["label"])
		assert.Equal(t, "text/csv", csvFormat["mime_type"])

		// Verify JSON format
		jsonFormat := formats[1].(map[string]interface{})
		assert.Equal(t, "json", jsonFormat["value"])
		assert.Equal(t, "JSON", jsonFormat["label"])
		assert.Equal(t, "application/json", jsonFormat["mime_type"])

		// Verify transactions data type
		transactionsType := dataTypes[0].(map[string]interface{})
		assert.Equal(t, "transactions", transactionsType["value"])
		assert.Equal(t, "Transactions", transactionsType["label"])
		assert.Equal(t, true, transactionsType["supports_date_range"])

		// Verify all data type
		allDataType := dataTypes[3].(map[string]interface{})
		assert.Equal(t, "all", allDataType["value"])
		assert.Equal(t, "All Data", allDataType["label"])
		assert.Equal(t, false, allDataType["supports_date_range"])
		supportedFormats := allDataType["supported_formats"].([]interface{})
		assert.Len(t, supportedFormats, 1)
		assert.Equal(t, "json", supportedFormats[0])
	})
}
