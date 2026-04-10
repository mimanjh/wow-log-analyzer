package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"wow-log-analyzer/services/ai-service/internal/config"
	"wow-log-analyzer/services/ai-service/internal/types"
)

type disabledModelClient struct{}

type openAIModelClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

type openAIResponsesRequest struct {
	Model string     `json:"model"`
	Input string     `json:"input"`
	Text  openAIText `json:"text"`
}

type openAIText struct {
	Format openAIFormat `json:"format"`
}

type openAIFormat struct {
	Type   string         `json:"type"`
	Name   string         `json:"name"`
	Schema map[string]any `json:"schema"`
	Strict bool           `json:"strict"`
}

type openAIResponsesResponse struct {
	Model      string `json:"model"`
	OutputText string `json:"output_text"`
	Error      *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func newModelClient(cfg config.Config) ModelClient {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if provider == "" || provider == "disabled" {
		return disabledModelClient{}
	}

	if provider == "openai" {
		if strings.TrimSpace(cfg.ModelAPIKey) == "" {
			return disabledModelClient{}
		}
		return openAIModelClient{
			apiKey:  cfg.ModelAPIKey,
			model:   cfg.Model,
			baseURL: "https://api.openai.com/v1/responses",
			httpClient: &http.Client{
				Timeout: 60 * time.Second,
			},
		}
	}

	return disabledModelClient{}
}

func (disabledModelClient) Generate(context.Context, string, types.InsightGenerationRequest) (*types.InsightGenerationResponse, error) {
	return nil, errors.New("model provider is not configured")
}

func (c openAIModelClient) Generate(ctx context.Context, prompt string, _ types.InsightGenerationRequest) (*types.InsightGenerationResponse, error) {
	requestBody := openAIResponsesRequest{
		Model: c.model,
		Input: prompt,
		Text: openAIText{
			Format: openAIFormat{
				Type:   "json_schema",
				Name:   "insight_generation_response",
				Schema: insightResponseSchema(),
				Strict: true,
			},
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai returned status %d: %s", resp.StatusCode, string(responseBody))
	}

	var response openAIResponsesResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, err
	}
	if response.Error != nil {
		return nil, errors.New(response.Error.Message)
	}
	if strings.TrimSpace(response.OutputText) == "" {
		return nil, errors.New("openai response did not include output_text")
	}

	var parsed types.InsightGenerationResponse
	if err := json.Unmarshal([]byte(response.OutputText), &parsed); err != nil {
		return nil, fmt.Errorf("failed to decode openai structured output: %w", err)
	}
	if response.Model != "" {
		parsed.Model = response.Model
	}
	return &parsed, nil
}

func insightResponseSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"insights": map[string]any{
				"type":     "array",
				"minItems": 3,
				"maxItems": 3,
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"properties": map[string]any{
						"metricKey": map[string]any{"type": "string"},
						"title":     map[string]any{"type": "string"},
						"summary":   map[string]any{"type": "string"},
						"confidence": map[string]any{
							"type": "string",
							"enum": []string{"high", "medium", "low"},
						},
						"caution": map[string]any{"type": "string"},
					},
					"required": []string{"metricKey", "title", "summary", "confidence", "caution"},
				},
			},
			"focusRecommendation": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"metricKey":      map[string]any{"type": "string"},
					"title":          map[string]any{"type": "string"},
					"recommendation": map[string]any{"type": "string"},
					"reasoning":      map[string]any{"type": "string"},
				},
				"required": []string{"metricKey", "title", "recommendation", "reasoning"},
			},
			"fallbackUsed": map[string]any{"type": "boolean"},
			"model":        map[string]any{"type": "string"},
		},
		"required": []string{"insights", "focusRecommendation", "fallbackUsed", "model"},
	}
}
