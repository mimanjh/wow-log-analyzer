package service

import (
	"math"
	"sort"
	"time"

	"wow-log-analyzer/services/analysis-service/internal/types"
)

type abilityUsageSummary struct {
	AbilityID        int
	AbilityName      string
	Count            int
	CastsPerMinute   float64
	FirstUseSeconds  float64
	HasFirstUse      bool
}

type buffUptimeSummary struct {
	AbilityID          int
	AbilityName        string
	UptimePct          float64
	FirstApplySeconds  float64
	HasFirstApply      bool
}

func (s *AnalysisService) CalculateAbilityUsageComparisons(playerData types.PlayerFightData, cohortData []types.PlayerFightData) []types.AbilityUsageComparison {
	playerSummary := summarizeAbilityUsage(playerData)
	cohortSummaries := make([]map[int]abilityUsageSummary, len(cohortData))
	abilityNames := make(map[int]string)

	for id, summary := range playerSummary {
		abilityNames[id] = summary.AbilityName
	}

	for i, entry := range cohortData {
		cohortSummaries[i] = summarizeAbilityUsage(entry)
		for id, summary := range cohortSummaries[i] {
			if _, exists := abilityNames[id]; !exists {
				abilityNames[id] = summary.AbilityName
			}
		}
	}

	comparisons := make([]types.AbilityUsageComparison, 0, len(abilityNames))
	for abilityID, abilityName := range abilityNames {
		counts := make([]float64, 0, len(cohortSummaries))
		perMinute := make([]float64, 0, len(cohortSummaries))
		firstUses := make([]float64, 0, len(cohortSummaries))

		for _, summaryMap := range cohortSummaries {
			summary, ok := summaryMap[abilityID]
			if !ok {
				counts = append(counts, 0)
				perMinute = append(perMinute, 0)
				continue
			}
			counts = append(counts, float64(summary.Count))
			perMinute = append(perMinute, summary.CastsPerMinute)
			if summary.HasFirstUse {
				firstUses = append(firstUses, summary.FirstUseSeconds)
			}
		}

		player := playerSummary[abilityID]
		cohortMedianCount := calculateMedianCopy(counts)
		cohortMedianPerMinute := calculateMedianCopy(perMinute)
		cohortMedianFirstUse := calculateMedianCopy(firstUses)
		percentile := s.calculatePercentileFromValues(player.CastsPerMinute, perMinute)

		confidence := "high"
		var caution string
		if len(counts) < 5 {
			confidence = "medium"
			caution = "Small cohort sample for this ability"
		}
		if player.Count == 0 && cohortMedianCount == 0 {
			confidence = "low"
			caution = "This ability was not commonly observed in the cohort"
		}

		comparisons = append(comparisons, types.AbilityUsageComparison{
			AbilityID:               abilityID,
			AbilityName:             abilityName,
			PlayerCount:             player.Count,
			PlayerCastsPerMinute:    round2(player.CastsPerMinute),
			PlayerFirstUseSeconds:   round2(player.FirstUseSeconds),
			CohortMedianCount:       round2(cohortMedianCount),
			CohortMedianPerMinute:   round2(cohortMedianPerMinute),
			CohortMedianFirstUseSec: round2(cohortMedianFirstUse),
			CountDelta:              round2(float64(player.Count) - cohortMedianCount),
			PerMinuteDelta:          round2(player.CastsPerMinute - cohortMedianPerMinute),
			FirstUseDeltaSeconds:    round2(player.FirstUseSeconds - cohortMedianFirstUse),
			Percentile:              round2(percentile),
			SampleSize:              len(counts),
			Confidence:              confidence,
			Caution:                 caution,
		})
	}

	sort.Slice(comparisons, func(i, j int) bool {
		if comparisons[i].CohortMedianCount == comparisons[j].CohortMedianCount {
			return comparisons[i].AbilityName < comparisons[j].AbilityName
		}
		return comparisons[i].CohortMedianCount > comparisons[j].CohortMedianCount
	})

	return filterTopAbilityComparisons(comparisons, 12)
}

func (s *AnalysisService) CalculateBuffUptimeComparisons(playerData types.PlayerFightData, cohortData []types.PlayerFightData) []types.BuffUptimeComparison {
	playerSummary := summarizeBuffUptime(playerData)
	cohortSummaries := make([]map[int]buffUptimeSummary, len(cohortData))
	buffNames := make(map[int]string)

	for id, summary := range playerSummary {
		buffNames[id] = summary.AbilityName
	}

	for i, entry := range cohortData {
		cohortSummaries[i] = summarizeBuffUptime(entry)
		for id, summary := range cohortSummaries[i] {
			if _, exists := buffNames[id]; !exists {
				buffNames[id] = summary.AbilityName
			}
		}
	}

	comparisons := make([]types.BuffUptimeComparison, 0, len(buffNames))
	for abilityID, abilityName := range buffNames {
		uptimes := make([]float64, 0, len(cohortSummaries))
		firstApplies := make([]float64, 0, len(cohortSummaries))

		for _, summaryMap := range cohortSummaries {
			summary, ok := summaryMap[abilityID]
			if !ok {
				uptimes = append(uptimes, 0)
				continue
			}
			uptimes = append(uptimes, summary.UptimePct)
			if summary.HasFirstApply {
				firstApplies = append(firstApplies, summary.FirstApplySeconds)
			}
		}

		player := playerSummary[abilityID]
		cohortMedianUptime := calculateMedianCopy(uptimes)
		cohortMedianFirstApply := calculateMedianCopy(firstApplies)
		percentile := s.calculatePercentileFromValues(player.UptimePct, uptimes)

		confidence := "high"
		var caution string
		if len(uptimes) < 5 {
			confidence = "medium"
			caution = "Small cohort sample for this buff"
		}
		if player.UptimePct == 0 && cohortMedianUptime == 0 {
			confidence = "low"
			caution = "This buff was not commonly observed in the cohort"
		}

		comparisons = append(comparisons, types.BuffUptimeComparison{
			AbilityID:               abilityID,
			AbilityName:             abilityName,
			PlayerUptimePct:         round2(player.UptimePct),
			PlayerFirstApplySeconds: round2(player.FirstApplySeconds),
			CohortMedianUptimePct:   round2(cohortMedianUptime),
			CohortMedianFirstApply:  round2(cohortMedianFirstApply),
			UptimeDelta:             round2(player.UptimePct - cohortMedianUptime),
			FirstApplyDeltaSeconds:  round2(player.FirstApplySeconds - cohortMedianFirstApply),
			Percentile:              round2(percentile),
			SampleSize:              len(uptimes),
			Confidence:              confidence,
			Caution:                 caution,
		})
	}

	sort.Slice(comparisons, func(i, j int) bool {
		if comparisons[i].CohortMedianUptimePct == comparisons[j].CohortMedianUptimePct {
			return comparisons[i].AbilityName < comparisons[j].AbilityName
		}
		return comparisons[i].CohortMedianUptimePct > comparisons[j].CohortMedianUptimePct
	})

	return filterTopBuffComparisons(comparisons, 10)
}

func summarizeAbilityUsage(data types.PlayerFightData) map[int]abilityUsageSummary {
	summaries := make(map[int]abilityUsageSummary)
	durationMinutes := data.FightEnd.Sub(data.FightStart).Minutes()
	if durationMinutes <= 0 {
		durationMinutes = 1
	}

	for _, event := range data.CastEvents {
		if event.Ability.ID == 0 || event.Ability.Name == "" {
			continue
		}

		summary := summaries[event.Ability.ID]
		summary.AbilityID = event.Ability.ID
		summary.AbilityName = event.Ability.Name
		summary.Count++
		if !summary.HasFirstUse || event.Timestamp.Before(data.FightStart.Add(time.Duration(summary.FirstUseSeconds*float64(time.Second)))) {
			summary.FirstUseSeconds = event.Timestamp.Sub(data.FightStart).Seconds()
			summary.HasFirstUse = true
		}
		summaries[event.Ability.ID] = summary
	}

	for _, event := range data.DamageEvents {
		if event.Ability.ID == 0 || event.Ability.Name == "" {
			continue
		}

		summary := summaries[event.Ability.ID]
		summary.AbilityID = event.Ability.ID
		summary.AbilityName = event.Ability.Name
		summary.Count++
		if !summary.HasFirstUse || event.Timestamp.Before(data.FightStart.Add(time.Duration(summary.FirstUseSeconds*float64(time.Second)))) {
			summary.FirstUseSeconds = event.Timestamp.Sub(data.FightStart).Seconds()
			summary.HasFirstUse = true
		}
		summaries[event.Ability.ID] = summary
	}

	for abilityID, summary := range summaries {
		summary.CastsPerMinute = float64(summary.Count) / durationMinutes
		summaries[abilityID] = summary
	}

	return summaries
}

func summarizeBuffUptime(data types.PlayerFightData) map[int]buffUptimeSummary {
	type state struct {
		name          string
		active        bool
		lastApplied   time.Time
		totalUptime   time.Duration
		firstApplied  float64
		hasFirstApply bool
	}

	durationSeconds := data.FightEnd.Sub(data.FightStart).Seconds()
	if durationSeconds <= 0 {
		durationSeconds = 1
	}

	states := make(map[int]*state)
	events := append([]types.BuffEvent(nil), data.BuffEvents...)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	for _, event := range events {
		if event.Ability.ID == 0 || event.Ability.Name == "" {
			continue
		}
		if event.TargetID != 0 && event.TargetID != data.PlayerID {
			continue
		}
		if event.Timestamp.IsZero() {
			continue
		}

		timestamp := clampTimestamp(event.Timestamp, data.FightStart, data.FightEnd)
		if timestamp.IsZero() {
			continue
		}

		entry, ok := states[event.Ability.ID]
		if !ok {
			entry = &state{name: event.Ability.Name}
			states[event.Ability.ID] = entry
		}

		switch event.EventType {
		case "apply":
			if !entry.hasFirstApply {
				entry.firstApplied = timestamp.Sub(data.FightStart).Seconds()
				entry.hasFirstApply = true
			}
			if !entry.active {
				entry.active = true
				entry.lastApplied = timestamp
			}
		case "refresh":
			if !entry.hasFirstApply {
				entry.firstApplied = timestamp.Sub(data.FightStart).Seconds()
				entry.hasFirstApply = true
			}
			if !entry.active {
				entry.active = true
				entry.lastApplied = timestamp
			}
		case "remove":
			if entry.active && !entry.lastApplied.IsZero() && timestamp.After(entry.lastApplied) {
				entry.totalUptime += timestamp.Sub(entry.lastApplied)
				entry.active = false
			}
		}
	}

	summaries := make(map[int]buffUptimeSummary, len(states))
	for abilityID, entry := range states {
		if entry.active && !entry.lastApplied.IsZero() && data.FightEnd.After(entry.lastApplied) {
			entry.totalUptime += data.FightEnd.Sub(entry.lastApplied)
		}
		if entry.totalUptime < 0 {
			entry.totalUptime = 0
		}
		if entry.totalUptime > data.FightEnd.Sub(data.FightStart) {
			entry.totalUptime = data.FightEnd.Sub(data.FightStart)
		}
		summaries[abilityID] = buffUptimeSummary{
			AbilityID:         abilityID,
			AbilityName:       entry.name,
			UptimePct:         (entry.totalUptime.Seconds() / durationSeconds) * 100,
			FirstApplySeconds: entry.firstApplied,
			HasFirstApply:     entry.hasFirstApply,
		}
	}

	return summaries
}

func calculateMedianCopy(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	count := len(sorted)
	middle := count / 2
	if count%2 == 0 {
		return (sorted[middle-1] + sorted[middle]) / 2
	}
	return sorted[middle]
}

func filterTopAbilityComparisons(values []types.AbilityUsageComparison, limit int) []types.AbilityUsageComparison {
	filtered := make([]types.AbilityUsageComparison, 0, len(values))
	for _, value := range values {
		if value.PlayerCount == 0 && value.CohortMedianCount == 0 {
			continue
		}
		filtered = append(filtered, value)
	}
	if len(filtered) <= limit {
		return filtered
	}
	return filtered[:limit]
}

func filterTopBuffComparisons(values []types.BuffUptimeComparison, limit int) []types.BuffUptimeComparison {
	filtered := make([]types.BuffUptimeComparison, 0, len(values))
	for _, value := range values {
		if value.PlayerUptimePct == 0 && value.CohortMedianUptimePct == 0 {
			continue
		}
		filtered = append(filtered, value)
	}
	if len(filtered) <= limit {
		return filtered
	}
	return filtered[:limit]
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}

func clampTimestamp(value, start, end time.Time) time.Time {
	if value.IsZero() || start.IsZero() || end.IsZero() {
		return time.Time{}
	}
	if value.Before(start) {
		return start
	}
	if value.After(end) {
		return end
	}
	return value
}
