package middleware

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware() gin.HandlerFunc {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		if raw != "" {
			path = path + "?" + raw
		}

		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if status >= 400 {
			logger.Error("HTTP Request",
				slog.Int("status", status),
				slog.String("method", method),
				slog.String("path", path),
				slog.String("ip", clientIP),
				slog.Duration("latency", latency),
				slog.String("error", errorMessage),
			)
		} else {
			logger.Info("HTTP Request",
				slog.Int("status", status),
				slog.String("method", method),
				slog.String("path", path),
				slog.String("ip", clientIP),
				slog.Duration("latency", latency),
			)
		}
	}
}
