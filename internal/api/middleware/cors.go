package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware adds Cross-Origin Resource Sharing headers to every response and
// handles OPTIONS preflight requests so the browser can call the API from
// a different origin (e.g. app.aegis.pre-alpha.local → api.aegis.pre-alpha.local).
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", c.GetHeader("Origin"))
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, Cache-Control, X-Requested-With")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
