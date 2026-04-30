package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnvReturnsDefaultWhenMissing(t *testing.T) {
	t.Setenv("MISSING_ENV_FOR_TEST", "")

	value := getEnv("MISSING_ENV_FOR_TEST_UNSET", "default-value")

	assert.Equal(t, "default-value", value)
}

func TestGetEnvReturnsEnvironmentValue(t *testing.T) {
	t.Setenv("APP_PORT", "9090")

	value := getEnv("APP_PORT", "8080")

	assert.Equal(t, "9090", value)
}

func TestLoadUsesEnvironmentValuesAndDefaults(t *testing.T) {
	t.Setenv("APP_PORT", "9090")
	t.Setenv("APP_ENV", "test")
	t.Setenv("DB_HOST", "db-host")
	t.Setenv("DB_PORT", "5544")
	t.Setenv("DB_USER", "db-user")
	t.Setenv("DB_PASSWORD", "db-pass")
	t.Setenv("DB_NAME", "db-name")
	t.Setenv("REDIS_HOST", "redis-host")
	t.Setenv("REDIS_PORT", "6380")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("JWT_EXPIRE", "30m")
	t.Setenv("JWT_REFRESH_SECRET", "refresh-secret")
	t.Setenv("JWT_REFRESH_EXP", "24h")

	cfg := Load()

	assert.Equal(t, "9090", cfg.AppPort)
	assert.Equal(t, "test", cfg.AppEnv)
	assert.Equal(t, "db-host", cfg.DBHost)
	assert.Equal(t, "5544", cfg.DBPort)
	assert.Equal(t, "db-user", cfg.DBUser)
	assert.Equal(t, "db-pass", cfg.DBPassword)
	assert.Equal(t, "db-name", cfg.DBName)
	assert.Equal(t, "redis-host", cfg.RedisHost)
	assert.Equal(t, "6380", cfg.RedisPort)
	assert.Equal(t, "", cfg.RedisPassword)
	assert.Equal(t, 0, cfg.RedisDB)
	assert.Equal(t, "secret", cfg.JWTSecret)
	assert.Equal(t, "30m", cfg.JWTExpire)
	assert.Equal(t, "refresh-secret", cfg.JWTRefreshSecret)
	assert.Equal(t, "24h", cfg.JWTRefreshExpire)
}
