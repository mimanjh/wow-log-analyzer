package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"wow-log-analyzer/services/log-service/internal/config"
	"wow-log-analyzer/services/log-service/internal/types"
)

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// WCLClient defines the interface for Warcraft Logs API interactions
type WCLClient interface {
	GetReportMetadata(reportID string) (*types.NormalizedReport, error)
	GetFights(reportID string) ([]types.NormalizedFight, error)
	GetCharacters(reportID string, fightID int) ([]types.CharacterOption, error)
	GetComparisonData(reportID string, fightID int, characterID int) (*types.ComparisonDataResponse, error)
}

// WCLHTTPClient implements WCLClient using HTTP requests
type WCLHTTPClient struct {
	config      config.WCLConfig
	httpClient  httpDoer
	tokenMu     sync.Mutex
	token       string
	tokenExpiry time.Time
}

// NewWCLClient creates a new Warcraft Logs client
func NewWCLClient(cfg config.WCLConfig) WCLClient {
	return &WCLHTTPClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetReportMetadata fetches report metadata and normalizes it
func (c *WCLHTTPClient) GetReportMetadata(reportID string) (*types.NormalizedReport, error) {
	query := fmt.Sprintf(`
		query {
			reportData {
				report(code: "%s") {
					code
					title
					startTime
					endTime
					zone {
						id
						name
					}
				}
			}
		}`, reportID)

	var response struct {
		Data struct {
			ReportData struct {
				Report WCLReport `json:"report"`
			} `json:"reportData"`
		} `json:"data"`
	}

	if err := c.makeGraphQLRequest(query, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch report metadata: %w", err)
	}

	return c.normalizeReport(response.Data.ReportData.Report), nil
}

// GetFights fetches fights for a report and normalizes them
func (c *WCLHTTPClient) GetFights(reportID string) ([]types.NormalizedFight, error) {
	query := fmt.Sprintf(`
		query {
			reportData {
				report(code: "%s") {
					fights {
						id
						name
						startTime
						endTime
						encounterID
						difficulty
						kill
						bossPercentage
					}
				}
			}
		}`, reportID)

	var response struct {
		Data struct {
			ReportData struct {
				Report struct {
					Fights []WCLFight `json:"fights"`
				} `json:"report"`
			} `json:"reportData"`
		} `json:"data"`
	}

	if err := c.makeGraphQLRequest(query, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch fights: %w", err)
	}

	return c.normalizeFights(response.Data.ReportData.Report.Fights), nil
}

func (c *WCLHTTPClient) GetCharacters(reportID string, fightID int) ([]types.CharacterOption, error) {
	actors, err := c.getPlayerActors(reportID)
	if err != nil {
		return nil, err
	}

	fights, err := c.GetFights(reportID)
	if err != nil {
		return nil, err
	}

	validFight := false
	for _, fight := range fights {
		if fight.ID == fightID {
			validFight = true
			break
		}
	}
	if !validFight {
		return nil, fmt.Errorf("fight %d not found in report", fightID)
	}

	characters := make([]types.CharacterOption, 0, len(actors))
	for _, actor := range actors {
		characters = append(characters, types.CharacterOption{
			ID:    actor.ID,
			Name:  actor.Name,
			Class: normalizeClass(actor.Type),
			Spec:  normalizeSpec(actor.SubType),
			Role:  inferRole(actor.SubType),
		})
	}

	sort.Slice(characters, func(i, j int) bool {
		return characters[i].Name < characters[j].Name
	})

	return characters, nil
}

func (c *WCLHTTPClient) GetComparisonData(reportID string, fightID int, characterID int) (*types.ComparisonDataResponse, error) {
	fights, err := c.GetFights(reportID)
	if err != nil {
		return nil, err
	}

	var selectedFight *types.NormalizedFight
	for _, fight := range fights {
		if fight.ID == fightID {
			fightCopy := fight
			selectedFight = &fightCopy
			break
		}
	}
	if selectedFight == nil {
		return nil, fmt.Errorf("fight %d not found in report", fightID)
	}

	actors, err := c.getPlayerActors(reportID)
	if err != nil {
		return nil, err
	}

	var selectedActor *WCLActor
	for _, actor := range actors {
		if actor.ID == characterID {
			actorCopy := actor
			selectedActor = &actorCopy
			break
		}
	}
	if selectedActor == nil {
		return nil, fmt.Errorf("character %d not found in report", characterID)
	}

	playerData, err := c.fetchPlayerFightData(reportID, *selectedFight, *selectedActor)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch player fight data: %w", err)
	}

	cohortActors := selectCohortActors(actors, *selectedActor)
	cohortData := make([]types.PlayerFightData, 0, len(cohortActors))
	for _, actor := range cohortActors {
		cohortEntry, err := c.fetchPlayerFightData(reportID, *selectedFight, actor)
		if err != nil {
			continue
		}
		cohortData = append(cohortData, cohortEntry)
		if len(cohortData) == 5 {
			break
		}
	}

	if len(cohortData) == 0 {
		return nil, fmt.Errorf("no cohort candidates found for fight %d", fightID)
	}

	return &types.ComparisonDataResponse{
		ReportID: reportID,
		Fight: types.FightSummary{
			ID:          selectedFight.ID,
			Name:        selectedFight.Name,
			Difficulty:  selectedFight.Difficulty,
			KillTime:    int(selectedFight.EndTime.Sub(selectedFight.StartTime).Seconds()),
			EncounterID: selectedFight.EncounterID,
		},
		PlayerData: playerData,
		CohortData: cohortData,
	}, nil
}

// makeGraphQLRequest performs a GraphQL request to the WCL API
func (c *WCLHTTPClient) makeGraphQLRequest(query string, response interface{}) error {
	requestBody := map[string]string{
		"query": query,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.config.BaseURL+"/client", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.config.ClientID != "" && c.config.ClientSecret != "" {
		token, err := c.getAccessToken()
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

func (c *WCLHTTPClient) getAccessToken() (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	if c.token != "" && time.Now().Before(c.tokenExpiry) {
		return c.token, nil
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", "https://www.warcraftlogs.com/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.config.ClientID, c.config.ClientSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request oauth token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("oauth token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResponse WCLTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("failed to decode oauth token response: %w", err)
	}
	if tokenResponse.AccessToken == "" {
		return "", fmt.Errorf("oauth token response did not include an access token")
	}

	c.token = tokenResponse.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResponse.ExpiresIn-60) * time.Second)
	return c.token, nil
}

// normalizeReport converts WCL API response to our internal format
func (c *WCLHTTPClient) normalizeReport(wclReport WCLReport) *types.NormalizedReport {
	return &types.NormalizedReport{
		ID:        wclReport.Code,
		Title:     wclReport.Title,
		StartTime: time.UnixMilli(wclReport.StartTime),
		EndTime:   time.UnixMilli(wclReport.EndTime),
		Zone: types.Zone{
			ID:   wclReport.Zone.ID,
			Name: wclReport.Zone.Name,
		},
	}
}

// normalizeFights converts WCL API fights to our internal format
func (c *WCLHTTPClient) normalizeFights(wclFights []WCLFight) []types.NormalizedFight {
	fights := make([]types.NormalizedFight, len(wclFights))
	for i, fight := range wclFights {
		fights[i] = types.NormalizedFight{
			ID:          fight.ID,
			Name:        fight.Name,
			StartTime:   time.UnixMilli(fight.StartTime),
			EndTime:     time.UnixMilli(fight.EndTime),
			EncounterID: fight.EncounterID,
			Difficulty:  c.normalizeDifficulty(fight.Difficulty),
			Kill:        fight.Kill,
			BossPercent: fight.BossPercentage,
		}
	}
	return fights
}

// normalizeDifficulty converts WCL difficulty numbers to readable strings
func (c *WCLHTTPClient) normalizeDifficulty(difficulty int) string {
	switch difficulty {
	case 1:
		return "LFR"
	case 2:
		return "Normal"
	case 3:
		return "Heroic"
	case 4:
		return "Mythic"
	default:
		return "Unknown"
	}
}

func (c *WCLHTTPClient) getPlayerActors(reportID string) ([]WCLActor, error) {
	query := fmt.Sprintf(`
		query {
			reportData {
				report(code: "%s") {
					masterData {
						actors(type: "Player") {
							id
							name
							server
							subType
							type
							petOwner
							gameID
						}
					}
				}
			}
		}`, reportID)

	var response struct {
		Data struct {
			ReportData struct {
				Report struct {
					MasterData struct {
						Actors []WCLActor `json:"actors"`
					} `json:"masterData"`
				} `json:"report"`
			} `json:"reportData"`
		} `json:"data"`
	}

	if err := c.makeGraphQLRequest(query, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch report actors: %w", err)
	}

	actors := make([]WCLActor, 0, len(response.Data.ReportData.Report.MasterData.Actors))
	for _, actor := range response.Data.ReportData.Report.MasterData.Actors {
		if actor.PetOwner != 0 || actor.ID == 0 || actor.Name == "" {
			continue
		}
		actors = append(actors, actor)
	}

	return actors, nil
}

func (c *WCLHTTPClient) fetchPlayerFightData(reportID string, fight types.NormalizedFight, actor WCLActor) (types.PlayerFightData, error) {
	startTime := fight.StartTime.UnixMilli()
	endTime := fight.EndTime.UnixMilli()

	castRaw, err := c.fetchEvents(reportID, "Casts", fight.ID, actor.ID, startTime, endTime)
	if err != nil {
		return types.PlayerFightData{}, err
	}
	damageRaw, err := c.fetchEvents(reportID, "DamageDone", fight.ID, actor.ID, startTime, endTime)
	if err != nil {
		return types.PlayerFightData{}, err
	}
	healRaw, err := c.fetchEvents(reportID, "Healing", fight.ID, actor.ID, startTime, endTime)
	if err != nil {
		return types.PlayerFightData{}, err
	}
	buffRaw, err := c.fetchBuffEvents(reportID, fight.ID, actor.ID, startTime, endTime)
	if err != nil {
		return types.PlayerFightData{}, err
	}

	return types.PlayerFightData{
		PlayerID:       actor.ID,
		FightID:        fight.ID,
		FightStart:     fight.StartTime,
		FightEnd:       fight.EndTime,
		CastEvents:     normalizeCastEvents(castRaw),
		DamageEvents:   normalizeDamageEvents(damageRaw),
		HealEvents:     normalizeHealEvents(healRaw),
		BuffEvents:     normalizeBuffEvents(buffRaw),
		CooldownEvents: normalizeCooldownEvents(castRaw),
	}, nil
}

func (c *WCLHTTPClient) fetchEvents(reportID, dataType string, fightID, sourceID int, startTime, endTime int64) ([]map[string]interface{}, error) {
	collected := make([]map[string]interface{}, 0)
	nextStart := startTime

	for {
		query := fmt.Sprintf(`
			query {
				reportData {
					report(code: "%s") {
						events(
							dataType: %s,
							fightIDs: [%d],
							sourceID: %d,
							startTime: %d,
							endTime: %d,
							limit: 10000
						) {
							data
							nextPageTimestamp
						}
					}
				}
			}`, reportID, dataType, fightID, sourceID, nextStart, endTime)

		var response struct {
			Data struct {
				ReportData struct {
					Report struct {
						Events struct {
							Data              []map[string]interface{} `json:"data"`
							NextPageTimestamp *float64                 `json:"nextPageTimestamp"`
						} `json:"events"`
					} `json:"report"`
				} `json:"reportData"`
			} `json:"data"`
		}

		if err := c.makeGraphQLRequest(query, &response); err != nil {
			return nil, fmt.Errorf("failed to fetch %s events: %w", dataType, err)
		}

		collected = append(collected, response.Data.ReportData.Report.Events.Data...)
		if response.Data.ReportData.Report.Events.NextPageTimestamp == nil || int64(*response.Data.ReportData.Report.Events.NextPageTimestamp) <= nextStart {
			break
		}
		nextStart = int64(*response.Data.ReportData.Report.Events.NextPageTimestamp)
	}

	return collected, nil
}

func (c *WCLHTTPClient) fetchBuffEvents(reportID string, fightID, actorID int, startTime, endTime int64) ([]map[string]interface{}, error) {
	collected := make([]map[string]interface{}, 0)
	nextStart := startTime

	for {
		query := fmt.Sprintf(`
			query {
				reportData {
					report(code: "%s") {
						events(
							dataType: Buffs,
							fightIDs: [%d],
							targetID: %d,
							startTime: %d,
							endTime: %d,
							limit: 10000
						) {
							data
							nextPageTimestamp
						}
					}
				}
			}`, reportID, fightID, actorID, nextStart, endTime)

		var response struct {
			Data struct {
				ReportData struct {
					Report struct {
						Events struct {
							Data              []map[string]interface{} `json:"data"`
							NextPageTimestamp *float64                 `json:"nextPageTimestamp"`
						} `json:"events"`
					} `json:"report"`
				} `json:"reportData"`
			} `json:"data"`
		}

		if err := c.makeGraphQLRequest(query, &response); err != nil {
			return nil, fmt.Errorf("failed to fetch buff events: %w", err)
		}

		collected = append(collected, response.Data.ReportData.Report.Events.Data...)
		if response.Data.ReportData.Report.Events.NextPageTimestamp == nil || int64(*response.Data.ReportData.Report.Events.NextPageTimestamp) <= nextStart {
			break
		}
		nextStart = int64(*response.Data.ReportData.Report.Events.NextPageTimestamp)
	}

	return collected, nil
}

func selectCohortActors(actors []WCLActor, selected WCLActor) []WCLActor {
	candidates := make([]WCLActor, 0, len(actors))
	for _, actor := range actors {
		if actor.ID == selected.ID {
			continue
		}
		if selected.SubType != "" && actor.SubType == selected.SubType {
			candidates = append(candidates, actor)
		}
	}
	if len(candidates) >= 2 {
		return candidates
	}

	for _, actor := range actors {
		if actor.ID == selected.ID {
			continue
		}
		if actor.Type == selected.Type && !containsActor(candidates, actor.ID) {
			candidates = append(candidates, actor)
		}
	}

	if len(candidates) >= 2 {
		return candidates
	}

	for _, actor := range actors {
		if actor.ID == selected.ID || containsActor(candidates, actor.ID) {
			continue
		}
		candidates = append(candidates, actor)
	}

	return candidates
}

func containsActor(actors []WCLActor, actorID int) bool {
	for _, actor := range actors {
		if actor.ID == actorID {
			return true
		}
	}
	return false
}

func normalizeCastEvents(raw []map[string]interface{}) []types.CastEvent {
	events := make([]types.CastEvent, 0, len(raw))
	for _, event := range raw {
		events = append(events, types.CastEvent{
			Timestamp: parseTimestamp(event["timestamp"]),
			Ability:   parseAbility(event),
			SourceID:  parseActorID(event, "sourceID", "source"),
		})
	}
	return events
}

func normalizeDamageEvents(raw []map[string]interface{}) []types.DamageEvent {
	events := make([]types.DamageEvent, 0, len(raw))
	for _, event := range raw {
		events = append(events, types.DamageEvent{
			Timestamp: parseTimestamp(event["timestamp"]),
			Ability:   parseAbility(event),
			SourceID:  parseActorID(event, "sourceID", "source"),
			TargetID:  parseActorID(event, "targetID", "target"),
			Amount:    parseInt(event["amount"]),
		})
	}
	return events
}

func normalizeHealEvents(raw []map[string]interface{}) []types.HealEvent {
	events := make([]types.HealEvent, 0, len(raw))
	for _, event := range raw {
		events = append(events, types.HealEvent{
			Timestamp: parseTimestamp(event["timestamp"]),
			Ability:   parseAbility(event),
			SourceID:  parseActorID(event, "sourceID", "source"),
			TargetID:  parseActorID(event, "targetID", "target"),
			Amount:    parseInt(event["amount"]),
		})
	}
	return events
}

func normalizeBuffEvents(raw []map[string]interface{}) []types.BuffEvent {
	events := make([]types.BuffEvent, 0, len(raw))
	for _, event := range raw {
		eventType := fmt.Sprintf("%v", event["type"])
		normalizedType := ""
		switch {
		case strings.HasPrefix(eventType, "apply"):
			normalizedType = "apply"
		case strings.HasPrefix(eventType, "refresh"):
			normalizedType = "refresh"
		case strings.HasPrefix(eventType, "remove"):
			normalizedType = "remove"
		default:
			continue
		}

		ability := parseAbility(event)
		ability.IsBuff = true
		events = append(events, types.BuffEvent{
			Timestamp: parseTimestamp(event["timestamp"]),
			Ability:   ability,
			SourceID:  parseActorID(event, "sourceID", "source"),
			TargetID:  parseActorID(event, "targetID", "target"),
			EventType: normalizedType,
		})
	}
	return events
}

func normalizeCooldownEvents(raw []map[string]interface{}) []types.CooldownEvent {
	events := make([]types.CooldownEvent, 0)
	seen := map[int]int{}

	for _, event := range raw {
		ability := parseAbility(event)
		if ability.ID == 0 {
			continue
		}
		seen[ability.ID]++
	}

	for _, event := range raw {
		ability := parseAbility(event)
		if ability.ID == 0 || seen[ability.ID] > 5 {
			continue
		}
		ability.IsMajorCD = true
		events = append(events, types.CooldownEvent{
			Timestamp: parseTimestamp(event["timestamp"]),
			Ability:   ability,
			SourceID:  parseActorID(event, "sourceID", "source"),
			EventType: "start",
		})
	}

	return events
}

func parseTimestamp(value interface{}) time.Time {
	ms := parseInt64(value)
	if ms <= 0 {
		return time.Time{}
	}
	return time.UnixMilli(ms)
}

func parseAbility(event map[string]interface{}) types.Ability {
	if abilityValue, ok := event["ability"]; ok {
		if abilityMap, ok := abilityValue.(map[string]interface{}); ok {
			return types.Ability{
				ID:   parseInt(abilityMap["gameID"]),
				Name: parseString(abilityMap["name"]),
			}
		}
	}

	return types.Ability{
		ID:   parseInt(event["abilityGameID"]),
		Name: parseString(event["abilityName"]),
	}
}

func parseActorID(event map[string]interface{}, idKey, objectKey string) int {
	if value, ok := event[idKey]; ok {
		return parseInt(value)
	}
	if objectValue, ok := event[objectKey]; ok {
		if actorMap, ok := objectValue.(map[string]interface{}); ok {
			return parseInt(actorMap["id"])
		}
	}
	return 0
}

func parseInt(value interface{}) int {
	return int(parseInt64(value))
}

func parseInt64(value interface{}) int64 {
	switch typed := value.(type) {
	case float64:
		return int64(typed)
	case int:
		return int64(typed)
	case int64:
		return typed
	case json.Number:
		parsed, _ := typed.Int64()
		return parsed
	case string:
		parsed, _ := strconv.ParseInt(typed, 10, 64)
		return parsed
	default:
		return 0
	}
}

func parseString(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return ""
	}
}

func normalizeClass(raw string) string {
	return strings.TrimSpace(raw)
}

func normalizeSpec(raw string) string {
	return strings.TrimSpace(raw)
}

func inferRole(spec string) string {
	switch strings.ToLower(strings.TrimSpace(spec)) {
	case "holy", "discipline", "restoration", "mistweaver", "preservation":
		return "Healer"
	case "protection", "blood", "guardian", "vengeance", "brewmaster":
		return "Tank"
	default:
		return "DPS"
	}
}
