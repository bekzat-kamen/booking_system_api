package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Under Limit", func(t *testing.T) {
		db, mock := redismock.NewClientMock()
		config := RateLimitConfig{
			Requests: 5,
			Window:   time.Minute,
		}

		router := gin.New()
		router.Use(RateLimitMiddleware(db, config))
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		mock.ExpectGet("ratelimit:127.0.0.1").RedisNil()

		mock.ExpectSet("ratelimit:127.0.0.1", 1, config.Window).SetVal("OK")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "ok", w.Body.String())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Blocked - Limit Exceeded", func(t *testing.T) {
		db, mock := redismock.NewClientMock()
		config := RateLimitConfig{
			Requests: 2,
			Window:   time.Minute,
		}

		router := gin.New()
		router.Use(RateLimitMiddleware(db, config))
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		mock.ExpectGet("ratelimit:127.0.0.1").SetVal("2")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Contains(t, w.Body.String(), "too many requests")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Bypass - Whitelist", func(t *testing.T) {
		db, _ := redismock.NewClientMock()
		config := RateLimitConfig{
			Requests:  1,
			Window:    time.Minute,
			Whitelist: []string{"192.168.1.1"},
		}

		router := gin.New()
		router.Use(RateLimitMiddleware(db, config))
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "ok", w.Body.String())
	})

	t.Run("Fail-Open - Redis Error", func(t *testing.T) {
		db, mock := redismock.NewClientMock()
		config := RateLimitConfig{
			Requests: 5,
			Window:   time.Minute,
		}

		router := gin.New()
		router.Use(RateLimitMiddleware(db, config))
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		mock.ExpectGet("ratelimit:127.0.0.1").SetErr(errors.New("redis down"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPerRouteRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success with Prefix", func(t *testing.T) {
		db, mock := redismock.NewClientMock()
		config := RateLimitConfig{
			Requests: 5,
			Window:   time.Minute,
		}

		router := gin.New()
		router.Use(PerRouteRateLimitMiddleware(db, "login", config))
		router.GET("/login", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		mock.ExpectGet("ratelimit:login:127.0.0.1").RedisNil()
		mock.ExpectSet("ratelimit:login:127.0.0.1", 1, config.Window).SetVal("OK")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/login", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Blocked with Prefix", func(t *testing.T) {
		db, mock := redismock.NewClientMock()
		config := RateLimitConfig{
			Requests: 1,
			Window:   time.Minute,
		}

		router := gin.New()
		router.Use(PerRouteRateLimitMiddleware(db, "api", config))
		router.GET("/api", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		mock.ExpectGet("ratelimit:api:127.0.0.1").SetVal("1")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Contains(t, w.Body.String(), "rate limit exceeded for this endpoint")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
