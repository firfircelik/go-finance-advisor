// Package config provides centralized configuration management for the application.
// It loads configuration from environment variables with validation and sensible defaults.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	API      APIConfig
	Server   ServerConfig
	Redis    RedisConfig
	Logging  LoggingConfig
}

// AppConfig contains general application settings
type AppConfig struct {
	Name        string
	Environment string
	Version     string
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	Driver          string
	Path            string // For SQLite
	Host            string // For PostgreSQL/MySQL
	Port            string
	Name            string
	User            string
	Password        string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

// JWTConfig contains JWT authentication settings
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

// APIConfig contains external API credentials
type APIConfig struct {
	AlphaVantageKey string
	CoinGeckoKey    string
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Port            string
	AllowedOrigins  []string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// RedisConfig contains Redis connection settings
type RedisConfig struct {
	URL      string
	Password string
	DB       int
	Enabled  bool
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level  string
	Format string // json or text
}

// Load reads configuration from environment variables with validation
func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "go-finance-advisor"),
			Environment: getEnv("APP_ENV", "development"),
			Version:     getEnv("APP_VERSION", "1.0.0"),
		},
		Database: DatabaseConfig{
			Driver:          getEnv("DB_DRIVER", "sqlite"),
			Path:            getEnv("DB_PATH", "finance.db"),
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			Name:            getEnv("DB_NAME", "finance_db"),
			User:            getEnv("DB_USER", "finance_user"),
			Password:        getEnv("DB_PASSWORD", ""),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 100),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 1*time.Hour),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", ""),
			Expiration: getEnvAsDuration("JWT_EXPIRATION", 24*time.Hour),
		},
		API: APIConfig{
			AlphaVantageKey: getEnv("ALPHA_VANTAGE_API_KEY", "demo"),
			CoinGeckoKey:    getEnv("COINGECKO_API_KEY", ""),
		},
		Server: ServerConfig{
			Port:            getEnv("APP_PORT", "8080"),
			AllowedOrigins:  getEnvAsSlice("ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
			ReadTimeout:     getEnvAsDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getEnvAsDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:     getEnvAsDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", "redis://localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			Enabled:  getEnvAsBool("REDIS_ENABLED", true),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	// Validate critical configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Validate checks that all required configuration is present
func (c *Config) Validate() error {
	// JWT Secret is required in production
	if c.App.Environment == "production" && c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required in production environment")
	}

	// Warn about default JWT secret in development
	if c.JWT.Secret == "" {
		fmt.Println("⚠️  WARNING: JWT_SECRET not set, using insecure default for development only")
		c.JWT.Secret = "development-secret-key-change-in-production"
	}

	// Validate database configuration
	if c.Database.Driver == "" {
		return fmt.Errorf("DB_DRIVER is required")
	}

	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("invalid LOG_LEVEL: %s (must be debug, info, warn, or error)", c.Logging.Level)
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development" || c.App.Environment == "dev"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production" || c.App.Environment == "prod"
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
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

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
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

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	// Simple comma-separated parsing
	result := []string{}
	current := ""
	for _, char := range valueStr {
		if char == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else if char != ' ' {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}

	if len(result) == 0 {
		return defaultValue
	}
	return result
}
