package client

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"

	"wow-log-analyzer/services/log-service/internal/config"
	"wow-log-analyzer/services/log-service/internal/types"
)

type mockHTTPClient struct {
	responseBody string
	statusCode   int
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(bytes.NewReader([]byte(m.responseBody))),
		Header:     make(http.Header),
	}, nil
}

func TestWCLHTTPClient_GetReportMetadata(t *testing.T) {
	mockResponse := `{
		"data": {
			"reportData": {
				"report": {
					"code": "ABC123",
					"title": "Test Report",
					"startTime": 1700000000000,
					"endTime": 1700003600000,
					"zone": {
						"id": 1001,
						"name": "Test Zone"
					}
				}
			}
		}
	}`

	client := &WCLHTTPClient{
		config: config.WCLConfig{
			BaseURL: "https://www.warcraftlogs.com/api/v2",
		},
		httpClient: &mockHTTPClient{
			responseBody: mockResponse,
			statusCode:   200,
		},
	}

	report, err := client.GetReportMetadata("ABC123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := &types.NormalizedReport{
		ID:        "ABC123",
		Title:     "Test Report",
		StartTime: time.UnixMilli(1700000000000),
		EndTime:   time.UnixMilli(1700003600000),
		Zone: types.Zone{
			ID:   1001,
			Name: "Test Zone",
		},
	}

	if report.ID != expected.ID {
		t.Errorf("Expected ID %s, got %s", expected.ID, report.ID)
	}
	if report.Title != expected.Title {
		t.Errorf("Expected Title %s, got %s", expected.Title, report.Title)
	}
	if report.StartTime != expected.StartTime {
		t.Errorf("Expected StartTime %v, got %v", expected.StartTime, report.StartTime)
	}
	if report.EndTime != expected.EndTime {
		t.Errorf("Expected EndTime %v, got %v", expected.EndTime, report.EndTime)
	}
	if report.Zone.ID != expected.Zone.ID {
		t.Errorf("Expected Zone ID %d, got %d", expected.Zone.ID, report.Zone.ID)
	}
	if report.Zone.Name != expected.Zone.Name {
		t.Errorf("Expected Zone Name %s, got %s", expected.Zone.Name, report.Zone.Name)
	}
}

func TestWCLHTTPClient_GetFights(t *testing.T) {
	mockResponse := `{
		"data": {
			"reportData": {
				"report": {
					"fights": [
						{
							"id": 1,
							"name": "Test Boss",
							"startTime": 1700000000000,
							"endTime": 1700003600000,
							"encounterID": 1234,
							"difficulty": 3,
							"kill": true,
							"bossPercentage": 0
						}
					]
				}
			}
		}
	}`

	client := &WCLHTTPClient{
		config: config.WCLConfig{
			BaseURL: "https://www.warcraftlogs.com/api/v2",
		},
		httpClient: &mockHTTPClient{
			responseBody: mockResponse,
			statusCode:   200,
		},
	}

	fights, err := client.GetFights("ABC123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(fights) != 1 {
		t.Fatalf("Expected 1 fight, got %d", len(fights))
	}

	fight := fights[0]
	expected := types.NormalizedFight{
		ID:          1,
		Name:        "Test Boss",
		StartTime:   time.UnixMilli(1700000000000),
		EndTime:     time.UnixMilli(1700003600000),
		EncounterID: 1234,
		Difficulty:  "Normal",
		Kill:        true,
		BossPercent: 0,
	}

	if fight.ID != expected.ID {
		t.Errorf("Expected ID %d, got %d", expected.ID, fight.ID)
	}
	if fight.Name != expected.Name {
		t.Errorf("Expected Name %s, got %s", expected.Name, fight.Name)
	}
	if fight.StartTime != expected.StartTime {
		t.Errorf("Expected StartTime %v, got %v", expected.StartTime, fight.StartTime)
	}
	if fight.EndTime != expected.EndTime {
		t.Errorf("Expected EndTime %v, got %v", expected.EndTime, fight.EndTime)
	}
	if fight.EncounterID != expected.EncounterID {
		t.Errorf("Expected EncounterID %d, got %d", expected.EncounterID, fight.EncounterID)
	}
	if fight.Difficulty != expected.Difficulty {
		t.Errorf("Expected Difficulty %s, got %s", expected.Difficulty, fight.Difficulty)
	}
	if fight.Kill != expected.Kill {
		t.Errorf("Expected Kill %t, got %t", expected.Kill, fight.Kill)
	}
	if fight.BossPercent != expected.BossPercent {
		t.Errorf("Expected BossPercent %f, got %f", expected.BossPercent, fight.BossPercent)
	}
}

func TestWCLHTTPClient_normalizeDifficulty(t *testing.T) {
	client := &WCLHTTPClient{}

	tests := []struct {
		input    int
		expected string
	}{
		{1, "LFR"},
		{2, "Normal"},
		{3, "Normal"},
		{4, "Heroic"},
		{5, "Mythic"},
		{17, "LFR"},
		{99, "Unknown"},
	}

	for _, test := range tests {
		result := client.normalizeDifficulty(test.input)
		if result != test.expected {
			t.Errorf("For input %d, expected %s, got %s", test.input, test.expected, result)
		}
	}
}

func TestNormalizeResourceEvents_IncludesNestedSourceClassResources(t *testing.T) {
	reportStart := int64(1700000000000)
	raw := []map[string]interface{}{
		{
			"timestamp":          float64(1000),
			"sourceID":           float64(10),
			"resourceChangeType": float64(6),
			"resourceChange":     float64(15),
			"waste":              float64(0),
			"sourceResources": map[string]interface{}{
				"resourceType":   float64(6),
				"resourceAmount": float64(45),
				"resourceCap":    float64(100),
				"additionalResources": []interface{}{
					map[string]interface{}{
						"resourceType":   float64(5),
						"resourceAmount": float64(3),
						"resourceCap":    float64(6),
						"resourceCost":   float64(1),
					},
				},
			},
		},
	}

	events := normalizeResourceEvents(reportStart, raw, 10)
	runicPower, ok := findResourceEvent(events, 6)
	if !ok {
		t.Fatalf("expected runic power resource event")
	}
	if runicPower.Amount != 45 {
		t.Fatalf("expected runic power amount 45, got %f", runicPower.Amount)
	}
	if runicPower.Change != 15 {
		t.Fatalf("expected runic power change 15, got %f", runicPower.Change)
	}
	if runicPower.MaxAmount != 100 {
		t.Fatalf("expected runic power max 100, got %f", runicPower.MaxAmount)
	}

	runes, ok := findResourceEvent(events, 5)
	if !ok {
		t.Fatalf("expected rune resource event from nested class resources")
	}
	if runes.ResourceType != "Runes" {
		t.Fatalf("expected rune resource type, got %s", runes.ResourceType)
	}
	if runes.Amount != 3 {
		t.Fatalf("expected rune amount 3, got %f", runes.Amount)
	}
	if runes.MaxAmount != 6 {
		t.Fatalf("expected rune max 6, got %f", runes.MaxAmount)
	}
	if runes.Change != -1 {
		t.Fatalf("expected rune cost as change -1, got %f", runes.Change)
	}
}

func TestNormalizeResourceEvents_IncludesNestedTargetComboPoints(t *testing.T) {
	reportStart := int64(1700000000000)
	raw := []map[string]interface{}{
		{
			"timestamp": float64(2000),
			"sourceID":  float64(99),
			"targetID":  float64(10),
			"targetResources": map[string]interface{}{
				"resourceType":   float64(3),
				"resourceAmount": float64(70),
				"resourceCap":    float64(100),
				"additionalResources": map[string]interface{}{
					"resourceType":   float64(4),
					"resourceAmount": float64(4),
					"resourceCap":    float64(5),
				},
			},
		},
	}

	events := normalizeResourceEvents(reportStart, raw, 10)
	comboPoints, ok := findResourceEvent(events, 4)
	if !ok {
		t.Fatalf("expected combo point resource event from nested target resources")
	}
	if comboPoints.ResourceType != "Combo Points" {
		t.Fatalf("expected combo point resource type, got %s", comboPoints.ResourceType)
	}
	if comboPoints.Amount != 4 {
		t.Fatalf("expected combo point amount 4, got %f", comboPoints.Amount)
	}
	if comboPoints.MaxAmount != 5 {
		t.Fatalf("expected combo point max 5, got %f", comboPoints.MaxAmount)
	}
}

func TestNormalizeResourceEvents_IncludesCastClassResourcesWithoutManaFallback(t *testing.T) {
	reportStart := int64(1700000000000)
	raw := []map[string]interface{}{
		{
			"timestamp": float64(3000),
			"type":      "cast",
			"sourceID":  float64(10),
			"classResources": []interface{}{
				map[string]interface{}{
					"type":   float64(5),
					"amount": float64(2),
					"max":    float64(6),
				},
				map[string]interface{}{
					"type":   "4",
					"amount": float64(5),
					"max":    float64(5),
				},
			},
		},
	}

	events := normalizeResourceEvents(reportStart, raw, 10)
	if _, ok := findResourceEvent(events, 0); ok {
		t.Fatalf("did not expect cast event type string to create a mana resource event")
	}
	runes, ok := findResourceEvent(events, 5)
	if !ok {
		t.Fatalf("expected rune resource event from cast class resources")
	}
	if runes.Amount != 2 {
		t.Fatalf("expected rune amount 2, got %f", runes.Amount)
	}
	comboPoints, ok := findResourceEvent(events, 4)
	if !ok {
		t.Fatalf("expected combo point resource event from cast class resources")
	}
	if comboPoints.Amount != 5 {
		t.Fatalf("expected combo point amount 5, got %f", comboPoints.Amount)
	}
}

func findResourceEvent(events []types.ResourceEvent, resourceTypeID int) (types.ResourceEvent, bool) {
	for _, event := range events {
		if event.ResourceTypeID == resourceTypeID {
			return event, true
		}
	}
	return types.ResourceEvent{}, false
}

func TestDeriveCharacterClassIDFromRecentReports_UsesExactCanonicalIDMatch(t *testing.T) {
	character := WCLUserCharacter{
		CanonicalID: 12345,
		ClassID:     6,
		RecentReports: WCLReportPagination{
			Data: []WCLReportSummary{
				{
					RankedCharacters: []WCLRankedCharacter{
						{CanonicalID: 99999, ClassID: 2},
						{CanonicalID: 12345, ClassID: 11},
					},
				},
			},
		},
	}

	classID, ok := deriveCharacterClassIDFromRecentReports(character)
	if !ok {
		t.Fatalf("expected exact canonicalID match to be found")
	}
	if classID != 11 {
		t.Fatalf("expected derived class id 11, got %d", classID)
	}
}

func TestDeriveCharacterClassIDFromRecentReports_IgnoresNonMatchingCharacters(t *testing.T) {
	character := WCLUserCharacter{
		CanonicalID: 12345,
		Name:        "Jaicher",
		ClassID:     2,
		Server: WCLServer{
			Name: "Tichondrius",
		},
		RecentReports: WCLReportPagination{
			Data: []WCLReportSummary{
				{
					RankedCharacters: []WCLRankedCharacter{
						{
							CanonicalID: 99999,
							Name:        "SomeoneElse",
							ClassID:     6,
							Server: WCLServer{
								Name: "Tichondrius",
							},
						},
					},
				},
			},
		},
	}

	_, ok := deriveCharacterClassIDFromRecentReports(character)
	if ok {
		t.Fatalf("expected no derived class id without exact canonicalID match")
	}
}

func TestDeriveCharacterClassIDFromRecentReports_FallsBackToExactNameAndServerMatch(t *testing.T) {
	character := WCLUserCharacter{
		CanonicalID: 12345,
		Name:        "Jaicherdk",
		ClassID:     1,
		Server: WCLServer{
			Name: "Tichondrius",
		},
		RecentReports: WCLReportPagination{
			Data: []WCLReportSummary{
				{
					RankedCharacters: []WCLRankedCharacter{
						{
							CanonicalID: 99999,
							Name:        "Jaicherdk",
							ClassID:     6,
							Server: WCLServer{
								Name: "Tichondrius",
							},
						},
					},
				},
			},
		},
	}

	classID, ok := deriveCharacterClassIDFromRecentReports(character)
	if !ok {
		t.Fatalf("expected exact name+server match to derive class id")
	}
	if classID != 6 {
		t.Fatalf("expected derived class id 6, got %d", classID)
	}
}

func TestNormalizeMatchKeys_IgnoreFormattingDifferences(t *testing.T) {
	if classMatchKey("Death Knight") != classMatchKey("DeathKnight") {
		t.Fatalf("expected class match keys to ignore spacing differences")
	}

	if specMatchKey("Beast Mastery") != specMatchKey("Beast-Mastery") {
		t.Fatalf("expected spec match keys to ignore punctuation differences")
	}
}
