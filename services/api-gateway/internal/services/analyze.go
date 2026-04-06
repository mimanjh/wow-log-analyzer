package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	reportIDRegex = regexp.MustCompile(`(?i)/reports/([a-zA-Z0-9]+)`)
	hostRegex     = regexp.MustCompile(`(?i)^https?://(?:www\.)?warcraftlogs\.com/reports/`)
)

type UrlValidationResult struct {
	IsValid  bool   `json:"isValid"`
	ReportID string `json:"reportId,omitempty"`
	Error    string `json:"error,omitempty"`
}

type AnalyzeIntakeRequest struct {
	Url string `json:"url"`
}

type AnalyzeIntakeResponse struct {
	ReportID   string             `json:"reportId"`
	Fights     []FightSummary     `json:"fights"`
	Characters []CharacterSummary `json:"characters"`
}

type FightSummary struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Difficulty  string `json:"difficulty"`
	KillTime    int    `json:"killTime"`
	EncounterID int    `json:"encounterId"`
}

type CharacterSummary struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Class string `json:"class"`
	Spec  string `json:"spec"`
	Role  string `json:"role"`
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
		BossPercent int       `json:"bossPercent"`
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
			KillTime:    killTime,
			EncounterID: nf.EncounterID,
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

	return UrlValidationResult{
		IsValid:  true,
		ReportID: reportID,
	}
}

func (s *AnalyzeService) ProcessIntake(req AnalyzeIntakeRequest) (AnalyzeIntakeResponse, error) {
	validation := s.ValidateAndParseUrl(req.Url)
	if !validation.IsValid {
		return AnalyzeIntakeResponse{}, fmt.Errorf("validation failed: %s", validation.Error)
	}

	// Call log-service to get fights
	fights, err := s.logClient.GetFights(validation.ReportID)
	if err != nil {
		log.Printf("Failed to get fights from log-service: %v", err)
		return AnalyzeIntakeResponse{}, fmt.Errorf("failed to retrieve fights: %w", err)
	}

	characters := []CharacterSummary{}
	if len(fights) > 0 {
		characters, err = s.logClient.GetCharacters(validation.ReportID, fights[0].ID)
		if err != nil {
			log.Printf("Failed to get characters from log-service: %v", err)
			return AnalyzeIntakeResponse{}, fmt.Errorf("failed to retrieve characters: %w", err)
		}
	}

	return AnalyzeIntakeResponse{
		ReportID:   validation.ReportID,
		Fights:     fights,
		Characters: characters,
	}, nil
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
