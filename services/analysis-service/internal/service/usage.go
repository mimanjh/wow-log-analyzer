package service

import (
	"math"
	"sort"
	"strings"
	"time"

	"wow-log-analyzer/services/analysis-service/types"
)


func (s *AnalysisService) CalculateAbilityUsageComparisons(playerData types.PlayerFightData, cohortData []types.PlayerFightData, characterClass, characterSpec string) []types.AbilityUsageComparison {
	playerSummary := summarizeAbilityUsage(playerData)
	cohortSummaries := make([]map[int]types.AbilityUsageSummary, len(cohortData))
	for i, entry := range cohortData {
		cohortSummaries[i] = summarizeAbilityUsage(entry)
	}
	return s.CompareAbilityUsages(playerSummary, cohortSummaries, characterClass, characterSpec)
}

// CompareAbilityUsages is the cached-input variant — same logic as
// CalculateAbilityUsageComparisons but skips the raw-event summarization step.
func (s *AnalysisService) CompareAbilityUsages(playerSummary map[int]types.AbilityUsageSummary, cohortSummaries []map[int]types.AbilityUsageSummary, characterClass, characterSpec string) []types.AbilityUsageComparison {
	abilityNames := make(map[int]string)
	for id, summary := range playerSummary {
		abilityNames[id] = summary.AbilityName
	}
	for _, summaryMap := range cohortSummaries {
		for id, summary := range summaryMap {
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

	return filterTopAbilityComparisons(comparisons, 12, trackedAbilityNamesForSpec(characterClass, characterSpec))
}

func (s *AnalysisService) CalculateBuffUptimeComparisons(playerData types.PlayerFightData, cohortData []types.PlayerFightData) []types.BuffUptimeComparison {
	playerSummary := summarizeBuffUptime(playerData)
	cohortSummaries := make([]map[int]types.BuffUptimeSummary, len(cohortData))
	for i, entry := range cohortData {
		cohortSummaries[i] = summarizeBuffUptime(entry)
	}
	return s.CompareBuffUptimes(playerSummary, cohortSummaries)
}

// CompareBuffUptimes is the cached-input variant of CalculateBuffUptimeComparisons.
func (s *AnalysisService) CompareBuffUptimes(playerSummary map[int]types.BuffUptimeSummary, cohortSummaries []map[int]types.BuffUptimeSummary) []types.BuffUptimeComparison {
	buffNames := make(map[int]string)
	for id, summary := range playerSummary {
		buffNames[id] = summary.AbilityName
	}
	for _, summaryMap := range cohortSummaries {
		for id, summary := range summaryMap {
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

func summarizeAbilityUsage(data types.PlayerFightData) map[int]types.AbilityUsageSummary {
	summaries := make(map[int]types.AbilityUsageSummary)
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

	castOrDamageAbilityIDs := make(map[int]bool, len(summaries))
	for abilityID := range summaries {
		castOrDamageAbilityIDs[abilityID] = true
	}
	for _, event := range data.CooldownEvents {
		if event.Ability.ID == 0 || event.Ability.Name == "" || event.EventType != "start" {
			continue
		}
		if castOrDamageAbilityIDs[event.Ability.ID] {
			continue
		}

		summary := summaries[event.Ability.ID]
		summary.AbilityID = event.Ability.ID
		summary.AbilityName = event.Ability.Name
		summary.Count++
		summary.FirstUseSeconds = event.Timestamp.Sub(data.FightStart).Seconds()
		summary.HasFirstUse = true
		summaries[event.Ability.ID] = summary
	}

	for abilityID, summary := range summaries {
		summary.CastsPerMinute = float64(summary.Count) / durationMinutes
		summaries[abilityID] = summary
	}

	return summaries
}

func summarizeBuffUptime(data types.PlayerFightData) map[int]types.BuffUptimeSummary {
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

	summaries := make(map[int]types.BuffUptimeSummary, len(states))
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
		summaries[abilityID] = types.BuffUptimeSummary{
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

func filterTopAbilityComparisons(values []types.AbilityUsageComparison, limit int, priorityNames map[string]bool) []types.AbilityUsageComparison {
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

	selected := append([]types.AbilityUsageComparison(nil), filtered[:limit]...)
	selectedNames := make(map[string]bool, len(selected))
	for _, value := range selected {
		selectedNames[normalizeAbilityName(value.AbilityName)] = true
	}
	for _, value := range filtered[limit:] {
		normalizedName := normalizeAbilityName(value.AbilityName)
		if !priorityNames[normalizedName] || selectedNames[normalizedName] {
			continue
		}
		selected[len(selected)-1] = value
		selectedNames[normalizedName] = true
	}
	sort.SliceStable(selected, func(i, j int) bool {
		leftPriority := priorityNames[normalizeAbilityName(selected[i].AbilityName)]
		rightPriority := priorityNames[normalizeAbilityName(selected[j].AbilityName)]
		if leftPriority != rightPriority {
			return leftPriority
		}
		if selected[i].CohortMedianCount == selected[j].CohortMedianCount {
			return selected[i].AbilityName < selected[j].AbilityName
		}
		return selected[i].CohortMedianCount > selected[j].CohortMedianCount
	})
	return selected
}

func trackedAbilityNamesForSpec(characterClass, characterSpec string) map[string]bool {
	key := normalizeAbilityName(strings.TrimSpace(characterSpec) + " " + strings.TrimSpace(characterClass))
	switch key {
	case "blood death knight", "blood deathknight":
		return map[string]bool{
			"dancing rune weapon": true,
		}
	default:
		return nil
	}
}

func normalizeAbilityName(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
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
