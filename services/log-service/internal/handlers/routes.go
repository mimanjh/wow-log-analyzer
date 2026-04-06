package handlers

import (
	"net/http"
	"strings"

	"wow-log-analyzer/services/log-service/internal/service"
)

func RegisterRoutes(mux *http.ServeMux, logService *service.LogService) {
	handler := NewHandler(logService)

	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/reports/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/comparison-data") {
			handler.GetComparisonData(w, r)
		} else if strings.HasSuffix(r.URL.Path, "/characters") {
			handler.GetCharacters(w, r)
		} else if strings.HasSuffix(r.URL.Path, "/fights") {
			handler.GetFights(w, r)
		} else {
			handler.GetReportMetadata(w, r)
		}
	})
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Log Service is running"))
}
