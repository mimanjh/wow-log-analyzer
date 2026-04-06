package service

import (
	"fmt"
	"math"
	"sort"
	"time"

	"wow-log-analyzer/services/analysis-service/internal/types"
)

// calculateCastsPerMinute calculates casts per minute
func (s *AnalysisService) calculateCastsPerMinute(casts []types.CastEvent, duration time.Duration) types.CastsPerMinuteMetric {
	totalCasts := len(casts)
	fightMinutes := duration.Minutes()

	if fightMinutes <= 0 {
		return types.CastsPerMinuteMetric{
			Value:         0,
			TotalCasts:    totalCasts,
			FightDuration: fightMinutes,
			Confidence:    "low",
			Caution:       "Fight duration is zero or negative",
		}
	}

	castsPerMin := float64(totalCasts) / fightMinutes

	confidence := "high"
	var caution string
	if totalCasts < 10 {
		confidence = "medium"
		caution = "Low sample size of casts"
	}

	return types.CastsPerMinuteMetric{
		Value:         math.Round(castsPerMin*100) / 100, // round to 2 decimal places
		TotalCasts:    totalCasts,
		FightDuration: fightMinutes,
		Confidence:    confidence,
		Caution:       caution,
	}
}

// calculateMajorCDCount calculates major cooldown usage count
func (s *AnalysisService) calculateMajorCDCount(cooldowns []types.CooldownEvent) types.MajorCDCountMetric {
	count := 0
	for _, cd := range cooldowns {
		if cd.EventType == "start" && cd.Ability.IsMajorCD {
			count++
		}
	}

	confidence := "high"
	var caution string
	if count == 0 {
		confidence = "medium"
		caution = "No major cooldowns detected"
	}

	return types.MajorCDCountMetric{
		Value:         count,
		TotalCooldowns: count,
		Confidence:    confidence,
		Caution:       caution,
	}
}

// calculateMajorCDDrift calculates timing drift of major cooldowns
func (s *AnalysisService) calculateMajorCDDrift(cooldowns []types.CooldownEvent, _, _ time.Time) types.MajorCDDriftMetric {
	var drifts []float64

	for i := 0; i < len(cooldowns)-1; i++ {
		if cooldowns[i].EventType == "start" && cooldowns[i].Ability.IsMajorCD {
			// Find the next major CD
			for j := i + 1; j < len(cooldowns); j++ {
				if cooldowns[j].EventType == "start" && cooldowns[j].Ability.IsMajorCD {
					// Calculate expected interval (assuming CDs are used every 3 minutes)
					expectedInterval := 180.0 // 3 minutes in seconds
					actualInterval := cooldowns[j].Timestamp.Sub(cooldowns[i].Timestamp).Seconds()

					drift := math.Abs(actualInterval - expectedInterval)
					drifts = append(drifts, drift)
					break
				}
			}
		}
	}

	if len(drifts) == 0 {
		return types.MajorCDDriftMetric{
			Value:         0,
			TotalDrift:    0,
			CooldownCount: 0,
			Confidence:    "low",
			Caution:       "No major cooldown pairs found to calculate drift",
		}
	}

	totalDrift := 0.0
	for _, drift := range drifts {
		totalDrift += drift
	}
	averageDrift := totalDrift / float64(len(drifts))

	confidence := "high"
	var caution string
	if len(drifts) < 3 {
		confidence = "medium"
		caution = "Low sample size of cooldown pairs"
	}

	return types.MajorCDDriftMetric{
		Value:         math.Round(averageDrift*100) / 100, // round to 2 decimal places
		TotalDrift:    math.Round(totalDrift*100) / 100,
		CooldownCount: len(drifts),
		Confidence:    confidence,
		Caution:       caution,
	}
}

// calculateBuffUptime calculates buff uptime percentage
func (s *AnalysisService) calculateBuffUptime(buffEvents []types.BuffEvent, duration time.Duration) types.BuffUptimeMetric {
	if len(buffEvents) == 0 {
		return types.BuffUptimeMetric{
			Value:         0,
			TotalUptime:   0,
			FightDuration: duration.Seconds(),
			Confidence:    "low",
			Caution:       "No buff events found",
		}
	}

	// Group buff events by ability and target
	buffUptimes := make(map[string]time.Duration)

	for _, event := range buffEvents {
		if event.Ability.IsBuff {
			key := fmt.Sprintf("%d-%d", event.Ability.ID, event.TargetID)
			if event.EventType == "apply" {
				// Find corresponding remove event
				for _, removeEvent := range buffEvents {
					if removeEvent.EventType == "remove" &&
					   removeEvent.Ability.ID == event.Ability.ID &&
					   removeEvent.TargetID == event.TargetID &&
					   removeEvent.Timestamp.After(event.Timestamp) {
						uptime := removeEvent.Timestamp.Sub(event.Timestamp)
						buffUptimes[key] += uptime
						break
					}
				}
			}
		}
	}

	totalUptime := 0.0
	for _, uptime := range buffUptimes {
		totalUptime += uptime.Seconds()
	}

	fightDuration := duration.Seconds()
	uptimePercentage := (totalUptime / fightDuration) * 100

	confidence := "high"
	var caution string
	if len(buffUptimes) < 5 {
		confidence = "medium"
		caution = "Low number of buff applications detected"
	}

	return types.BuffUptimeMetric{
		Value:         math.Round(uptimePercentage*100) / 100, // round to 2 decimal places
		TotalUptime:   math.Round(totalUptime*100) / 100,
		FightDuration: fightDuration,
		Confidence:    confidence,
		Caution:       caution,
	}
}

// calculateDowntimePercentage calculates percentage of fight spent not doing damage/healing
func (s *AnalysisService) calculateDowntimePercentage(damageEvents []types.DamageEvent, healEvents []types.HealEvent, duration time.Duration) types.DowntimePercentageMetric {
	// Combine damage and heal events
	var allEvents []time.Time
	for _, event := range damageEvents {
		allEvents = append(allEvents, event.Timestamp)
	}
	for _, event := range healEvents {
		allEvents = append(allEvents, event.Timestamp)
	}

	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Before(allEvents[j])
	})

	if len(allEvents) == 0 {
		return types.DowntimePercentageMetric{
			Value:         100,
			TotalDowntime: duration.Seconds(),
			FightDuration: duration.Seconds(),
			Confidence:    "low",
			Caution:       "No damage or healing events found",
		}
	}

	// Calculate gaps between events (assuming 5 seconds of activity per event)
	activeTime := float64(len(allEvents)) * 5.0 // 5 seconds per event
	fightDuration := duration.Seconds()

	if activeTime > fightDuration {
		activeTime = fightDuration
	}

	downtime := fightDuration - activeTime
	downtimePercentage := (downtime / fightDuration) * 100

	confidence := "medium"
	var caution string
	if len(allEvents) < 50 {
		confidence = "low"
		caution = "Low sample size of damage/healing events"
	}

	return types.DowntimePercentageMetric{
		Value:         math.Round(downtimePercentage*100) / 100, // round to 2 decimal places
		TotalDowntime: math.Round(downtime*100) / 100,
		FightDuration: fightDuration,
		Confidence:    confidence,
		Caution:       caution,
	}
}
