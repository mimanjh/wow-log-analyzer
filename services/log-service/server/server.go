package server

import (
	"fmt"
	"log"
	"net/http"

	"wow-log-analyzer/services/log-service/internal/config"
	"wow-log-analyzer/services/log-service/internal/handlers"
	"wow-log-analyzer/services/log-service/internal/service"
)

func Run() error {
	cfg := config.Load()
	logService := service.NewLogService(cfg.WCL)

	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux, logService)

	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("log-service starting on %s", serverAddr)
	return http.ListenAndServe(serverAddr, mux)
}
