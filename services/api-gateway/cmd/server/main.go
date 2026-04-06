package main

import (
	"fmt"
	"log"
	"net/http"

	"wow-log-analyzer/services/api-gateway/internal/config"
	"wow-log-analyzer/services/api-gateway/internal/handlers"
	"wow-log-analyzer/services/api-gateway/internal/services"
)

func main() {
	cfg := config.Load()
	mux := http.NewServeMux()
	analyzeService := services.NewAnalyzeService(cfg.LogServiceURL)
	reportService := services.NewReportService(cfg.LogServiceURL, cfg.AnalysisServiceURL, cfg.AIServiceURL)
	handlers.RegisterRoutes(mux, analyzeService, reportService)

	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("api-gateway starting on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalf("api-gateway server failed: %v", err)
	}
}
