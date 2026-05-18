package handlers

import (
	"net/http"
	"strings"

	"wow-log-analyzer/services/api-gateway/internal/config"
	"wow-log-analyzer/services/api-gateway/internal/services"
)

func RegisterRoutes(mux *http.ServeMux, cfg config.Config, analyzeService *services.AnalyzeService, reportService *services.ReportService, authService *services.AuthService, browserService *services.BrowserService) {
	analyzeHandler := NewAnalyzeHandler(analyzeService)
	reportHandler := NewReportHandler(reportService)
	authHandler := NewAuthHandler(authService, browserService, cfg)
	browserHandler := NewBrowserHandler(authService, browserService)

	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/auth/login", authHandler.Login)
	mux.HandleFunc("/api/auth/callback", authHandler.Callback)
	mux.HandleFunc("/api/auth/status", authHandler.Status)
	mux.HandleFunc("/api/auth/logout", authHandler.Logout)
	mux.HandleFunc("/api/browser/characters", browserHandler.GetCharacters)
	mux.HandleFunc("/api/browser/characters/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/reports") {
			browserHandler.GetCharacterReports(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/api/analyze/intake", analyzeHandler.HandleIntake)
	mux.HandleFunc("/api/analyze/characters", analyzeHandler.HandleCharacters)
	mux.HandleFunc("/api/report/jobs", reportHandler.CreateJob)
	mux.HandleFunc("/api/report/jobs/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/ability-timeline") {
			reportHandler.GetAbilityTimeline(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/buff-timeline") {
			reportHandler.GetBuffTimeline(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/resource-timeline") {
			reportHandler.GetResourceTimeline(w, r)
			return
		}
		reportHandler.GetJob(w, r)
	})
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("API Gateway is running"))
}
