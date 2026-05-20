package services

import (
	"testing"
	"time"
)

func TestReportService_GetBuffTimelineBuildsPlayerAndEliteWindows(t *testing.T) {
	service := NewReportService("", "", "", nil)
	start := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	jobID := "job-buff-timeline"

	service.setJob(ReportJob{
		ID:        jobID,
		Status:    ReportJobCompleted,
		Fight:     FightSummary{ID: 1, Name: "Boss"},
		Character: CharacterSummary{ID: 10, Name: "Player", Class: "Death Knight", Spec: "Blood"},
		timeline: &reportTimelineData{
			Character: CharacterSummary{ID: 10, Name: "Player", Class: "Death Knight", Spec: "Blood"},
			PlayerData: timelineFightData{
				FightStart: start,
				FightEnd:   start.Add(60 * time.Second),
				BuffEvents: []timelineBuffEvent{
					{Timestamp: start.Add(5 * time.Second), Ability: timelineAbility{ID: 81256, Name: "Dancing Rune Weapon"}, EventType: "apply"},
					{Timestamp: start.Add(25 * time.Second), Ability: timelineAbility{ID: 81256, Name: "Dancing Rune Weapon"}, EventType: "remove"},
					{Timestamp: start.Add(40 * time.Second), Ability: timelineAbility{ID: 81256, Name: "Dancing Rune Weapon"}, EventType: "apply"},
				},
			},
			EliteData: []timelineFightData{
				{
					FightStart: start,
					FightEnd:   start.Add(60 * time.Second),
					BuffEvents: []timelineBuffEvent{
						{Timestamp: start.Add(10 * time.Second), Ability: timelineAbility{ID: 81256, Name: "Dancing Rune Weapon"}, EventType: "apply"},
						{Timestamp: start.Add(30 * time.Second), Ability: timelineAbility{ID: 81256, Name: "Dancing Rune Weapon"}, EventType: "remove"},
					},
				},
			},
			EliteEntries: []CohortEntry{
				{Name: "Elite", Class: "Death Knight", Spec: "Blood", Server: "Area 52", ReportURL: "https://example.test/report"},
			},
		},
	})

	timeline, err := service.GetBuffTimeline(jobID, 81256)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if timeline.AbilityName != "Dancing Rune Weapon" {
		t.Fatalf("expected buff name Dancing Rune Weapon, got %s", timeline.AbilityName)
	}
	if timeline.FightDurationMS != 60000 {
		t.Fatalf("expected duration 60000ms, got %d", timeline.FightDurationMS)
	}
	if len(timeline.Player.Windows) != 2 {
		t.Fatalf("expected two player buff windows, got %d", len(timeline.Player.Windows))
	}
	if timeline.Player.Windows[0].StartMS != 5000 || timeline.Player.Windows[0].EndMS != 25000 {
		t.Fatalf("unexpected first player window: %#v", timeline.Player.Windows[0])
	}
	if timeline.Player.Windows[1].StartMS != 40000 || timeline.Player.Windows[1].EndMS != 60000 {
		t.Fatalf("unexpected second player window: %#v", timeline.Player.Windows[1])
	}
	if len(timeline.Elite) != 1 {
		t.Fatalf("expected one elite series, got %d", len(timeline.Elite))
	}
	if timeline.Elite[0].Windows[0].StartMS != 10000 || timeline.Elite[0].Windows[0].EndMS != 30000 {
		t.Fatalf("unexpected elite window: %#v", timeline.Elite[0].Windows[0])
	}
}
