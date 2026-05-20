package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"wow-log-analyzer/services/api-gateway/internal/services"
)

type ReportHandler struct {
	reportService *services.ReportService
	authService   *services.AuthService
}

func NewReportHandler(reportService *services.ReportService, authService *services.AuthService) *ReportHandler {
	return &ReportHandler{reportService: reportService, authService: authService}
}

func (h *ReportHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req services.GenerateReportRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		log.Printf("Failed to decode report request: %v", err)
		http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	// Enforce daily usage limit for authenticated users, but skip if result is already cached.
	if cookie, err := r.Cookie("wowlog_session"); err == nil {
		if session, ok := h.authService.GetSession(cookie.Value); ok && session.User != nil {
			if !h.reportService.HasCachedResult(req) {
				limit := services.TierDailyLimit(session.AccountTier)
				allowed, _, usageErr := h.reportService.CheckAndIncrementDailyUsage(r.Context(), session.User.ID, limit)
				if usageErr != nil {
					log.Printf("Usage check failed for user %d: %v", session.User.ID, usageErr)
				} else if !allowed {
					http.Error(w, fmt.Sprintf("Daily analysis limit of %d reached. Try again tomorrow.", limit), http.StatusTooManyRequests)
					return
				}
			}
		}
	}

	response, err := h.reportService.CreateJob(req)
	if err != nil {
		log.Printf("Failed to create report job: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode report response: %v", err)
	}
}

func (h *ReportHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobID := strings.TrimPrefix(r.URL.Path, "/api/report/jobs/")
	if strings.TrimSpace(jobID) == "" {
		http.Error(w, "jobId is required", http.StatusBadRequest)
		return
	}

	job, err := h.reportService.GetJob(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(job); err != nil {
		log.Printf("Failed to encode report job response: %v", err)
	}
}

func (h *ReportHandler) GetAbilityTimeline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobID := strings.TrimPrefix(r.URL.Path, "/api/report/jobs/")
	jobID = strings.TrimSuffix(jobID, "/ability-timeline")
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		http.Error(w, "jobId is required", http.StatusBadRequest)
		return
	}

	abilityID, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("abilityId")))
	if err != nil || abilityID == 0 {
		http.Error(w, "abilityId is required", http.StatusBadRequest)
		return
	}

	response, err := h.reportService.GetAbilityTimeline(jobID, abilityID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode ability timeline response: %v", err)
	}
}

func (h *ReportHandler) GetBuffTimeline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobID := strings.TrimPrefix(r.URL.Path, "/api/report/jobs/")
	jobID = strings.TrimSuffix(jobID, "/buff-timeline")
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		http.Error(w, "jobId is required", http.StatusBadRequest)
		return
	}

	abilityID, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("abilityId")))
	if err != nil || abilityID == 0 {
		http.Error(w, "abilityId is required", http.StatusBadRequest)
		return
	}

	response, err := h.reportService.GetBuffTimeline(jobID, abilityID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode buff timeline response: %v", err)
	}
}

func (h *ReportHandler) GetResourceTimeline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobID := strings.TrimPrefix(r.URL.Path, "/api/report/jobs/")
	jobID = strings.TrimSuffix(jobID, "/resource-timeline")
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		http.Error(w, "jobId is required", http.StatusBadRequest)
		return
	}

	resourceTypeID, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("resourceTypeId")))
	if err != nil {
		http.Error(w, "resourceTypeId is required", http.StatusBadRequest)
		return
	}

	response, err := h.reportService.GetResourceTimeline(jobID, resourceTypeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode resource timeline response: %v", err)
	}
}
