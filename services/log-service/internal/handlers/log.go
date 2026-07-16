package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"wow-log-analyzer/services/log-service/internal/service"
	"wow-log-analyzer/services/log-service/internal/types"
)

// WCL report codes are alphanumeric; anonymous reports carry an "a:" prefix.
var reportIDPattern = regexp.MustCompile(`^[A-Za-z0-9:]{4,64}$`)

func validReportID(reportID string) bool {
	return reportIDPattern.MatchString(reportID)
}

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
	if !validReportID(reportID) {
		http.Error(w, "A valid report ID is required", http.StatusBadRequest)
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
	if !validReportID(reportID) {
		http.Error(w, "A valid report ID is required", http.StatusBadRequest)
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
	if !validReportID(reportID) {
		http.Error(w, "A valid report ID is required", http.StatusBadRequest)
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
	if !validReportID(reportID) {
		http.Error(w, "A valid report ID is required", http.StatusBadRequest)
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

func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	accessToken, ok := bearerToken(r)
	if !ok {
		http.Error(w, "authorization bearer token is required", http.StatusUnauthorized)
		return
	}

	user, err := h.logService.GetCurrentUser(accessToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) GetOwnedCharacters(w http.ResponseWriter, r *http.Request) {
	accessToken, ok := bearerToken(r)
	if !ok {
		http.Error(w, "authorization bearer token is required", http.StatusUnauthorized)
		return
	}

	characters, err := h.logService.GetOwnedCharacters(accessToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(characters)
}

func (h *Handler) GetCharacterReports(w http.ResponseWriter, r *http.Request) {
	accessToken, ok := bearerToken(r)
	if !ok {
		http.Error(w, "authorization bearer token is required", http.StatusUnauthorized)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/user/characters/")
	path = strings.TrimSuffix(path, "/reports")
	characterID, err := strconv.Atoi(path)
	if err != nil || characterID <= 0 {
		http.Error(w, "invalid character id", http.StatusBadRequest)
		return
	}

	limit := 10
	if value := r.URL.Query().Get("limit"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	page, err := h.logService.GetCharacterReports(accessToken, characterID, r.URL.Query().Get("cursor"), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(page)
}

func bearerToken(r *http.Request) (string, bool) {
	value := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(strings.ToLower(value), "bearer ") {
		return "", false
	}
	return strings.TrimSpace(value[7:]), true
}

func (h *Handler) GetPlayerData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reportID := strings.TrimPrefix(r.URL.Path, "/reports/")
	reportID = strings.TrimSuffix(reportID, "/player-data")
	if !validReportID(reportID) {
		http.Error(w, "A valid report ID is required", http.StatusBadRequest)
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

	if !validReportID(req.Candidate.ReportID) || req.Candidate.FightID == 0 {
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
