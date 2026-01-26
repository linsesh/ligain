package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupVersionTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestVersionCheck_NoEnvVar(t *testing.T) {
	// Ensure MIN_APP_VERSION is not set
	os.Unsetenv("MIN_APP_VERSION")

	router := setupVersionTestRouter()
	router.Use(VersionCheck())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["message"])
}

func TestVersionCheck_NoHeader(t *testing.T) {
	os.Setenv("MIN_APP_VERSION", "1.4.0")
	defer os.Unsetenv("MIN_APP_VERSION")

	router := setupVersionTestRouter()
	router.Use(VersionCheck())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	// No X-App-Version header
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUpgradeRequired, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "App update required", response["error"])
	assert.Equal(t, "VERSION_OUTDATED", response["code"])
	assert.Equal(t, "1.4.0", response["min_version"])
	assert.Equal(t, AppStoreURL, response["store_url"])
}

func TestVersionCheck_VersionBelowMinimum(t *testing.T) {
	os.Setenv("MIN_APP_VERSION", "1.4.0")
	defer os.Unsetenv("MIN_APP_VERSION")

	router := setupVersionTestRouter()
	router.Use(VersionCheck())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-App-Version", "1.3.0")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUpgradeRequired, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "App update required", response["error"])
	assert.Equal(t, "VERSION_OUTDATED", response["code"])
}

func TestVersionCheck_VersionEqualMinimum(t *testing.T) {
	os.Setenv("MIN_APP_VERSION", "1.4.0")
	defer os.Unsetenv("MIN_APP_VERSION")

	router := setupVersionTestRouter()
	router.Use(VersionCheck())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-App-Version", "1.4.0")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["message"])
}

func TestVersionCheck_VersionAboveMinimum(t *testing.T) {
	os.Setenv("MIN_APP_VERSION", "1.4.0")
	defer os.Unsetenv("MIN_APP_VERSION")

	router := setupVersionTestRouter()
	router.Use(VersionCheck())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-App-Version", "1.5.0")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["message"])
}

// TestCompareVersions tests the version comparison function
func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{
			name:     "Basic comparison: major v1 < v2",
			v1:       "1.0.0",
			v2:       "2.0.0",
			expected: -1,
		},
		{
			name:     "Basic comparison: major v1 > v2",
			v1:       "2.0.0",
			v2:       "1.0.0",
			expected: 1,
		},
		{
			name:     "Minor version: v1 < v2",
			v1:       "1.1.0",
			v2:       "1.2.0",
			expected: -1,
		},
		{
			name:     "Minor version: v1 > v2",
			v1:       "1.2.0",
			v2:       "1.1.0",
			expected: 1,
		},
		{
			name:     "Patch version: v1 < v2",
			v1:       "1.0.1",
			v2:       "1.0.2",
			expected: -1,
		},
		{
			name:     "Patch version: v1 > v2",
			v1:       "1.0.2",
			v2:       "1.0.1",
			expected: 1,
		},
		{
			name:     "Equal versions",
			v1:       "1.4.0",
			v2:       "1.4.0",
			expected: 0,
		},
		{
			name:     "Edge case: 1.9.0 < 1.10.0 (string comparison would fail)",
			v1:       "1.9.0",
			v2:       "1.10.0",
			expected: -1,
		},
		{
			name:     "Edge case: 1.10.0 > 1.9.0",
			v1:       "1.10.0",
			v2:       "1.9.0",
			expected: 1,
		},
		{
			name:     "Incomplete versions: 1.0 vs 1.0.0",
			v1:       "1.0",
			v2:       "1.0.0",
			expected: 0,
		},
		{
			name:     "Incomplete versions: 1 vs 1.0.0",
			v1:       "1",
			v2:       "1.0.0",
			expected: 0,
		},
		{
			name:     "Incomplete versions: 1.0 < 1.0.1",
			v1:       "1.0",
			v2:       "1.0.1",
			expected: -1,
		},
		{
			name:     "Complex comparison: 2.1.3 > 1.9.9",
			v1:       "2.1.3",
			v2:       "1.9.9",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareVersions(tt.v1, tt.v2)
			assert.Equal(t, tt.expected, result, "compareVersions(%s, %s) should be %d", tt.v1, tt.v2, tt.expected)
		})
	}
}

// TestParseVersion tests the version parsing function
func TestParseVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected [3]int
	}{
		{
			name:     "Full version",
			version:  "1.4.0",
			expected: [3]int{1, 4, 0},
		},
		{
			name:     "Two parts",
			version:  "1.4",
			expected: [3]int{1, 4, 0},
		},
		{
			name:     "One part",
			version:  "1",
			expected: [3]int{1, 0, 0},
		},
		{
			name:     "Empty string",
			version:  "",
			expected: [3]int{0, 0, 0},
		},
		{
			name:     "High numbers",
			version:  "10.20.30",
			expected: [3]int{10, 20, 30},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseVersion(tt.version)
			assert.Equal(t, tt.expected, result, "parseVersion(%s) should be %v", tt.version, tt.expected)
		})
	}
}

// TestVersionCheckResponseFields verifies the response body contains correct fields
func TestVersionCheckResponseFields(t *testing.T) {
	os.Setenv("MIN_APP_VERSION", "2.0.0")
	defer os.Unsetenv("MIN_APP_VERSION")

	router := setupVersionTestRouter()
	router.Use(VersionCheck())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-App-Version", "1.0.0")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUpgradeRequired, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify all required fields are present
	assert.Contains(t, response, "error")
	assert.Contains(t, response, "code")
	assert.Contains(t, response, "min_version")
	assert.Contains(t, response, "store_url")

	// Verify field values
	assert.Equal(t, "App update required", response["error"])
	assert.Equal(t, "VERSION_OUTDATED", response["code"])
	assert.Equal(t, "2.0.0", response["min_version"])
	assert.Equal(t, "https://apps.apple.com/fr/app/ligain/id6748531523", response["store_url"])
}
