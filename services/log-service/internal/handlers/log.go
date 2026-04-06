package handlers

import (
	"encoding/json"
	"fmt"
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

func (h *Handler) GetCharacters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reportID := strings.TrimPrefix(r.URL.Path, "/reports/")
	reportID = strings.TrimSuffix(reportID, "/characters")
	if reportID == "" {
		http.Error(w, "Report ID is required", http.StatusBadRequest)
		return
	}

	fightIDParam := r.URL.Query().Get("fightId")
	if fightIDParam == "" {
		http.Error(w, "fightId is required", http.StatusBadRequest)
		return
	}

	var fightID int
	if _, err := fmt.Sscanf(fightIDParam, "%d", &fightID); err != nil || fightID == 0 {
		http.Error(w, "fightId must be a valid integer", http.StatusBadRequest)
		return
	}

	characters, err := h.logService.GetCharacters(reportID, fightID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(characters)
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

	if req.Fight.ID == 0 || req.CharacterID == 0 {
		http.Error(w, "fight.id and characterId are required", http.StatusBadRequest)
		return
	}

	response, err := h.logService.GetComparisonData(reportID, req.Fight, req.CharacterID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetPlayerData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reportID := strings.TrimPrefix(r.URL.Path, "/reports/")
	reportID = strings.TrimSuffix(reportID, "/player-data")
	if reportID == "" {
		http.Error(w, "Report ID is required", http.StatusBadRequest)
		return
	}

	var req types.PlayerDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Fight.ID == 0 || req.CharacterID == 0 {
		http.Error(w, "fight.id and characterId are required", http.StatusBadRequest)
		return
	}

	response, err := h.logService.GetPlayerFightData(reportID, req.Fight, req.CharacterID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetRankingCandidates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req types.RankingCandidatesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Fight.ID == 0 || strings.TrimSpace(req.CharacterClass) == "" || strings.TrimSpace(req.CharacterSpec) == "" {
		http.Error(w, "fight, characterClass, and characterSpec are required", http.StatusBadRequest)
		return
	}

	candidates, err := h.logService.GetRankingCandidates(req.Fight, req.CharacterClass, req.CharacterSpec, req.Limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(candidates)
}

func (h *Handler) GetCohortMemberData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req types.CohortMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Candidate.ReportID == "" || req.Candidate.FightID == 0 {
		http.Error(w, "candidate reportId and fightId are required", http.StatusBadRequest)
		return
	}

	response, err := h.logService.GetCohortMemberData(req.Candidate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
