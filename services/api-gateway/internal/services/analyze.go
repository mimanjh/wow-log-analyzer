package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	neturl "net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	reportIDRegex = regexp.MustCompile(`(?i)/reports/([a-zA-Z0-9]+)`)
	hostRegex     = regexp.MustCompile(`(?i)^https?://(?:www\.)?warcraftlogs\.com/reports/`)
)

type UrlValidationResult struct {
	IsValid          bool   `json:"isValid"`
	ReportID         string `json:"reportId,omitempty"`
	PreferredFightID int    `json:"preferredFightId,omitempty"`
	Error            string `json:"error,omitempty"`
}

type AnalyzeIntakeRequest struct {
	Url string `json:"url"`
}

type AnalyzeIntakeResponse struct {
	ReportID         string             `json:"reportId"`
	PreferredFightID int                `json:"preferredFightId,omitempty"`
	Fights           []FightSummary     `json:"fights"`
	Characters       []CharacterSummary `json:"characters"`
}

type FightSummary struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Difficulty  string    `json:"difficulty"`
	Kill        bool      `json:"kill"`
	KillTime    int       `json:"killTime"`
	EncounterID int       `json:"encounterId"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime"`
	BossPercent float64   `json:"bossPercent,omitempty"`
}

type CharacterSummary struct {
	ID                  int    `json:"id"`
	Name                string `json:"name"`
	Class               string `json:"class"`
	Spec                string `json:"spec"`
	Role                string `json:"role"`
	ServerName          string `json:"serverName,omitempty"`
	TalentImportCode    string `json:"talentImportCode,omitempty"`
	TalentCalculatorURL string `json:"talentCalculatorUrl,omitempty"`
}

type LogServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

type logDataClient interface {
	GetFights(reportID string) ([]FightSummary, error)
	GetCharacters(reportID string, fightID int) ([]CharacterSummary, error)
}

func NewLogServiceClient(baseURL string) *LogServiceClient {
	return &LogServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *LogServiceClient) GetFights(reportID string) ([]FightSummary, error) {
	url := fmt.Sprintf("%s/reports/%s/fights", c.baseURL, reportID)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call log-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("log-service returned status %d: %s", resp.StatusCode, string(body))
	}

	var normalizedFights []struct {
		ID          int       `json:"id"`
		Name        string    `json:"name"`
		StartTime   time.Time `json:"startTime"`
		EndTime     time.Time `json:"endTime"`
		EncounterID int       `json:"encounterId"`
		Difficulty  string    `json:"difficulty"`
		Kill        bool      `json:"kill"`
		BossPercent float64   `json:"bossPercent"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&normalizedFights); err != nil {
		return nil, fmt.Errorf("failed to decode fights response: %w", err)
	}

	fights := make([]FightSummary, len(normalizedFights))
	for i, nf := range normalizedFights {
		killTime := int(nf.EndTime.Sub(nf.StartTime).Seconds())

		fights[i] = FightSummary{
			ID:          nf.ID,
			Name:        nf.Name,
			Difficulty:  nf.Difficulty,
			Kill:        nf.Kill,
			KillTime:    killTime,
			EncounterID: nf.EncounterID,
			StartTime:   nf.StartTime,
			EndTime:     nf.EndTime,
			BossPercent: nf.BossPercent,
		}
	}

	return fights, nil
}

func (c *LogServiceClient) GetCharacters(reportID string, fightID int) ([]CharacterSummary, error) {
	url := fmt.Sprintf("%s/reports/%s/characters?fightId=%d", c.baseURL, reportID, fightID)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call log-service characters endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("log-service returned status %d: %s", resp.StatusCode, string(body))
	}

	var characters []CharacterSummary
	if err := json.NewDecoder(resp.Body).Decode(&characters); err != nil {
		return nil, fmt.Errorf("failed to decode characters response: %w", err)
	}

	return characters, nil
}

type AnalyzeService struct {
	logClient logDataClient
}

func NewAnalyzeService(logServiceURL string) *AnalyzeService {
	return &AnalyzeService{
		logClient: NewLogServiceClient(logServiceURL),
	}
}

func (s *AnalyzeService) ValidateAndParseUrl(rawUrl string) UrlValidationResult {
	trimmed := strings.TrimSpace(rawUrl)

	if trimmed == "" {
		return UrlValidationResult{
			IsValid: false,
			Error:   "URL is required",
		}
	}

	if !hostRegex.MatchString(trimmed) {
		return UrlValidationResult{
			IsValid: false,
			Error:   "URL must be from warcraftlogs.com/reports/",
		}
	}

	match := reportIDRegex.FindStringSubmatch(trimmed)
	if match == nil || len(match) < 2 {
		return UrlValidationResult{
			IsValid: false,
			Error:   "URL must contain a valid report code",
		}
	}

	reportID := match[1]
	if len(reportID) < 6 {
		return UrlValidationResult{
			IsValid: false,
			Error:   "Report code appears too short",
		}
	}

	if len(reportID) > 16 {
		return UrlValidationResult{
			IsValid: false,
			Error:   "Report code appears too long",
		}
	}

	preferredFightID := extractFightID(trimmed)

	return UrlValidationResult{
		IsValid:          true,
		ReportID:         reportID,
		PreferredFightID: preferredFightID,
	}
}

func (s *AnalyzeService) ProcessIntake(req AnalyzeIntakeRequest) (AnalyzeIntakeResponse, error) {
	validation := s.ValidateAndParseUrl(req.Url)
	if !validation.IsValid {
		return AnalyzeIntakeResponse{}, fmt.Errorf("validation failed: %s", validation.Error)
	}

	return AnalyzeIntakeResponse{
		ReportID:         validation.ReportID,
		PreferredFightID: validation.PreferredFightID,
		Fights:           []FightSummary{},
		Characters:       []CharacterSummary{},
	}, nil
}

func (s *AnalyzeService) GetFightsForReport(reportID string, preferredFightID int) ([]FightSummary, error) {
	if strings.TrimSpace(reportID) == "" {
		return nil, fmt.Errorf("reportId is required")
	}

	fights, err := s.logClient.GetFights(reportID)
	if err != nil {
		log.Printf("Failed to get fights from log-service: %v", err)
		return nil, fmt.Errorf("failed to retrieve fights: %w", err)
	}

	fights = filterRelevantFights(fights)
	fights = prioritizeFight(fights, preferredFightID)

	return fights, nil
}

func (s *AnalyzeService) GetCharactersForFight(reportID string, fightID int) ([]CharacterSummary, error) {
	if strings.TrimSpace(reportID) == "" {
		return nil, fmt.Errorf("reportId is required")
	}
	if fightID == 0 {
		return nil, fmt.Errorf("fightId is required")
	}

	characters, err := s.logClient.GetCharacters(reportID, fightID)
	if err != nil {
		log.Printf("Failed to get characters from log-service: %v", err)
		return nil, fmt.Errorf("failed to retrieve characters: %w", err)
	}

	return characters, nil
}

func extractFightID(rawURL string) int {
	parsed, err := neturl.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return 0
	}

	if fightID := parseFightID(parsed.Query().Get("fight")); fightID != 0 {
		return fightID
	}

	fragment := strings.TrimPrefix(parsed.Fragment, "#")
	hashParams, err := neturl.ParseQuery(fragment)
	if err != nil {
		return 0
	}

	return parseFightID(hashParams.Get("fight"))
}

func parseFightID(value string) int {
	if strings.TrimSpace(value) == "" {
		return 0
	}

	var fightID int
	if _, err := fmt.Sscanf(value, "%d", &fightID); err != nil || fightID <= 0 {
		return 0
	}

	return fightID
}

func prioritizeFight(fights []FightSummary, preferredFightID int) []FightSummary {
	if preferredFightID == 0 || len(fights) < 2 {
		return fights
	}

	prioritized := append([]FightSummary(nil), fights...)
	sort.SliceStable(prioritized, func(i, j int) bool {
		if prioritized[i].ID == preferredFightID {
			return true
		}
		if prioritized[j].ID == preferredFightID {
			return false
		}
		return false
	})

	return prioritized
}

func filterRelevantFights(fights []FightSummary) []FightSummary {
	filtered := make([]FightSummary, 0, len(fights))
	for _, fight := range fights {
		if fight.EncounterID == 0 {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(fight.Difficulty), "Unknown") {
			continue
		}
		filtered = append(filtered, fight)
	}

	return filtered
}
