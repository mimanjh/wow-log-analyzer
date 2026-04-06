package main

import (
	"fmt"
	"log"
	"net/http"

	"wow-log-analyzer/services/analysis-service/internal/config"
	"wow-log-analyzer/services/analysis-service/internal/handlers"
	"wow-log-analyzer/services/analysis-service/internal/service"
)

func main() {
	cfg := config.Load()
	analysisService := service.NewAnalysisService()

	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux, analysisService)

	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("analysis-service starting on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalf("analysis-service server failed: %v", err)
	}
}
