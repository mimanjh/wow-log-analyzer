package service

import (
	"wow-log-analyzer/services/log-service/internal/client"
	"wow-log-analyzer/services/log-service/internal/config"
	"wow-log-analyzer/services/log-service/internal/types"
)

// LogService provides business logic for log operations
type LogService struct {
	wclClient client.WCLClient
}

// NewLogService creates a new log service
func NewLogService(cfg config.WCLConfig) *LogService {
	return &LogService{
		wclClient: client.NewWCLClient(cfg),
	}
}

// GetReportMetadata retrieves report metadata for a given report ID
func (s *LogService) GetReportMetadata(reportID string) (*types.NormalizedReport, error) {
	return s.wclClient.GetReportMetadata(reportID)
}

// GetFights retrieves fights for a given report ID
func (s *LogService) GetFights(reportID string) ([]types.NormalizedFight, error) {
	return s.wclClient.GetFights(reportID)
}

func (s *LogService) GetCharacters(reportID string, fightID int) ([]types.CharacterOption, error) {
	return s.wclClient.GetCharacters(reportID, fightID)
}

func (s *LogService) GetComparisonData(reportID string, fightID int, characterID int) (*types.ComparisonDataResponse, error) {
	return s.wclClient.GetComparisonData(reportID, fightID, characterID)
}
