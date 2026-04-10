package service

import (
	"sort"

	"wow-log-analyzer/services/analysis-service/internal/types"
)

type resourceUsageSummary struct {
	ResourceTypeID     int
	ResourceType       string
	GeneratedPerMinute float64
	WastePerMinute     float64
	WastePct           float64
}

func (s *AnalysisService) CalculateResourceUsageComparisons(playerData types.PlayerFightData, cohortData []types.PlayerFightData) []types.ResourceUsageComparison {
	playerSummary := summarizeResourceUsage(playerData)
	cohortSummaries := make([]map[int]resourceUsageSummary, len(cohortData))
	resourceNames := make(map[int]string)

	for id, summary := range playerSummary {
		resourceNames[id] = summary.ResourceType
	}

	for i, entry := range cohortData {
		cohortSummaries[i] = summarizeResourceUsage(entry)
		for id, summary := range cohortSummaries[i] {
			if _, exists := resourceNames[id]; !exists {
				resourceNames[id] = summary.ResourceType
			}
		}
	}

	comparisons := make([]types.ResourceUsageComparison, 0, len(resourceNames))
	for resourceTypeID, resourceType := range resourceNames {
		generatedValues := make([]float64, 0, len(cohortSummaries))
		wasteValues := make([]float64, 0, len(cohortSummaries))
		wastePctValues := make([]float64, 0, len(cohortSummaries))

		for _, summaryMap := range cohortSummaries {
			summary, ok := summaryMap[resourceTypeID]
			if !ok {
				generatedValues = append(generatedValues, 0)
				wasteValues = append(wasteValues, 0)
				wastePctValues = append(wastePctValues, 0)
				continue
			}
			generatedValues = append(generatedValues, summary.GeneratedPerMinute)
			wasteValues = append(wasteValues, summary.WastePerMinute)
			wastePctValues = append(wastePctValues, summary.WastePct)
		}

		player := playerSummary[resourceTypeID]
		cohortGenerated := calculateMedianCopy(generatedValues)
		cohortWaste := calculateMedianCopy(wasteValues)
		cohortWastePct := calculateMedianCopy(wastePctValues)

		confidence := "high"
		var caution string
		if len(generatedValues) < 5 {
			confidence = "medium"
			caution = "Small cohort sample for this resource"
		}
		if player.GeneratedPerMinute == 0 && cohortGenerated == 0 && player.WastePerMinute == 0 && cohortWaste == 0 {
			confidence = "low"
			caution = "This resource was not commonly observed in the cohort"
		}

		comparisons = append(comparisons, types.ResourceUsageComparison{
			ResourceTypeID:                 resourceTypeID,
			ResourceType:                   resourceType,
			PlayerGeneratedPerMinute:       round2(player.GeneratedPerMinute),
			CohortMedianGeneratedPerMinute: round2(cohortGenerated),
			GeneratedDelta:                 round2(player.GeneratedPerMinute - cohortGenerated),
			PlayerWastePerMinute:           round2(player.WastePerMinute),
			CohortMedianWastePerMinute:     round2(cohortWaste),
			WasteDelta:                     round2(player.WastePerMinute - cohortWaste),
			PlayerWastePct:                 round2(player.WastePct),
			CohortMedianWastePct:           round2(cohortWastePct),
			WastePctDelta:                  round2(player.WastePct - cohortWastePct),
			SampleSize:                     len(generatedValues),
			Confidence:                     confidence,
			Caution:                        caution,
		})
	}

	sort.Slice(comparisons, func(i, j int) bool {
		if comparisons[i].CohortMedianGeneratedPerMinute == comparisons[j].CohortMedianGeneratedPerMinute {
			return comparisons[i].ResourceType < comparisons[j].ResourceType
		}
		return comparisons[i].CohortMedianGeneratedPerMinute > comparisons[j].CohortMedianGeneratedPerMinute
	})

	return filterTopResourceComparisons(comparisons, 10)
}

func summarizeResourceUsage(data types.PlayerFightData) map[int]resourceUsageSummary {
	durationMinutes := data.FightEnd.Sub(data.FightStart).Minutes()
	if durationMinutes <= 0 {
		durationMinutes = 1
	}

	type aggregate struct {
		name      string
		generated float64
		waste     float64
	}

	aggregates := make(map[int]*aggregate)
	for _, event := range data.ResourceEvents {
		if event.ResourceTypeID == 0 && event.ResourceType == "" {
			continue
		}

		entry, ok := aggregates[event.ResourceTypeID]
		if !ok {
			entry = &aggregate{name: event.ResourceType}
			aggregates[event.ResourceTypeID] = entry
		}
		if entry.name == "" {
			entry.name = event.ResourceType
		}
		if event.Change > 0 {
			entry.generated += event.Change
		}
		if event.Waste > 0 {
			entry.waste += event.Waste
		}
	}

	summaries := make(map[int]resourceUsageSummary, len(aggregates))
	for resourceTypeID, entry := range aggregates {
		wastePct := 0.0
		if entry.generated > 0 {
			wastePct = (entry.waste / entry.generated) * 100
		}
		summaries[resourceTypeID] = resourceUsageSummary{
			ResourceTypeID:     resourceTypeID,
			ResourceType:       entry.name,
			GeneratedPerMinute: entry.generated / durationMinutes,
			WastePerMinute:     entry.waste / durationMinutes,
			WastePct:           wastePct,
		}
	}

	return summaries
}

func filterTopResourceComparisons(values []types.ResourceUsageComparison, limit int) []types.ResourceUsageComparison {
	filtered := make([]types.ResourceUsageComparison, 0, len(values))
	for _, value := range values {
		if value.PlayerGeneratedPerMinute == 0 &&
			value.CohortMedianGeneratedPerMinute == 0 &&
			value.PlayerWastePerMinute == 0 &&
			value.CohortMedianWastePerMinute == 0 {
			continue
		}
		filtered = append(filtered, value)
	}
	if len(filtered) <= limit {
		return filtered
	}
	return filtered[:limit]
}
