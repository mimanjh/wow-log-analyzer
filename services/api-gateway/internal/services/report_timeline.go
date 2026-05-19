package services

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func (s *ReportService) GetAbilityTimeline(jobID string, abilityID int) (AbilityTimelineResponse, error) {
	if abilityID == 0 {
		return AbilityTimelineResponse{}, fmt.Errorf("abilityId is required")
	}

	s.jobMu.RLock()
	job, ok := s.jobs[jobID]
	s.jobMu.RUnlock()
	if !ok {
		return AbilityTimelineResponse{}, fmt.Errorf("report job %s not found", jobID)
	}
	if job.timeline == nil {
		return AbilityTimelineResponse{}, fmt.Errorf("ability timeline is not available for this job yet")
	}

	playerSeries := buildAbilityTimelineSeries(
		job.timeline.PlayerData,
		abilityID,
		job.timeline.Character.Name,
		fmt.Sprintf("%s %s", job.timeline.Character.Spec, job.timeline.Character.Class),
		"",
	)
	if len(playerSeries.CastsMS) == 0 {
		return AbilityTimelineResponse{}, fmt.Errorf("no cast timeline was available for this ability")
	}

	eliteSeries := make([]AbilityTimelineSeries, 0, len(job.timeline.EliteData))
	for index, eliteData := range job.timeline.EliteData {
		entry := job.timeline.EliteEntries[index]
		subtitle := strings.TrimSpace(fmt.Sprintf("%s %s", entry.Spec, entry.Class))
		if entry.Server != "" {
			subtitle = strings.TrimSpace(fmt.Sprintf("%s • %s", subtitle, entry.Server))
		}
		series := buildAbilityTimelineSeries(
			eliteData,
			abilityID,
			entry.Name,
			subtitle,
			entry.ReportURL,
		)
		if len(series.CastsMS) > 0 {
			eliteSeries = append(eliteSeries, series)
		}
	}

	abilityName := findAbilityName(job.timeline.PlayerData, abilityID)
	if abilityName == "" {
		for _, eliteData := range job.timeline.EliteData {
			abilityName = findAbilityName(eliteData, abilityID)
			if abilityName != "" {
				break
			}
		}
	}
	if abilityName == "" {
		abilityName = "Selected Ability"
	}

	return AbilityTimelineResponse{
		AbilityID:       abilityID,
		AbilityName:     abilityName,
		FightDurationMS: job.timeline.PlayerData.FightEnd.Sub(job.timeline.PlayerData.FightStart).Milliseconds(),
		Player:          playerSeries,
		Elite:           eliteSeries,
	}, nil
}

func (s *ReportService) GetResourceTimeline(jobID string, resourceTypeID int) (ResourceTimelineResponse, error) {
	s.jobMu.RLock()
	job, ok := s.jobs[jobID]
	s.jobMu.RUnlock()
	if !ok {
		return ResourceTimelineResponse{}, fmt.Errorf("report job %s not found", jobID)
	}
	if job.timeline == nil {
		return ResourceTimelineResponse{}, fmt.Errorf("resource timeline is not available for this job yet")
	}

	resourceType := findResourceType(job.timeline.PlayerData, resourceTypeID)
	if resourceType == "" {
		for _, eliteData := range job.timeline.EliteData {
			resourceType = findResourceType(eliteData, resourceTypeID)
			if resourceType != "" {
				break
			}
		}
	}
	if resourceType == "" {
		resourceType = "Selected Resource"
	}

	playerSeries := buildResourceTimelineSeries(
		job.timeline.PlayerData,
		resourceTypeID,
		job.timeline.Character.Name,
		fmt.Sprintf("%s %s", job.timeline.Character.Spec, job.timeline.Character.Class),
		"",
	)
	if len(playerSeries.Samples) == 0 {
		return ResourceTimelineResponse{}, fmt.Errorf("no resource timeline was available for this resource")
	}

	eliteSeries := make([]ResourceTimelineSeries, 0, len(job.timeline.EliteData))
	for index, eliteData := range job.timeline.EliteData {
		entry := job.timeline.EliteEntries[index]
		subtitle := strings.TrimSpace(fmt.Sprintf("%s %s", entry.Spec, entry.Class))
		if entry.Server != "" {
			subtitle = strings.TrimSpace(fmt.Sprintf("%s - %s", subtitle, entry.Server))
		}
		series := buildResourceTimelineSeries(
			eliteData,
			resourceTypeID,
			entry.Name,
			subtitle,
			entry.ReportURL,
		)
		if len(series.Samples) > 0 {
			eliteSeries = append(eliteSeries, series)
		}
	}

	return ResourceTimelineResponse{
		ResourceTypeID:  resourceTypeID,
		ResourceType:    resourceType,
		FightDurationMS: job.timeline.PlayerData.FightEnd.Sub(job.timeline.PlayerData.FightStart).Milliseconds(),
		Player:          playerSeries,
		Elite:           eliteSeries,
	}, nil
}

func (s *ReportService) GetBuffTimeline(jobID string, abilityID int) (BuffTimelineResponse, error) {
	if abilityID == 0 {
		return BuffTimelineResponse{}, fmt.Errorf("abilityId is required")
	}

	s.jobMu.RLock()
	job, ok := s.jobs[jobID]
	s.jobMu.RUnlock()
	if !ok {
		return BuffTimelineResponse{}, fmt.Errorf("report job %s not found", jobID)
	}
	if job.timeline == nil {
		return BuffTimelineResponse{}, fmt.Errorf("buff timeline is not available for this job yet")
	}

	playerSeries := buildBuffTimelineSeries(
		job.timeline.PlayerData,
		abilityID,
		job.timeline.Character.Name,
		fmt.Sprintf("%s %s", job.timeline.Character.Spec, job.timeline.Character.Class),
		"",
	)

	eliteSeries := make([]BuffTimelineSeries, 0, len(job.timeline.EliteData))
	for index, eliteData := range job.timeline.EliteData {
		entry := job.timeline.EliteEntries[index]
		subtitle := strings.TrimSpace(fmt.Sprintf("%s %s", entry.Spec, entry.Class))
		if entry.Server != "" {
			subtitle = strings.TrimSpace(fmt.Sprintf("%s - %s", subtitle, entry.Server))
		}
		series := buildBuffTimelineSeries(
			eliteData,
			abilityID,
			entry.Name,
			subtitle,
			entry.ReportURL,
		)
		if len(series.Windows) > 0 {
			eliteSeries = append(eliteSeries, series)
		}
	}
	if len(playerSeries.Windows) == 0 && len(eliteSeries) == 0 {
		return BuffTimelineResponse{}, fmt.Errorf("no buff timeline was available for this ability")
	}

	abilityName := findBuffName(job.timeline.PlayerData, abilityID)
	if abilityName == "" {
		for _, eliteData := range job.timeline.EliteData {
			abilityName = findBuffName(eliteData, abilityID)
			if abilityName != "" {
				break
			}
		}
	}
	if abilityName == "" {
		abilityName = "Selected Buff"
	}

	return BuffTimelineResponse{
		AbilityID:       abilityID,
		AbilityName:     abilityName,
		FightDurationMS: job.timeline.PlayerData.FightEnd.Sub(job.timeline.PlayerData.FightStart).Milliseconds(),
		Player:          playerSeries,
		Elite:           eliteSeries,
	}, nil
}

func buildAbilityTimelineSeries(data timelineFightData, abilityID int, label, subtitle, reportURL string) AbilityTimelineSeries {
	casts := make([]int64, 0)
	for _, event := range data.CastEvents {
		if event.Ability.ID != abilityID {
			continue
		}
		casts = append(casts, event.Timestamp.Sub(data.FightStart).Milliseconds())
	}
	if len(casts) == 0 {
		for _, event := range data.CooldownEvents {
			if event.Ability.ID != abilityID {
				continue
			}
			casts = append(casts, event.Timestamp.Sub(data.FightStart).Milliseconds())
		}
	}
	if len(casts) == 0 {
		for _, event := range data.DamageEvents {
			if event.Ability.ID != abilityID {
				continue
			}
			casts = append(casts, event.Timestamp.Sub(data.FightStart).Milliseconds())
		}
	}

	return AbilityTimelineSeries{
		Label:     label,
		Subtitle:  subtitle,
		ReportURL: reportURL,
		CastsMS:   casts,
	}
}

func findAbilityName(data timelineFightData, abilityID int) string {
	for _, event := range data.CastEvents {
		if event.Ability.ID == abilityID {
			return event.Ability.Name
		}
	}
	for _, event := range data.DamageEvents {
		if event.Ability.ID == abilityID {
			return event.Ability.Name
		}
	}
	for _, event := range data.CooldownEvents {
		if event.Ability.ID == abilityID {
			return event.Ability.Name
		}
	}
	return ""
}

func buildBuffTimelineSeries(data timelineFightData, abilityID int, label, subtitle, reportURL string) BuffTimelineSeries {
	windows := buffWindows(data, abilityID)
	timelineWindows := make([]BuffTimelineWindow, 0, len(windows))
	durationMS := data.FightEnd.Sub(data.FightStart).Milliseconds()
	if durationMS < 0 {
		durationMS = 0
	}

	for _, window := range windows {
		startMS := window.start.Sub(data.FightStart).Milliseconds()
		endMS := window.end.Sub(data.FightStart).Milliseconds()
		if startMS < 0 {
			startMS = 0
		}
		if endMS < 0 {
			endMS = 0
		}
		if durationMS > 0 {
			if startMS > durationMS {
				startMS = durationMS
			}
			if endMS > durationMS {
				endMS = durationMS
			}
		}
		if endMS <= startMS {
			continue
		}
		timelineWindows = append(timelineWindows, BuffTimelineWindow{
			StartMS: startMS,
			EndMS:   endMS,
		})
	}

	return BuffTimelineSeries{
		Label:     label,
		Subtitle:  subtitle,
		ReportURL: reportURL,
		Windows:   timelineWindows,
	}
}

func findBuffName(data timelineFightData, abilityID int) string {
	for _, event := range data.BuffEvents {
		if event.Ability.ID == abilityID {
			return event.Ability.Name
		}
	}
	return ""
}

func buildResourceTimelineSeries(data timelineFightData, resourceTypeID int, label, subtitle, reportURL string) ResourceTimelineSeries {
	samples := make([]ResourceTimelineSample, 0)
	wasteMarkers := make([]int64, 0)

	for _, event := range data.ResourceEvents {
		if event.ResourceTypeID != resourceTypeID {
			continue
		}

		timestampMS := event.Timestamp.Sub(data.FightStart).Milliseconds()
		if timestampMS < 0 {
			timestampMS = 0
		}
		sample := ResourceTimelineSample{
			TimestampMS: timestampMS,
			Value:       event.Amount,
			MaxValue:    event.MaxAmount,
			Waste:       event.Waste,
		}
		if sample.Value == 0 && event.Change > 0 {
			sample.Value = event.Change
		}
		maxValue := sample.MaxValue
		if maxValue <= 0 {
			maxValue = defaultResourceMaxValue(event.ResourceTypeID, event.ResourceType)
			sample.MaxValue = maxValue
		}
		if event.Waste > 0 && maxValue > 0 {
			sample.Value = maxValue
		}
		samples = append(samples, sample)

		if event.Waste > 0 || (maxValue > 0 && sample.Value >= maxValue) {
			wasteMarkers = append(wasteMarkers, timestampMS)
		}
	}

	sort.Slice(samples, func(i, j int) bool {
		return samples[i].TimestampMS < samples[j].TimestampMS
	})
	sort.Slice(wasteMarkers, func(i, j int) bool {
		return wasteMarkers[i] < wasteMarkers[j]
	})

	return ResourceTimelineSeries{
		Label:        label,
		Subtitle:     subtitle,
		ReportURL:    reportURL,
		DurationMS:   data.FightEnd.Sub(data.FightStart).Milliseconds(),
		Samples:      samples,
		WasteMarkers: wasteMarkers,
	}
}

func defaultResourceMaxValue(resourceTypeID int, resourceType string) float64 {
	switch resourceTypeID {
	case 1, 2, 3, 6, 8, 13, 18:
		return 100
	case 4, 7, 9:
		return 5
	case 5, 12, 19:
		return 6
	case 11:
		return 10
	case 16:
		return 4
	case 17:
		return 120
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

func findResourceType(data timelineFightData, resourceTypeID int) string {
	for _, event := range data.ResourceEvents {
		if event.ResourceTypeID == resourceTypeID && event.ResourceType != "" {
			return event.ResourceType
		}
	}
	return ""
}

type buffWindow struct {
	start time.Time
	end   time.Time
}

func buffWindows(data timelineFightData, abilityID int) []buffWindow {
	if len(data.BuffEvents) == 0 {
		return nil
	}

	windows := make([]buffWindow, 0)
	active := false
	start := data.FightStart

	for _, event := range data.BuffEvents {
		if event.Ability.ID != abilityID {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(event.EventType)) {
		case "apply":
			if !active {
				active = true
				start = event.Timestamp
			}
		case "refresh":
			if !active {
				active = true
				start = event.Timestamp
			}
		case "remove":
			if active {
				windows = append(windows, buffWindow{start: start, end: event.Timestamp})
				active = false
			}
		}
	}

	if active {
		windows = append(windows, buffWindow{start: start, end: data.FightEnd})
	}

	return windows
}
