package services

import (
	"sort"
	"strings"
)

func buildAbilityHighlights(values []AbilityUsageComparison, limit int, characterClass, characterSpec string, playerData timelineFightData, eliteData []timelineFightData) []insightHighlight {
	if len(values) == 0 || limit <= 0 {
		return nil
	}

	prioritySet := specPriorityFor(characterClass, characterSpec)
	orderedValues, categories := orderAbilityHighlights(values, characterClass, characterSpec)
	filteredValues := filterAbilityHighlightsForAI(orderedValues, prioritySet)
	if len(filteredValues) == 0 {
		filteredValues = orderedValues
	}

	highlights := make([]insightHighlight, 0, limit)
	for _, value := range filteredValues {
		playerUseTimes := abilityUseTimesSeconds(playerData, value.AbilityID, 3)
		eliteUseTimes := eliteMedianUseTimesSeconds(eliteData, value.AbilityID, 3)
		if isHighFrequencyAbility(value) {
			playerUseTimes = nil
			eliteUseTimes = nil
		}

		highlights = append(highlights, insightHighlight{
			Name:                  value.AbilityName,
			PlayerValue:           value.PlayerCastsPerMinute,
			EliteValue:            value.CohortMedianPerMinute,
			Difference:            value.PerMinuteDelta,
			Unit:                  "/min",
			PlayerRawCount:        float64(value.PlayerCount),
			EliteRawCount:         value.CohortMedianCount,
			PlayerTimingSeconds:   value.PlayerFirstUseSeconds,
			EliteTimingSeconds:    value.CohortMedianFirstUseSec,
			TimingDeltaSeconds:    value.FirstUseDeltaSeconds,
			TimingLabel:           "first use",
			PlayerUseTimesSeconds: playerUseTimes,
			EliteUseTimesSeconds:  eliteUseTimes,
			Category:              categories[normalizeTrackedCooldownName(value.AbilityName)],
		})
		if len(highlights) == limit {
			break
		}
	}

	return highlights
}

func buildBuffHighlights(values []BuffUptimeComparison, limit int, characterClass, characterSpec string, playerData timelineFightData, eliteData []timelineFightData) []insightHighlight {
	if len(values) == 0 || limit <= 0 {
		return nil
	}

	prioritySet := specPriorityFor(characterClass, characterSpec)
	orderedValues, categories := orderBuffHighlights(values, characterClass, characterSpec)
	filteredValues := filterBuffHighlightsForAI(orderedValues, prioritySet)
	if len(filteredValues) == 0 {
		filteredValues = orderedValues
	}

	highlights := make([]insightHighlight, 0, limit)
	for _, value := range filteredValues {
		highlights = append(highlights, insightHighlight{
			Name:                value.AbilityName,
			PlayerValue:         value.PlayerUptimePct,
			EliteValue:          value.CohortMedianUptimePct,
			Difference:          value.UptimeDelta,
			Unit:                "%",
			PlayerTimingSeconds: value.PlayerFirstApplySeconds,
			EliteTimingSeconds:  value.CohortMedianFirstApply,
			TimingDeltaSeconds:  value.FirstApplyDeltaSeconds,
			TimingLabel:         "first apply",
			PlayerLargestGapSec: largestBuffGapSeconds(playerData, value.AbilityID),
			EliteLargestGapSec:  medianLargestBuffGapSeconds(eliteData, value.AbilityID),
			Category:            categories[normalizeTrackedCooldownName(value.AbilityName)],
		})
		if len(highlights) == limit {
			break
		}
	}

	return highlights
}

func buildResourceHighlights(values []ResourceUsageComparison, limit int) []insightHighlight {
	if len(values) == 0 || limit <= 0 {
		return nil
	}

	ordered := append([]ResourceUsageComparison(nil), values...)
	sort.SliceStable(ordered, func(i, j int) bool {
		leftConcern := ordered[i].TimeAtMaxDeltaSeconds + maxFloat64(0, -ordered[i].SpentDelta/10)
		rightConcern := ordered[j].TimeAtMaxDeltaSeconds + maxFloat64(0, -ordered[j].SpentDelta/10)
		if leftConcern == rightConcern {
			return ordered[i].ResourceType < ordered[j].ResourceType
		}
		return leftConcern > rightConcern
	})

	highlights := make([]insightHighlight, 0, limit)
	for _, value := range ordered {
		if value.PlayerAveragePct == 0 && value.CohortMedianAveragePct == 0 && value.PlayerSpent == 0 && value.CohortMedianSpent == 0 {
			continue
		}
		highlights = append(highlights, insightHighlight{
			Name:                value.ResourceType,
			PlayerValue:         value.PlayerAveragePct,
			EliteValue:          value.CohortMedianAveragePct,
			Difference:          value.AveragePctDelta,
			Unit:                "%",
			PlayerRawCount:      value.PlayerSpent,
			EliteRawCount:       value.CohortMedianSpent,
			PlayerTimingSeconds: value.PlayerTimeAtMaxSeconds,
			EliteTimingSeconds:  value.CohortMedianTimeAtMaxSeconds,
			TimingDeltaSeconds:  value.TimeAtMaxDeltaSeconds,
			TimingLabel:         "time at max",
			Category:            "resource",
		})
		if len(highlights) == limit {
			break
		}
	}

	return highlights
}

func abilityUseTimesSeconds(data timelineFightData, abilityID int, limit int) []float64 {
	if limit <= 0 {
		return nil
	}

	casts := buildAbilityTimelineSeries(data, abilityID, "", "", "").CastsMS
	if len(casts) == 0 {
		return nil
	}
	if len(casts) > limit {
		casts = casts[:limit]
	}

	values := make([]float64, 0, len(casts))
	for _, cast := range casts {
		values = append(values, float64(cast)/1000)
	}
	return values
}

func eliteMedianUseTimesSeconds(values []timelineFightData, abilityID int, limit int) []float64 {
	if limit <= 0 {
		return nil
	}

	useTimesByElite := make([][]float64, 0, len(values))
	maxUses := 0
	for _, value := range values {
		useTimes := abilityUseTimesSeconds(value, abilityID, limit)
		if len(useTimes) == 0 {
			continue
		}
		useTimesByElite = append(useTimesByElite, useTimes)
		if len(useTimes) > maxUses {
			maxUses = len(useTimes)
		}
	}

	if len(useTimesByElite) == 0 {
		return nil
	}

	medians := make([]float64, 0, maxUses)
	for useIndex := 0; useIndex < maxUses; useIndex++ {
		samples := make([]float64, 0, len(useTimesByElite))
		for _, useTimes := range useTimesByElite {
			if useIndex < len(useTimes) {
				samples = append(samples, useTimes[useIndex])
			}
		}
		if len(samples) == 0 {
			continue
		}
		medians = append(medians, medianFloat64(samples))
	}

	return medians
}

func largestBuffGapSeconds(data timelineFightData, abilityID int) float64 {
	windows := buffWindows(data, abilityID)
	if len(windows) == 0 {
		return data.FightEnd.Sub(data.FightStart).Seconds()
	}

	largestGap := windows[0].start.Sub(data.FightStart).Seconds()
	previousEnd := data.FightStart
	for _, window := range windows {
		gap := window.start.Sub(previousEnd).Seconds()
		if gap > largestGap {
			largestGap = gap
		}
		previousEnd = window.end
	}
	finalGap := data.FightEnd.Sub(previousEnd).Seconds()
	if finalGap > largestGap {
		largestGap = finalGap
	}
	if largestGap < 0 {
		return 0
	}
	return largestGap
}

func medianLargestBuffGapSeconds(values []timelineFightData, abilityID int) float64 {
	samples := make([]float64, 0, len(values))
	for _, value := range values {
		samples = append(samples, largestBuffGapSeconds(value, abilityID))
	}
	if len(samples) == 0 {
		return 0
	}
	return medianFloat64(samples)
}

func medianFloat64(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	ordered := append([]float64(nil), values...)
	sort.Float64s(ordered)
	middle := len(ordered) / 2
	if len(ordered)%2 == 1 {
		return ordered[middle]
	}
	return (ordered[middle-1] + ordered[middle]) / 2
}

func maxFloat64(left, right float64) float64 {
	if left > right {
		return left
	}
	return right
}

func filterAbilityHighlightsForAI(values []AbilityUsageComparison, prioritySet specPrioritySet) []AbilityUsageComparison {
	if len(values) == 0 {
		return nil
	}

	tracked := make(map[string]struct{}, len(prioritySet.Offensives)+len(prioritySet.Defensives)+1)
	for _, value := range prioritySet.Offensives {
		tracked[normalizeTrackedCooldownName(value)] = struct{}{}
	}
	for _, value := range prioritySet.Defensives {
		tracked[normalizeTrackedCooldownName(value)] = struct{}{}
	}
	tracked["melee"] = struct{}{}

	filtered := make([]AbilityUsageComparison, 0, len(values))
	for _, value := range values {
		if _, ok := tracked[normalizeTrackedCooldownName(value.AbilityName)]; ok {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func filterBuffHighlightsForAI(values []BuffUptimeComparison, prioritySet specPrioritySet) []BuffUptimeComparison {
	if len(values) == 0 {
		return nil
	}

	tracked := make(map[string]struct{}, len(prioritySet.Offensives)+len(prioritySet.Defensives))
	for _, value := range prioritySet.Offensives {
		tracked[normalizeTrackedCooldownName(value)] = struct{}{}
	}
	for _, value := range prioritySet.Defensives {
		tracked[normalizeTrackedCooldownName(value)] = struct{}{}
	}

	filtered := make([]BuffUptimeComparison, 0, len(values))
	for _, value := range values {
		if _, ok := tracked[normalizeTrackedCooldownName(value.AbilityName)]; ok {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func isHighFrequencyAbility(value AbilityUsageComparison) bool {
	return normalizeTrackedCooldownName(value.AbilityName) == "melee" || value.PlayerCount > 300 || value.CohortMedianCount > 300
}

func orderAbilityHighlights(values []AbilityUsageComparison, characterClass, characterSpec string) ([]AbilityUsageComparison, map[string]string) {
	prioritySet := specPriorityFor(characterClass, characterSpec)
	return orderTrackedAbilities(values, prioritySet), buildCooldownCategoryMap(prioritySet)
}

func orderBuffHighlights(values []BuffUptimeComparison, characterClass, characterSpec string) ([]BuffUptimeComparison, map[string]string) {
	prioritySet := specPriorityFor(characterClass, characterSpec)
	return orderTrackedBuffs(values, prioritySet), buildCooldownCategoryMap(prioritySet)
}

func orderTrackedAbilities(values []AbilityUsageComparison, prioritySet specPrioritySet) []AbilityUsageComparison {
	order := append(append([]string{}, prioritySet.Offensives...), prioritySet.Defensives...)
	indexByName := make(map[string]int, len(order))
	for index, value := range order {
		indexByName[normalizeTrackedCooldownName(value)] = index
	}

	ordered := append([]AbilityUsageComparison(nil), values...)
	sort.SliceStable(ordered, func(i, j int) bool {
		leftIndex, leftTracked := indexByName[normalizeTrackedCooldownName(ordered[i].AbilityName)]
		rightIndex, rightTracked := indexByName[normalizeTrackedCooldownName(ordered[j].AbilityName)]
		if leftTracked && rightTracked {
			return leftIndex < rightIndex
		}
		if leftTracked != rightTracked {
			return leftTracked
		}
		return strings.ToLower(ordered[i].AbilityName) < strings.ToLower(ordered[j].AbilityName)
	})

	return ordered
}

func orderTrackedBuffs(values []BuffUptimeComparison, prioritySet specPrioritySet) []BuffUptimeComparison {
	order := append(append([]string{}, prioritySet.Offensives...), prioritySet.Defensives...)
	indexByName := make(map[string]int, len(order))
	for index, value := range order {
		indexByName[normalizeTrackedCooldownName(value)] = index
	}

	ordered := append([]BuffUptimeComparison(nil), values...)
	sort.SliceStable(ordered, func(i, j int) bool {
		leftIndex, leftTracked := indexByName[normalizeTrackedCooldownName(ordered[i].AbilityName)]
		rightIndex, rightTracked := indexByName[normalizeTrackedCooldownName(ordered[j].AbilityName)]
		if leftTracked && rightTracked {
			return leftIndex < rightIndex
		}
		if leftTracked != rightTracked {
			return leftTracked
		}
		return strings.ToLower(ordered[i].AbilityName) < strings.ToLower(ordered[j].AbilityName)
	})

	return ordered
}
