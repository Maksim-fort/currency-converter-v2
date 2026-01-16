package app

import (
	"context"
	"currency-converter-v2/internal/config"
	"currency-converter-v2/internal/handler"
	"currency-converter-v2/internal/middleware"
	"currency-converter-v2/internal/service"
	"currency-converter-v2/pkg/cache"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Application struct {
	config *config.Config
	router *gin.Engine
	logger *zap.Logger
	redis  *cache.RedisClient
	server *http.Server
}

func New(cfg *config.Config) *Application {
	logger := initLogger(&cfg.Logging)
	reddisClient, err := cache.NewRedisClient(cfg.Redis, logger)
	if err != nil {
		logger.Error("Failed to create Redis client", zap.Error(err))
		// Продолжаем без Redis
	}
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
		logger.Info("Running in RELEASE mode")
	} else {
		gin.SetMode(gin.DebugMode)
		logger.Info("Running in DEBUG mode")
	}
	router := gin.New()
	currencyService := service.NewCurrencyService(cfg, reddisClient, logger)
	currencyHandler := handler.NewCurrencyHandler(currencyService)
	app := &Application{
		config: cfg,
		router: router,
		logger: logger,
		redis:  reddisClient,
	}
	app.setupMiddleware()
	app.setupRouter(currencyHandler)
	logger.Info("Application initialized",
		zap.String("host", cfg.Server.Host),
		zap.String("port", cfg.Server.Port),
		zap.Bool("redis_connected", reddisClient != nil),
	)
	return app

}
func initLogger(cfg *config.LoggingConfig) *zap.Logger {
	var logger *zap.Logger
	var err error
	if cfg.Format == "json" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	switch cfg.Level {
	case "debug":
		logger = logger.WithOptions(zap.IncreaseLevel(zap.DebugLevel))
	case "info":
		logger = logger.WithOptions(zap.IncreaseLevel(zap.InfoLevel))
	case "warn":
		logger = logger.WithOptions(zap.IncreaseLevel(zap.WarnLevel))
	case "error":
		logger = logger.WithOptions(zap.IncreaseLevel(zap.ErrorLevel))
	}
	return logger
}
func (a *Application) setupMiddleware() {
	a.router.Use(middleware.RecoveryMiddleware(a.logger))
	a.router.Use(middleware.LoggingMiddleware(a.logger))
	a.router.Use(middleware.CORSMiddleware())
	a.logger.Debug("Middleware configured")
}
func (a *Application) setupRouter(currencyHandler *handler.CurrencyHandler) {
	a.router.GET("/health", handler.HealthCheck)
	apiV1 := a.router.Group("/api/v1")
	apiV1.GET("/convert", currencyHandler.Convert)
	a.router.Static("/ui", "/app/frontend")
	a.router.StaticFile("/", "/app/frontend/index.html")
	a.logger.Debug("Routes configured",
		zap.String("health", "GET /health"),
		zap.String("convert", "GET /api/v1/convert"),
		zap.String("frontend", "GET /ui"),
	)
}
func (a *Application) Run() error {
	a.server = &http.Server{
		Addr:    a.config.Server.Addr(),
		Handler: a.router,
	}

	// Канал для ошибки сервера
	serverErr := make(chan error, 1)

	// Запускаем сервер в горутине
	go func() {
		a.logger.Info("🚀 Server starting",
			zap.String("address", a.server.Addr),
			zap.String("mode", a.config.Server.Mode),
		)

		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
		close(serverErr)
	}()

	// Настраиваем graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Ждем либо ошибку сервера, либо сигнал shutdown
	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)

	case sig := <-quit:
		a.logger.Info("🛑 Received shutdown signal", zap.String("signal", sig.String()))

		// Graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := a.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}

		a.logger.Info("✅ Server stopped gracefully")
		return nil
	}
}

// Shutdown корректно останавливает сервер
func (a *Application) Shutdown() {
	a.logger.Info("Starting graceful shutdown...")

	// Даем серверу 5 секунд на завершение текущих запросов
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Останавливаем прием новых соединений
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("Failed to shutdown HTTP server", zap.Error(err))
	}

	// Закрываем соединение с Redis
	if a.redis != nil {
		a.redis.Close()
	}

	// Сбрасываем буферы логгера
	a.logger.Sync()

	a.logger.Info("Server stopped gracefully")
}
