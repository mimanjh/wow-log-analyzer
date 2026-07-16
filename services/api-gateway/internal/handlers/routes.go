package handlers

import (
	"net/http"
	"strings"

	"wow-log-analyzer/services/api-gateway/internal/config"
	"wow-log-analyzer/services/api-gateway/internal/services"
)

func RegisterRoutes(mux *http.ServeMux, cfg config.Config, analyzeService *services.AnalyzeService, reportService *services.ReportService, authService *services.AuthService, browserService *services.BrowserService, accountService *services.AccountService, billingService *services.BillingService, healthHandler *HealthHandler) {
	analyzeHandler := NewAnalyzeHandler(analyzeService)
	reportHandler := NewReportHandler(reportService, authService, accountService)
	authHandler := NewAuthHandler(authService, browserService, accountService, cfg)
	browserHandler := NewBrowserHandler(authService, browserService)
	billingHandler := NewBillingHandler(authService, accountService, billingService, cfg.FrontendURL)

	mux.Handle("/", NewStaticHandler())
	mux.HandleFunc("/health", healthHandler.Handle)
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
	mux.HandleFunc("/api/analyze/fights", analyzeHandler.HandleFights)
	mux.HandleFunc("/api/analyze/characters", analyzeHandler.HandleCharacters)
	mux.HandleFunc("/api/report/jobs", reportHandler.CreateJob)
	mux.HandleFunc("/api/reports", reportHandler.List)
	mux.HandleFunc("/api/billing/checkout", billingHandler.CreateCheckout)
	mux.HandleFunc("/api/billing/status", billingHandler.GetStatus)
	mux.HandleFunc("/api/billing/portal", billingHandler.CreatePortal)
	mux.HandleFunc("/api/billing/webhook", billingHandler.Webhook)
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
