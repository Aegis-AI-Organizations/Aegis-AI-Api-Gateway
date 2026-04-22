package api

import (
	"log"
	"os"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/handlers"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/middleware"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc"
	"github.com/gin-gonic/gin"
)

func NewRouter(gc *agrpc.Client) *gin.Engine {
	r := gin.Default()

	// Apply CORS middleware
	r.Use(middleware.CORSMiddleware())

	// Apply Rate Limiting (10 req/s, burst of 20)
	r.Use(middleware.RateLimiter(10, 20))

	h := &handlers.API{
		GRPCClient: gc,
	}

	// Basic public routes
	r.GET("/health", h.HealthHandler)
	r.GET("/", h.RootHandler)

	// Public Auth routes
	r.POST("/auth/login", h.LoginHandler)
	r.POST("/auth/refresh", h.RefreshHandler)

	// Protected routes group
	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.POST("/auth/logout", middleware.RequirePermission(middleware.ScopeAuthRead), h.LogoutHandler)
		auth.GET("/auth/me", middleware.RequirePermission(middleware.ScopeAuthRead), h.GetMeHandler)

		// User profile management
		auth.PUT("/users/me/profile", middleware.RequirePermission(middleware.ScopeAuthRead), h.UpdateProfileHandler)
		auth.DELETE("/users/me/profile/avatar", middleware.RequirePermission(middleware.ScopeAuthRead), h.RemoveAvatarHandler)
		auth.PUT("/users/me/email", middleware.RequirePermission(middleware.ScopeAuthRead), h.UpdateEmailHandler)
		auth.PUT("/users/me/password", middleware.RequirePermission(middleware.ScopeAuthRead), h.UpdatePasswordHandler)

		// Company management
		auth.GET("/companies", middleware.RequirePermission(middleware.ScopeCompanyRead), h.ListCompaniesHandler)
		auth.POST("/companies", middleware.RequirePermission(middleware.ScopeCompanyWrite), h.CreateCompanyHandler)
		auth.POST("/companies/onboard", middleware.RequirePermission(middleware.ScopeCompanyWrite), h.OnboardCompanyHandler)

		// Administrative search (Best for Production)
		admin := auth.Group("/admin")
		{
			admin.GET("/companies", middleware.RequirePermission(middleware.ScopeCompanyRead), h.SearchCompaniesHandler)
			admin.GET("/users", middleware.RequirePermission(middleware.ScopeAuthRead), h.SearchUsersHandler)
			admin.POST("/users", middleware.RequirePermission(middleware.ScopeUserWrite), h.CreateUserHandler)
			admin.GET("/teams/stream", middleware.RequirePermission(middleware.ScopeCompanyRead), h.TeamStreamHandler)
			admin.GET("/audit-logs", middleware.RequirePermission(middleware.ScopeCompanyRead), h.ListAuditLogsHandler)
		}

		// Scan routes
		auth.POST("/scans", middleware.RequirePermission(middleware.ScopeScanWrite), h.CreateScanHandler)
		auth.GET("/scans", middleware.RequirePermission(middleware.ScopeScanRead), h.GetScansHandler)
		auth.GET("/scans/:id", middleware.RequirePermission(middleware.ScopeScanRead), h.GetScanByIDHandler)
		auth.GET("/scans/:id/vulnerabilities", middleware.RequirePermission(middleware.ScopeVulnerabilityRead), h.GetVulnerabilitiesHandler)
		auth.GET("/scans/:id/report", middleware.RequirePermission(middleware.ScopeReportRead), h.GetScanReportHandler)

		// Vulnerability routes
		auth.GET("/vulnerabilities/:id/evidences", middleware.RequirePermission(middleware.ScopeVulnerabilityRead), h.GetEvidencesHandler)

		// Streaming routes
		auth.GET("/scans/stream", middleware.RequirePermission(middleware.ScopeScanRead), h.ScanStreamHandler)
		auth.GET("/scans/:id/stream", middleware.RequirePermission(middleware.ScopeScanRead), h.ScanStreamHandler)
	}

	return r
}

func Start(gc *agrpc.Client) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := NewRouter(gc)

	log.Printf("🌍 Aegis AI Web API Gateway (Gin) listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
