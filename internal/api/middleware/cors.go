package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

const DashboardOrigin = "https://app.aegis-ai.fr"

// CORSMiddleware adds Cross-Origin Resource Sharing headers to every response and
// handles OPTIONS preflight requests so the browser can call the API from
// the public dashboard origin.
func CORSMiddleware() gin.HandlerFunc {
	allowedOrigins := allowedCORSOrigins()

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if _, ok := allowedOrigins[origin]; ok {
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

func allowedCORSOrigins() map[string]struct{} {
	origins := map[string]struct{}{
		DashboardOrigin: {},
	}

	for _, origin := range strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			origins[origin] = struct{}{}
		}
	}

	return origins
}
