package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/grpc"
)

func main() {
	fmt.Println("🧪 Testing gRPC connection to Brain...")

	client, err := grpc.NewClient("localhost:50051")
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Failed to close client: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	message, err := client.Ping(ctx)
	if err != nil {
		log.Printf("❌ Ping failed: %v", err)
		fmt.Println("Note: Make sure the Brain gRPC server is running on localhost:50051")
		return
	}

	fmt.Printf("✅ Ping successful! Response: %s\n", message)
}
