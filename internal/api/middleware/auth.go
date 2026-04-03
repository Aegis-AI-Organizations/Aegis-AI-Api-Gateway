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
	TokenKey     ContextKey = "token"
)

// AuthMiddleware validates the JWT token and injects claims into context.
func AuthMiddleware() gin.HandlerFunc {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		fmt.Println("ERROR: JWT_SECRET not set in environment. All authenticated requests will fail.")
		return func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication service misconfigured"})
			c.Abort()
		}
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenString := ""

		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}
		if tokenString == "" {
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication token is required (header or query param)"})
			c.Abort()
			return
		}

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

		userID, ok1 := claims["sub"].(string)
		companyID, ok2 := claims["company_id"].(string)
		role, ok3 := claims["role"].(string)

		if !ok1 || !ok2 || !ok3 || userID == "" || companyID == "" || role == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token missing required identity claims"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("company_id", companyID)
		c.Set("role", role)
		c.Set("token", tokenString)

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, UserIDKey, userID)
		ctx = context.WithValue(ctx, CompanyIDKey, companyID)
		ctx = context.WithValue(ctx, RoleKey, role)
		ctx = context.WithValue(ctx, TokenKey, tokenString)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
