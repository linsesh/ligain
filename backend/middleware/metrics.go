package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// MetricsMiddleware logs request metrics in a structured format for Cloud Monitoring
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)
		durationMs := float64(duration.Milliseconds())

		// Get route pattern (e.g., "/api/games/:id" instead of "/api/games/123")
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path // Fallback for unmatched routes
		}

		// Log structured metrics
		log.WithFields(log.Fields{
			"route":       route,
			"method":      c.Request.Method,
			"status":      c.Writer.Status(),
			"duration_ms": durationMs,
			"path":        c.Request.URL.Path,
			"client_ip":   c.ClientIP(),
			"user_agent":  c.Request.UserAgent(),
			"metric_type": "http_request",
		}).Info("request completed")
	}
}
