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
		auth.POST("/auth/logout", h.LogoutHandler)
		auth.GET("/auth/me", h.GetMeHandler)

		// Scan routes
		auth.POST("/scans", h.CreateScanHandler)
		auth.GET("/scans", h.GetScansHandler)
		auth.GET("/scans/:id", h.GetScanByIDHandler)
		auth.GET("/scans/:id/vulnerabilities", h.GetVulnerabilitiesHandler)
		auth.GET("/scans/:id/report", h.GetScanReportHandler)

		// Vulnerability routes
		auth.GET("/vulnerabilities/:id/evidences", h.GetEvidencesHandler)

		// Streaming routes
		auth.GET("/scans/stream", h.ScanStreamHandler)
		auth.GET("/scans/:id/stream", h.ScanStreamHandler)
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
