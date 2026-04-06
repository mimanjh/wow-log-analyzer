package service

import (
	"context"
	"errors"

	"wow-log-analyzer/services/ai-service/internal/config"
	"wow-log-analyzer/services/ai-service/internal/types"
)

type disabledModelClient struct{}

func newModelClient(cfg config.Config) ModelClient {
	if cfg.Provider == "" || cfg.Provider == "disabled" {
		return disabledModelClient{}
	}

	return disabledModelClient{}
}

func (disabledModelClient) Generate(context.Context, string, types.InsightGenerationRequest) (*types.InsightGenerationResponse, error) {
	return nil, errors.New("model provider is not configured")
}
