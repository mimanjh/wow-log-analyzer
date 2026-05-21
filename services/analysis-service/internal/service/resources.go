package service

import (
	"sort"
	"strings"

	"wow-log-analyzer/services/analysis-service/types"
)


type resourceAggregate struct {
	name              string
	sampleCount       int
	fullMarkerCount   int
	fullWindowSeconds float64
	generated         float64
	waste             float64
	spent             float64
	weightedPctTotal  float64
	timeAtMaxSeconds  float64
	observedSeconds   float64
	samples           []types.ResourceEvent
}

func (s *AnalysisService) CalculateResourceUsageComparisons(playerData types.PlayerFightData, cohortData []types.PlayerFightData, characterClass, characterSpec string) []types.ResourceUsageComparison {
	strategy := resolveResourceStrategy(characterClass, characterSpec)
	playerSummary := summarizeResourceUsage(playerData, strategy)
	cohortSummaries := make([]map[int]types.ResourceUsageSummary, len(cohortData))
	for i, entry := range cohortData {
		cohortSummaries[i] = summarizeResourceUsage(entry, strategy)
	}
	return s.CompareResourceUsages(playerSummary, cohortSummaries, characterClass, characterSpec)
}

// CompareResourceUsages is the cached-input variant of CalculateResourceUsageComparisons.
func (s *AnalysisService) CompareResourceUsages(playerSummary map[int]types.ResourceUsageSummary, cohortSummaries []map[int]types.ResourceUsageSummary, characterClass, characterSpec string) []types.ResourceUsageComparison {
	strategy := resolveResourceStrategy(characterClass, characterSpec)
	resourceNames := make(map[int]string)
	for id, summary := range playerSummary {
		resourceNames[id] = summary.ResourceType
	}
	for _, summaryMap := range cohortSummaries {
		for id, summary := range summaryMap {
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
		averagePctValues := make([]float64, 0, len(cohortSummaries))
		timeAtMaxValues := make([]float64, 0, len(cohortSummaries))
		spentValues := make([]float64, 0, len(cohortSummaries))
		sampleCountValues := make([]float64, 0, len(cohortSummaries))
		fullMarkerValues := make([]float64, 0, len(cohortSummaries))
		fullWindowValues := make([]float64, 0, len(cohortSummaries))

		for _, summaryMap := range cohortSummaries {
			summary, ok := summaryMap[resourceTypeID]
			if !ok {
				generatedValues = append(generatedValues, 0)
				wasteValues = append(wasteValues, 0)
				wastePctValues = append(wastePctValues, 0)
				averagePctValues = append(averagePctValues, 0)
				timeAtMaxValues = append(timeAtMaxValues, 0)
				spentValues = append(spentValues, 0)
				sampleCountValues = append(sampleCountValues, 0)
				fullMarkerValues = append(fullMarkerValues, 0)
				fullWindowValues = append(fullWindowValues, 0)
				continue
			}
			generatedValues = append(generatedValues, summary.GeneratedPerMinute)
			wasteValues = append(wasteValues, summary.WastePerMinute)
			wastePctValues = append(wastePctValues, summary.WastePct)
			averagePctValues = append(averagePctValues, summary.AveragePct)
			timeAtMaxValues = append(timeAtMaxValues, summary.TimeAtMaxSeconds)
			spentValues = append(spentValues, summary.Spent)
			sampleCountValues = append(sampleCountValues, float64(summary.SampleCount))
			fullMarkerValues = append(fullMarkerValues, float64(summary.FullMarkerCount))
			fullWindowValues = append(fullWindowValues, summary.FullWindowSeconds)
		}

		player := playerSummary[resourceTypeID]
		cohortGenerated := calculateMedianCopy(generatedValues)
		cohortWaste := calculateMedianCopy(wasteValues)
		cohortWastePct := calculateMedianCopy(wastePctValues)
		cohortAveragePct := calculateMedianCopy(averagePctValues)
		cohortTimeAtMax := calculateMedianCopy(timeAtMaxValues)
		cohortSpent := calculateMedianCopy(spentValues)
		cohortSampleCount := calculateMedianCopy(sampleCountValues)
		cohortFullMarkerCount := calculateMedianCopy(fullMarkerValues)
		cohortFullWindowSeconds := calculateMedianCopy(fullWindowValues)

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
			PlayerSampleCount:              player.SampleCount,
			CohortMedianSampleCount:        round2(cohortSampleCount),
			SampleCountDelta:               round2(float64(player.SampleCount) - cohortSampleCount),
			PlayerFullMarkerCount:          player.FullMarkerCount,
			CohortMedianFullMarkerCount:    round2(cohortFullMarkerCount),
			FullMarkerDelta:                round2(float64(player.FullMarkerCount) - cohortFullMarkerCount),
			PlayerFullWindowSeconds:        round2(player.FullWindowSeconds),
			CohortMedianFullWindowSeconds:  round2(cohortFullWindowSeconds),
			FullWindowDeltaSeconds:         round2(player.FullWindowSeconds - cohortFullWindowSeconds),
			PlayerAveragePct:               round2(player.AveragePct),
			CohortMedianAveragePct:         round2(cohortAveragePct),
			AveragePctDelta:                round2(player.AveragePct - cohortAveragePct),
			PlayerTimeAtMaxSeconds:         round2(player.TimeAtMaxSeconds),
			CohortMedianTimeAtMaxSeconds:   round2(cohortTimeAtMax),
			TimeAtMaxDeltaSeconds:          round2(player.TimeAtMaxSeconds - cohortTimeAtMax),
			PlayerSpent:                    round2(player.Spent),
			CohortMedianSpent:              round2(cohortSpent),
			SpentDelta:                     round2(player.Spent - cohortSpent),
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
		if comparisons[i].ResourceTypeID == comparisons[j].ResourceTypeID {
			return comparisons[i].ResourceType < comparisons[j].ResourceType
		}
		return comparisons[i].ResourceTypeID < comparisons[j].ResourceTypeID
	})

	return comparisons
}

func summarizeResourceUsage(data types.PlayerFightData, strategy resourceStrategy) map[int]types.ResourceUsageSummary {
	durationMinutes := data.FightEnd.Sub(data.FightStart).Minutes()
	if durationMinutes <= 0 {
		durationMinutes = 1
	}

	aggregates := make(map[int]*resourceAggregate)
	for _, event := range data.ResourceEvents {
		if event.ResourceTypeID == 0 && event.ResourceType == "" {
			continue
		}

		entry, ok := aggregates[event.ResourceTypeID]
		if !ok {
			entry = &resourceAggregate{name: event.ResourceType}
			aggregates[event.ResourceTypeID] = entry
		}
		if entry.name == "" {
			entry.name = event.ResourceType
		}
		if event.Change > 0 {
			entry.generated += event.Change
		}
		if event.Change < 0 {
			entry.spent += -event.Change
		}
		if event.Waste > 0 {
			entry.waste += event.Waste
		}
		entry.sampleCount++
		entry.samples = append(entry.samples, event)
	}
	addDerivedRuneSpend(aggregates, data, strategy)

	summaries := make(map[int]types.ResourceUsageSummary, len(aggregates))
	for resourceTypeID, entry := range aggregates {
		sort.Slice(entry.samples, func(i, j int) bool {
			return entry.samples[i].Timestamp.Before(entry.samples[j].Timestamp)
		})
		for index, sample := range entry.samples {
			sampleAmount := resourceSampleAmount(sample)
			maxAmount := sample.MaxAmount
			if maxAmount <= 0 {
				maxAmount = defaultResourceMax(sample.ResourceTypeID, sample.ResourceType)
			}
			if sample.Waste > 0 {
				entry.fullMarkerCount++
				if maxAmount > 0 {
					sampleAmount = maxAmount
				}
			}
			if index > 0 && sample.Change == 0 {
				previous := entry.samples[index-1]
				previousAmount := resourceSampleAmount(previous)
				if previousAmount > sampleAmount {
					entry.spent += previousAmount - sampleAmount
				}
			}

			nextTimestamp := data.FightEnd
			if index+1 < len(entry.samples) {
				nextTimestamp = entry.samples[index+1].Timestamp
			}
			seconds := nextTimestamp.Sub(sample.Timestamp).Seconds()
			if seconds <= 0 {
				continue
			}
			if maxAmount > 0 {
				pct := (sampleAmount / maxAmount) * 100
				if pct < 0 {
					pct = 0
				}
				if pct > 100 {
					pct = 100
				}
				entry.weightedPctTotal += pct * seconds
				entry.observedSeconds += seconds
				if sampleAmount >= maxAmount {
					if sample.Waste == 0 {
						entry.fullMarkerCount++
					}
					entry.fullWindowSeconds += seconds
					entry.timeAtMaxSeconds += seconds
				}
			}
		}

		averagePct := 0.0
		if entry.observedSeconds > 0 {
			averagePct = entry.weightedPctTotal / entry.observedSeconds
		}
		wastePct := 0.0
		if entry.generated > 0 {
			wastePct = (entry.waste / entry.generated) * 100
		}
		summaries[resourceTypeID] = types.ResourceUsageSummary{
			ResourceTypeID:     resourceTypeID,
			ResourceType:       entry.name,
			SampleCount:        entry.sampleCount,
			FullMarkerCount:    entry.fullMarkerCount,
			FullWindowSeconds:  entry.fullWindowSeconds,
			AveragePct:         averagePct,
			TimeAtMaxSeconds:   entry.timeAtMaxSeconds,
			Spent:              entry.spent,
			GeneratedPerMinute: entry.generated / durationMinutes,
			WastePerMinute:     entry.waste / durationMinutes,
			WastePct:           wastePct,
		}
	}

	return summaries
}

func addDerivedRuneSpend(aggregates map[int]*resourceAggregate, data types.PlayerFightData, strategy resourceStrategy) {
	if !strategy.resourceIDs[5] {
		return
	}
	if existing, ok := aggregates[5]; ok && len(existing.samples) > 0 {
		return
	}

	runeEvents := make([]types.ResourceEvent, 0)
	for _, cast := range data.CastEvents {
		runeCost := deathKnightRuneCost(cast.Ability.Name)
		if runeCost <= 0 {
			continue
		}
		runeEvents = append(runeEvents, types.ResourceEvent{
			Timestamp:      cast.Timestamp,
			SourceID:       cast.SourceID,
			ResourceTypeID: 5,
			ResourceType:   "Runes",
			Change:         -float64(runeCost),
			MaxAmount:      6,
		})
	}
	if len(runeEvents) == 0 {
		return
	}

	entry := &resourceAggregate{name: "Runes"}
	for _, event := range runeEvents {
		entry.spent += -event.Change
		entry.sampleCount++
		entry.samples = append(entry.samples, event)
	}
	aggregates[5] = entry
}

func deathKnightRuneCost(abilityName string) int {
	switch strings.ToLower(strings.TrimSpace(abilityName)) {
	case "heart strike",
		"death's caress",
		"death and decay",
		"soul reaper",
		"scourge strike",
		"clawing shadows",
		"festering strike",
		"howling blast",
		"frostscythe":
		return 1
	case "marrowrend",
		"obliterate":
		return 2
	default:
		return 0
	}
}

func resourceSampleAmount(sample types.ResourceEvent) float64 {
	if sample.Amount != 0 {
		return sample.Amount
	}
	if sample.Change > 0 {
		return sample.Change
	}
	return 0
}

func defaultResourceMax(resourceTypeID int, resourceType string) float64 {
	switch resourceTypeID {
	case 1:
		return 100
	case 2:
		return 100
	case 3:
		return 100
	case 4:
		return 5
	case 5:
		return 6
	case 6:
		return 100
	case 7:
		return 5
	case 8:
		return 100
	case 9:
		return 5
	case 11:
		return 10
	case 12:
		return 6
	case 13:
		return 100
	case 16:
		return 4
	case 17:
		return 120
	case 18:
		return 100
	case 19:
		return 6
	}

	switch strings.ToLower(strings.TrimSpace(resourceType)) {
	case "rage", "focus", "energy", "runic power", "lunar power", "insanity", "pain":
		return 100
	case "fury":
		return 120
	case "combo points", "soul shards", "holy power":
		return 5
	case "runes", "chi", "essence":
		return 6
	case "maelstrom":
		return 10
	case "arcane charges":
		return 4
	default:
		return 0
	}
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

func buildResourceCaution(player types.ResourceUsageSummary, cohortGenerated, cohortWastePct float64, strategy resourceStrategy) string {
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
