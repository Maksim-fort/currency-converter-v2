package service

import (
	"context"
	"currency-converter-v2/internal/config"
	"currency-converter-v2/pkg/cache"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// CurrencyServiceInterface - интерфейс для тестирования
type CurrencyServiceInterface interface {
	Convert(ctx context.Context, from, to string, amount float64) (float64, float64, error)
	GetExchangeRate(ctx context.Context, from, to string) (float64, error)
}
type CurrencyService struct {
	config     *config.Config
	redis      *cache.RedisClient
	logger     *zap.Logger
	httpClient *http.Client

	// TODO: Добавить поля структуры
}

func NewCurrencyService(cfg *config.Config, redisClient *cache.RedisClient, logger *zap.Logger) *CurrencyService {
	client := &http.Client{
		Timeout: cfg.API.Timeout,
	}
	return &CurrencyService{
		config:     cfg,
		redis:      redisClient,
		logger:     logger,
		httpClient: client,
	}
}

type DataConvert struct {
	Data map[string]float64 `json:"data"`
	// TODO: Реализовать конструктор
}
type ConversionResult struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
	Rate   float64 `json:"rate"`
	Result float64 `json:"result"`
}

func (s *CurrencyService) GetExchangeRate(ctx context.Context, from, to string) (float64, error) {
	if from == "" || to == "" {
		return 0, fmt.Errorf("currency codes cannot be empty")
	}
	if len(from) != 3 || len(to) != 3 {
		return 0, fmt.Errorf("currency codes must be 3 characters")
	}
	if strings.ToUpper(from) == strings.ToUpper(to) {
		return 1.0, nil
	}
	rate, err := s.redis.GetExchangeRate(ctx, from, to)
	if err == nil {
		s.logger.Debug("Cache hit",
			zap.String("from", from),
			zap.String("to", to),
			zap.Float64("rate", rate),
		)
		return rate, nil
	}
	if !strings.Contains(err.Error(), "not found") {
		// Реальная ошибка Redis (не "не найден")
		s.logger.Warn("Redis error (will try API)",
			zap.String("from", from),
			zap.String("to", to),
			zap.Error(err),
		)
	} else {
		// Cache miss - нормально
		s.logger.Debug("Cache miss",
			zap.String("from", from),
			zap.String("to", to),
		)
	}
	rateAPI, err := s.FetchRateFromAPI(ctx, from, to)
	if err != nil {
		return 0, fmt.Errorf("failed to get rate from API: %w", err)
	}
	go func() {
		caheCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err := s.redis.SetExchangeRate(caheCtx, from, to, rateAPI)
		if err != nil {
			s.logger.Warn("Failed to cache rate (non-critical)",
				zap.String("from", from),
				zap.String("to", to),
				zap.Error(err),
			)
		}
		s.logger.Debug("Rate cached in Redis",
			zap.String("from", from),
			zap.String("to", to),
			zap.Float64("rate", rateAPI),
			zap.Duration("ttl", s.config.Redis.TTL),
		)

	}()

	return rateAPI, nil
}

// TODO: Получить курс (сначала из кеша, потом из API)

func (s *CurrencyService) FetchRateFromAPI(ctx context.Context, from, to string) (float64, error) {
	s.logger.Error("🔍 DEBUG - Проверка конфигурации:",
		zap.String("CurrencyAPIURL", s.config.API.CurrencyAPIURL),
		zap.String("CurrencyKeyAPI", "["+s.config.API.CurrencyKeyAPI+"]"),
		zap.Int("KeyLength", len(s.config.API.CurrencyKeyAPI)),
	)

	// Формируем URL
	apiURL := fmt.Sprintf("%s/v6/%s/latest/%s",
		s.config.API.CurrencyAPIURL,
		s.config.API.CurrencyKeyAPI,
		from,
	)

	// 🔥 ДОБАВЬ ЭТУ СТРОКУ
	s.logger.Error("🔍 DEBUG - Сформированный URL:", zap.String("full_url", apiURL))

	// Маскируем ключ для логов
	maskedURL := strings.Replace(apiURL, s.config.API.CurrencyKeyAPI, "***", 1)

	s.logger.Debug("Fetching rate from ExchangeRate-API",
		zap.String("from", from),
		zap.String("to", to),
		zap.String("url", maskedURL),
	)

	// === ВСЕ ЭТО ОСТАЕТСЯ БЕЗ ИЗМЕНЕНИЙ ===
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		s.logger.Error("Failed to create HTTP request",
			zap.String("from", from),
			zap.String("to", to),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			s.logger.Error("API request timeout",
				zap.String("from", from),
				zap.String("to", to),
				zap.Duration("timeout", s.config.API.Timeout),
			)
			return 0, fmt.Errorf("API request timeout after %v", s.config.API.Timeout)
		}
		if ctx.Err() == context.Canceled {
			s.logger.Warn("API request canceled by client",
				zap.String("from", from),
				zap.String("to", to),
			)
			return 0, fmt.Errorf("API request canceled")
		}

		s.logger.Error("API request failed",
			zap.String("from", from),
			zap.String("to", to),
			zap.Error(err),
		)
		return 0, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error("API returned error status",
			zap.String("from", from),
			zap.String("to", to),
			zap.Int("status_code", resp.StatusCode),
			zap.String("status", resp.Status),
			zap.String("response", string(body)),
		)
		return 0, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("Failed to read API response",
			zap.String("from", from),
			zap.String("to", to),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	// === ИЗМЕНЕНО: Новая структура для ExchangeRate-API ===
	var apiResponse struct {
		Result          string             `json:"result"`
		BaseCode        string             `json:"base_code"`
		ConversionRates map[string]float64 `json:"conversion_rates"`
	}

	if err := json.Unmarshal(data, &apiResponse); err != nil {
		s.logger.Error("Invalid JSON from ExchangeRate-API",
			zap.String("from", from),
			zap.String("to", to),
			zap.String("response", string(data)),
			zap.Error(err),
		)
		return 0, fmt.Errorf("invalid JSON response: %w", err)
	}

	// === ИЗМЕНЕНО: Проверяем поле "result" ===
	if apiResponse.Result != "success" {
		s.logger.Error("ExchangeRate-API returned error",
			zap.String("from", from),
			zap.String("to", to),
			zap.String("result", apiResponse.Result),
			zap.String("response", string(data)),
		)
		return 0, fmt.Errorf("ExchangeRate-API error: %s", apiResponse.Result)
	}

	// === ИЗМЕНЕНО: Берем из conversion_rates вместо data ===
	rate, exists := apiResponse.ConversionRates[to]
	if !exists {
		var availableCurrencies []string
		for currency := range apiResponse.ConversionRates {
			availableCurrencies = append(availableCurrencies, currency)
		}

		s.logger.Error("Currency not found in ExchangeRate-API response",
			zap.String("from", from),
			zap.String("to", to),
			zap.Strings("available_currencies", availableCurrencies),
		)
		return 0, fmt.Errorf("currency %s not found in API response", to)
	}

	s.logger.Debug("Rate successfully fetched from ExchangeRate-API",
		zap.String("from", from),
		zap.String("to", to),
		zap.Float64("rate", rate),
	)

	return rate, nil
}
func (s *CurrencyService) Convert(ctx context.Context, from, to string, amount float64) (result, rate float64, err error) {
	// Валидация суммы
	if amount <= 0 {
		return 0, 0, fmt.Errorf("amount must be positive, got: %.2f", amount)
	}

	// Получаем курс
	rate, err = s.GetExchangeRate(ctx, from, to)
	if err != nil {
		return 0, 0, err
	}

	// Вычисляем результат
	result = amount * rate

	s.logger.Info("Currency conversion completed",
		zap.String("from", from),
		zap.String("to", to),
		zap.Float64("amount", amount),
		zap.Float64("rate", rate),
		zap.Float64("result", result),
	)

	return result, rate, nil
}

var _ CurrencyServiceInterface = (*CurrencyService)(nil)
