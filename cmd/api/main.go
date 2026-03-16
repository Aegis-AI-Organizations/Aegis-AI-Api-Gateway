package main

import (
	"fmt"
	"log"

	"os"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/db"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/grpc"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/temporalclient"
)

func main() {
	fmt.Println("🚀 Aegis AI API Gateway is starting...")

	database, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()
	fmt.Println("✅ Connected to PostgreSQL database")

	tc, err := temporalclient.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to Temporal: %v", err)
	}
	defer tc.Close()
	fmt.Println("✅ Connected to Temporal orchestrator")

	brainAddr := os.Getenv("BRAIN_GRPC_ADDR")
	if brainAddr == "" {
		brainAddr = "localhost:50051"
	}
	gc, err := grpc.NewClient(brainAddr)
	if err != nil {
		log.Fatalf("Failed to connect to Brain gRPC: %v", err)
	}
	defer func() {
		if err := gc.Close(); err != nil {
			log.Printf("Failed to close Brain gRPC client: %v", err)
		}
	}()
	fmt.Printf("✅ Connected to Brain gRPC at %s\n", brainAddr)

	api.Start(database, tc, gc)
}
