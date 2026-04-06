package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"wow-log-analyzer/services/ai-service/internal/service"
	"wow-log-analyzer/services/ai-service/internal/types"
)

type InsightHandler struct {
	insightService *service.InsightService
}

func NewInsightHandler(insightService *service.InsightService) *InsightHandler {
	return &InsightHandler{insightService: insightService}
}

func (h *InsightHandler) Generate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req types.InsightGenerationRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		log.Printf("Failed to decode insight request: %v", err)
		http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	response, err := h.insightService.GenerateInsights(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode insight response: %v", err)
	}
}
