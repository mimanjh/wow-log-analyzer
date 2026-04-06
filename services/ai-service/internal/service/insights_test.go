package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"wow-log-analyzer/services/ai-service/internal/config"
	"wow-log-analyzer/services/ai-service/internal/types"
)

type stubModelClient struct {
	response *types.InsightGenerationResponse
	err      error
}

func (s stubModelClient) Generate(context.Context, string, types.InsightGenerationRequest) (*types.InsightGenerationResponse, error) {
	return s.response, s.err
}

func validRequest() types.InsightGenerationRequest {
	return types.InsightGenerationRequest{
		Context: types.InsightContext{
			EncounterName:    "Queen Ansurek",
			Difficulty:       "Heroic",
			CharacterName:    "Testmage",
			CharacterClass:   "Mage",
			CharacterSpec:    "Fire",
			FightDurationSec: 312,
			CohortSize:       25,
		},
		Metrics: []types.InsightMetric{
			{
				Key:            "castsPerMin",
				Label:          "Casts per Minute",
				HigherIsBetter: true,
				PlayerValue:    2.4,
				CohortValue:    2.8,
				Difference:     -0.4,
				Percentile:     32,
				Confidence:     "high",
			},
			{
				Key:            "majorCdDrift",
				Label:          "Major Cooldown Timing Drift",
				Unit:           "s",
				HigherIsBetter: false,
				PlayerValue:    6.2,
				CohortValue:    3.9,
				Difference:     2.3,
				Percentile:     70,
				Confidence:     "medium",
				Caution:        "Low sample size of cooldown pairs",
			},
			{
				Key:            "buffUptime",
				Label:          "Buff Uptime",
				Unit:           "%",
				HigherIsBetter: true,
				PlayerValue:    88,
				CohortValue:    85,
				Difference:     3,
				Percentile:     61,
				Confidence:     "high",
			},
		},
	}
}

func TestInsightService_GenerateInsightsFallsBack(t *testing.T) {
	service := NewInsightService(config.Config{})
	service.modelClient = stubModelClient{err: errors.New("provider failed")}

	response, err := service.GenerateInsights(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !response.FallbackUsed {
		t.Fatal("expected fallback response")
	}
	if len(response.Insights) != 3 {
		t.Fatalf("expected 3 insights, got %d", len(response.Insights))
	}
	if response.FocusRecommendation.MetricKey == "" {
		t.Fatal("expected focus recommendation metric key")
	}
}

func TestInsightService_GenerateInsightsUsesModelWhenAvailable(t *testing.T) {
	service := NewInsightService(config.Config{Model: "test-model"})
	service.modelClient = stubModelClient{
		response: &types.InsightGenerationResponse{
			Insights: []types.AIInsight{
				{MetricKey: "castsPerMin", Title: "Keep uptime higher", Summary: "Example", Confidence: "medium"},
				{MetricKey: "majorCdDrift", Title: "Tighten cooldown timings", Summary: "Example", Confidence: "medium"},
				{MetricKey: "buffUptime", Title: "Buff coverage is stable", Summary: "Example", Confidence: "high"},
			},
			FocusRecommendation: types.FocusRecommendation{
				MetricKey:      "castsPerMin",
				Title:          "Focus on casts per minute",
				Recommendation: "Example",
				Reasoning:      "Example",
			},
		},
	}

	response, err := service.GenerateInsights(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response.FallbackUsed {
		t.Fatal("expected model response")
	}
	if response.Model != "test-model" {
		t.Fatalf("expected model name to be preserved, got %s", response.Model)
	}
}

func TestInsightService_RejectsInvalidRequest(t *testing.T) {
	service := NewInsightService(config.Config{})
	req := validRequest()
	req.Metrics[0].Confidence = "certain"

	_, err := service.GenerateInsights(context.Background(), req)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestBuildPromptExcludesRawLogReferences(t *testing.T) {
	prompt := buildPrompt(validRequest())

	if strings.Contains(strings.ToLower(prompt), "inspect raw logs") {
		t.Fatal("prompt should not ask the model to inspect raw logs")
	}
	if !strings.Contains(prompt, "Use only these deterministic comparison outputs") {
		t.Fatal("prompt should emphasize deterministic-only input")
	}
}
