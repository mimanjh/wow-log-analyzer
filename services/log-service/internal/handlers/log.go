package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"wow-log-analyzer/services/log-service/internal/service"
	"wow-log-analyzer/services/log-service/internal/types"
)

type Handler struct {
	logService *service.LogService
}

func NewHandler(logService *service.LogService) *Handler {
	return &Handler{
		logService: logService,
	}
}

func (h *Handler) GetReportMetadata(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reportID := strings.TrimPrefix(r.URL.Path, "/reports/")
	if reportID == "" {
		http.Error(w, "Report ID is required", http.StatusBadRequest)
		return
	}

	report, err := h.logService.GetReportMetadata(reportID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (h *Handler) GetFights(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reportID := strings.TrimPrefix(r.URL.Path, "/reports/")
	reportID = strings.TrimSuffix(reportID, "/fights")
	if reportID == "" {
		http.Error(w, "Report ID is required", http.StatusBadRequest)
		return
	}

	fights, err := h.logService.GetFights(reportID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fights)
}

func (h *Handler) GetComparisonData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reportID := strings.TrimPrefix(r.URL.Path, "/reports/")
	reportID = strings.TrimSuffix(reportID, "/comparison-data")
	if reportID == "" {
		http.Error(w, "Report ID is required", http.StatusBadRequest)
		return
	}

	var req types.ComparisonDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.FightID == 0 || req.CharacterID == 0 {
		http.Error(w, "fightId and characterId are required", http.StatusBadRequest)
		return
	}

	response, err := h.logService.GetComparisonData(reportID, req.FightID, req.CharacterID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
