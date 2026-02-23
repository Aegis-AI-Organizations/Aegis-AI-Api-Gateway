package main

import (
	"fmt"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/gateway"
)

func main() {
	fmt.Println("Hello, world! Aegis AI API Gateway is starting...")
	gateway.Start()
}
