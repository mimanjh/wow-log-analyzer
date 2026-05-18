package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"wow-log-analyzer/services/api-gateway/internal/services"
)

type ReportHandler struct {
	reportService *services.ReportService
}

func NewReportHandler(reportService *services.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
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
