package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	browserCharactersCacheTTL = 5 * time.Minute
	browserReportsCacheTTL    = 5 * time.Minute
)

type BrowserCharacter struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Class        string `json:"class"`
	ServerName   string `json:"serverName"`
	ServerRegion string `json:"serverRegion"`
	ServerSlug   string `json:"serverSlug,omitempty"`
}

type CharacterFightSummary struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Difficulty  string `json:"difficulty"`
	Kill        bool   `json:"kill"`
	KillTime    int    `json:"killTime"`
	EncounterID int    `json:"encounterId"`
}

type CharacterReportSummary struct {
	Code      string                 `json:"code"`
	Title     string                 `json:"title"`
	ZoneName  string                 `json:"zoneName,omitempty"`
	BossNames []string               `json:"bossNames,omitempty"`
	StartTime time.Time              `json:"startTime"`
	EndTime   time.Time              `json:"endTime"`
	Fights    []CharacterFightSummary `json:"fights,omitempty"`
}

type CharacterReportsPage struct {
	Reports    []CharacterReportSummary `json:"reports"`
	NextCursor string                   `json:"nextCursor,omitempty"`
	HasMore    bool                     `json:"hasMore"`
}

type BrowserService struct {
	baseURL     string
	httpClient  *http.Client
	redisClient *redis.Client
}

func NewBrowserService(logServiceURL string, redisClient *redis.Client) *BrowserService {
	return &BrowserService{
		baseURL:     logServiceURL,
		httpClient:  &http.Client{Timeout: 150 * time.Second},
		redisClient: redisClient,
	}
}

func (s *BrowserService) GetCurrentUser(accessToken string) (*AuthUser, error) {
	req, err := http.NewRequest(http.MethodGet, s.baseURL+"/user/profile", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("log-service returned status %d: %s", resp.StatusCode, string(body))
	}

	var user AuthUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *BrowserService) GetCharacters(accessToken, sessionID string) ([]BrowserCharacter, error) {
	cacheKey := "browser:characters:" + sessionID
	if s.redisClient != nil && sessionID != "" {
		if data, err := s.redisClient.Get(context.Background(), cacheKey).Bytes(); err == nil {
			var cached []BrowserCharacter
			if json.Unmarshal(data, &cached) == nil {
				return cached, nil
			}
		}
	}

	req, err := http.NewRequest(http.MethodGet, s.baseURL+"/user/characters", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("log-service returned status %d: %s", resp.StatusCode, string(body))
	}

	var characters []BrowserCharacter
	if err := json.NewDecoder(resp.Body).Decode(&characters); err != nil {
		return nil, err
	}

	if s.redisClient != nil && sessionID != "" {
		if data, err := json.Marshal(characters); err == nil {
			s.redisClient.Set(context.Background(), cacheKey, data, browserCharactersCacheTTL)
		}
	}

	return characters, nil
}

func (s *BrowserService) GetCharacterReports(accessToken, sessionID string, characterID int, cursor string, limit int) (*CharacterReportsPage, error) {
	if characterID == 0 {
		return nil, fmt.Errorf("characterId is required")
	}
	if limit <= 0 {
		limit = 10
	}

	cacheKey := fmt.Sprintf("browser:reports:%s:%d:%s:%d", sessionID, characterID, cursor, limit)
	if s.redisClient != nil && sessionID != "" {
		if data, err := s.redisClient.Get(context.Background(), cacheKey).Bytes(); err == nil {
			var cached CharacterReportsPage
			if json.Unmarshal(data, &cached) == nil {
				return &cached, nil
			}
		}
	}

	url := fmt.Sprintf("%s/user/characters/%d/reports?limit=%d", s.baseURL, characterID, limit)
	if strings.TrimSpace(cursor) != "" {
		url += "&cursor=" + cursor
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("log-service returned status %d: %s", resp.StatusCode, string(body))
	}

	var page CharacterReportsPage
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, err
	}

	if s.redisClient != nil && sessionID != "" {
		if data, err := json.Marshal(page); err == nil {
			s.redisClient.Set(context.Background(), cacheKey, data, browserReportsCacheTTL)
		}
	}

	return &page, nil
}
