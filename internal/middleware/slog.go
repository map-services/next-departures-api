package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// SlogMiddleware returns a Gin middleware that uses log/slog for structured logging.
func SlogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		// Skip logging for healthz and metrics
		if path == "/healthz" || path == "/metrics" {
			return
		}

		end := time.Now()
		latency := end.Sub(start)

		logger.Info("request",
			"status", c.Writer.Status(),
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"latency", latency,
			"client_ip", c.ClientIP(),
		)
	}
}
