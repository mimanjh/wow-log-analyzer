package server

import (
	"fmt"
	"log"
	"net/http"

	"wow-log-analyzer/services/ai-service/internal/config"
	"wow-log-analyzer/services/ai-service/internal/handlers"
	"wow-log-analyzer/services/ai-service/internal/service"
)

func Run() error {
	cfg := config.Load()
	mux := http.NewServeMux()
	insightService := service.NewInsightService(cfg)
	handlers.RegisterRoutes(mux, insightService)

	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf(
		"ai-service config: provider=%s model=%s apiKeyConfigured=%t liveModelEnabled=%t",
		cfg.Provider,
		cfg.Model,
		cfg.ModelAPIKey != "",
		cfg.LiveModelEnabled,
	)
	log.Printf("ai-service starting on %s", serverAddr)
	return http.ListenAndServe(serverAddr, mux)
}
