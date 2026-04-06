package handlers

import (
	"net/http"

	"wow-log-analyzer/services/api-gateway/internal/services"
)

func RegisterRoutes(mux *http.ServeMux, analyzeService *services.AnalyzeService, reportService *services.ReportService) {
	analyzeHandler := NewAnalyzeHandler(analyzeService)
	reportHandler := NewReportHandler(reportService)

	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/analyze/intake", analyzeHandler.HandleIntake)
	mux.HandleFunc("/api/report/generate", reportHandler.Generate)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("API Gateway is running"))
}
