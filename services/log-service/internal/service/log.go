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

func (s *LogService) GetComparisonData(reportID string, fight types.FightSelection, characterID int) (*types.ComparisonDataResponse, error) {
	return s.wclClient.GetComparisonData(reportID, fight, characterID)
}

func (s *LogService) GetPlayerFightData(reportID string, fight types.FightSelection, characterID int) (*types.PlayerFightData, error) {
	return s.wclClient.GetPlayerFightData(reportID, fight, characterID)
}

func (s *LogService) GetRankingCandidates(fight types.FightSelection, characterClass, characterSpec string, limit int) ([]types.RankingCandidate, error) {
	return s.wclClient.GetRankingCandidates(fight, characterClass, characterSpec, limit)
}

func (s *LogService) GetCohortMemberData(candidate types.RankingCandidate) (*types.PlayerFightData, error) {
	return s.wclClient.GetCohortMemberData(candidate)
}
