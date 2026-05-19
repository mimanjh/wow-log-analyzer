package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
	"wow-log-analyzer/services/api-gateway/internal/config"
	"wow-log-analyzer/services/api-gateway/internal/handlers"
	"wow-log-analyzer/services/api-gateway/internal/services"
)

func main() {
	cfg := config.Load()
	mux := http.NewServeMux()

	var redisClient *redis.Client
	if cfg.RedisAddr != "" {
		redisClient = redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
		log.Printf("Redis configured at %s", cfg.RedisAddr)
	}

	analyzeService := services.NewAnalyzeService(cfg.LogServiceURL)
	reportService := services.NewReportService(cfg.LogServiceURL, cfg.AnalysisServiceURL, cfg.AIServiceURL, redisClient)
	authService := services.NewAuthService(cfg)
	browserService := services.NewBrowserService(cfg.LogServiceURL)
	handlers.RegisterRoutes(mux, cfg, analyzeService, reportService, authService, browserService)

	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("api-gateway starting on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalf("api-gateway server failed: %v", err)
	}
}
