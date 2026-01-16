package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	Redis     RedisConfig
	Database  DatabaseConfig
	API       APIConfig
	JWT       JWTConfig
	RateLimit RateLimitConfig
	Cache     CacheConfig
	Logging   LoggingConfig
}
type ServerConfig struct {
	Port         string
	Host         string
	Mode         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	TTL      time.Duration
}
type DatabaseConfig struct {
	DSN string // Connection string для PostgreSQL
}
type APIConfig struct {
	CurrencyKeyAPI string
	CurrencyAPIURL string
	Timeout        time.Duration
}
type JWTConfig struct {
	JWTSecret  string
	Expiration time.Duration
}
type RateLimitConfig struct {
	Free    RateLimitPlan
	Basic   RateLimitPlan
	Premium RateLimitPlan
}
type RateLimitPlan struct {
	Requests int
	Period   time.Duration
}
type CacheConfig struct {
	DefaultTTL      time.Duration
	CleanupInterval time.Duration
}
type LoggingConfig struct {
	Level  string // "debug", "info", "warn", "error"
	Format string // "json" или "text"
	Output string // "stdout" или файл
}

// Метод для получения адреса сервера
func (s *ServerConfig) Addr() string {
	return s.Host + ":" + s.Port
}
func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
func Load() *Config {
	// 🔥 ПРАВИЛЬНАЯ ЗАГРУЗКА .env файла
	fmt.Println("=== LOADING CONFIGURATION ===")

	// Вариант 1: Загрузить .env без указания пути (godotenv найдет сам)
	if err := godotenv.Load(); err != nil {
		fmt.Printf("⚠️ godotenv.Load() failed: %v\n", err)

		// Вариант 2: Попробуем с абсолютным путем
		cwd, _ := os.Getwd()
		envPath := filepath.Join(cwd, ".env")
		fmt.Printf("📁 Trying absolute path: %s\n", envPath)

		if err := godotenv.Load(envPath); err != nil {
			fmt.Printf("❌ ERROR: Cannot load .env file from %s: %v\n", envPath, err)
			fmt.Println("📝 Using default values instead...")
		} else {
			fmt.Printf("✅ .env loaded from: %s\n", envPath)
		}
	} else {
		fmt.Println("✅ .env loaded successfully")
	}

	// 🔥 ПРОВЕРКА: что переменные загрузились
	fmt.Println("=== ENVIRONMENT VARIABLES ===")
	fmt.Printf("CURRENCY_KEY_API: '%s' (length: %d)\n",
		os.Getenv("CURRENCY_KEY_API"),
		len(os.Getenv("CURRENCY_KEY_API")))
	fmt.Printf("CURRENCY_API_URL: '%s'\n", os.Getenv("CURRENCY_API_URL"))
	fmt.Printf("PORT: '%s'\n", os.Getenv("PORT"))
	fmt.Println("=============================")

	// Используем значения из .env или дефолтные
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			Host:         getEnv("HOST", "0.0.0.0"),
			Mode:         getEnv("GIN_MODE", "debug"),
			ReadTimeout:  getEnvAsDuration("READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getEnvAsDuration("WRITE_TIMEOUT", 10*time.Second),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			TTL:      getEnvAsDuration("REDIS_TTL", 30*time.Minute),
		},
		Database: DatabaseConfig{
			DSN: getEnv("DATABASE_URL", ""),
		},
		API: APIConfig{
			// 🔥 ВАЖНО: проверь правильное имя переменной
			CurrencyKeyAPI: getEnv("CURRENCY_KEY_API", "4cd60470d61ac235ae2e1f77"),
			CurrencyAPIURL: getEnv("CURRENCY_API_URL", "https://v6.exchangerate-api.com"),
			Timeout:        getEnvAsDuration("API_TIMEOUT", 10*time.Second),
		},
		JWT: JWTConfig{
			JWTSecret:  getEnv("JWT_SECRET", "your-super-secret-key-change-this-in-production"),
			Expiration: getEnvAsDuration("JWT_EXPIRATION", 24*time.Hour),
		},
		RateLimit: RateLimitConfig{
			Free: RateLimitPlan{
				Requests: getEnvAsInt("RATE_LIMIT_FREE", 100),
				Period:   getEnvAsDuration("RATE_LIMIT_FREE_PERIOD", 24*time.Hour),
			},
			Basic: RateLimitPlan{
				Requests: getEnvAsInt("RATE_LIMIT_BASIC", 1000),
				Period:   getEnvAsDuration("RATE_LIMIT_BASIC_PERIOD", 24*time.Hour),
			},
			Premium: RateLimitPlan{
				Requests: getEnvAsInt("RATE_LIMIT_PREMIUM", 10000),
				Period:   getEnvAsDuration("RATE_LIMIT_PREMIUM_PERIOD", 24*time.Hour),
			},
		},
		Cache: CacheConfig{
			DefaultTTL:      getEnvAsDuration("CACHE_TTL", 30*time.Minute),
			CleanupInterval: getEnvAsDuration("CACHE_CLEANUP", 1*time.Hour),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
		},
	}
}
