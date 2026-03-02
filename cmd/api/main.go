package main

import (
	"fmt"
	"log"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/db"
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

	api.Start(database, tc)
}
