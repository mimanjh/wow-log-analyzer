package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"wow-log-analyzer/services/analysis-service/internal/service"
	"wow-log-analyzer/services/analysis-service/internal/types"
)

type AnalysisHandler struct {
	analysisService *service.AnalysisService
}

func NewAnalysisHandler(analysisService *service.AnalysisService) *AnalysisHandler {
	return &AnalysisHandler{
		analysisService: analysisService,
	}
}

func (h *AnalysisHandler) AnalyzeFight(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var playerData types.PlayerFightData
	if err := json.NewDecoder(r.Body).Decode(&playerData); err != nil {
		log.Printf("Failed to decode request: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	metrics, err := h.analysisService.AnalyzePlayerFight(playerData)
	if err != nil {
		log.Printf("Failed to analyze fight: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (h *AnalysisHandler) CompareFight(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PlayerData     types.PlayerFightData   `json:"playerData"`
		CohortData     []types.PlayerFightData `json:"cohortData"`
		CharacterClass string                  `json:"characterClass"`
		CharacterSpec  string                  `json:"characterSpec"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Analyze player metrics
	playerMetrics, err := h.analysisService.AnalyzePlayerFight(req.PlayerData)
	if err != nil {
		log.Printf("Failed to analyze player fight: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Analyze cohort metrics
	cohortMetrics := make([]types.PlayerFightMetrics, len(req.CohortData))
	for i, cohortData := range req.CohortData {
		metrics, err := h.analysisService.AnalyzePlayerFight(cohortData)
		if err != nil {
			log.Printf("Failed to analyze cohort fight %d: %v", i, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cohortMetrics[i] = *metrics
	}

	// Compare against cohort
	comparison, err := h.analysisService.CompareAgainstCohort(*playerMetrics, cohortMetrics)
	if err != nil {
		log.Printf("Failed to compare against cohort: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	comparison.AbilityUsage = h.analysisService.CalculateAbilityUsageComparisons(req.PlayerData, req.CohortData, req.CharacterClass, req.CharacterSpec)
	comparison.BuffUptimes = h.analysisService.CalculateBuffUptimeComparisons(req.PlayerData, req.CohortData)
	comparison.ResourceUsage = h.analysisService.CalculateResourceUsageComparisons(req.PlayerData, req.CohortData, req.CharacterClass, req.CharacterSpec)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comparison)
}
