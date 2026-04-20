package main

import (
	"fmt"
	"log"

	"os"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc"
)

func main() {
	fmt.Println("🚀 Aegis AI API Gateway is starting...")

	brainAddr := os.Getenv("BRAIN_GRPC_ADDR")
	if brainAddr == "" {
		brainAddr = "localhost:50051"
	}

	// TLS Configuration for Brain gRPC
	tlsConf := agrpc.TLSConfig{
		Enable:     os.Getenv("BRAIN_TLS_ENABLE") == "true",
		CAPath:     os.Getenv("BRAIN_TLS_CA_CERT"),
		CertPath:   os.Getenv("BRAIN_TLS_CLIENT_CERT"),
		KeyPath:    os.Getenv("BRAIN_TLS_CLIENT_KEY"),
		ServerName: os.Getenv("BRAIN_TLS_SERVER_NAME"),
	}

	if tlsConf.CAPath == "" {
		tlsConf.CAPath = "/etc/brain/certs/ca.crt"
	}
	if tlsConf.CertPath == "" {
		tlsConf.CertPath = "/etc/brain/certs/tls.crt"
	}
	if tlsConf.KeyPath == "" {
		tlsConf.KeyPath = "/etc/brain/certs/tls.key"
	}

	gc, err := agrpc.NewClient(brainAddr, tlsConf)
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
