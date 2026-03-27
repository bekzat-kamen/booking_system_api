package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

// RateLimitConfig конфигурация лимитов
type RateLimitConfig struct {
	Requests  int           // Максимум запросов
	Window    time.Duration // Окно времени
	Whitelist []string      // IP без лимитов (опционально)
}

// DefaultRateLimitConfig конфигурация по умолчанию
var DefaultRateLimitConfig = RateLimitConfig{
	Requests:  100,         // 100 запросов
	Window:    time.Minute, // в минуту
	Whitelist: []string{},
}

// RateLimitMiddleware создаёт middleware для ограничения запросов
func RateLimitMiddleware(client *redis.Client, config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем IP клиента
		ip := c.ClientIP()

		// Проверяем whitelist
		for _, whitelistIP := range config.Whitelist {
			if ip == whitelistIP {
				c.Next()
				return
			}
		}

		// Ключ для Redis
		key := "ratelimit:" + ip

		// Получаем текущее счётчик
		count, err := client.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			// Ошибка Redis — пропускаем запрос (fail-open)
			c.Next()
			return
		}

		// Если ключа нет — создаём
		if err == redis.Nil {
			client.Set(ctx, key, 1, config.Window)
			c.Next()
			return
		}

		// Проверяем лимит
		if count >= config.Requests {
			// Возвращаем заголовки о лимитах
			c.Header("X-RateLimit-Limit", string(rune(config.Requests)))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", time.Now().Add(config.Window).Format(time.RFC3339))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "too many requests",
				"message":     "rate limit exceeded, please try again later",
				"retry_after": config.Window.Seconds(),
			})
			c.Abort()
			return
		}

		// Увеличиваем счётчик
		client.Incr(ctx, key)

		// Устанавливаем время жизни, если ключ новый
		ttl, _ := client.TTL(ctx, key).Result()
		if ttl <= 0 {
			client.Expire(ctx, key, config.Window)
		}

		// Возвращаем заголовки о лимитах
		c.Header("X-RateLimit-Limit", string(rune(config.Requests)))
		c.Header("X-RateLimit-Remaining", string(rune(config.Requests-count-1)))
		c.Header("X-RateLimit-Reset", time.Now().Add(config.Window).Format(time.RFC3339))

		c.Next()
	}
}

// PerRouteRateLimitMiddleware для разных лимитов на разные роуты
func PerRouteRateLimitMiddleware(client *redis.Client, prefix string, config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := "ratelimit:" + prefix + ":" + ip

		count, err := client.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			c.Next()
			return
		}

		if err == redis.Nil {
			client.Set(ctx, key, 1, config.Window)
			c.Next()
			return
		}

		if count >= config.Requests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "too many requests",
				"message":     "rate limit exceeded for this endpoint",
				"retry_after": config.Window.Seconds(),
			})
			c.Abort()
			return
		}

		client.Incr(ctx, key)
		ttl, _ := client.TTL(ctx, key).Result()
		if ttl <= 0 {
			client.Expire(ctx, key, config.Window)
		}

		c.Next()
	}
}
