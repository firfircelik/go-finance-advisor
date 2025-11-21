package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Set test environment variables
	os.Setenv("APP_NAME", "test-app")
	os.Setenv("APP_ENV", "test")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("APP_PORT", "9090")

	defer func() {
		os.Unsetenv("APP_NAME")
		os.Unsetenv("APP_ENV")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("APP_PORT")
	}()

	cfg, err := Load()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "test-app", cfg.App.Name)
	assert.Equal(t, "test", cfg.App.Environment)
	assert.Equal(t, "test-secret", cfg.JWT.Secret)
	assert.Equal(t, "9090", cfg.Server.Port)
}

func TestValidate_ProductionRequiresJWTSecret(t *testing.T) {
	cfg := &Config{
		App: AppConfig{
			Environment: "production",
		},
		JWT: JWTConfig{
			Secret: "",
		},
		Database: DatabaseConfig{
			Driver: "sqlite",
		},
		Logging: LoggingConfig{
			Level: "info",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET is required")
}

func TestValidate_InvalidLogLevel(t *testing.T) {
	cfg := &Config{
		App: AppConfig{
			Environment: "development",
		},
		JWT: JWTConfig{
			Secret: "test",
		},
		Database: DatabaseConfig{
			Driver: "sqlite",
		},
		Logging: LoggingConfig{
			Level: "invalid",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid LOG_LEVEL")
}

func TestIsDevelopment(t *testing.T) {
	cfg := &Config{
		App: AppConfig{Environment: "development"},
	}
	assert.True(t, cfg.IsDevelopment())

	cfg.App.Environment = "dev"
	assert.True(t, cfg.IsDevelopment())

	cfg.App.Environment = "production"
	assert.False(t, cfg.IsDevelopment())
}

func TestIsProduction(t *testing.T) {
	cfg := &Config{
		App: AppConfig{Environment: "production"},
	}
	assert.True(t, cfg.IsProduction())

	cfg.App.Environment = "prod"
	assert.True(t, cfg.IsProduction())

	cfg.App.Environment = "development"
	assert.False(t, cfg.IsProduction())
}

func TestGetEnvAsInt(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	result := getEnvAsInt("TEST_INT", 10)
	assert.Equal(t, 42, result)

	result = getEnvAsInt("NONEXISTENT", 10)
	assert.Equal(t, 10, result)
}

func TestGetEnvAsBool(t *testing.T) {
	os.Setenv("TEST_BOOL", "true")
	defer os.Unsetenv("TEST_BOOL")

	result := getEnvAsBool("TEST_BOOL", false)
	assert.True(t, result)

	result = getEnvAsBool("NONEXISTENT", false)
	assert.False(t, result)
}

func TestGetEnvAsDuration(t *testing.T) {
	os.Setenv("TEST_DURATION", "5m")
	defer os.Unsetenv("TEST_DURATION")

	result := getEnvAsDuration("TEST_DURATION", 1*time.Minute)
	assert.Equal(t, 5*time.Minute, result)

	result = getEnvAsDuration("NONEXISTENT", 1*time.Minute)
	assert.Equal(t, 1*time.Minute, result)
}

func TestGetEnvAsSlice(t *testing.T) {
	os.Setenv("TEST_SLICE", "a,b,c")
	defer os.Unsetenv("TEST_SLICE")

	result := getEnvAsSlice("TEST_SLICE", []string{"default"})
	assert.Equal(t, []string{"a", "b", "c"}, result)

	result = getEnvAsSlice("NONEXISTENT", []string{"default"})
	assert.Equal(t, []string{"default"}, result)
}
