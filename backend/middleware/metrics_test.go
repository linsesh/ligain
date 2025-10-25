package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMetricsMiddleware(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetFormatter(&log.JSONFormatter{})

	// Create test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(MetricsMiddleware())

	// Add test route
	router.GET("/test/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test/123", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse log output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	// Assert log fields
	assert.Equal(t, "/test/:id", logEntry["route"])
	assert.Equal(t, "GET", logEntry["method"])
	assert.Equal(t, float64(200), logEntry["status"])
	assert.Equal(t, "/test/123", logEntry["path"])
	assert.Equal(t, "http_request", logEntry["metric_type"])
	assert.Contains(t, logEntry, "duration_ms")
	assert.GreaterOrEqual(t, logEntry["duration_ms"].(float64), 0.0)
}

func TestMetricsMiddleware_UnmatchedRoute(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetFormatter(&log.JSONFormatter{})

	// Create test router with no routes
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(MetricsMiddleware())

	// Make request to unmatched route
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert 404
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Parse log output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	// For unmatched routes, path should be used as route
	assert.Equal(t, "/nonexistent", logEntry["route"])
	assert.Equal(t, float64(404), logEntry["status"])
}
