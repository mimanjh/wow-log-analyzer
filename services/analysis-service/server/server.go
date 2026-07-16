package server

import (
	"fmt"
	"log"
	"net/http"

	"wow-log-analyzer/internal/httpserver"
	"wow-log-analyzer/services/analysis-service/internal/config"
	"wow-log-analyzer/services/analysis-service/internal/handlers"
	"wow-log-analyzer/services/analysis-service/internal/service"
)

func Run() error {
	cfg := config.Load()
	analysisService := service.NewAnalysisService()

	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux, analysisService)

	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("analysis-service starting on %s", serverAddr)
	return httpserver.ListenAndServe("analysis-service", serverAddr, mux)
}
