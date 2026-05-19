package services

import "testing"

func TestBuildInsightRequestAttachesSpecProfile(t *testing.T) {
	req := GenerateReportRequest{
		Fight: FightSummary{
			Name:       "Test Boss",
			Difficulty: "Heroic",
			KillTime:   300,
		},
		Character: CharacterSummary{
			Name:  "Testdk",
			Class: "Death Knight",
			Spec:  "Blood",
		},
	}
	comparison := ComparisonResult{
		CohortStats: CohortStatistics{SampleSize: 10},
	}

	insightReq := buildInsightRequest(req, comparison, timelineFightData{}, nil)
	if insightReq.Context.SpecProfile.Label != "Blood Death Knight" {
		t.Fatalf("expected Blood Death Knight spec profile, got %s", insightReq.Context.SpecProfile.Label)
	}
	if insightReq.Context.SpecProfile.Role != "Tank" {
		t.Fatalf("expected Tank role, got %s", insightReq.Context.SpecProfile.Role)
	}
	if len(insightReq.Context.SpecProfile.KeyMechanics) == 0 {
		t.Fatalf("expected key mechanics")
	}
}
