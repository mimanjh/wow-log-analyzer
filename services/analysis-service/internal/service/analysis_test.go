package service

import (
	"testing"
	"time"

	"wow-log-analyzer/services/analysis-service/internal/types"
)

func TestAnalysisService_AnalyzePlayerFight(t *testing.T) {
	service := NewAnalysisService()

	// Create test data
	playerData := types.PlayerFightData{
		PlayerID:   1,
		FightID:    100,
		FightStart: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		FightEnd:   time.Date(2024, 1, 1, 12, 5, 0, 0, time.UTC), // 5 minutes
		CastEvents: []types.CastEvent{
			{Timestamp: time.Date(2024, 1, 1, 12, 0, 30, 0, time.UTC), Ability: types.Ability{ID: 1, Name: "Spell 1"}, SourceID: 1},
			{Timestamp: time.Date(2024, 1, 1, 12, 1, 0, 0, time.UTC), Ability: types.Ability{ID: 2, Name: "Spell 2"}, SourceID: 1},
			{Timestamp: time.Date(2024, 1, 1, 12, 2, 0, 0, time.UTC), Ability: types.Ability{ID: 3, Name: "Spell 3"}, SourceID: 1},
			{Timestamp: time.Date(2024, 1, 1, 12, 3, 0, 0, time.UTC), Ability: types.Ability{ID: 4, Name: "Spell 4"}, SourceID: 1},
			{Timestamp: time.Date(2024, 1, 1, 12, 4, 0, 0, time.UTC), Ability: types.Ability{ID: 5, Name: "Spell 5"}, SourceID: 1},
		},
		DamageEvents: []types.DamageEvent{
			{Timestamp: time.Date(2024, 1, 1, 12, 0, 35, 0, time.UTC), Ability: types.Ability{ID: 1}, SourceID: 1, TargetID: 2, Amount: 1000},
			{Timestamp: time.Date(2024, 1, 1, 12, 1, 5, 0, time.UTC), Ability: types.Ability{ID: 2}, SourceID: 1, TargetID: 2, Amount: 1200},
		},
		HealEvents: []types.HealEvent{
			{Timestamp: time.Date(2024, 1, 1, 12, 2, 5, 0, time.UTC), Ability: types.Ability{ID: 10}, SourceID: 1, TargetID: 1, Amount: 500},
		},
		BuffEvents: []types.BuffEvent{
			{Timestamp: time.Date(2024, 1, 1, 12, 0, 10, 0, time.UTC), Ability: types.Ability{ID: 20, Name: "Buff", IsBuff: true}, SourceID: 1, TargetID: 1, EventType: "apply"},
			{Timestamp: time.Date(2024, 1, 1, 12, 4, 50, 0, time.UTC), Ability: types.Ability{ID: 20, Name: "Buff", IsBuff: true}, SourceID: 1, TargetID: 1, EventType: "remove"},
		},
		CooldownEvents: []types.CooldownEvent{
			{Timestamp: time.Date(2024, 1, 1, 12, 0, 15, 0, time.UTC), Ability: types.Ability{ID: 30, Name: "Major CD", IsMajorCD: true}, SourceID: 1, EventType: "start"},
			{Timestamp: time.Date(2024, 1, 1, 12, 3, 20, 0, time.UTC), Ability: types.Ability{ID: 31, Name: "Major CD 2", IsMajorCD: true}, SourceID: 1, EventType: "start"},
		},
	}

	metrics, err := service.AnalyzePlayerFight(playerData)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test casts per minute (5 casts in 5 minutes = 1 cast per minute)
	expectedCastsPerMin := 1.0
	if metrics.CastsPerMin.Value != expectedCastsPerMin {
		t.Errorf("Expected casts per minute %f, got %f", expectedCastsPerMin, metrics.CastsPerMin.Value)
	}

	// Test major CD count (2 major CDs)
	expectedMajorCDCount := 2
	if metrics.MajorCDCount.Value != expectedMajorCDCount {
		t.Errorf("Expected major CD count %d, got %d", expectedMajorCDCount, metrics.MajorCDCount.Value)
	}

	// Test buff uptime (4 minutes 40 seconds out of 5 minutes = 93.33%)
	expectedBuffUptime := 93.33
	if metrics.BuffUptime.Value < expectedBuffUptime-1 || metrics.BuffUptime.Value > expectedBuffUptime+1 {
		t.Errorf("Expected buff uptime around %f, got %f", expectedBuffUptime, metrics.BuffUptime.Value)
	}

	// Test downtime percentage (3 events * 5 seconds = 15 seconds active out of 300 = 95% downtime)
	expectedDowntimePct := 95.0
	if metrics.DowntimePct.Value < expectedDowntimePct-1 || metrics.DowntimePct.Value > expectedDowntimePct+1 {
		t.Errorf("Expected downtime percentage around %f, got %f", expectedDowntimePct, metrics.DowntimePct.Value)
	}
}

func TestAnalysisService_CompareAgainstCohort(t *testing.T) {
	service := NewAnalysisService()

	// Create player data
	playerData := types.PlayerFightData{
		PlayerID:   1,
		FightID:    100,
		FightStart: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		FightEnd:   time.Date(2024, 1, 1, 12, 5, 0, 0, time.UTC),
		CastEvents: []types.CastEvent{
			{Timestamp: time.Date(2024, 1, 1, 12, 0, 30, 0, time.UTC), Ability: types.Ability{ID: 1}, SourceID: 1},
			{Timestamp: time.Date(2024, 1, 1, 12, 1, 0, 0, time.UTC), Ability: types.Ability{ID: 2}, SourceID: 1},
		},
	}

	// Create cohort data (3 players with different performance)
	cohortData := []types.PlayerFightData{
		{
			PlayerID:   2,
			FightID:    100,
			FightStart: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			FightEnd:   time.Date(2024, 1, 1, 12, 5, 0, 0, time.UTC),
			CastEvents: []types.CastEvent{
				{Timestamp: time.Date(2024, 1, 1, 12, 0, 30, 0, time.UTC), Ability: types.Ability{ID: 1}, SourceID: 2},
			},
		},
		{
			PlayerID:   3,
			FightID:    100,
			FightStart: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			FightEnd:   time.Date(2024, 1, 1, 12, 5, 0, 0, time.UTC),
			CastEvents: []types.CastEvent{
				{Timestamp: time.Date(2024, 1, 1, 12, 0, 30, 0, time.UTC), Ability: types.Ability{ID: 1}, SourceID: 3},
				{Timestamp: time.Date(2024, 1, 1, 12, 1, 0, 0, time.UTC), Ability: types.Ability{ID: 2}, SourceID: 3},
				{Timestamp: time.Date(2024, 1, 1, 12, 2, 0, 0, time.UTC), Ability: types.Ability{ID: 3}, SourceID: 3},
			},
		},
		{
			PlayerID:   4,
			FightID:    100,
			FightStart: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			FightEnd:   time.Date(2024, 1, 1, 12, 5, 0, 0, time.UTC),
			CastEvents: []types.CastEvent{
				{Timestamp: time.Date(2024, 1, 1, 12, 0, 30, 0, time.UTC), Ability: types.Ability{ID: 1}, SourceID: 4},
				{Timestamp: time.Date(2024, 1, 1, 12, 1, 0, 0, time.UTC), Ability: types.Ability{ID: 2}, SourceID: 4},
				{Timestamp: time.Date(2024, 1, 1, 12, 2, 0, 0, time.UTC), Ability: types.Ability{ID: 3}, SourceID: 4},
				{Timestamp: time.Date(2024, 1, 1, 12, 3, 0, 0, time.UTC), Ability: types.Ability{ID: 4}, SourceID: 4},
			},
		},
	}

	playerMetrics, err := service.AnalyzePlayerFight(playerData)
	if err != nil {
		t.Fatalf("Failed to analyze player: %v", err)
	}

	cohortMetrics := make([]types.PlayerFightMetrics, len(cohortData))
	for i, data := range cohortData {
		metrics, err := service.AnalyzePlayerFight(data)
		if err != nil {
			t.Fatalf("Failed to analyze cohort player %d: %v", i, err)
		}
		cohortMetrics[i] = *metrics
	}

	comparison, err := service.CompareAgainstCohort(*playerMetrics, cohortMetrics)
	if err != nil {
		t.Fatalf("Failed to compare against cohort: %v", err)
	}

	// Player has 0.4 casts/min, cohort median should be around 0.4-0.8
	if comparison.CohortStats.SampleSize != 3 {
		t.Errorf("Expected cohort size 3, got %d", comparison.CohortStats.SampleSize)
	}

	// Player should be around 50th percentile for casts per minute
	if comparison.Deltas.CastsPerMin.Percentile < 30 || comparison.Deltas.CastsPerMin.Percentile > 70 {
		t.Errorf("Expected player percentile around 50, got %f", comparison.Deltas.CastsPerMin.Percentile)
	}
}

func TestAnalysisService_UsageComparisons(t *testing.T) {
	service := NewAnalysisService()

	playerData := types.PlayerFightData{
		PlayerID:   1,
		FightID:    100,
		FightStart: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		FightEnd:   time.Date(2024, 1, 1, 12, 5, 0, 0, time.UTC),
		CastEvents: []types.CastEvent{
			{Timestamp: time.Date(2024, 1, 1, 12, 0, 10, 0, time.UTC), Ability: types.Ability{ID: 1, Name: "Shadowstrike"}, SourceID: 1},
			{Timestamp: time.Date(2024, 1, 1, 12, 1, 10, 0, time.UTC), Ability: types.Ability{ID: 1, Name: "Shadowstrike"}, SourceID: 1},
		},
		BuffEvents: []types.BuffEvent{
			{Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), Ability: types.Ability{ID: 2, Name: "Slice and Dice", IsBuff: true}, SourceID: 1, TargetID: 1, EventType: "apply"},
			{Timestamp: time.Date(2024, 1, 1, 12, 4, 0, 0, time.UTC), Ability: types.Ability{ID: 2, Name: "Slice and Dice", IsBuff: true}, SourceID: 1, TargetID: 1, EventType: "remove"},
		},
	}

	cohortData := []types.PlayerFightData{
		{
			PlayerID:   2,
			FightID:    100,
			FightStart: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			FightEnd:   time.Date(2024, 1, 1, 12, 5, 0, 0, time.UTC),
			CastEvents: []types.CastEvent{
				{Timestamp: time.Date(2024, 1, 1, 12, 0, 5, 0, time.UTC), Ability: types.Ability{ID: 1, Name: "Shadowstrike"}, SourceID: 2},
				{Timestamp: time.Date(2024, 1, 1, 12, 0, 45, 0, time.UTC), Ability: types.Ability{ID: 1, Name: "Shadowstrike"}, SourceID: 2},
				{Timestamp: time.Date(2024, 1, 1, 12, 1, 25, 0, time.UTC), Ability: types.Ability{ID: 1, Name: "Shadowstrike"}, SourceID: 2},
			},
			BuffEvents: []types.BuffEvent{
				{Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), Ability: types.Ability{ID: 2, Name: "Slice and Dice", IsBuff: true}, SourceID: 2, TargetID: 2, EventType: "apply"},
				{Timestamp: time.Date(2024, 1, 1, 12, 5, 0, 0, time.UTC), Ability: types.Ability{ID: 2, Name: "Slice and Dice", IsBuff: true}, SourceID: 2, TargetID: 2, EventType: "remove"},
			},
		},
	}

	abilityComparisons := service.CalculateAbilityUsageComparisons(playerData, cohortData)
	if len(abilityComparisons) == 0 {
		t.Fatalf("expected ability comparisons to be populated")
	}
	if abilityComparisons[0].AbilityName != "Shadowstrike" {
		t.Fatalf("expected Shadowstrike comparison, got %s", abilityComparisons[0].AbilityName)
	}
	if abilityComparisons[0].PlayerCount != 2 {
		t.Fatalf("expected player cast count 2, got %d", abilityComparisons[0].PlayerCount)
	}

	buffComparisons := service.CalculateBuffUptimeComparisons(playerData, cohortData)
	if len(buffComparisons) == 0 {
		t.Fatalf("expected buff comparisons to be populated")
	}
	if buffComparisons[0].AbilityName != "Slice and Dice" {
		t.Fatalf("expected Slice and Dice comparison, got %s", buffComparisons[0].AbilityName)
	}
	if buffComparisons[0].PlayerUptimePct <= 0 {
		t.Fatalf("expected positive player uptime, got %f", buffComparisons[0].PlayerUptimePct)
	}
}

func TestAnalysisService_BuffUptimeComparisonClampsInvalidTimestamps(t *testing.T) {
	service := NewAnalysisService()

	playerData := types.PlayerFightData{
		PlayerID:   1,
		FightID:    100,
		FightStart: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		FightEnd:   time.Date(2024, 1, 1, 12, 3, 0, 0, time.UTC),
		BuffEvents: []types.BuffEvent{
			{Timestamp: time.Time{}, Ability: types.Ability{ID: 2, Name: "Slice and Dice", IsBuff: true}, SourceID: 1, TargetID: 1, EventType: "apply"},
			{Timestamp: time.Date(2024, 1, 1, 12, 2, 0, 0, time.UTC), Ability: types.Ability{ID: 2, Name: "Slice and Dice", IsBuff: true}, SourceID: 1, TargetID: 1, EventType: "remove"},
		},
	}

	cohortData := []types.PlayerFightData{
		{
			PlayerID:   2,
			FightID:    100,
			FightStart: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			FightEnd:   time.Date(2024, 1, 1, 12, 3, 0, 0, time.UTC),
			BuffEvents: []types.BuffEvent{
				{Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), Ability: types.Ability{ID: 2, Name: "Slice and Dice", IsBuff: true}, SourceID: 2, TargetID: 2, EventType: "apply"},
				{Timestamp: time.Date(2024, 1, 1, 12, 3, 0, 0, time.UTC), Ability: types.Ability{ID: 2, Name: "Slice and Dice", IsBuff: true}, SourceID: 2, TargetID: 2, EventType: "remove"},
			},
		},
	}

	buffComparisons := service.CalculateBuffUptimeComparisons(playerData, cohortData)
	if len(buffComparisons) == 0 {
		t.Fatalf("expected buff comparisons to be populated")
	}
	if buffComparisons[0].PlayerUptimePct > 100 {
		t.Fatalf("expected clamped uptime <= 100, got %f", buffComparisons[0].PlayerUptimePct)
	}
}

func TestAnalysisService_ResourceUsageUsesAverageCapAndSpent(t *testing.T) {
	service := NewAnalysisService()
	start := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	playerData := types.PlayerFightData{
		PlayerID:   1,
		FightID:    100,
		FightStart: start,
		FightEnd:   start.Add(60 * time.Second),
		ResourceEvents: []types.ResourceEvent{
			{Timestamp: start, ResourceTypeID: 3, ResourceType: "Energy", Amount: 100, MaxAmount: 100},
			{Timestamp: start.Add(30 * time.Second), ResourceTypeID: 3, ResourceType: "Energy", Amount: 50, MaxAmount: 100, Change: -50},
		},
	}
	cohortData := []types.PlayerFightData{
		{
			PlayerID:   2,
			FightID:    100,
			FightStart: start,
			FightEnd:   start.Add(60 * time.Second),
			ResourceEvents: []types.ResourceEvent{
				{Timestamp: start, ResourceTypeID: 3, ResourceType: "Energy", Amount: 50, MaxAmount: 100},
				{Timestamp: start.Add(30 * time.Second), ResourceTypeID: 3, ResourceType: "Energy", Amount: 25, MaxAmount: 100, Change: -75},
			},
		},
	}

	comparisons := service.CalculateResourceUsageComparisons(playerData, cohortData, "Rogue", "Outlaw")
	if len(comparisons) != 1 {
		t.Fatalf("expected one resource comparison, got %d", len(comparisons))
	}

	got := comparisons[0]
	if got.PlayerAveragePct != 75 {
		t.Fatalf("expected player average pct 75, got %f", got.PlayerAveragePct)
	}
	if got.PlayerTimeAtMaxSeconds != 30 {
		t.Fatalf("expected player time at max 30s, got %f", got.PlayerTimeAtMaxSeconds)
	}
	if got.PlayerSpent != 50 {
		t.Fatalf("expected player spent 50, got %f", got.PlayerSpent)
	}
	if got.CohortMedianSpent != 75 {
		t.Fatalf("expected cohort spent 75, got %f", got.CohortMedianSpent)
	}
}
