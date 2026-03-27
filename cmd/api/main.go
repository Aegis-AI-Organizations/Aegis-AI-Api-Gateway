package main

import (
	"fmt"
	"log"

	"os"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/grpc"
)

func main() {
	fmt.Println("🚀 Aegis AI API Gateway is starting...")

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

	api.Start(gc)
}
