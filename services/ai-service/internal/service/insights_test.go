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

func TestBuildPromptIncludesSpecProfileContext(t *testing.T) {
	req := validRequest()
	req.Context.CharacterClass = "Death Knight"
	req.Context.CharacterSpec = "Blood"
	req.Context.SpecProfile = types.SpecProfile{
		Label:          "Blood Death Knight",
		Role:           "Tank",
		StatPriorities: []string{"Haste", "Mastery"},
		KeyMechanics: []string{
			"Build and spend Runic Power deliberately.",
			"Manage rune availability so core spenders are not delayed.",
		},
		Rotation: []types.SpecGuideSection{
			{
				Context:    "Single Target",
				HeroTalent: "Deathbringer",
				Steps: []types.SpecGuideStep{
					{Text: "Use spell 206930 to build Runic Power without running out of runes.", SpellIDs: []string{"206930"}},
					{Text: "Use spell 49998 for self-healing when needed.", SpellIDs: []string{"49998"}},
				},
			},
		},
	}

	prompt := buildPrompt(req)
	if !strings.Contains(prompt, "Spec guide context:") {
		t.Fatal("expected spec guide context section")
	}
	if !strings.Contains(prompt, "Blood Death Knight") {
		t.Fatal("expected spec label in prompt")
	}
	if !strings.Contains(prompt, "Use spec guide context only to explain why a deterministic gap matters") {
		t.Fatal("expected prompt guardrail for spec context")
	}
	if !strings.Contains(prompt, "Runic Power") || !strings.Contains(strings.ToLower(prompt), "runes") {
		t.Fatal("expected Blood DK resource context in prompt")
	}
}

func TestNewModelClientRequiresLiveAnthropicConfig(t *testing.T) {
	disabled := newModelClient(config.Config{
		Provider:         "anthropic",
		Model:            "claude-sonnet-4-6",
		ModelAPIKey:      "test-key",
		LiveModelEnabled: false,
	})
	if _, ok := disabled.(disabledModelClient); !ok {
		t.Fatalf("expected disabled model client when live model is disabled")
	}

	missingKey := newModelClient(config.Config{
		Provider:         "anthropic",
		Model:            "claude-sonnet-4-6",
		LiveModelEnabled: true,
	})
	if _, ok := missingKey.(disabledModelClient); !ok {
		t.Fatalf("expected disabled model client when api key is empty")
	}

	wrongProvider := newModelClient(config.Config{
		Provider:         "openai",
		Model:            "gpt-5-mini",
		ModelAPIKey:      "test-key",
		LiveModelEnabled: true,
	})
	if _, ok := wrongProvider.(disabledModelClient); !ok {
		t.Fatalf("expected disabled model client for unsupported provider %q", "openai")
	}

	enabled := newModelClient(config.Config{
		Provider:         "anthropic",
		Model:            "claude-sonnet-4-6",
		ModelAPIKey:      "test-key",
		LiveModelEnabled: true,
	})
	if _, ok := enabled.(anthropicModelClient); !ok {
		t.Fatalf("expected anthropic model client when provider, key, and live flag are set")
	}
}

func TestInsightResponseSchemaShape(t *testing.T) {
	schema := insightResponseSchema()
	if schema["type"] != "object" {
		t.Fatalf("expected top-level type=object")
	}
	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatalf("expected required as []string")
	}
	wantRequired := map[string]bool{"insights": true, "focusRecommendation": true, "fallbackUsed": true, "model": true}
	for _, r := range required {
		delete(wantRequired, r)
	}
	if len(wantRequired) != 0 {
		t.Fatalf("missing required fields: %v", wantRequired)
	}
}

func TestSummarizeModelErrorCompactsLongMessages(t *testing.T) {
	message := summarizeModelError(errors.New(strings.Repeat("provider failure\n", 40)))
	if strings.Contains(message, "\n") {
		t.Fatalf("expected newlines to be removed, got %q", message)
	}
	if len(message) > 303 {
		t.Fatalf("expected compact message, got length %d", len(message))
	}
}
