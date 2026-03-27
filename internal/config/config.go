package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort          string
	AppEnv           string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	JWTSecret        string
	JWTExpire        string
	JWTRefreshSecret string
	JWTRefreshExpire string
	RedisHost        string
	RedisPort        string
	RedisPassword    string
	RedisDB          int
}

func Load() *Config {

	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using system environment variables")
	}

	return &Config{
		AppPort:          getEnv("APP_PORT", "8080"),
		AppEnv:           getEnv("APP_ENV", "development"),
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBUser:           getEnv("DB_USER", "postgres"),
		DBPassword:       getEnv("DB_PASSWORD", "postgres"),
		DBName:           getEnv("DB_NAME", "ticket_booking"),
		RedisHost:        getEnv("REDIS_HOST", "localhost"),
		RedisPort:        getEnv("REDIS_PORT", "6379"),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		RedisDB:          0,
		JWTSecret:        getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpire:        getEnv("JWT_EXPIRE", "15m"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "change-me-refresh-secret"),
		JWTRefreshExpire: getEnv("JWT_REFRESH_EXP", "168h"),
	}
}

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultVal
}
