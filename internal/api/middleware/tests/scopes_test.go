package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/middleware"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHasScope(t *testing.T) {
	tests := []struct {
		name          string
		role          types.UserRole
		requiredScope string
		want          bool
	}{
		{"Viewer can read scans", types.RoleViewer, middleware.ScopeScanRead, true},
		{"Viewer cannot write scans", types.RoleViewer, middleware.ScopeScanWrite, false},
		{"Operator can write scans", types.RoleOperator, middleware.ScopeScanWrite, true},
		{"Operator can execute scans", types.RoleOperator, middleware.ScopeScanExecute, true},
		{"Owner can read users", types.RoleOwner, middleware.ScopeUserRead, true},
		{"Owner can write users", types.RoleOwner, middleware.ScopeUserWrite, true},
		{"SuperAdmin has all access", types.RoleSuperAdmin, "any:scope", true},
		{"Viewer has auth read", types.RoleViewer, middleware.ScopeAuthRead, true},
		{"Unknown role has no access", "unknown", middleware.ScopeScanRead, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, middleware.HasScope(tt.role, tt.requiredScope))
		})
	}
}

func TestRequirePermissionMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           string
		requiredScope  string
		expectedStatus int
	}{
		{
			name:           "Valid scope for Viewer",
			role:           string(types.RoleViewer),
			requiredScope:  middleware.ScopeScanRead,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid scope for Viewer",
			role:           string(types.RoleViewer),
			requiredScope:  middleware.ScopeScanWrite,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Valid scope for Operator",
			role:           string(types.RoleOperator),
			requiredScope:  middleware.ScopeScanWrite,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing role in context",
			role:           "",
			requiredScope:  middleware.ScopeScanRead,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "SuperAdmin bypass",
			role:           string(types.RoleSuperAdmin),
			requiredScope:  "anything",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)

			r.Use(func(c *gin.Context) {
				if tt.role != "" {
					c.Set("role", tt.role)
				}
				c.Next()
			})
			r.Use(middleware.RequirePermission(tt.requiredScope))
			r.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			c.Request, _ = http.NewRequest(http.MethodGet, "/test", nil)
			r.ServeHTTP(w, c.Request)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
