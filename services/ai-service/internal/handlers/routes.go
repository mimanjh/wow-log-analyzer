package handlers

import (
	"net/http"

	"wow-log-analyzer/services/ai-service/internal/service"
)

func RegisterRoutes(mux *http.ServeMux, insightService *service.InsightService) {
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/insights/generate", NewInsightHandler(insightService).Generate)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("AI Service is running"))
}
