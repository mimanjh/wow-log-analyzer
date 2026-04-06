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
		Difficulty:  "Heroic",
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
		t.Errorf("Expected BossPercent %d, got %d", expected.BossPercent, fight.BossPercent)
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
		{3, "Heroic"},
		{4, "Mythic"},
		{5, "Unknown"},
	}

	for _, test := range tests {
		result := client.normalizeDifficulty(test.input)
		if result != test.expected {
			t.Errorf("For input %d, expected %s, got %s", test.input, test.expected, result)
		}
	}
}
