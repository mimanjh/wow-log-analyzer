package services

import (
	"testing"
	"time"
)

type stubLogClient struct {
	fights     []FightSummary
	characters []CharacterSummary
	err        error
}

func (s *stubLogClient) GetFights(string) ([]FightSummary, error) {
	return s.fights, s.err
}

func (s *stubLogClient) GetCharacters(string, int) ([]CharacterSummary, error) {
	return s.characters, s.err
}

func TestAnalyzeService_ValidateAndParseUrl(t *testing.T) {
	service := NewAnalyzeService("http://example.com")

	tests := []struct {
		name string
		url  string
		want UrlValidationResult
	}{
		{
			name: "valid URL",
			url:  "https://www.warcraftlogs.com/reports/abc123?fight=4",
			want: UrlValidationResult{
				IsValid:          true,
				ReportID:         "abc123",
				PreferredFightID: 4,
			},
		},
		{
			name: "valid URL without www",
			url:  "https://warcraftlogs.com/reports/def456",
			want: UrlValidationResult{
				IsValid:  true,
				ReportID: "def456",
			},
		},
		{
			name: "valid HTTP URL",
			url:  "http://www.warcraftlogs.com/reports/ghi789",
			want: UrlValidationResult{
				IsValid:  true,
				ReportID: "ghi789",
			},
		},
		{
			name: "empty string",
			url:  "",
			want: UrlValidationResult{
				IsValid: false,
				Error:   "URL is required",
			},
		},
		{
			name: "whitespace only",
			url:  "   ",
			want: UrlValidationResult{
				IsValid: false,
				Error:   "URL is required",
			},
		},
		{
			name: "non-Warcraft Logs URL",
			url:  "https://google.com",
			want: UrlValidationResult{
				IsValid: false,
				Error:   "URL must be from warcraftlogs.com/reports/",
			},
		},
		{
			name: "Warcraft Logs URL without reports path",
			url:  "https://www.warcraftlogs.com/",
			want: UrlValidationResult{
				IsValid: false,
				Error:   "URL must be from warcraftlogs.com/reports/",
			},
		},
		{
			name: "URL with invalid report code",
			url:  "https://www.warcraftlogs.com/reports/",
			want: UrlValidationResult{
				IsValid: false,
				Error:   "URL must contain a valid report code",
			},
		},
		{
			name: "report code too short",
			url:  "https://www.warcraftlogs.com/reports/abc",
			want: UrlValidationResult{
				IsValid: false,
				Error:   "Report code appears too short",
			},
		},
		{
			name: "report code too long",
			url:  "https://www.warcraftlogs.com/reports/abcdefghijklmnopqrstuv",
			want: UrlValidationResult{
				IsValid: false,
				Error:   "Report code appears too long",
			},
		},
		{
			name: "URL with extra path segments",
			url:  "https://www.warcraftlogs.com/reports/abc123/fight/1",
			want: UrlValidationResult{
				IsValid:  true,
				ReportID: "abc123",
			},
		},
		{
			name: "URL with query parameters",
			url:  "https://www.warcraftlogs.com/reports/abc123?ref=raidbots",
			want: UrlValidationResult{
				IsValid:  true,
				ReportID: "abc123",
			},
		},
		{
			name: "mixed case report code",
			url:  "https://www.warcraftlogs.com/reports/AbC123",
			want: UrlValidationResult{
				IsValid:  true,
				ReportID: "AbC123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.ValidateAndParseUrl(tt.url)
			if got.IsValid != tt.want.IsValid {
				t.Errorf("ValidateAndParseUrl() IsValid = %v, want %v", got.IsValid, tt.want.IsValid)
			}
			if got.ReportID != tt.want.ReportID {
				t.Errorf("ValidateAndParseUrl() ReportID = %v, want %v", got.ReportID, tt.want.ReportID)
			}
			if got.Error != tt.want.Error {
				t.Errorf("ValidateAndParseUrl() Error = %v, want %v", got.Error, tt.want.Error)
			}
			if got.PreferredFightID != tt.want.PreferredFightID {
				t.Errorf("ValidateAndParseUrl() PreferredFightID = %v, want %v", got.PreferredFightID, tt.want.PreferredFightID)
			}
		})
	}
}

func TestAnalyzeService_ProcessIntake(t *testing.T) {
	service := &AnalyzeService{
		logClient: &stubLogClient{
			fights: []FightSummary{
				{ID: 1, Name: "Boss One", Difficulty: "Heroic", Kill: true, KillTime: 300, EncounterID: 101, StartTime: time.Unix(0, 0), EndTime: time.Unix(300, 0)},
				{ID: 2, Name: "Boss Two", Difficulty: "Mythic", Kill: true, KillTime: 420, EncounterID: 102, StartTime: time.Unix(0, 0), EndTime: time.Unix(420, 0)},
				{ID: 3, Name: "Dungeon Boss", Difficulty: "Unknown", Kill: true, KillTime: 180, EncounterID: 103, StartTime: time.Unix(0, 0), EndTime: time.Unix(180, 0)},
				{ID: 4, Name: "Wipe Boss", Difficulty: "Heroic", Kill: false, KillTime: 240, EncounterID: 104, StartTime: time.Unix(0, 0), EndTime: time.Unix(240, 0)},
			},
			characters: []CharacterSummary{
				{ID: 1, Name: "Tester", Class: "Mage", Spec: "Fire", Role: "DPS"},
			},
		},
	}

	t.Run("valid request", func(t *testing.T) {
		req := AnalyzeIntakeRequest{
			Url: "https://www.warcraftlogs.com/reports/abc123?fight=2",
		}

		resp, err := service.ProcessIntake(req)
		if err != nil {
			t.Errorf("ProcessIntake() error = %v", err)
			return
		}

		if resp.ReportID != "abc123" {
			t.Errorf("ProcessIntake() ReportID = %v, want %v", resp.ReportID, "abc123")
		}

		if len(resp.Fights) != 3 {
			t.Errorf("ProcessIntake() expected 3 fights, got %d", len(resp.Fights))
		}

		if len(resp.Characters) != 1 {
			t.Errorf("ProcessIntake() expected 1 character, got %d", len(resp.Characters))
		}

		if resp.PreferredFightID != 2 {
			t.Errorf("ProcessIntake() PreferredFightID = %v, want %v", resp.PreferredFightID, 2)
		}

		if len(resp.Fights) > 0 && resp.Fights[0].ID != 2 {
			t.Errorf("ProcessIntake() expected preferred fight to be first, got %d", resp.Fights[0].ID)
		}
		if len(resp.Fights) > 1 && resp.Fights[1].ID != 1 {
			t.Errorf("ProcessIntake() expected non-preferred raid fights to remain, got %#v", resp.Fights)
		}
		if len(resp.Fights) > 2 && resp.Fights[2].ID != 4 {
			t.Errorf("ProcessIntake() expected valid wipe fights to remain, got %#v", resp.Fights)
		}
	})

	t.Run("invalid request", func(t *testing.T) {
		req := AnalyzeIntakeRequest{
			Url: "https://google.com",
		}

		_, err := service.ProcessIntake(req)
		if err == nil {
			t.Error("ProcessIntake() expected error for invalid URL")
		}
	})
}
