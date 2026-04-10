package service

import (
	"sort"
	"strings"

	"wow-log-analyzer/services/analysis-service/internal/types"
)

type resourceUsageSummary struct {
	ResourceTypeID     int
	ResourceType       string
	GeneratedPerMinute float64
	WastePerMinute     float64
	WastePct           float64
}

func (s *AnalysisService) CalculateResourceUsageComparisons(playerData types.PlayerFightData, cohortData []types.PlayerFightData, characterClass, characterSpec string) []types.ResourceUsageComparison {
	strategy := resolveResourceStrategy(characterClass, characterSpec)
	playerSummary := summarizeResourceUsage(playerData, strategy)
	cohortSummaries := make([]map[int]resourceUsageSummary, len(cohortData))
	resourceNames := make(map[int]string)

	for id, summary := range playerSummary {
		resourceNames[id] = summary.ResourceType
	}

	for i, entry := range cohortData {
		cohortSummaries[i] = summarizeResourceUsage(entry, strategy)
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

		var caution string
		if len(generatedValues) < 5 {
			caution = "Small cohort sample for this resource"
		}
		if player.GeneratedPerMinute == 0 && cohortGenerated == 0 && player.WastePerMinute == 0 && cohortWaste == 0 {
			caution = "This resource was not commonly observed in the cohort"
		}
		if caution == "" {
			caution = buildResourceCaution(player, cohortGenerated, cohortWastePct, strategy)
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

func summarizeResourceUsage(data types.PlayerFightData, strategy resourceStrategy) map[int]resourceUsageSummary {
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
		if len(strategy.resourceIDs) > 0 && !strategy.resourceIDs[event.ResourceTypeID] {
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

type resourceStrategy struct {
	resourceIDs map[int]bool
	primaryID   int
}

func resolveResourceStrategy(characterClass, characterSpec string) resourceStrategy {
	key := strings.ToLower(strings.TrimSpace(characterSpec + " " + characterClass))
	switch key {
	case "blood deathknight", "blood death knight", "frost deathknight", "frost death knight", "unholy deathknight", "unholy death knight":
		return resourceStrategy{resourceIDs: resourceSet(5, 6), primaryID: 6}
	case "assassination rogue", "outlaw rogue", "subtlety rogue":
		return resourceStrategy{resourceIDs: resourceSet(3, 4), primaryID: 3}
	case "feral druid":
		return resourceStrategy{resourceIDs: resourceSet(3, 4), primaryID: 3}
	case "guardian druid":
		return resourceStrategy{resourceIDs: resourceSet(1), primaryID: 1}
	case "balance druid":
		return resourceStrategy{resourceIDs: resourceSet(8), primaryID: 8}
	case "retribution paladin", "protection paladin", "holy paladin":
		return resourceStrategy{resourceIDs: resourceSet(9), primaryID: 9}
	case "arms warrior", "fury warrior", "protection warrior":
		return resourceStrategy{resourceIDs: resourceSet(1), primaryID: 1}
	case "havoc demonhunter", "havoc demon hunter":
		return resourceStrategy{resourceIDs: resourceSet(17), primaryID: 17}
	case "vengeance demonhunter", "vengeance demon hunter", "devourer demonhunter", "devourer demon hunter":
		return resourceStrategy{resourceIDs: resourceSet(18), primaryID: 18}
	case "windwalker monk":
		return resourceStrategy{resourceIDs: resourceSet(3, 12), primaryID: 12}
	case "brewmaster monk":
		return resourceStrategy{resourceIDs: resourceSet(3), primaryID: 3}
	case "mistweaver monk":
		return resourceStrategy{resourceIDs: resourceSet(3, 12), primaryID: 12}
	case "beast mastery hunter", "marksmanship hunter", "survival hunter":
		return resourceStrategy{resourceIDs: resourceSet(2), primaryID: 2}
	case "enhancement shaman", "elemental shaman":
		return resourceStrategy{resourceIDs: resourceSet(11), primaryID: 11}
	case "shadow priest":
		return resourceStrategy{resourceIDs: resourceSet(13), primaryID: 13}
	case "arcane mage":
		return resourceStrategy{resourceIDs: resourceSet(16), primaryID: 16}
	case "augmentation evoker", "devastation evoker":
		return resourceStrategy{resourceIDs: resourceSet(19), primaryID: 19}
	case "affliction warlock", "demonology warlock", "destruction warlock":
		return resourceStrategy{resourceIDs: resourceSet(7), primaryID: 7}
	default:
		return resourceStrategy{}
	}
}

func resourceSet(ids ...int) map[int]bool {
	values := make(map[int]bool, len(ids))
	for _, id := range ids {
		values[id] = true
	}
	return values
}

func buildResourceCaution(player resourceUsageSummary, cohortGenerated, cohortWastePct float64, strategy resourceStrategy) string {
	isPrimary := strategy.primaryID != 0 && player.ResourceTypeID == strategy.primaryID
	switch {
	case player.WastePct-cohortWastePct >= 5:
		if isPrimary {
			return "Primary resource overcap is higher than the elite baseline."
		}
		return "Resource waste is higher than the elite baseline."
	case cohortGenerated-player.GeneratedPerMinute >= maxFloat(5, cohortGenerated*0.1):
		if isPrimary {
			return "Primary resource generation is behind the elite baseline, which can indicate fewer spend windows or idle time."
		}
		return "Resource generation is behind the elite baseline."
	default:
		return ""
	}
}

func maxFloat(left, right float64) float64 {
	if left > right {
		return left
	}
	return right
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
