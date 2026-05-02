package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/db"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/types"
	"github.com/gin-gonic/gin"
)

// AgentAuthMiddleware validates Agent deployment tokens against the Brain and caches the result.
func AgentAuthMiddleware(grpcClient *agrpc.Client, redisClient *db.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		token := ""

		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Agent deployment token required in Authorization header"})
			c.Abort()
			return
		}

		// Check Redis cache for valid token
		cacheKey := fmt.Sprintf("agent_token:%s", token)
		tenantID, err := redisClient.Client.Get(c.Request.Context(), cacheKey).Result()
		if err == nil && tenantID != "" {
			// Cache hit - valid token
			c.Set(string(types.AgentTenantIDKey), tenantID)
			c.Set(string(types.AgentTokenKey), token)
			c.Next()
			return
		}

		// Cache miss or expired - query Brain
		resp, err := grpcClient.VerifyToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify token with auth service"})
			c.Abort()
			return
		}

		if !resp.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or inactive deployment token"})
			c.Abort()
			return
		}

		// Cache successful validation for 5 minutes as per requirements
		redisClient.Client.Set(c.Request.Context(), cacheKey, resp.TenantId, 5*time.Minute)

		// Add claims to context for subsequent handlers
		c.Set(string(types.AgentTenantIDKey), resp.TenantId)
		c.Set(string(types.AgentTokenKey), token)

		ctx := context.WithValue(c.Request.Context(), types.AgentTenantIDKey, resp.TenantId)
		ctx = context.WithValue(ctx, types.AgentTokenKey, token)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
