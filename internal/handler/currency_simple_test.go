package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockCurrencyService struct {
	// Конфигурация возвращаемых значений
	ShouldReturnError bool
	MockResult        float64
	MockRate          float64
	MockError         error

	// Для отслеживания вызовов
	Called     bool
	LastFrom   string
	LastTo     string
	LastAmount float64
	CallCount  int
}

func (m *MockCurrencyService) Convert(ctx context.Context, from, to string, amount float64) (float64, float64, error) {
	m.Called = true
	m.CallCount++
	m.LastFrom = from
	m.LastTo = to
	m.LastAmount = amount
	if m.ShouldReturnError {
		return 0, 0, m.MockError
	}
	return m.MockResult, m.MockRate, nil
}
func (m *MockCurrencyService) GetExchangeRate(ctx context.Context, from, to string) (float64, error) {
	return 0.0, nil
}

// setupTestRouter создаёт тестовый роутер с хендлером
func setupTestRouter(service *MockCurrencyService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := NewCurrencyHandler(service)
	router.GET("/convert", handler.Convert)

	return router
}

// performRequest выполняет тестовый запрос
func performRequest(router *gin.Engine, method, url string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, url, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestCurrencyHandler_Convert_Success(t *testing.T) {
	mockService := &MockCurrencyService{
		ShouldReturnError: false,
		MockResult:        85.23,
		MockRate:          0.8523,
	}
	router := setupTestRouter(mockService)
	w := performRequest(router, "GET", "/convert?from=USD&to=EUR&amount=100")
	assert.Equal(t, http.StatusOK, w.Code, "Ожидался статус 200 OK, получили: %d", w.Code)
	var response struct {
		From   string  `json:"from"`
		To     string  `json:"to"`
		Amount float64 `json:"amount"`
		Rate   float64 `json:"rate"`
		Result float64 `json:"result"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Ошибка парсинга JSON: %s", err)
	assert.Equal(t, "USD", response.From)
	assert.Equal(t, "EUR", response.To)
	assert.Equal(t, 100.0, response.Amount)
	assert.Equal(t, 0.8523, response.Rate)
	assert.Equal(t, 85.23, response.Result)
	assert.True(t, mockService.Called, "Сервис не был вызван")
	assert.Equal(t, "USD", mockService.LastFrom)
	assert.Equal(t, "EUR", mockService.LastTo)
	assert.Equal(t, 100.0, mockService.LastAmount)
	t.Log("ТЕСТ 1 ПРОЙДЕН: Успешная конвертация работает!")
}
func TestCurrencyHandler_Convert_ServiceError(t *testing.T) {
	mockService := &MockCurrencyService{
		ShouldReturnError: true,
		MockError:         assert.AnError,
	}
	router := setupTestRouter(mockService)
	w := performRequest(router, "GET", "/convert?from=USD&to=EUR&amount=100")
	assert.Equal(t, http.StatusInternalServerError, w.Code,
		"Ожидался статус 500, получили: %d", w.Code)

	var errorResponse struct {
		Error   string `json:"error"`
		Details string `json:"details"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)

	assert.Equal(t, "Conversion failed", errorResponse.Error)
	assert.Contains(t, errorResponse.Details, "assert.AnError")

	t.Log("ТЕСТ 2 ПРОЙДЕН: Обработка ошибки сервиса работает!")
}
func TestCurrencyHandler_Convert_ValidationError_CurrencyLength(t *testing.T) {
	mockService := &MockCurrencyService{}
	router := setupTestRouter(mockService)
	w := performRequest(router, "GET", "/convert?from=US&to=EUR&amount=100")
	assert.Equal(t, http.StatusBadRequest, w.Code,
		"Ожидался статус 400 при неправильной валюте")

	var errorResponse struct {
		Error   string `json:"error"`
		Details string `json:"details"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)

	assert.Equal(t, "Invalid request", errorResponse.Error)
	// Проверяем что в ошибке есть информация о длине
	assert.Contains(t, errorResponse.Details, "len",
		"Должна быть ошибка о длине валюты. Получено: %s", errorResponse.Details)
	assert.False(t, mockService.Called,
		"Сервис не должен вызываться при ошибке валидации")

	t.Log("ТЕСТ 3 ПРОЙДЕН: Валидация длины валюты работает!")
}
func TestCurrencyHandler_Convert_ValidationError_NegativeAmount(t *testing.T) {
	mockService := &MockCurrencyService{}
	router := setupTestRouter(mockService)
	w := performRequest(router, "GET", "/convert?from=USD&to=EUR&amount=-100")
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResponse struct {
		Error   string `json:"error"`
		Details string `json:"details"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Invalid request", errorResponse.Error)
	assert.Contains(t, errorResponse.Details, "min",
		"Должна быть ошибка о минимальной сумме. Получено: %s", errorResponse.Details)
	assert.False(t, mockService.Called)

	t.Log("ТЕСТ 4 ПРОЙДЕН: Валидация суммы (нельзя отрицательную) работает!")
}
func TestCurrencyHandler_Convert_SmallAmount(t *testing.T) {

	mockService := &MockCurrencyService{
		ShouldReturnError: false,
		MockResult:        0.008523,
		MockRate:          0.8523,
	}

	router := setupTestRouter(mockService)

	w := performRequest(router, "GET", "/convert?from=USD&to=EUR&amount=0.01")

	assert.Equal(t, http.StatusOK, w.Code,
		"Сумма 0.01 должна быть допустимой (min=0.01)")

	var response struct {
		Result float64 `json:"result"`
		Rate   float64 `json:"rate"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.InDelta(t, 0.008523, response.Result, 0.000001)
	assert.Equal(t, 0.8523, response.Rate)

	t.Log("ТЕСТ 5 ПРОЙДЕН: Конвертация маленькой суммы работает!")
}

func TestCurrencyHandler_Convert_MissingParameters(t *testing.T) {
	testCases := []struct {
		name        string
		url         string
		description string
	}{
		{
			name:        "MissingFrom",
			url:         "/convert?to=EUR&amount=100",
			description: "Отсутствует параметр 'from'",
		},
		{
			name:        "MissingTo",
			url:         "/convert?from=USD&amount=100",
			description: "Отсутствует параметр 'to'",
		},
		{
			name:        "MissingAmount",
			url:         "/convert?from=USD&to=EUR",
			description: "Отсутствует параметр 'amount'",
		},
		{
			name:        "AllMissing",
			url:         "/convert",
			description: "Все параметры отсутствуют",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ARRANGE
			mockService := &MockCurrencyService{}
			router := setupTestRouter(mockService)

			// ACT
			w := performRequest(router, "GET", tc.url)

			// ASSERT
			assert.Equal(t, http.StatusBadRequest, w.Code,
				"Для случая '%s' ожидался статус 400", tc.description)

			// Сервис не должен вызываться
			assert.False(t, mockService.Called,
				"Для случая '%s' сервис не должен вызываться", tc.description)
		})
	}

	t.Log("ТЕСТ 6 ПРОЙДЕН: Проверка обязательных параметров работает!")
}

// ============================================
// ТЕСТ 7: Некорректный формат числа
// ============================================

func TestCurrencyHandler_Convert_InvalidNumberFormat(t *testing.T) {

	mockService := &MockCurrencyService{}
	router := setupTestRouter(mockService)

	w := performRequest(router, "GET", "/convert?from=USD&to=EUR&amount=abc")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResponse struct {
		Error   string `json:"error"`
		Details string `json:"details"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)

	assert.Equal(t, "Invalid request", errorResponse.Error)
	assert.Contains(t, errorResponse.Details, "parsing")

	assert.False(t, mockService.Called)

	t.Log("ТЕСТ 7 ПРОЙДЕН: Валидация формата числа работает!")
}
