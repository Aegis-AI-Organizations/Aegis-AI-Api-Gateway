package api

import (
	"log"
	"os"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/handlers"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/middleware"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/grpc"
	"github.com/gin-gonic/gin"
)

func NewRouter(gc *grpc.Client) *gin.Engine {
	r := gin.Default()

	// Apply CORS middleware
	r.Use(middleware.CORSMiddleware())

	h := &handlers.API{
		GRPCClient: gc,
	}

	// Basic routes
	r.GET("/health", h.HealthHandler)
	r.GET("/", h.RootHandler)

	// Auth routes
	r.POST("/auth/login", h.LoginHandler)
	r.POST("/auth/refresh", h.RefreshHandler)
	r.POST("/auth/logout", h.LogoutHandler)

	// Scan routes
	r.POST("/scans", h.CreateScanHandler)
	r.GET("/scans", h.GetScansHandler)
	r.GET("/scans/:id", h.GetScanByIDHandler)
	r.GET("/scans/:id/vulnerabilities", h.GetVulnerabilitiesHandler)
	r.GET("/scans/:id/report", h.GetScanReportHandler)

	// Vulnerability routes
	r.GET("/vulnerabilities/:id/evidences", h.GetEvidencesHandler)

	// Streaming routes
	r.GET("/scans/stream", h.ScanStreamHandler)
	r.GET("/scans/:id/stream", h.ScanStreamHandler)

	return r
}

func Start(gc *grpc.Client) {
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
