package middleware

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// AppStoreURL is the URL to the Ligain app on the App Store
	AppStoreURL = "https://apps.apple.com/fr/app/ligain/id6748531523"
)

// VersionCheck validates X-App-Version header against MIN_APP_VERSION env var
// - Returns HTTP 426 Upgrade Required if version is below minimum
// - Skips check if MIN_APP_VERSION is not set (graceful rollout)
// - Response includes store_url and min_version for frontend
func VersionCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		minVersion := os.Getenv("MIN_APP_VERSION")

		// Skip check if MIN_APP_VERSION is not set (graceful rollout)
		if minVersion == "" {
			c.Next()
			return
		}

		appVersion := c.GetHeader("X-App-Version")

		// If no version header, require update
		if appVersion == "" {
			respondUpgradeRequired(c, minVersion)
			return
		}

		// Compare versions
		if compareVersions(appVersion, minVersion) < 0 {
			respondUpgradeRequired(c, minVersion)
			return
		}

		c.Next()
	}
}

// respondUpgradeRequired sends HTTP 426 response with update instructions
func respondUpgradeRequired(c *gin.Context, minVersion string) {
	c.JSON(http.StatusUpgradeRequired, gin.H{
		"error":       "App update required",
		"code":        "VERSION_OUTDATED",
		"min_version": minVersion,
		"store_url":   AppStoreURL,
	})
	c.Abort()
}

// compareVersions compares two semantic version strings
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	parts1 := parseVersion(v1)
	parts2 := parseVersion(v2)

	for i := 0; i < 3; i++ {
		if parts1[i] < parts2[i] {
			return -1
		}
		if parts1[i] > parts2[i] {
			return 1
		}
	}
	return 0
}

// parseVersion parses a version string into [major, minor, patch]
// Handles incomplete versions like "1.0" by treating missing parts as 0
func parseVersion(v string) [3]int {
	parts := strings.Split(v, ".")
	result := [3]int{0, 0, 0}

	for i := 0; i < len(parts) && i < 3; i++ {
		num, err := strconv.Atoi(parts[i])
		if err == nil {
			result[i] = num
		}
	}

	return result
}
