package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/db"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/types"
	"github.com/gin-gonic/gin"
)

const agentDeploymentTokenPrefix = "ag_"
const agentDeploymentTokenBodyMinLength = 43

// AgentAuthMiddleware validates Agent credentials (deployment token or secret) and caches the result.
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Agent credentials required in Authorization header"})
			c.Abort()
			return
		}

		isDeploymentToken := isAgentDeploymentToken(token)
		isRegisterRoute := strings.HasSuffix(c.Request.URL.Path, "/register")

		// 1. Enforce route separation
		if isDeploymentToken && !isRegisterRoute {
			c.JSON(http.StatusForbidden, gin.H{"error": "Deployment token is only allowed for agent registration"})
			c.Abort()
			return
		}
		if !isDeploymentToken && isRegisterRoute {
			c.JSON(http.StatusForbidden, gin.H{"error": "Agent secrets cannot be used for registration"})
			c.Abort()
			return
		}

		// 2. Cross-validation for operational routes
		agentID := ""
		if !isRegisterRoute {
			agentID = c.Param("id")
			if agentID == "" {
				// Fallback to body if not in URL? (most routes have it in URL)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Agent ID missing in request path"})
				c.Abort()
				return
			}
		}

		// Check Redis cache
		var cacheKey string
		if isRegisterRoute {
			cacheKey = fmt.Sprintf("agent_token:%s", token)
		} else {
			cacheKey = fmt.Sprintf("agent_secret:%s:%s", agentID, token)
		}

		if redisClient != nil && redisClient.Client != nil {
			tenantID, err := redisClient.Client.Get(c.Request.Context(), cacheKey).Result()
			if err == nil && tenantID != "" {
				if agentID != "" {
					c.Set(string(types.AgentIDKey), agentID)
				}
				c.Set(string(types.AgentTenantIDKey), tenantID)
				c.Set(string(types.AgentTokenKey), token)
				c.Next()
				return
			}
		}

		var tenantID string
		if isRegisterRoute {
			// Deployment token validation
			resp, err := grpcClient.VerifyToken(c.Request.Context(), token)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal auth service error"})
				c.Abort()
				return
			}
			if !resp.Valid {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid deployment token"})
				c.Abort()
				return
			}
			tenantID = resp.TenantId
		} else {
			// Agent secret validation (cross-validated with agentID)
			resp, err := grpcClient.VerifyAgentSecret(c.Request.Context(), agentID, token)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal auth service error"})
				c.Abort()
				return
			}
			if !resp.Valid {
				c.JSON(http.StatusForbidden, gin.H{"error": "Invalid agent secret or unauthorized access to agent"})
				c.Abort()
				return
			}
			tenantID = resp.TenantId
		}

		// Cache result
		if redisClient != nil && redisClient.Client != nil {
			redisClient.Client.Set(c.Request.Context(), cacheKey, tenantID, 30*time.Minute)
		}

		// Set context
		if agentID != "" {
			c.Set(string(types.AgentIDKey), agentID)
		}
		c.Set(string(types.AgentTenantIDKey), tenantID)
		c.Set(string(types.AgentTokenKey), token)
		c.Next()
	}
}

func isAgentDeploymentToken(token string) bool {
	body, ok := strings.CutPrefix(token, agentDeploymentTokenPrefix)
	if !ok || len(body) < agentDeploymentTokenBodyMinLength {
		return false
	}

	for _, char := range body {
		if char >= 'A' && char <= 'Z' {
			continue
		}
		if char >= 'a' && char <= 'z' {
			continue
		}
		if char >= '0' && char <= '9' {
			continue
		}
		if char == '_' || char == '-' {
			continue
		}
		return false
	}

	return true
}
