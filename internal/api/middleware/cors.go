package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware adds Cross-Origin Resource Sharing headers to every response and
// handles OPTIONS preflight requests so the browser can call the API from
// a different origin (e.g. app.aegis.pre-alpha.local → api.aegis.pre-alpha.local).
func CORSMiddleware() gin.HandlerFunc {
	var allowedOrigins []string

	// Load from environment if available (Overrides defaults)
	if envOrigins := os.Getenv("ALLOWED_ORIGINS"); envOrigins != "" {
		for _, o := range strings.Split(envOrigins, ",") {
			trimmed := strings.TrimSpace(o)
			if trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	} else {
		// Default allowed origins for development if NO override provided
		allowedOrigins = []string{
			"http://localhost:3000",
			"http://app.aegis.pre-alpha.local",
			"https://app.aegis.pre-alpha.local",
		}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		isAllowed := false

		if origin != "" {
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					isAllowed = true
					break
				}
			}
		}

		if isAllowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, Cache-Control, X-Requested-With")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
