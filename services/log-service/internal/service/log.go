package service

import (
	"context"

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
func (s *LogService) GetReportMetadata(ctx context.Context, reportID string) (*types.NormalizedReport, error) {
	return s.wclClient.GetReportMetadata(ctx, reportID)
}

// GetFights retrieves fights for a given report ID
func (s *LogService) GetFights(ctx context.Context, reportID string) ([]types.NormalizedFight, error) {
	return s.wclClient.GetFights(ctx, reportID)
}

func (s *LogService) GetCharacters(ctx context.Context, reportID string, fightID int) ([]types.CharacterOption, error) {
	return s.wclClient.GetCharacters(ctx, reportID, fightID)
}

func (s *LogService) GetComparisonData(ctx context.Context, reportID string, fight types.FightSelection, characterID int) (*types.ComparisonDataResponse, error) {
	return s.wclClient.GetComparisonData(ctx, reportID, fight, characterID)
}

func (s *LogService) GetPlayerFightData(ctx context.Context, reportID string, fight types.FightSelection, characterID int) (*types.PlayerFightData, error) {
	return s.wclClient.GetPlayerFightData(ctx, reportID, fight, characterID)
}

func (s *LogService) GetRankingCandidates(ctx context.Context, fight types.FightSelection, characterClass, characterSpec string, limit int) ([]types.RankingCandidate, error) {
	return s.wclClient.GetRankingCandidates(ctx, fight, characterClass, characterSpec, limit)
}

func (s *LogService) GetCohortMemberData(ctx context.Context, candidate types.RankingCandidate) (*types.PlayerFightData, error) {
	return s.wclClient.GetCohortMemberData(ctx, candidate)
}

func (s *LogService) GetCurrentUser(ctx context.Context, accessToken string) (*types.UserProfile, error) {
	return s.wclClient.GetCurrentUser(ctx, accessToken)
}

func (s *LogService) GetOwnedCharacters(ctx context.Context, accessToken string) ([]types.OwnedCharacter, error) {
	return s.wclClient.GetOwnedCharacters(ctx, accessToken)
}

func (s *LogService) GetCharacterReports(ctx context.Context, accessToken string, characterID int, cursor string, limit int) (*types.CharacterReportsPage, error) {
	return s.wclClient.GetCharacterReports(ctx, accessToken, characterID, cursor, limit)
}
