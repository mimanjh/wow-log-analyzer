package service

import (
	"fmt"
	"time"

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

func (s *LogService) GetComparisonData(reportID string, fightID int, characterID int) (*types.ComparisonDataResponse, error) {
	fights, err := s.GetFights(reportID)
	if err != nil {
		return nil, err
	}

	var selectedFight *types.NormalizedFight
	for _, fight := range fights {
		if fight.ID == fightID {
			fightCopy := fight
			selectedFight = &fightCopy
			break
		}
	}

	if selectedFight == nil {
		return nil, fmt.Errorf("fight %d not found in report", fightID)
	}

	playerData := buildPlayerFightData(*selectedFight, characterID, 0)
	cohortData := []types.PlayerFightData{
		buildPlayerFightData(*selectedFight, characterID+100, 1),
		buildPlayerFightData(*selectedFight, characterID+200, 2),
		buildPlayerFightData(*selectedFight, characterID+300, 3),
		buildPlayerFightData(*selectedFight, characterID+400, 4),
	}

	return &types.ComparisonDataResponse{
		ReportID: reportID,
		Fight: types.FightSummary{
			ID:          selectedFight.ID,
			Name:        selectedFight.Name,
			Difficulty:  selectedFight.Difficulty,
			KillTime:    int(selectedFight.EndTime.Sub(selectedFight.StartTime).Seconds()),
			EncounterID: selectedFight.EncounterID,
		},
		PlayerData: playerData,
		CohortData: cohortData,
	}, nil
}

func buildPlayerFightData(fight types.NormalizedFight, playerID int, profile int) types.PlayerFightData {
	start := fight.StartTime
	duration := fight.EndTime.Sub(fight.StartTime)
	if duration <= 0 {
		duration = 5 * time.Minute
	}
	end := start.Add(duration)

	casts := 12 + profile*2
	damageEvents := 48 + profile*5
	healEvents := 6 + profile
	buffDuration := duration - time.Duration((40-profile*5))*time.Second
	if buffDuration < 2*time.Minute {
		buffDuration = 2 * time.Minute
	}

	castEvents := make([]types.CastEvent, 0, casts)
	for i := 0; i < casts; i++ {
		castEvents = append(castEvents, types.CastEvent{
			Timestamp: start.Add(time.Duration((i + 1) * int(duration/time.Duration(casts+1)))),
			Ability: types.Ability{
				ID:   1000 + i%4,
				Name: fmt.Sprintf("Spell %d", i%4+1),
			},
			SourceID: playerID,
		})
	}

	damage := make([]types.DamageEvent, 0, damageEvents)
	for i := 0; i < damageEvents; i++ {
		damage = append(damage, types.DamageEvent{
			Timestamp: start.Add(time.Duration((i + 1) * int(duration/time.Duration(damageEvents+2)))),
			Ability:   types.Ability{ID: 2000 + i%3, Name: "Damage"},
			SourceID:  playerID,
			TargetID:  9000 + i%2,
			Amount:    1000 + profile*50 + i*7,
		})
	}

	heals := make([]types.HealEvent, 0, healEvents)
	for i := 0; i < healEvents; i++ {
		heals = append(heals, types.HealEvent{
			Timestamp: start.Add(time.Duration((i + 2) * int(duration/time.Duration(healEvents+4)))),
			Ability:   types.Ability{ID: 3000 + i%2, Name: "Heal"},
			SourceID:  playerID,
			TargetID:  playerID,
			Amount:    400 + profile*30 + i*15,
		})
	}

	cooldownBase := 180 + profile*8
	cooldowns := []types.CooldownEvent{
		{
			Timestamp: start.Add(15 * time.Second),
			Ability:   types.Ability{ID: 4001, Name: "Major Cooldown", IsMajorCD: true},
			SourceID:  playerID,
			EventType: "start",
		},
		{
			Timestamp: start.Add(time.Duration(cooldownBase) * time.Second),
			Ability:   types.Ability{ID: 4002, Name: "Major Cooldown", IsMajorCD: true},
			SourceID:  playerID,
			EventType: "start",
		},
	}
	if duration > 5*time.Minute {
		cooldowns = append(cooldowns, types.CooldownEvent{
			Timestamp: start.Add(time.Duration(cooldownBase*2) * time.Second),
			Ability:   types.Ability{ID: 4003, Name: "Major Cooldown", IsMajorCD: true},
			SourceID:  playerID,
			EventType: "start",
		})
	}

	buffs := []types.BuffEvent{
		{
			Timestamp: start.Add(10 * time.Second),
			Ability:   types.Ability{ID: 5001, Name: "Primary Buff", IsBuff: true},
			SourceID:  playerID,
			TargetID:  playerID,
			EventType: "apply",
		},
		{
			Timestamp: start.Add(10 * time.Second).Add(buffDuration),
			Ability:   types.Ability{ID: 5001, Name: "Primary Buff", IsBuff: true},
			SourceID:  playerID,
			TargetID:  playerID,
			EventType: "remove",
		},
	}

	return types.PlayerFightData{
		PlayerID:       playerID,
		FightID:        fight.ID,
		FightStart:     start,
		FightEnd:       end,
		CastEvents:     castEvents,
		DamageEvents:   damage,
		HealEvents:     heals,
		BuffEvents:     buffs,
		CooldownEvents: cooldowns,
	}
}
