package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"wow-log-analyzer/services/api-gateway/internal/services"
)

type ReportHandler struct {
	reportService *services.ReportService
}

func NewReportHandler(reportService *services.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

func (h *ReportHandler) Generate(w http.ResponseWriter, r *http.Request) {
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

	response, err := h.reportService.Generate(r.Context(), req)
	if err != nil {
		log.Printf("Failed to generate report: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode report response: %v", err)
	}
}
