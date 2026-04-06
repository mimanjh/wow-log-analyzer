package handlers

import (
	"net/http"

	"wow-log-analyzer/services/analysis-service/internal/service"
)

func RegisterRoutes(mux *http.ServeMux, analysisService *service.AnalysisService) {
	handler := NewAnalysisHandler(analysisService)

	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/analyze/fight", handler.AnalyzeFight)
	mux.HandleFunc("/analyze/compare", handler.CompareFight)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Analysis Service is running"))
}
