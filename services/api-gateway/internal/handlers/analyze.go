package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"wow-log-analyzer/services/api-gateway/internal/services"
)

type AnalyzeHandler struct {
	analyzeService *services.AnalyzeService
}

func NewAnalyzeHandler(analyzeService *services.AnalyzeService) *AnalyzeHandler {
	return &AnalyzeHandler{
		analyzeService: analyzeService,
	}
}

func (h *AnalyzeHandler) HandleIntake(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req services.AnalyzeIntakeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Url == "" {
		http.Error(w, "url field is required", http.StatusBadRequest)
		return
	}

	response, err := h.analyzeService.ProcessIntake(req)
	if err != nil {
		log.Printf("Failed to process intake: %v", err)
		// For validation errors, return 400 with the error message
		if validationErr, ok := err.(interface{ Error() string }); ok && len(validationErr.Error()) > 18 && validationErr.Error()[:18] == "validation failed:" {
			http.Error(w, validationErr.Error()[20:], http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (h *AnalyzeHandler) HandleCharacters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reportID := r.URL.Query().Get("reportId")
	fightIDParam := r.URL.Query().Get("fightId")
	if reportID == "" || fightIDParam == "" {
		http.Error(w, "reportId and fightId are required", http.StatusBadRequest)
		return
	}

	fightID, err := strconv.Atoi(fightIDParam)
	if err != nil || fightID == 0 {
		http.Error(w, "fightId must be a valid integer", http.StatusBadRequest)
		return
	}

	characters, err := h.analyzeService.GetCharactersForFight(reportID, fightID)
	if err != nil {
		log.Printf("Failed to get characters: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(characters); err != nil {
		log.Printf("Failed to encode characters response: %v", err)
	}
}
