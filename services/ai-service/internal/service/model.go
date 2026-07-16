package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"wow-log-analyzer/services/ai-service/internal/config"
	"wow-log-analyzer/services/ai-service/internal/types"
)

type disabledModelClient struct{}

type anthropicModelClient struct {
	client anthropic.Client
	model  anthropic.Model
}

const insightToolName = "emit_insights"

const insightSystemPrompt = `You are a deterministic-data-grounded World of Warcraft raid analysis coach.

Your job: produce 3 short, cautious coaching insights and 1 focus recommendation based ONLY on the deterministic comparison data the user provides. Return your response by calling the emit_insights tool exactly once.

Rules:
- Use ONLY the comparison data in the user message. Never invent causality. Never speculate about events the data does not show.
- Spec guide context is for explaining WHY a deterministic gap matters for the player's spec — not as evidence the player made a mistake.
- Each insight should be concrete, 1-3 sentences, and grounded in the specific delta the data shows.
- Prioritize: ability usage timing, buff uptime gaps, missed cooldown windows, resource overcap, and clear strengths.
- Never mention raw logs. Never claim to know the player's intent.
- Confidence: "high" when the delta is clear with a solid sample; "medium" when there's noise or smaller sample; "low" when data is sparse.
- You MUST call emit_insights exactly once. Do not write prose outside the tool call.`

func newModelClient(cfg config.Config) ModelClient {
	if !cfg.LiveModelEnabled {
		return disabledModelClient{}
	}
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if provider == "" || provider == "disabled" {
		return disabledModelClient{}
	}
	if provider != "anthropic" {
		return disabledModelClient{}
	}
	if strings.TrimSpace(cfg.ModelAPIKey) == "" {
		return disabledModelClient{}
	}

	model := strings.TrimSpace(cfg.Model)
	if model == "" || model == "fallback-only" {
		model = string(anthropic.ModelClaudeSonnet4_6)
	}

	return anthropicModelClient{
		client: anthropic.NewClient(
			option.WithAPIKey(cfg.ModelAPIKey),
			// Retry transient failures (429/5xx) with the SDK's backoff.
			option.WithMaxRetries(3),
		),
		model: anthropic.Model(model),
	}
}

func (disabledModelClient) Generate(context.Context, string, types.InsightGenerationRequest) (*types.InsightGenerationResponse, error) {
	return nil, errors.New("model provider is not configured")
}

func (c anthropicModelClient) Generate(ctx context.Context, prompt string, _ types.InsightGenerationRequest) (*types.InsightGenerationResponse, error) {
	schema := insightResponseSchema()
	properties, _ := schema["properties"].(map[string]any)
	required, _ := schema["required"].([]string)

	tool := anthropic.ToolParam{
		Name:        insightToolName,
		Description: anthropic.String("Emit exactly 3 coaching insights and 1 focus recommendation derived from the deterministic comparison data."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: properties,
			Required:   required,
		},
	}

	resp, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     c.model,
		MaxTokens: 8192,
		System: []anthropic.TextBlockParam{{
			Text:         insightSystemPrompt,
			CacheControl: anthropic.NewCacheControlEphemeralParam(),
		}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
		Tools: []anthropic.ToolUnionParam{{OfTool: &tool}},
		Thinking: anthropic.ThinkingConfigParamUnion{
			OfAdaptive: &anthropic.ThinkingConfigAdaptiveParam{},
		},
	})
	if err != nil {
		return nil, err
	}

	for _, block := range resp.Content {
		toolUse, ok := block.AsAny().(anthropic.ToolUseBlock)
		if !ok || toolUse.Name != insightToolName {
			continue
		}
		var parsed types.InsightGenerationResponse
		if err := json.Unmarshal([]byte(toolUse.JSON.Input.Raw()), &parsed); err != nil {
			return nil, fmt.Errorf("failed to decode anthropic tool output: %w", err)
		}
		if string(resp.Model) != "" {
			parsed.Model = string(resp.Model)
		}
		return &parsed, nil
	}

	return nil, errors.New("anthropic response did not include emit_insights tool call")
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
