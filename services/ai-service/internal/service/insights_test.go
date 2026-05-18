package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
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

func TestNewModelClientRequiresLiveOpenAIConfig(t *testing.T) {
	disabled := newModelClient(config.Config{
		Provider:         "openai",
		Model:            "gpt-5-mini",
		ModelAPIKey:      "test-key",
		LiveModelEnabled: false,
	})
	if _, ok := disabled.(disabledModelClient); !ok {
		t.Fatalf("expected disabled model client when live model is disabled")
	}

	enabled := newModelClient(config.Config{
		Provider:         "openai",
		Model:            "gpt-5-mini",
		ModelAPIKey:      "test-key",
		LiveModelEnabled: true,
	})
	if _, ok := enabled.(openAIModelClient); !ok {
		t.Fatalf("expected OpenAI model client when provider, key, and live flag are set")
	}
}

func TestResponseTextReadsResponsesOutputContent(t *testing.T) {
	response := openAIResponsesResponse{
		Output: []struct {
			Type    string `json:"type"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		}{
			{
				Type: "message",
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{
					{Type: "output_text", Text: `{"insights":[]`},
					{Type: "output_text", Text: `}`},
				},
			},
		},
	}

	if got := responseText(response); got != `{"insights":[]}` {
		t.Fatalf("expected concatenated output content, got %s", got)
	}
}

func TestOpenAIModelClientGenerateParsesResponsesOutputContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("expected bearer auth header")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"model":"gpt-5-mini",
			"output":[
				{
					"type":"message",
					"content":[
						{
							"type":"output_text",
							"text":"{\"insights\":[{\"metricKey\":\"castsPerMin\",\"title\":\"Tighten casts\",\"summary\":\"Review the cast gap using deterministic comparisons only.\",\"confidence\":\"medium\"},{\"metricKey\":\"buff:uptime\",\"title\":\"Buff window\",\"summary\":\"Buff timing trailed the elite sample.\",\"confidence\":\"high\"},{\"metricKey\":\"resource:runes\",\"title\":\"Rune use\",\"summary\":\"Rune use was below the elite reference.\",\"confidence\":\"medium\"}],\"focusRecommendation\":{\"metricKey\":\"castsPerMin\",\"title\":\"Focus casts\",\"recommendation\":\"Review cast pacing next pull.\",\"reasoning\":\"It is the clearest deterministic gap.\"},\"fallbackUsed\":false,\"model\":\"gpt-5-mini\"}"
						}
					]
				}
			]
		}`))
	}))
	defer server.Close()

	client := openAIModelClient{
		apiKey:     "test-key",
		model:      "gpt-5-mini",
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	response, err := client.Generate(context.Background(), "test prompt", validRequest())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response.FallbackUsed {
		t.Fatal("expected model response, got fallback")
	}
	if response.Model != "gpt-5-mini" {
		t.Fatalf("expected model gpt-5-mini, got %s", response.Model)
	}
	if len(response.Insights) != 3 {
		t.Fatalf("expected three insights, got %d", len(response.Insights))
	}
	if response.FocusRecommendation.MetricKey == "" {
		t.Fatal("expected focus recommendation")
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
