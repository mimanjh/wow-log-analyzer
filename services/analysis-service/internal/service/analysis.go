package service

import (
	"fmt"
	"math"
	"sort"
	"time"

	"wow-log-analyzer/services/analysis-service/types"
)

// AnalysisService provides fight analysis and comparison functionality
type AnalysisService struct{}

// NewAnalysisService creates a new analysis service
func NewAnalysisService() *AnalysisService {
	return &AnalysisService{}
}

// AnalyzeMember derives every cacheable summary for one player's fight in a single
// pass: aggregate metrics, per-ability usage, per-buff uptime, per-resource
// usage, plus the per-ability use-time and per-buff gap arrays the AI insight
// builder needs. The returned MemberContext is what api-gateway caches in place
// of the raw 1MB fight JSON — same logical content, ~5-20KB serialized.
func (s *AnalysisService) AnalyzeMember(data types.PlayerFightData, characterClass, characterSpec string) (*types.MemberContext, error) {
	metrics, err := s.AnalyzePlayerFight(data)
	if err != nil {
		return nil, err
	}
	strategy := resolveResourceStrategy(characterClass, characterSpec)
	return &types.MemberContext{
		Metrics:               *metrics,
		AbilityUsage:          summarizeAbilityUsage(data),
		BuffUptime:            summarizeBuffUptime(data),
		ResourceUsage:         summarizeResourceUsage(data, strategy),
		TalentImportCode:      data.TalentImportCode,
		TalentCalculatorURL:   data.TalentCalculatorURL,
		AbilityFirstUseTimes:  collectAbilityFirstUseTimes(data, 3),
		BuffLargestGapSeconds: collectBuffLargestGaps(data),
	}, nil
}

// collectAbilityFirstUseTimes returns up to `limit` earliest cast timestamps
// (seconds into the fight) for every ability the player cast. Used by the AI
// insight builder to compare timing against the cohort median.
func collectAbilityFirstUseTimes(data types.PlayerFightData, limit int) map[int][]float64 {
	if limit <= 0 {
		return nil
	}
	result := make(map[int][]float64)
	for _, event := range data.CastEvents {
		if event.Ability.ID == 0 {
			continue
		}
		offset := event.Timestamp.Sub(data.FightStart).Seconds()
		if offset < 0 {
			offset = 0
		}
		existing := result[event.Ability.ID]
		if len(existing) >= limit {
			continue
		}
		result[event.Ability.ID] = append(existing, offset)
	}
	return result
}

// collectBuffLargestGaps returns the longest stretch (seconds) the player went
// without each buff being active. Computed once per fight so the AI insight
// builder doesn't have to re-scan raw buff events.
func collectBuffLargestGaps(data types.PlayerFightData) map[int]float64 {
	type windowState struct {
		windows [][2]time.Time // start, end pairs
		active  bool
		opened  time.Time
	}
	states := make(map[int]*windowState)
	for _, event := range data.BuffEvents {
		if event.Ability.ID == 0 {
			continue
		}
		if event.TargetID != 0 && event.TargetID != data.PlayerID {
			continue
		}
		ts := event.Timestamp
		if ts.Before(data.FightStart) {
			ts = data.FightStart
		} else if ts.After(data.FightEnd) {
			ts = data.FightEnd
		}
		state, ok := states[event.Ability.ID]
		if !ok {
			state = &windowState{}
			states[event.Ability.ID] = state
		}
		switch event.EventType {
		case "apply", "refresh":
			if !state.active {
				state.active = true
				state.opened = ts
			}
		case "remove":
			if state.active && ts.After(state.opened) {
				state.windows = append(state.windows, [2]time.Time{state.opened, ts})
				state.active = false
			}
		}
	}
	result := make(map[int]float64, len(states))
	for abilityID, state := range states {
		if state.active && data.FightEnd.After(state.opened) {
			state.windows = append(state.windows, [2]time.Time{state.opened, data.FightEnd})
		}
		largestGap := 0.0
		if len(state.windows) == 0 {
			largestGap = data.FightEnd.Sub(data.FightStart).Seconds()
		} else {
			previous := data.FightStart
			for _, window := range state.windows {
				gap := window[0].Sub(previous).Seconds()
				if gap > largestGap {
					largestGap = gap
				}
				previous = window[1]
			}
			tailGap := data.FightEnd.Sub(previous).Seconds()
			if tailGap > largestGap {
				largestGap = tailGap
			}
		}
		if largestGap < 0 {
			largestGap = 0
		}
		result[abilityID] = largestGap
	}
	return result
}

// CompareMemberContexts builds a full ComparisonResult from precomputed member
// contexts — no raw fight events required. This is the cached-input variant of
// the legacy raw-input compare path.
func (s *AnalysisService) CompareMemberContexts(player types.MemberContext, cohort []types.MemberContext, characterClass, characterSpec string) (*types.ComparisonResult, error) {
	cohortMetrics := make([]types.PlayerFightMetrics, len(cohort))
	cohortAbility := make([]map[int]types.AbilityUsageSummary, len(cohort))
	cohortBuff := make([]map[int]types.BuffUptimeSummary, len(cohort))
	cohortResource := make([]map[int]types.ResourceUsageSummary, len(cohort))
	for i, c := range cohort {
		cohortMetrics[i] = c.Metrics
		cohortAbility[i] = c.AbilityUsage
		cohortBuff[i] = c.BuffUptime
		cohortResource[i] = c.ResourceUsage
	}

	comparison, err := s.CompareAgainstCohort(player.Metrics, cohortMetrics)
	if err != nil {
		return nil, err
	}

	comparison.AbilityUsage = s.CompareAbilityUsages(player.AbilityUsage, cohortAbility, characterClass, characterSpec)
	comparison.BuffUptimes = s.CompareBuffUptimes(player.BuffUptime, cohortBuff)
	comparison.ResourceUsage = s.CompareResourceUsages(player.ResourceUsage, cohortResource, characterClass, characterSpec)
	return comparison, nil
}

// AnalyzePlayerFight calculates all metrics for a single player's fight
func (s *AnalysisService) AnalyzePlayerFight(data types.PlayerFightData) (*types.PlayerFightMetrics, error) {
	if data.PlayerID == 0 {
		return nil, fmt.Errorf("player ID is required")
	}
	if data.FightID == 0 {
		return nil, fmt.Errorf("fight ID is required")
	}

	duration := data.FightEnd.Sub(data.FightStart)

	metrics := &types.PlayerFightMetrics{
		PlayerID:   data.PlayerID,
		FightID:    data.FightID,
		FightStart: data.FightStart,
		FightEnd:   data.FightEnd,
		Duration:   duration,
	}

	// Calculate each metric
	metrics.CastsPerMin = s.calculateCastsPerMinute(data.CastEvents, duration)
	metrics.MajorCDCount = s.calculateMajorCDCount(data.CooldownEvents)
	metrics.MajorCDDrift = s.calculateMajorCDDrift(data.CooldownEvents, data.FightStart, data.FightEnd)
	metrics.BuffUptime = s.calculateBuffUptime(data)
	metrics.DowntimePct = s.calculateDowntimePercentage(data.DamageEvents, data.HealEvents, duration)

	return metrics, nil
}

// CompareAgainstCohort compares player metrics against a cohort
func (s *AnalysisService) CompareAgainstCohort(playerMetrics types.PlayerFightMetrics, cohortMetrics []types.PlayerFightMetrics) (*types.ComparisonResult, error) {
	if len(cohortMetrics) == 0 {
		return nil, fmt.Errorf("cohort must contain at least one player")
	}

	cohortStats := s.calculateCohortStatistics(cohortMetrics)
	deltas := s.calculateDeltas(playerMetrics, cohortStats)
	rankings := s.calculateRankings(playerMetrics, cohortMetrics)

	return &types.ComparisonResult{
		PlayerMetrics: playerMetrics,
		CohortStats:   cohortStats,
		Deltas:        deltas,
		Rankings:      rankings,
	}, nil
}

// calculateCohortStatistics calculates statistics for each metric across the cohort
func (s *AnalysisService) calculateCohortStatistics(cohortMetrics []types.PlayerFightMetrics) types.CohortStatistics {
	stats := types.CohortStatistics{
		SampleSize: len(cohortMetrics),
	}

	if len(cohortMetrics) == 0 {
		return stats
	}

	// Extract values for each metric
	castsPerMin := make([]float64, len(cohortMetrics))
	majorCDCount := make([]float64, len(cohortMetrics))
	majorCDDrift := make([]float64, len(cohortMetrics))
	buffUptime := make([]float64, len(cohortMetrics))
	downtimePct := make([]float64, len(cohortMetrics))

	for i, metric := range cohortMetrics {
		castsPerMin[i] = metric.CastsPerMin.Value
		majorCDCount[i] = float64(metric.MajorCDCount.Value)
		majorCDDrift[i] = metric.MajorCDDrift.Value
		buffUptime[i] = metric.BuffUptime.Value
		downtimePct[i] = metric.DowntimePct.Value
	}

	stats.CastsPerMin = s.calculateMetricStats(castsPerMin)
	stats.MajorCDCount = s.calculateMetricStats(majorCDCount)
	stats.MajorCDDrift = s.calculateMetricStats(majorCDDrift)
	stats.BuffUptime = s.calculateMetricStats(buffUptime)
	stats.DowntimePct = s.calculateMetricStats(downtimePct)

	return stats
}

// calculateMetricStats calculates statistical measures for a slice of values
func (s *AnalysisService) calculateMetricStats(values []float64) types.CohortMetricStats {
	if len(values) == 0 {
		return types.CohortMetricStats{}
	}

	sort.Float64s(values)

	mean := s.calculateMean(values)
	median := s.calculateMedian(values)
	stdDev := s.calculateStdDev(values, mean)
	min := values[0]
	max := values[len(values)-1]
	p25 := s.calculatePercentile(values, 25)
	p75 := s.calculatePercentile(values, 75)
	p95 := s.calculatePercentile(values, 95)

	return types.CohortMetricStats{
		Mean:   math.Round(mean*100) / 100,
		Median: math.Round(median*100) / 100,
		StdDev: math.Round(stdDev*100) / 100,
		Min:    math.Round(min*100) / 100,
		Max:    math.Round(max*100) / 100,
		P25:    math.Round(p25*100) / 100,
		P75:    math.Round(p75*100) / 100,
		P95:    math.Round(p95*100) / 100,
	}
}

// calculateDeltas calculates the differences between player and cohort
func (s *AnalysisService) calculateDeltas(playerMetrics types.PlayerFightMetrics, cohortStats types.CohortStatistics) types.MetricDeltas {
	return types.MetricDeltas{
		CastsPerMin:  s.calculateMetricDelta(playerMetrics.CastsPerMin.Value, cohortStats.CastsPerMin.Median, playerMetrics.CastsPerMin.Confidence, cohortStats.CastsPerMin),
		MajorCDCount: s.calculateMetricDelta(float64(playerMetrics.MajorCDCount.Value), cohortStats.MajorCDCount.Median, playerMetrics.MajorCDCount.Confidence, cohortStats.MajorCDCount),
		MajorCDDrift: s.calculateMetricDelta(playerMetrics.MajorCDDrift.Value, cohortStats.MajorCDDrift.Median, playerMetrics.MajorCDDrift.Confidence, cohortStats.MajorCDDrift),
		BuffUptime:   s.calculateMetricDelta(playerMetrics.BuffUptime.Value, cohortStats.BuffUptime.Median, playerMetrics.BuffUptime.Confidence, cohortStats.BuffUptime),
		DowntimePct:  s.calculateMetricDelta(playerMetrics.DowntimePct.Value, cohortStats.DowntimePct.Median, playerMetrics.DowntimePct.Confidence, cohortStats.DowntimePct),
	}
}

// calculateMetricDelta calculates the delta for a single metric
func (s *AnalysisService) calculateMetricDelta(playerValue, cohortValue float64, confidence string, cohortStats types.CohortMetricStats) types.MetricDelta {
	difference := playerValue - cohortValue

	// Calculate percentile based on the cohort distribution
	var percentile float64
	if cohortStats.StdDev > 0 {
		// Use z-score to estimate percentile
		zScore := (playerValue - cohortStats.Mean) / cohortStats.StdDev
		percentile = s.normalCDF(zScore) * 100
	} else {
		// If no variance, check if above/below/equal to mean
		if playerValue > cohortStats.Mean {
			percentile = 100
		} else if playerValue < cohortStats.Mean {
			percentile = 0
		} else {
			percentile = 50
		}
	}

	return types.MetricDelta{
		PlayerValue: playerValue,
		CohortValue: cohortValue,
		Difference:  math.Round(difference*100) / 100,
		Percentile:  math.Round(percentile*100) / 100,
		Confidence:  confidence,
	}
}

// calculateRankings calculates percentile rankings for each metric
func (s *AnalysisService) calculateRankings(playerMetrics types.PlayerFightMetrics, cohortMetrics []types.PlayerFightMetrics) types.MetricRankings {
	return types.MetricRankings{
		CastsPerMin:  s.calculatePercentileRanking(playerMetrics.CastsPerMin.Value, cohortMetrics, func(m types.PlayerFightMetrics) float64 { return m.CastsPerMin.Value }),
		MajorCDCount: s.calculatePercentileRanking(float64(playerMetrics.MajorCDCount.Value), cohortMetrics, func(m types.PlayerFightMetrics) float64 { return float64(m.MajorCDCount.Value) }),
		MajorCDDrift: s.calculatePercentileRanking(playerMetrics.MajorCDDrift.Value, cohortMetrics, func(m types.PlayerFightMetrics) float64 { return m.MajorCDDrift.Value }),
		BuffUptime:   s.calculatePercentileRanking(playerMetrics.BuffUptime.Value, cohortMetrics, func(m types.PlayerFightMetrics) float64 { return m.BuffUptime.Value }),
		DowntimePct:  s.calculatePercentileRanking(playerMetrics.DowntimePct.Value, cohortMetrics, func(m types.PlayerFightMetrics) float64 { return m.DowntimePct.Value }),
	}
}

// calculatePercentileRanking calculates the percentile ranking for a value
func (s *AnalysisService) calculatePercentileRanking(playerValue float64, cohortMetrics []types.PlayerFightMetrics, extractor func(types.PlayerFightMetrics) float64) float64 {
	values := make([]float64, len(cohortMetrics))
	for i, metric := range cohortMetrics {
		values[i] = extractor(metric)
	}

	return s.calculatePercentileFromValues(playerValue, values)
}
