package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"currency-converter-v2/internal/config"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type RedisClient struct {
	// TODO: Добавить поля структуры
	client *redis.Client
	config config.RedisConfig
	logger *zap.Logger
}

// NewRedisClient создает новый Redis клиент
func NewRedisClient(cfg config.RedisConfig, logger *zap.Logger) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("✅ Connected to Redis",
		zap.String("addr", cfg.Addr),
		zap.Int("db", cfg.DB),
	)

	return &RedisClient{
		client: client,
		config: cfg,
		logger: logger,
	}, nil
}

// GetExchangeRate получает курс валюты из Redis
func (r *RedisClient) GetExchangeRate(ctx context.Context, from, to string) (float64, error) {
	key := fmt.Sprintf("rate:%s:%s", from, to)
	valueStr, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, fmt.Errorf("exchange rate not found for %s to %s", from, to)
		}
		r.logger.Error("Redis GET error",
			zap.String("key", key),
			zap.Error(err))
		return 0, err
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		r.logger.Error("Redis GET error",
			zap.String("key", key),
			zap.Error(err))
		return 0, fmt.Errorf("invalid exchange rate format: %w", err)
	}
	// TODO: Реализовать получение курса
	return value, nil
}

// SetExchangeRate устанавливает курс валюты в Redis
func (r *RedisClient) SetExchangeRate(ctx context.Context, from, to string, rate float64) error {
	key := fmt.Sprintf("rate:%s:%s", from, to)
	err := r.client.Set(ctx, key, rate, r.config.TTL).Err()
	if err != nil {
		r.logger.Error("Redis SET error",
			zap.String("key", key),
			zap.Error(err))
		return fmt.Errorf("caching error: %w", err)
	}
	r.logger.Debug("Exchange rate saved to Redis",
		zap.String("key", key),
		zap.Float64("rate", rate),
		zap.Duration("ttl", r.config.TTL),
	)
	// TODO: Реализовать сохранение курса
	return nil
}

// DeleteExchangeRate удаляет курс из Redis
func (r *RedisClient) DeleteExchangeRate(ctx context.Context, from, to string) error {
	key := fmt.Sprintf("rate:%s:%s", from, to)

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		r.logger.Error("Failed to delete exchange rate",
			zap.String("key", key),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete rate: %w", err)
	}

	r.logger.Debug("Exchange rate deleted from cache",
		zap.String("key", key),
	)

	return nil
}

// HealthCheck проверяет доступность Redis
func (r *RedisClient) HealthCheck(ctx context.Context) error {
	err := r.client.Ping(ctx).Err()
	if err != nil {
		r.logger.Warn("Redis health check failed",
			zap.Error(err),
		)
		return fmt.Errorf("redis health check failed: %w", err)
	}

	return nil
}

// Close закрывает подключение к Redis
func (r *RedisClient) Close() {
	if r == nil || r.client == nil {
		return
	}
	r.client.Close() // Просто вызываем, не проверяем ошибку
}
