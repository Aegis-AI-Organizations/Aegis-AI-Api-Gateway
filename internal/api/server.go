package api

import (
	"log"
	"os"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/handlers"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/middleware"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/db"
	"github.com/gin-gonic/gin"
)

func NewRouter(gc *agrpc.Client, rdb *db.RedisClient, mclient *db.MinioClient) *gin.Engine {
	r := gin.Default()

	// Apply CORS middleware
	r.Use(middleware.CORSMiddleware())

	// Apply Rate Limiting (Distributed via Redis)
	r.Use(middleware.RedisRateLimiter(rdb))

	h := &handlers.API{
		GRPCClient: gc,
	}

	mh := &handlers.MinioHandler{
		MinioClient: mclient,
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

		// Billing routes
		auth.GET("/billing/balance", middleware.RequirePermission(middleware.ScopeBillingRead), h.GetBalanceHandler)
		auth.GET("/billing/ledger", middleware.RequirePermission(middleware.ScopeBillingRead), h.GetLedgerHandler)
		auth.GET("/billing/stats", middleware.RequirePermission(middleware.ScopeBillingRead), h.GetUsageStatsHandler)

		// Admin billing adjustment and management
		admin.POST("/companies/:id/tokens/adjust", middleware.RequirePermission(middleware.ScopeAdminWrite), h.AdjustTokensHandler)
		admin.GET("/companies/:id/billing/balance", middleware.RequirePermission(middleware.ScopeAdminRead), h.GetBalanceHandler)
		admin.GET("/companies/:id/billing/ledger", middleware.RequirePermission(middleware.ScopeAdminRead), h.GetLedgerHandler)
		admin.GET("/companies/:id/billing/stats", middleware.RequirePermission(middleware.ScopeAdminRead), h.GetUsageStatsHandler)

		// Vulnerability routes
		auth.GET("/vulnerabilities/:id/evidences", middleware.RequirePermission(middleware.ScopeVulnerabilityRead), h.GetEvidencesHandler)

		// Streaming routes
		auth.GET("/scans/stream", middleware.RequirePermission(middleware.ScopeScanRead), h.ScanStreamHandler)
		auth.GET("/scans/:id/stream", middleware.RequirePermission(middleware.ScopeScanRead), h.ScanStreamHandler)

		// File storage (MinIO)
		auth.GET("/storage/upload-url", mh.GetUploadURLHandler)
	}

	// Agent-specific routes
	// 1. Onboarding (Uses deployment token)
	r.POST("/agents/register", h.RegisterAgentHandler)

	// 2. Persistent Agents Management (Uses Agent ID)
	agent := r.Group("/agents")
	// AgentAuthMiddleware ensures the request is coming from a valid agent
	agent.Use(middleware.AgentAuthMiddleware(gc, rdb))
	{
		// Agents can report their operational status (IDLE, UPLOADING, ERROR, etc.)
		agent.POST("/:id/status", h.UpdateAgentStatusHandler)

		// Agents request presigned URLs to upload infrastructure configs/logs
		agent.GET("/:id/upload-url", h.GetUploadLinkHandler)
	}

	return r
}

func Start(gc *agrpc.Client, rdb *db.RedisClient, mclient *db.MinioClient) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := NewRouter(gc, rdb, mclient)

	log.Printf("🌍 Aegis AI Web API Gateway (Gin) listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
