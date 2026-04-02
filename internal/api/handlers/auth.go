package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoginHandler handles user login and sets the refresh token cookie.
func (a *API) LoginHandler(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := a.GRPCClient.Login(ctx, req.Email, req.Password)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.Unauthenticated {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		} else if ok && st.Code() == codes.PermissionDenied {
			c.JSON(http.StatusForbidden, gin.H{"error": "Account forbidden"})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Authentication service unavailable"})
		}
		return
	}

	// Set Refresh Token in HttpOnly cookie
	// MaxAge: 7 days (604800 seconds)
	secure := gin.Mode() == gin.ReleaseMode
	if secure {
		c.SetSameSite(http.SameSiteNoneMode)
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
	}
	c.SetCookie("refresh_token", resp.RefreshToken, 604800, "/", "", secure, true)

	c.JSON(http.StatusOK, gin.H{
		"access_token": resp.AccessToken,
	})
}

// RefreshHandler handles token refresh using the cookie.
func (a *API) RefreshHandler(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token missing"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := a.GRPCClient.Refresh(ctx, refreshToken)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && (st.Code() == codes.Unauthenticated || st.Code() == codes.PermissionDenied) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Authentication service unavailable"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": resp.AccessToken,
	})
}

// LogoutHandler handles user logout and clears the cookie.
func (a *API) LogoutHandler(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err == nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		_, _ = a.GRPCClient.Logout(ctx, refreshToken)
	}

	// Clear the cookie
	secure := gin.Mode() == gin.ReleaseMode
	if secure {
		c.SetSameSite(http.SameSiteNoneMode)
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
	}
	c.SetCookie("refresh_token", "", -1, "/", "", secure, true)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// GetMeHandler retrieves the current user's profile information.
func (a *API) GetMeHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := a.GRPCClient.GetMe(ctx)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else if ok && st.Code() == codes.Unauthenticated {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Authentication service unavailable"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}
