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
	service := NewAnalyzeService("http://example.com", nil, "")

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

		if len(resp.Fights) != 0 {
			t.Errorf("ProcessIntake() expected fights to be lazy-loaded, got %d", len(resp.Fights))
		}

		if len(resp.Characters) != 0 {
			t.Errorf("ProcessIntake() expected characters to be lazy-loaded, got %d", len(resp.Characters))
		}

		if resp.PreferredFightID != 2 {
			t.Errorf("ProcessIntake() PreferredFightID = %v, want %v", resp.PreferredFightID, 2)
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

func TestAnalyzeService_GetFightsForReport(t *testing.T) {
	service := &AnalyzeService{
		logClient: &stubLogClient{
			fights: []FightSummary{
				{ID: 1, Name: "Boss One", Difficulty: "Heroic", Kill: true, KillTime: 300, EncounterID: 101, StartTime: time.Unix(0, 0), EndTime: time.Unix(300, 0)},
				{ID: 2, Name: "Boss Two", Difficulty: "Mythic", Kill: true, KillTime: 420, EncounterID: 102, StartTime: time.Unix(0, 0), EndTime: time.Unix(420, 0)},
				{ID: 3, Name: "Dungeon Boss", Difficulty: "Unknown", Kill: true, KillTime: 180, EncounterID: 103, StartTime: time.Unix(0, 0), EndTime: time.Unix(180, 0)},
				{ID: 4, Name: "Wipe Boss", Difficulty: "Heroic", Kill: false, KillTime: 240, EncounterID: 104, StartTime: time.Unix(0, 0), EndTime: time.Unix(240, 0)},
			},
		},
	}

	fights, err := service.GetFightsForReport("abc123", 2, CharacterFightFilter{})
	if err != nil {
		t.Fatalf("GetFightsForReport() error = %v", err)
	}

	if len(fights) != 3 {
		t.Fatalf("GetFightsForReport() expected 3 fights, got %d", len(fights))
	}
	if fights[0].ID != 2 {
		t.Errorf("GetFightsForReport() expected preferred fight first, got %d", fights[0].ID)
	}
	if fights[1].ID != 1 {
		t.Errorf("GetFightsForReport() expected non-preferred raid fights to remain, got %#v", fights)
	}
	if fights[2].ID != 4 {
		t.Errorf("GetFightsForReport() expected valid wipe fights to remain, got %#v", fights)
	}
}

func TestAnalyzeService_GetFightsForReport_FiltersByCharacter(t *testing.T) {
	service := &AnalyzeService{
		logClient: &stubLogClient{
			fights: []FightSummary{
				{
					ID:          1,
					Name:        "Boss One",
					Difficulty:  "Heroic",
					Kill:        true,
					EncounterID: 101,
					FriendlyPlayers: []FightParticipant{
						{Name: "Jaicherdk", ServerName: "Tichondrius", Class: "Death Knight"},
					},
				},
				{
					ID:          2,
					Name:        "Boss Two",
					Difficulty:  "Heroic",
					Kill:        true,
					EncounterID: 102,
					FriendlyPlayers: []FightParticipant{
						{Name: "Otherplayer", ServerName: "Tichondrius", Class: "Druid"},
					},
				},
				{
					ID:          3,
					Name:        "Boss Three",
					Difficulty:  "Heroic",
					Kill:        true,
					EncounterID: 103,
					FriendlyPlayers: []FightParticipant{
						{Name: "Jaicherdk", ServerName: "tichondrius", Class: "DeathKnight"},
					},
				},
			},
		},
	}

	fights, err := service.GetFightsForReport("abc123", 0, CharacterFightFilter{
		Name:       "Jaicherdk",
		ServerName: "Tichondrius",
		ClassName:  "Death Knight",
	})
	if err != nil {
		t.Fatalf("GetFightsForReport() error = %v", err)
	}

	if len(fights) != 2 {
		t.Fatalf("GetFightsForReport() expected 2 fights, got %d", len(fights))
	}
	if fights[0].ID != 1 || fights[1].ID != 3 {
		t.Fatalf("GetFightsForReport() returned wrong fights: %#v", fights)
	}
}

func TestAnalyzeService_GetFightsForReport_AllowsGenericPlayerClass(t *testing.T) {
	service := &AnalyzeService{
		logClient: &stubLogClient{
			fights: []FightSummary{
				{
					ID:          1,
					Name:        "Boss One",
					Difficulty:  "Heroic",
					Kill:        true,
					EncounterID: 101,
					FriendlyPlayers: []FightParticipant{
						{Name: "Jaicherdk", ServerName: "Tichondrius", Class: "Player"},
					},
				},
			},
		},
	}

	fights, err := service.GetFightsForReport("abc123", 0, CharacterFightFilter{
		Name:       "Jaicherdk",
		ServerName: "Tichondrius",
		ClassName:  "Death Knight",
	})
	if err != nil {
		t.Fatalf("GetFightsForReport() error = %v", err)
	}

	if len(fights) != 1 {
		t.Fatalf("GetFightsForReport() expected 1 fight, got %d", len(fights))
	}
}
