package service

import (
	"testing"
	"time"

	"wow-log-analyzer/services/analysis-service/types"
)

func TestCalculateCastsPerMinute(t *testing.T) {
	s := NewAnalysisService()
	start := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("zero duration returns low confidence", func(t *testing.T) {
		result := s.calculateCastsPerMinute(nil, 0)
		if result.Value != 0 {
			t.Errorf("expected 0, got %f", result.Value)
		}
		if result.Confidence != "low" {
			t.Errorf("expected low confidence, got %q", result.Confidence)
		}
	})

	t.Run("five casts in five minutes", func(t *testing.T) {
		casts := make([]types.CastEvent, 5)
		for i := range casts {
			casts[i] = types.CastEvent{
				Timestamp: start.Add(time.Duration(i) * time.Minute),
				Ability:   types.Ability{ID: 1, Name: "Spell"},
				SourceID:  1,
			}
		}
		result := s.calculateCastsPerMinute(casts, 5*time.Minute)
		if result.Value != 1.0 {
			t.Errorf("expected 1.0 casts/min, got %f", result.Value)
		}
		if result.Confidence != "medium" {
			t.Errorf("expected medium confidence (< 10 casts), got %q", result.Confidence)
		}
		if result.TotalCasts != 5 {
			t.Errorf("expected TotalCasts 5, got %d", result.TotalCasts)
		}
	})

	t.Run("ten casts in two minutes gives high confidence", func(t *testing.T) {
		casts := make([]types.CastEvent, 10)
		for i := range casts {
			casts[i] = types.CastEvent{
				Timestamp: start.Add(time.Duration(i) * 12 * time.Second),
				Ability:   types.Ability{ID: 1, Name: "Spell"},
				SourceID:  1,
			}
		}
		result := s.calculateCastsPerMinute(casts, 2*time.Minute)
		if result.Value != 5.0 {
			t.Errorf("expected 5.0 casts/min, got %f", result.Value)
		}
		if result.Confidence != "high" {
			t.Errorf("expected high confidence, got %q", result.Confidence)
		}
	})

	t.Run("zero casts with valid duration", func(t *testing.T) {
		result := s.calculateCastsPerMinute(nil, 5*time.Minute)
		if result.Value != 0 {
			t.Errorf("expected 0 casts/min, got %f", result.Value)
		}
		if result.Confidence != "medium" {
			t.Errorf("expected medium confidence, got %q", result.Confidence)
		}
	})
}

func TestCalculateMajorCDCount(t *testing.T) {
	s := NewAnalysisService()
	start := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("no cooldowns returns zero with medium confidence", func(t *testing.T) {
		result := s.calculateMajorCDCount(nil)
		if result.Value != 0 {
			t.Errorf("expected 0, got %d", result.Value)
		}
		if result.Confidence != "medium" {
			t.Errorf("expected medium confidence, got %q", result.Confidence)
		}
	})

	t.Run("only non-major cooldowns", func(t *testing.T) {
		cooldowns := []types.CooldownEvent{
			{Timestamp: start, Ability: types.Ability{ID: 1, Name: "Minor CD", IsMajorCD: false}, EventType: "start"},
		}
		result := s.calculateMajorCDCount(cooldowns)
		if result.Value != 0 {
			t.Errorf("expected 0 major CDs, got %d", result.Value)
		}
		if result.Confidence != "medium" {
			t.Errorf("expected medium confidence, got %q", result.Confidence)
		}
	})

	t.Run("end events not counted", func(t *testing.T) {
		cooldowns := []types.CooldownEvent{
			{Timestamp: start, Ability: types.Ability{ID: 1, Name: "Big CD", IsMajorCD: true}, EventType: "end"},
		}
		result := s.calculateMajorCDCount(cooldowns)
		if result.Value != 0 {
			t.Errorf("expected 0 (end events ignored), got %d", result.Value)
		}
	})

	t.Run("two major CD start events", func(t *testing.T) {
		cooldowns := []types.CooldownEvent{
			{Timestamp: start, Ability: types.Ability{ID: 1, Name: "Big CD", IsMajorCD: true}, EventType: "start"},
			{Timestamp: start.Add(3 * time.Minute), Ability: types.Ability{ID: 1, Name: "Big CD", IsMajorCD: true}, EventType: "start"},
		}
		result := s.calculateMajorCDCount(cooldowns)
		if result.Value != 2 {
			t.Errorf("expected 2 major CDs, got %d", result.Value)
		}
		if result.Confidence != "high" {
			t.Errorf("expected high confidence, got %q", result.Confidence)
		}
	})
}

func TestCalculateMajorCDDrift(t *testing.T) {
	s := NewAnalysisService()
	start := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	end := start.Add(10 * time.Minute)

	t.Run("no cooldowns returns low confidence", func(t *testing.T) {
		result := s.calculateMajorCDDrift(nil, start, end)
		if result.Value != 0 {
			t.Errorf("expected 0 drift, got %f", result.Value)
		}
		if result.Confidence != "low" {
			t.Errorf("expected low confidence, got %q", result.Confidence)
		}
	})

	t.Run("single major CD has no pairs", func(t *testing.T) {
		cooldowns := []types.CooldownEvent{
			{Timestamp: start, Ability: types.Ability{ID: 1, IsMajorCD: true}, EventType: "start"},
		}
		result := s.calculateMajorCDDrift(cooldowns, start, end)
		if result.Confidence != "low" {
			t.Errorf("expected low confidence for no pairs, got %q", result.Confidence)
		}
	})

	t.Run("two CDs at exact 180s interval gives zero drift", func(t *testing.T) {
		cooldowns := []types.CooldownEvent{
			{Timestamp: start, Ability: types.Ability{ID: 1, IsMajorCD: true}, EventType: "start"},
			{Timestamp: start.Add(180 * time.Second), Ability: types.Ability{ID: 1, IsMajorCD: true}, EventType: "start"},
		}
		result := s.calculateMajorCDDrift(cooldowns, start, end)
		if result.Value != 0 {
			t.Errorf("expected 0 drift at exact interval, got %f", result.Value)
		}
		if result.Confidence != "medium" {
			t.Errorf("expected medium confidence (1 pair < 3), got %q", result.Confidence)
		}
	})

	t.Run("two CDs 60 seconds late gives 60s drift", func(t *testing.T) {
		cooldowns := []types.CooldownEvent{
			{Timestamp: start, Ability: types.Ability{ID: 1, IsMajorCD: true}, EventType: "start"},
			{Timestamp: start.Add(240 * time.Second), Ability: types.Ability{ID: 1, IsMajorCD: true}, EventType: "start"},
		}
		result := s.calculateMajorCDDrift(cooldowns, start, end)
		if result.Value != 60 {
			t.Errorf("expected 60s drift, got %f", result.Value)
		}
	})

	t.Run("three or more pairs gives high confidence", func(t *testing.T) {
		cooldowns := []types.CooldownEvent{
			{Timestamp: start, Ability: types.Ability{ID: 1, IsMajorCD: true}, EventType: "start"},
			{Timestamp: start.Add(180 * time.Second), Ability: types.Ability{ID: 1, IsMajorCD: true}, EventType: "start"},
			{Timestamp: start.Add(360 * time.Second), Ability: types.Ability{ID: 1, IsMajorCD: true}, EventType: "start"},
			{Timestamp: start.Add(540 * time.Second), Ability: types.Ability{ID: 1, IsMajorCD: true}, EventType: "start"},
		}
		result := s.calculateMajorCDDrift(cooldowns, start, end)
		if result.Confidence != "high" {
			t.Errorf("expected high confidence for 3 pairs, got %q", result.Confidence)
		}
	})
}

func TestCalculateBuffUptime(t *testing.T) {
	s := NewAnalysisService()
	start := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("no buff events returns zero with low confidence", func(t *testing.T) {
		data := types.PlayerFightData{
			PlayerID:   1,
			FightID:    1,
			FightStart: start,
			FightEnd:   start.Add(5 * time.Minute),
		}
		result := s.calculateBuffUptime(data)
		if result.Value != 0 {
			t.Errorf("expected 0 uptime, got %f", result.Value)
		}
		if result.Confidence != "low" {
			t.Errorf("expected low confidence, got %q", result.Confidence)
		}
	})

	t.Run("single buff fully active gives 100 percent uptime", func(t *testing.T) {
		data := types.PlayerFightData{
			PlayerID:   1,
			FightID:    1,
			FightStart: start,
			FightEnd:   start.Add(5 * time.Minute),
			BuffEvents: []types.BuffEvent{
				{Timestamp: start, Ability: types.Ability{ID: 1, Name: "Power Word: Fortitude", IsBuff: true}, SourceID: 1, TargetID: 1, EventType: "apply"},
				{Timestamp: start.Add(5 * time.Minute), Ability: types.Ability{ID: 1, Name: "Power Word: Fortitude", IsBuff: true}, SourceID: 1, TargetID: 1, EventType: "remove"},
			},
		}
		result := s.calculateBuffUptime(data)
		if result.Value < 99 || result.Value > 100 {
			t.Errorf("expected ~100 percent uptime, got %f", result.Value)
		}
		if result.Confidence != "medium" {
			t.Errorf("expected medium confidence (1 buff < 3), got %q", result.Confidence)
		}
	})

	t.Run("three distinct buffs gives high confidence", func(t *testing.T) {
		data := types.PlayerFightData{
			PlayerID:   1,
			FightID:    1,
			FightStart: start,
			FightEnd:   start.Add(5 * time.Minute),
			BuffEvents: []types.BuffEvent{
				{Timestamp: start, Ability: types.Ability{ID: 1, Name: "Buff A", IsBuff: true}, SourceID: 1, TargetID: 1, EventType: "apply"},
				{Timestamp: start.Add(5 * time.Minute), Ability: types.Ability{ID: 1, Name: "Buff A", IsBuff: true}, SourceID: 1, TargetID: 1, EventType: "remove"},
				{Timestamp: start, Ability: types.Ability{ID: 2, Name: "Buff B", IsBuff: true}, SourceID: 1, TargetID: 1, EventType: "apply"},
				{Timestamp: start.Add(5 * time.Minute), Ability: types.Ability{ID: 2, Name: "Buff B", IsBuff: true}, SourceID: 1, TargetID: 1, EventType: "remove"},
				{Timestamp: start, Ability: types.Ability{ID: 3, Name: "Buff C", IsBuff: true}, SourceID: 1, TargetID: 1, EventType: "apply"},
				{Timestamp: start.Add(5 * time.Minute), Ability: types.Ability{ID: 3, Name: "Buff C", IsBuff: true}, SourceID: 1, TargetID: 1, EventType: "remove"},
			},
		}
		result := s.calculateBuffUptime(data)
		if result.Confidence != "high" {
			t.Errorf("expected high confidence for 3 buffs, got %q", result.Confidence)
		}
	})
}

func TestCalculateDowntimePercentage(t *testing.T) {
	s := NewAnalysisService()
	start := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("no events returns 100 percent downtime", func(t *testing.T) {
		result := s.calculateDowntimePercentage(nil, nil, 5*time.Minute)
		if result.Value != 100 {
			t.Errorf("expected 100 percent downtime, got %f", result.Value)
		}
		if result.Confidence != "low" {
			t.Errorf("expected low confidence, got %q", result.Confidence)
		}
	})

	t.Run("few events gives low confidence", func(t *testing.T) {
		damage := []types.DamageEvent{
			{Timestamp: start, Ability: types.Ability{ID: 1}, SourceID: 1, TargetID: 2, Amount: 100},
		}
		result := s.calculateDowntimePercentage(damage, nil, 5*time.Minute)
		if result.Confidence != "low" {
			t.Errorf("expected low confidence for 1 event, got %q", result.Confidence)
		}
	})

	t.Run("fifty or more events gives medium confidence", func(t *testing.T) {
		damage := make([]types.DamageEvent, 50)
		for i := range damage {
			damage[i] = types.DamageEvent{
				Timestamp: start.Add(time.Duration(i) * 3 * time.Second),
				Ability:   types.Ability{ID: 1},
				SourceID:  1,
				TargetID:  2,
				Amount:    100,
			}
		}
		result := s.calculateDowntimePercentage(damage, nil, 10*time.Minute)
		if result.Confidence != "medium" {
			t.Errorf("expected medium confidence for 50 events, got %q", result.Confidence)
		}
		if result.Value < 0 || result.Value > 100 {
			t.Errorf("downtime percentage out of range: %f", result.Value)
		}
	})

	t.Run("heal events count toward active time", func(t *testing.T) {
		heals := []types.HealEvent{
			{Timestamp: start, Ability: types.Ability{ID: 1}, SourceID: 1, TargetID: 1, Amount: 500},
		}
		resultWithHeal := s.calculateDowntimePercentage(nil, heals, 5*time.Minute)
		resultWithout := s.calculateDowntimePercentage(nil, nil, 5*time.Minute)
		if resultWithHeal.Value >= resultWithout.Value {
			t.Errorf("heal events should reduce downtime: withHeal=%f withoutHeal=%f", resultWithHeal.Value, resultWithout.Value)
		}
	})
}
