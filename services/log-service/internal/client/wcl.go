package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"wow-log-analyzer/services/log-service/internal/config"
	"wow-log-analyzer/services/log-service/internal/types"
)

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type graphqlErrorResponse struct {
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// WCLClient defines the interface for Warcraft Logs API interactions
type WCLClient interface {
	GetReportMetadata(reportID string) (*types.NormalizedReport, error)
	GetFights(reportID string) ([]types.NormalizedFight, error)
	GetCharacters(reportID string, fightID int) ([]types.CharacterOption, error)
	GetComparisonData(reportID string, fight types.FightSelection, characterID int) (*types.ComparisonDataResponse, error)
	GetPlayerFightData(reportID string, fight types.FightSelection, characterID int) (*types.PlayerFightData, error)
	GetRankingCandidates(fight types.FightSelection, characterClass, characterSpec string, limit int) ([]types.RankingCandidate, error)
	GetCohortMemberData(candidate types.RankingCandidate) (*types.PlayerFightData, error)
	GetCurrentUser(accessToken string) (*types.UserProfile, error)
	GetOwnedCharacters(accessToken string) ([]types.OwnedCharacter, error)
	GetCharacterReports(accessToken string, characterID int, cursor string, limit int) (*types.CharacterReportsPage, error)
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
	const requestTimeout = 120 * time.Second

	return &WCLHTTPClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: requestTimeout,
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
					startTime
					fights(killType: Encounters) {
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
					StartTime int64      `json:"startTime"`
					Fights    []WCLFight `json:"fights"`
				} `json:"report"`
			} `json:"reportData"`
		} `json:"data"`
	}

	if err := c.makeGraphQLRequest(query, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch fights: %w", err)
	}

	return c.normalizeFights(
		response.Data.ReportData.Report.StartTime,
		response.Data.ReportData.Report.Fights,
	), nil
}

func (c *WCLHTTPClient) GetCharacters(reportID string, fightID int) ([]types.CharacterOption, error) {
	return c.getFightCharacters(reportID, fightID)
}

func (c *WCLHTTPClient) GetComparisonData(reportID string, fight types.FightSelection, characterID int) (*types.ComparisonDataResponse, error) {
	selectedFight := normalizedFightFromSelection(fight)

	playerData, err := c.GetPlayerFightData(reportID, fight, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch player fight data: %w", err)
	}

	characters, err := c.getFightCharacters(reportID, fight.ID)
	if err != nil {
		return nil, err
	}

	var selectedCharacter *types.CharacterOption
	for _, character := range characters {
		if character.ID == characterID {
			characterCopy := character
			selectedCharacter = &characterCopy
			break
		}
	}
	if selectedCharacter == nil {
		return nil, fmt.Errorf("character %d not found in report", characterID)
	}

	candidates, err := c.GetRankingCandidates(
		fight,
		selectedCharacter.Class,
		selectedCharacter.Spec,
		10,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ranking candidates: %w", err)
	}

	cohortData := make([]types.PlayerFightData, 0, len(candidates))
	for _, candidate := range candidates {
		cohortEntry, err := c.GetCohortMemberData(candidate)
		if err != nil {
			continue
		}
		cohortData = append(cohortData, *cohortEntry)
	}

	if len(cohortData) == 0 {
		return nil, fmt.Errorf("no cohort candidates found for fight %d", fight.ID)
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
		PlayerData: *playerData,
		CohortData: cohortData,
	}, nil
}

func (c *WCLHTTPClient) GetPlayerFightData(reportID string, fight types.FightSelection, characterID int) (*types.PlayerFightData, error) {
	selectedFight := normalizedFightFromSelection(fight)
	playerData, err := c.fetchPlayerFightData(reportID, selectedFight, characterID)
	if err != nil {
		return nil, err
	}

	return &playerData, nil
}

func (c *WCLHTTPClient) GetRankingCandidates(fight types.FightSelection, characterClass, characterSpec string, limit int) ([]types.RankingCandidate, error) {
	selectedFight := normalizedFightFromSelection(fight)
	if limit <= 0 {
		limit = 10
	}

	rankings, err := c.getEncounterRankings(selectedFight.EncounterID, selectedFight.Difficulty, characterClass, characterSpec, limit)
	if err != nil {
		return nil, err
	}

	capacity := limit
	if len(rankings) < capacity {
		capacity = len(rankings)
	}
	candidates := make([]types.RankingCandidate, 0, capacity)
	fallbackCandidates := make([]types.RankingCandidate, 0, capacity)
	targetDurationMS := fight.KillTime * 1000
	for _, ranking := range rankings {
		candidate := types.RankingCandidate{
			Name:         ranking.Name,
			Class:        normalizeClass(ranking.Class),
			Spec:         normalizeSpec(ranking.Spec),
			Server:       ranking.Server.Name,
			ServerRegion: ranking.Server.Region,
			ReportID:     ranking.Report.Code,
			FightID:      ranking.Report.FightID,
			RankValue:    ranking.Amount,
			DurationMS:   ranking.Duration,
		}
		if isSimilarFightDuration(targetDurationMS, ranking.Duration) {
			candidates = append(candidates, candidate)
		} else {
			fallbackCandidates = append(fallbackCandidates, candidate)
		}
		if len(candidates) == limit {
			break
		}
	}

	for _, candidate := range fallbackCandidates {
		if len(candidates) == limit {
			break
		}
		candidates = append(candidates, candidate)
	}

	return candidates, nil
}

func (c *WCLHTTPClient) GetCohortMemberData(candidate types.RankingCandidate) (*types.PlayerFightData, error) {
	ranking := WCLRankingEntry{
		Name:     candidate.Name,
		Class:    candidate.Class,
		Spec:     candidate.Spec,
		Amount:   candidate.RankValue,
		Duration: candidate.DurationMS,
		Report: WCLRankingReport{
			Code:    candidate.ReportID,
			FightID: candidate.FightID,
		},
		Server: WCLRankingServer{
			Name:   candidate.Server,
			Region: candidate.ServerRegion,
		},
	}

	data, err := c.fetchRankedPlayerFightData(ranking)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// makeGraphQLRequest performs a GraphQL request to the WCL API
func (c *WCLHTTPClient) makeGraphQLRequest(query string, response interface{}) error {
	return c.makeAuthorizedGraphQLRequest(c.config.BaseURL+"/client", "", query, response)
}

func (c *WCLHTTPClient) makeUserGraphQLRequest(accessToken, query string, response interface{}) error {
	return c.makeAuthorizedGraphQLRequest(c.config.BaseURL+"/user", accessToken, query, response)
}

func (c *WCLHTTPClient) makeAuthorizedGraphQLRequest(endpoint, accessToken, query string, response interface{}) error {
	requestBody := map[string]string{
		"query": query,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	} else if c.config.ClientID != "" && c.config.ClientSecret != "" {
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var graphqlErr graphqlErrorResponse
	if err := json.Unmarshal(body, &graphqlErr); err == nil && len(graphqlErr.Errors) > 0 {
		messages := make([]string, 0, len(graphqlErr.Errors))
		for _, graphqlMessage := range graphqlErr.Errors {
			if strings.TrimSpace(graphqlMessage.Message) != "" {
				messages = append(messages, graphqlMessage.Message)
			}
		}
		if len(messages) > 0 {
			return fmt.Errorf("graphql errors: %s", strings.Join(messages, "; "))
		}
		return fmt.Errorf("graphql request failed with unknown errors")
	}

	if err := json.Unmarshal(body, response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

func (c *WCLHTTPClient) GetCurrentUser(accessToken string) (*types.UserProfile, error) {
	query := `
		query {
			userData {
				currentUser {
					id
					name
					avatar
					battleTag
				}
			}
		}`

	var response struct {
		Data struct {
			UserData struct {
				CurrentUser WCLCurrentUser `json:"currentUser"`
			} `json:"userData"`
		} `json:"data"`
	}

	if err := c.makeUserGraphQLRequest(accessToken, query, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch current user: %w", err)
	}

	user := response.Data.UserData.CurrentUser
	return &types.UserProfile{
		ID:        user.ID,
		Name:      user.Name,
		Avatar:    user.Avatar,
		BattleTag: user.BattleTag,
	}, nil
}

func (c *WCLHTTPClient) GetOwnedCharacters(accessToken string) ([]types.OwnedCharacter, error) {
	query := `
		query {
			userData {
				currentUser {
					characters {
						canonicalID
						name
						classID
						server {
							name
							slug
							region {
								name
							}
						}
						recentReports(limit: 10, page: 1) {
							data {
								rankedCharacters {
									canonicalID
									name
									classID
									server {
										name
									}
								}
							}
						}
					}
				}
			}
		}`

	var response struct {
		Data struct {
			UserData struct {
				CurrentUser WCLCurrentUser `json:"currentUser"`
			} `json:"userData"`
		} `json:"data"`
	}

	if err := c.makeUserGraphQLRequest(accessToken, query, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch owned characters: %w", err)
	}

	rawCharacters := response.Data.UserData.CurrentUser.Characters
	characters := make([]types.OwnedCharacter, 0, len(rawCharacters))
	for _, character := range rawCharacters {
		className := classNameFromID(character.ClassID)
		if derivedClassID, ok := deriveCharacterClassIDFromRecentReports(character); ok {
			className = classNameFromID(derivedClassID)
		}

		characters = append(characters, types.OwnedCharacter{
			ID:           character.CanonicalID,
			Name:         character.Name,
			Class:        className,
			ServerName:   character.Server.Name,
			ServerRegion: character.Server.Region.Name,
			ServerSlug:   character.Server.Slug,
		})
	}

	sort.Slice(characters, func(i, j int) bool {
		if characters[i].Name == characters[j].Name {
			return characters[i].ServerName < characters[j].ServerName
		}
		return characters[i].Name < characters[j].Name
	})

	return characters, nil
}

func (c *WCLHTTPClient) GetCharacterReports(accessToken string, characterID int, cursor string, limit int) (*types.CharacterReportsPage, error) {
	if characterID == 0 {
		return nil, fmt.Errorf("character id is required")
	}
	if limit <= 0 {
		limit = 10
	}

	page, offset := parseCharacterReportsCursor(cursor)
	rawLimit := limit * 3
	if rawLimit < 25 {
		rawLimit = 25
	}
	if rawLimit > 50 {
		rawLimit = 50
	}
	cutoff := time.Now().AddDate(0, 0, -30)

	collected := make([]types.CharacterReportSummary, 0, limit)
	currentPage := page
	currentOffset := offset

	for len(collected) < limit {
		character, err := c.getCurrentUserCharacterReports(accessToken, characterID, currentPage, rawLimit)
		if err != nil {
			return nil, err
		}

		matchingReports := make([]WCLReportSummary, 0, len(character.RecentReports.Data))
		reachedCutoff := false
		for _, report := range character.RecentReports.Data {
			reportStart := time.UnixMilli(report.StartTime)
			if reportStart.Before(cutoff) {
				reachedCutoff = true
				break
			}
			if !isAllowedRaidReportTitle(report.Title) {
				continue
			}
			if !hasSuccessfulBossKill(report.Fights) {
				continue
			}
			matchingReports = append(matchingReports, report)
		}

		for index := currentOffset; index < len(matchingReports); index++ {
			report := matchingReports[index]

			zoneName := ""
			if report.Zone != nil {
				zoneName = report.Zone.Name
			}

			collected = append(collected, types.CharacterReportSummary{
				Code:      report.Code,
				Title:     report.Title,
				ZoneName:  zoneName,
				BossNames: extractKilledBossNames(report.Fights),
				StartTime: time.UnixMilli(report.StartTime),
				EndTime:   time.UnixMilli(report.EndTime),
			})
			if len(collected) == limit {
				nextOffset := index + 1
				nextCursor := ""
				if nextOffset < len(matchingReports) {
					nextCursor = formatCharacterReportsCursor(currentPage, nextOffset)
				} else if character.RecentReports.HasMorePages {
					nextCursor = formatCharacterReportsCursor(currentPage+1, 0)
				}

				return &types.CharacterReportsPage{
					Reports:    collected,
					NextCursor: nextCursor,
					HasMore:    nextCursor != "",
				}, nil
			}
		}

		if reachedCutoff || !character.RecentReports.HasMorePages {
			break
		}
		currentPage++
		currentOffset = 0
	}

	nextCursor := ""
	return &types.CharacterReportsPage{
		Reports:    collected,
		NextCursor: nextCursor,
		HasMore:    false,
	}, nil
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
func (c *WCLHTTPClient) normalizeFights(reportStartTime int64, wclFights []WCLFight) []types.NormalizedFight {
	fights := make([]types.NormalizedFight, len(wclFights))
	for i, fight := range wclFights {
		startTime := absoluteReportTimestamp(reportStartTime, fight.StartTime)
		endTime := absoluteReportTimestamp(reportStartTime, fight.EndTime)
		fights[i] = types.NormalizedFight{
			ID:          fight.ID,
			Name:        fight.Name,
			StartTime:   time.UnixMilli(startTime),
			EndTime:     time.UnixMilli(endTime),
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
	case 17, 1:
		return "LFR"
	case 2, 3:
		return "Normal"
	case 4:
		return "Heroic"
	case 5:
		return "Mythic"
	default:
		return "Unknown"
	}
}

func (c *WCLHTTPClient) fetchTopRankedCohortData(selectedFight types.NormalizedFight, selectedActor WCLActor) ([]types.PlayerFightData, error) {
	rankings, err := c.getEncounterRankings(selectedFight.EncounterID, selectedFight.Difficulty, selectedActor.Type, selectedActor.SubType, 10)
	if err != nil {
		return nil, err
	}

	type cohortResult struct {
		index int
		data  types.PlayerFightData
		err   error
	}

	results := make(chan cohortResult, len(rankings))
	var wg sync.WaitGroup

	for index, ranking := range rankings {
		wg.Add(1)
		go func(index int, ranking WCLRankingEntry) {
			defer wg.Done()

			data, err := c.fetchRankedPlayerFightData(ranking)
			results <- cohortResult{
				index: index,
				data:  data,
				err:   err,
			}
		}(index, ranking)
	}

	wg.Wait()
	close(results)

	ordered := make([]types.PlayerFightData, 0, 10)
	byIndex := make(map[int]types.PlayerFightData, len(rankings))
	for result := range results {
		if result.err != nil {
			continue
		}
		byIndex[result.index] = result.data
	}

	for index := range rankings {
		data, ok := byIndex[index]
		if !ok {
			continue
		}
		ordered = append(ordered, data)
		if len(ordered) == 10 {
			break
		}
	}

	return ordered, nil
}

func (c *WCLHTTPClient) getEncounterRankings(encounterID int, difficultyName, className, specName string, limit int) ([]WCLRankingEntry, error) {
	const maxRankingPage = 20
	if limit <= 0 {
		limit = 10
	}

	filteredClass := rankingClassFilterValue(className)
	filteredSpec := rankingSpecFilterValue(specName)

	return c.collectEncounterRankings(encounterID, difficultyName, className, filteredClass, filteredSpec, maxRankingPage, limit)
}

func (c *WCLHTTPClient) collectEncounterRankings(encounterID int, difficultyName, className, classFilter, specFilter string, maxRankingPage int, limit int) ([]WCLRankingEntry, error) {
	page := 1
	results := make([]WCLRankingEntry, 0, limit)

	for len(results) < limit && page <= maxRankingPage {
		rankingPage, err := c.fetchEncounterRankingsPage(encounterID, difficultyName, classFilter, specFilter, page)
		if err != nil {
			return nil, err
		}

		for _, ranking := range rankingPage.Rankings {
			if className != "" && classMatchKey(ranking.Class) != classMatchKey(className) {
				continue
			}
			if specFilter != "" && specMatchKey(ranking.Spec) != specMatchKey(specFilter) {
				continue
			}
			results = append(results, ranking)
			if len(results) == limit {
				break
			}
		}

		if !rankingPage.HasMorePages || len(rankingPage.Rankings) == 0 {
			break
		}
		page++
	}

	return results, nil
}

func (c *WCLHTTPClient) fetchEncounterRankingsPage(encounterID int, difficultyName, classFilter, specFilter string, page int) (*WCLCharacterRankingsResponse, error) {
	difficultyArg := ""
	if difficultyID := rankingDifficultyID(difficultyName); difficultyID != 0 {
		difficultyArg = fmt.Sprintf(", difficulty: %d", difficultyID)
	}
	classArg := ""
	if classFilter != "" {
		classArg = fmt.Sprintf(`, className: "%s"`, classFilter)
	}
	specArg := ""
	if specFilter != "" {
		specArg = fmt.Sprintf(`, specName: "%s"`, specFilter)
	}

	query := fmt.Sprintf(`
		query {
			worldData {
				encounter(id: %d) {
					characterRankings(page: %d%s%s%s, includeCombatantInfo: true, metric: default)
				}
			}
		}`, encounterID, page, difficultyArg, classArg, specArg)

	var response struct {
		Data struct {
			WorldData struct {
				Encounter struct {
					CharacterRankings WCLCharacterRankingsResponse `json:"characterRankings"`
				} `json:"encounter"`
			} `json:"worldData"`
		} `json:"data"`
	}

	var err error
	for attempt := 1; attempt <= 3; attempt++ {
		err = c.makeGraphQLRequest(query, &response)
		if err == nil {
			break
		}
		if !isRetryableWCLFailure(err) || attempt == 3 {
			return nil, fmt.Errorf("failed to fetch encounter rankings: %w", err)
		}
		time.Sleep(time.Duration(attempt) * 750 * time.Millisecond)
	}
	if response.Data.WorldData.Encounter.CharacterRankings.Error != "" {
		return nil, fmt.Errorf("warcraft logs rankings query failed: %s", response.Data.WorldData.Encounter.CharacterRankings.Error)
	}

	return &response.Data.WorldData.Encounter.CharacterRankings, nil
}

func (c *WCLHTTPClient) fetchRankedPlayerFightData(ranking WCLRankingEntry) (types.PlayerFightData, error) {
	fights, err := c.GetFights(ranking.Report.Code)
	if err != nil {
		return types.PlayerFightData{}, err
	}

	var fight *types.NormalizedFight
	for _, entry := range fights {
		if entry.ID == ranking.Report.FightID {
			entryCopy := entry
			fight = &entryCopy
			break
		}
	}
	if fight == nil {
		return types.PlayerFightData{}, fmt.Errorf("fight %d not found in report %s", ranking.Report.FightID, ranking.Report.Code)
	}

	characters, err := c.getFightCharacters(ranking.Report.Code, ranking.Report.FightID)
	if err != nil {
		return types.PlayerFightData{}, err
	}

	character, found := matchRankingToCharacter(characters, ranking)
	if !found {
		return types.PlayerFightData{}, fmt.Errorf("actor match not found for %s in report %s", ranking.Name, ranking.Report.Code)
	}

	data, err := c.fetchPlayerFightData(ranking.Report.Code, *fight, character.ID)
	if err != nil {
		return types.PlayerFightData{}, err
	}
	data.TalentImportCode = character.TalentImportCode
	data.TalentCalculatorURL = character.TalentCalculatorURL

	return data, nil
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

func (c *WCLHTTPClient) getReportAbilityNames(reportID string) (map[int]string, error) {
	query := fmt.Sprintf(`
		query {
			reportData {
				report(code: "%s") {
					masterData {
						abilities {
							gameID
							name
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
						Abilities []WCLAbility `json:"abilities"`
					} `json:"masterData"`
				} `json:"report"`
			} `json:"reportData"`
		} `json:"data"`
	}

	if err := c.makeGraphQLRequest(query, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch report abilities: %w", err)
	}

	abilityNames := make(map[int]string, len(response.Data.ReportData.Report.MasterData.Abilities))
	for _, ability := range response.Data.ReportData.Report.MasterData.Abilities {
		if ability.GameID == 0 || strings.TrimSpace(ability.Name) == "" {
			continue
		}
		abilityNames[ability.GameID] = ability.Name
	}

	return abilityNames, nil
}

func (c *WCLHTTPClient) getFightCharacters(reportID string, fightID int) ([]types.CharacterOption, error) {
	actors, err := c.getPlayerActors(reportID)
	if err != nil {
		return nil, err
	}

	serverByID := make(map[int]string, len(actors))
	for _, actor := range actors {
		if actor.ID == 0 {
			continue
		}
		serverByID[actor.ID] = actor.Server
	}

	query := fmt.Sprintf(`
		query {
			reportData {
				report(code: "%s") {
					playerDetails(
						fightIDs: [%d],
						includeCombatantInfo: false
					)
				}
			}
		}`, reportID, fightID)

	var response struct {
		Data struct {
			ReportData struct {
				Report struct {
					PlayerDetails json.RawMessage `json:"playerDetails"`
				} `json:"report"`
			} `json:"reportData"`
		} `json:"data"`
	}

	if err := c.makeGraphQLRequest(query, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch fight player details: %w", err)
	}

	characters, err := parseFightCharacters(
		response.Data.ReportData.Report.PlayerDetails,
		actors,
		serverByID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse fight player details: %w", err)
	}

	c.addTalentBuilds(reportID, fightID, characters)

	sort.Slice(characters, func(i, j int) bool {
		return characters[i].Name < characters[j].Name
	})

	return characters, nil
}

func (c *WCLHTTPClient) addTalentBuilds(reportID string, fightID int, characters []types.CharacterOption) {
	if len(characters) == 0 {
		return
	}

	fields := make([]string, 0, len(characters))
	aliases := make(map[string]int, len(characters))
	for index, character := range characters {
		lookupIDs := character.TalentLookupIDs
		if len(lookupIDs) == 0 {
			lookupIDs = []int{character.ID}
		}
		for candidateIndex, actorID := range lookupIDs {
			if actorID == 0 {
				continue
			}
			alias := fmt.Sprintf("talent%d_%d", index, candidateIndex)
			aliases[alias] = index
			fields = append(fields, fmt.Sprintf("%s: talentImportCode(actorID: %d)", alias, actorID))
		}
	}
	if len(fields) == 0 {
		return
	}

	query := fmt.Sprintf(`
		query {
			reportData {
				report(code: "%s") {
					fights(fightIDs: [%d]) {
						%s
					}
				}
			}
		}`, reportID, fightID, strings.Join(fields, "\n\t\t\t\t\t\t"))

	var response struct {
		Data struct {
			ReportData struct {
				Report struct {
					Fights []map[string]*string `json:"fights"`
				} `json:"report"`
			} `json:"reportData"`
		} `json:"data"`
	}

	if err := c.makeGraphQLRequest(query, &response); err != nil {
		return
	}
	if len(response.Data.ReportData.Report.Fights) == 0 {
		return
	}

	for alias, code := range response.Data.ReportData.Report.Fights[0] {
		index, ok := aliases[alias]
		if !ok || code == nil || strings.TrimSpace(*code) == "" {
			continue
		}
		if characters[index].TalentImportCode != "" {
			continue
		}
		trimmedCode := strings.TrimSpace(*code)
		characters[index].TalentImportCode = trimmedCode
		characters[index].TalentCalculatorURL = "https://www.wowhead.com/talent-calc/blizzard/" + url.PathEscape(trimmedCode)
	}
}

func (c *WCLHTTPClient) fetchPlayerFightData(reportID string, fight types.NormalizedFight, actorID int) (types.PlayerFightData, error) {
	reportMetadata, err := c.GetReportMetadata(reportID)
	if err != nil {
		return types.PlayerFightData{}, err
	}

	abilityNames, err := c.getReportAbilityNames(reportID)
	if err != nil {
		return types.PlayerFightData{}, err
	}

	castRaw, err := c.fetchEvents(reportID, "Casts", fight.ID, actorID)
	if err != nil {
		return types.PlayerFightData{}, err
	}
	damageRaw, err := c.fetchEvents(reportID, "DamageDone", fight.ID, actorID)
	if err != nil {
		return types.PlayerFightData{}, err
	}
	healRaw, err := c.fetchEvents(reportID, "Healing", fight.ID, actorID)
	if err != nil {
		return types.PlayerFightData{}, err
	}
	resourceRaw, err := c.fetchEvents(reportID, "Resources", fight.ID, actorID)
	if err != nil {
		return types.PlayerFightData{}, err
	}
	buffRaw, err := c.fetchBuffEvents(reportID, fight.ID, actorID)
	if err != nil {
		return types.PlayerFightData{}, err
	}

	return types.PlayerFightData{
		PlayerID:       actorID,
		FightID:        fight.ID,
		FightStart:     fight.StartTime,
		FightEnd:       fight.EndTime,
		CastEvents:     normalizeCastEvents(reportMetadata.StartTime.UnixMilli(), castRaw, abilityNames),
		DamageEvents:   normalizeDamageEvents(reportMetadata.StartTime.UnixMilli(), damageRaw, abilityNames),
		HealEvents:     normalizeHealEvents(reportMetadata.StartTime.UnixMilli(), healRaw, abilityNames),
		BuffEvents:     normalizeBuffEvents(reportMetadata.StartTime.UnixMilli(), buffRaw, abilityNames),
		CooldownEvents: normalizeCooldownEvents(reportMetadata.StartTime.UnixMilli(), castRaw, abilityNames),
		ResourceEvents: normalizeResourceEvents(reportMetadata.StartTime.UnixMilli(), resourceRaw, actorID),
	}, nil
}

func (c *WCLHTTPClient) fetchEvents(reportID, dataType string, fightID, sourceID int) ([]map[string]interface{}, error) {
	collected := make([]map[string]interface{}, 0)
	var nextStart *int64

	for {
		startClause := ""
		if nextStart != nil {
			startClause = fmt.Sprintf("\n\t\t\t\t\t\t\tstartTime: %d,", *nextStart)
		}

		query := fmt.Sprintf(`
			query {
				reportData {
					report(code: "%s") {
						events(
							dataType: %s,
							fightIDs: [%d],
							sourceID: %d,
%s
							limit: 10000
						) {
							data
							nextPageTimestamp
						}
					}
				}
			}`, reportID, dataType, fightID, sourceID, startClause)

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
		if response.Data.ReportData.Report.Events.NextPageTimestamp == nil {
			break
		}
		nextValue := int64(*response.Data.ReportData.Report.Events.NextPageTimestamp)
		if nextStart != nil && nextValue <= *nextStart {
			break
		}
		nextStart = &nextValue
	}

	return collected, nil
}

func (c *WCLHTTPClient) fetchBuffEvents(reportID string, fightID, actorID int) ([]map[string]interface{}, error) {
	collected := make([]map[string]interface{}, 0)
	var nextStart *int64

	for {
		startClause := ""
		if nextStart != nil {
			startClause = fmt.Sprintf("\n\t\t\t\t\t\t\tstartTime: %d,", *nextStart)
		}

		query := fmt.Sprintf(`
			query {
				reportData {
					report(code: "%s") {
						events(
							dataType: Buffs,
							fightIDs: [%d],
							targetID: %d,
%s
							limit: 10000
						) {
							data
							nextPageTimestamp
						}
					}
				}
			}`, reportID, fightID, actorID, startClause)

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
		if response.Data.ReportData.Report.Events.NextPageTimestamp == nil {
			break
		}
		nextValue := int64(*response.Data.ReportData.Report.Events.NextPageTimestamp)
		if nextStart != nil && nextValue <= *nextStart {
			break
		}
		nextStart = &nextValue
	}

	return collected, nil
}

func (c *WCLHTTPClient) getCurrentUserCharacterReports(accessToken string, characterID, page, limit int) (*WCLUserCharacter, error) {
	query := fmt.Sprintf(`
		query {
			userData {
				currentUser {
					characters {
						canonicalID
						name
						classID
						server {
							name
							slug
							region {
								name
							}
						}
						recentReports(limit: %d, page: %d) {
							data {
								code
								title
								startTime
								endTime
								zone {
									id
									name
								}
								fights {
									id
									name
									encounterID
									difficulty
									kill
								}
							}
							current_page
							last_page
							has_more_pages
						}
					}
				}
			}
		}`, limit, page)

	var response struct {
		Data struct {
			UserData struct {
				CurrentUser struct {
					Characters []WCLUserCharacter `json:"characters"`
				} `json:"currentUser"`
			} `json:"userData"`
		} `json:"data"`
	}

	if err := c.makeUserGraphQLRequest(accessToken, query, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch character recent reports: %w", err)
	}

	for _, character := range response.Data.UserData.CurrentUser.Characters {
		if character.CanonicalID == characterID {
			characterCopy := character
			return &characterCopy, nil
		}
	}

	return nil, fmt.Errorf("character %d was not found for the current user", characterID)
}
