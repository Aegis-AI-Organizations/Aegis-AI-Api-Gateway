package gateway

import (
	"fmt"
	"log"
	"net/http"
)

func Start() {
	fmt.Println("Aegis AI Web API Gateway started.")

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
			log.Printf("health write error: %v", err)
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"service":"aegis-api-gateway","version":"pre-alpha"}`)); err != nil {
			log.Printf("root write error: %v", err)
		}
	})

	log.Println("Listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
