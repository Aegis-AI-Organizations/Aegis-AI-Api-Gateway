package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ContextKey defines the type for storing values in the standard context.
type ContextKey string

const (
	UserIDKey    ContextKey = "user_id"
	CompanyIDKey ContextKey = "company_id"
	RoleKey      ContextKey = "role"
)

// AuthMiddleware validates the JWT token and injects claims into context.
func AuthMiddleware() gin.HandlerFunc {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		fmt.Println("WARNING: JWT_SECRET not set in environment. Authentication will fail.")
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format. Expected 'Bearer <token>'"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims := jwt.MapClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}
		userID, _ := claims["sub"].(string)
		companyID, _ := claims["company_id"].(string)
		role, _ := claims["role"].(string)

		c.Set("user_id", userID)
		c.Set("company_id", companyID)
		c.Set("role", role)

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, UserIDKey, userID)
		ctx = context.WithValue(ctx, CompanyIDKey, companyID)
		ctx = context.WithValue(ctx, RoleKey, role)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
