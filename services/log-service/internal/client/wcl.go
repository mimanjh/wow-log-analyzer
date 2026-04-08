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

	rankings, err := c.getEncounterRankings(selectedFight.EncounterID, selectedFight.Difficulty, characterClass, characterSpec)
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
		fmt.Printf(
			"WCL owned character raw: canonicalID=%d name=%s classID=%d server=%s region=%s recentReports=%d\n",
			character.CanonicalID,
			character.Name,
			character.ClassID,
			character.Server.Name,
			character.Server.Region.Name,
			len(character.RecentReports.Data),
		)

		classID := character.ClassID
		className := classNameFromID(character.ClassID)
		classSource := "top-level classID"
		if derivedClassID, ok := deriveCharacterClassIDFromRecentReports(character); ok {
			classID = derivedClassID
			className = classNameFromID(derivedClassID)
			classSource = "recentReports exact match"
		}
		fmt.Printf(
			"Owned character class mapping: name=%s canonicalID=%d server=%s topLevelClassID=%d topLevelClass=%s finalClassID=%d finalClass=%s source=%s\n",
			character.Name,
			character.CanonicalID,
			character.Server.Name,
			character.ClassID,
			classNameFromID(character.ClassID),
			classID,
			className,
			classSource,
		)

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
	rankings, err := c.getEncounterRankings(selectedFight.EncounterID, selectedFight.Difficulty, selectedActor.Type, selectedActor.SubType)
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

func (c *WCLHTTPClient) getEncounterRankings(encounterID int, difficultyName, className, specName string) ([]WCLRankingEntry, error) {
	page := 1
	results := make([]WCLRankingEntry, 0, 10)

	for len(results) < 10 {
		rankingPage, err := c.fetchEncounterRankingsPage(encounterID, difficultyName, page)
		if err != nil {
			return nil, err
		}

		for _, ranking := range rankingPage.Rankings {
			if normalizeClass(ranking.Class) != normalizeClass(className) {
				continue
			}
			if normalizeSpec(ranking.Spec) != normalizeSpec(specName) {
				continue
			}
			results = append(results, ranking)
			if len(results) == 10 {
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

func (c *WCLHTTPClient) fetchEncounterRankingsPage(encounterID int, difficultyName string, page int) (*WCLCharacterRankingsResponse, error) {
	difficultyArg := ""
	if difficultyID := rankingDifficultyID(difficultyName); difficultyID != 0 {
		difficultyArg = fmt.Sprintf(", difficulty: %d", difficultyID)
	}

	query := fmt.Sprintf(`
		query {
			worldData {
				encounter(id: %d) {
					characterRankings(page: %d%s, includeCombatantInfo: true, metric: default)
				}
			}
		}`, encounterID, page, difficultyArg)

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

	return c.fetchPlayerFightData(ranking.Report.Code, *fight, character.ID)
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

	sort.Slice(characters, func(i, j int) bool {
		return characters[i].Name < characters[j].Name
	})

	return characters, nil
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

func matchRankingToCharacter(characters []types.CharacterOption, ranking WCLRankingEntry) (types.CharacterOption, bool) {
	rankingName := strings.TrimSpace(ranking.Name)
	rankingClass := normalizeClass(ranking.Class)
	rankingSpec := normalizeSpec(ranking.Spec)

	for _, character := range characters {
		if strings.TrimSpace(character.Name) != rankingName {
			continue
		}
		if rankingClass != "" && normalizeClass(character.Class) != rankingClass {
			continue
		}
		if rankingSpec != "" && normalizeSpec(character.Spec) != rankingSpec {
			continue
		}
		return character, true
	}

	for _, character := range characters {
		if strings.TrimSpace(character.Name) == rankingName {
			return character, true
		}
	}

	return types.CharacterOption{}, false
}

func normalizeCastEvents(reportStartTime int64, raw []map[string]interface{}, abilityNames map[int]string) []types.CastEvent {
	events := make([]types.CastEvent, 0, len(raw))
	for _, event := range raw {
		events = append(events, types.CastEvent{
			Timestamp: parseTimestamp(reportStartTime, event["timestamp"]),
			Ability:   parseAbility(event, abilityNames),
			SourceID:  parseActorID(event, "sourceID", "source"),
		})
	}
	return events
}

func normalizeDamageEvents(reportStartTime int64, raw []map[string]interface{}, abilityNames map[int]string) []types.DamageEvent {
	events := make([]types.DamageEvent, 0, len(raw))
	for _, event := range raw {
		events = append(events, types.DamageEvent{
			Timestamp: parseTimestamp(reportStartTime, event["timestamp"]),
			Ability:   parseAbility(event, abilityNames),
			SourceID:  parseActorID(event, "sourceID", "source"),
			TargetID:  parseActorID(event, "targetID", "target"),
			Amount:    parseInt(event["amount"]),
		})
	}
	return events
}

func normalizeHealEvents(reportStartTime int64, raw []map[string]interface{}, abilityNames map[int]string) []types.HealEvent {
	events := make([]types.HealEvent, 0, len(raw))
	for _, event := range raw {
		events = append(events, types.HealEvent{
			Timestamp: parseTimestamp(reportStartTime, event["timestamp"]),
			Ability:   parseAbility(event, abilityNames),
			SourceID:  parseActorID(event, "sourceID", "source"),
			TargetID:  parseActorID(event, "targetID", "target"),
			Amount:    parseInt(event["amount"]),
		})
	}
	return events
}

func normalizeBuffEvents(reportStartTime int64, raw []map[string]interface{}, abilityNames map[int]string) []types.BuffEvent {
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

		ability := parseAbility(event, abilityNames)
		ability.IsBuff = true
		events = append(events, types.BuffEvent{
			Timestamp: parseTimestamp(reportStartTime, event["timestamp"]),
			Ability:   ability,
			SourceID:  parseActorID(event, "sourceID", "source"),
			TargetID:  parseActorID(event, "targetID", "target"),
			EventType: normalizedType,
		})
	}
	return events
}

func normalizeCooldownEvents(reportStartTime int64, raw []map[string]interface{}, abilityNames map[int]string) []types.CooldownEvent {
	events := make([]types.CooldownEvent, 0)
	seen := map[int]int{}

	for _, event := range raw {
		ability := parseAbility(event, abilityNames)
		if ability.ID == 0 {
			continue
		}
		seen[ability.ID]++
	}

	for _, event := range raw {
		ability := parseAbility(event, abilityNames)
		if ability.ID == 0 || seen[ability.ID] > 5 {
			continue
		}
		ability.IsMajorCD = true
		events = append(events, types.CooldownEvent{
			Timestamp: parseTimestamp(reportStartTime, event["timestamp"]),
			Ability:   ability,
			SourceID:  parseActorID(event, "sourceID", "source"),
			EventType: "start",
		})
	}

	return events
}

func parseTimestamp(reportStartTime int64, value interface{}) time.Time {
	ms := absoluteReportTimestamp(reportStartTime, parseInt64(value))
	if ms <= 0 {
		return time.Time{}
	}
	return time.UnixMilli(ms)
}

func parseAbility(event map[string]interface{}, abilityNames map[int]string) types.Ability {
	if abilityValue, ok := event["ability"]; ok {
		if abilityMap, ok := abilityValue.(map[string]interface{}); ok {
			return types.Ability{
				ID:   parseInt(abilityMap["gameID"]),
				Name: parseString(abilityMap["name"]),
			}
		}
	}

	abilityID := parseInt(event["abilityGameID"])
	abilityName := parseString(event["abilityName"])
	if abilityName == "" && abilityID != 0 {
		abilityName = abilityNames[abilityID]
	}

	return types.Ability{
		ID:   abilityID,
		Name: abilityName,
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

func rankingDifficultyID(name string) int {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "lfr":
		return 17
	case "normal":
		return 3
	case "heroic":
		return 4
	case "mythic":
		return 5
	default:
		return 0
	}
}

func isSimilarFightDuration(targetMS, candidateMS int) bool {
	if targetMS <= 0 || candidateMS <= 0 {
		return true
	}

	allowedDifference := int(float64(targetMS) * 0.15)
	if allowedDifference < 15000 {
		allowedDifference = 15000
	}

	difference := targetMS - candidateMS
	if difference < 0 {
		difference = -difference
	}

	return difference <= allowedDifference
}

func normalizedFightFromSelection(selection types.FightSelection) types.NormalizedFight {
	return types.NormalizedFight{
		ID:          selection.ID,
		Name:        selection.Name,
		StartTime:   selection.StartTime,
		EndTime:     selection.EndTime,
		EncounterID: selection.EncounterID,
		Difficulty:  selection.Difficulty,
		BossPercent: selection.BossPercent,
	}
}

func isRetryableWCLFailure(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "status 504") ||
		strings.Contains(message, "status 503") ||
		strings.Contains(message, "status 502") ||
		strings.Contains(message, "timeout")
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
						className
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

func classNameFromID(classID int) string {
	switch classID {
	case 1:
		return "Death Knight"
	case 2:
		return "Druid"
	case 3:
		return "Hunter"
	case 4:
		return "Mage"
	case 5:
		return "Monk"
	case 6:
		return "Paladin"
	case 7:
		return "Priest"
	case 8:
		return "Rogue"
	case 9:
		return "Shaman"
	case 10:
		return "Warlock"
	case 11:
		return "Warrior"
	case 12:
		return "Demon Hunter"
	case 13:
		return "Evoker"
	default:
		return "Unknown"
	}
}

func deriveCharacterClassIDFromRecentReports(character WCLUserCharacter) (int, bool) {
	canonicalMatches := make(map[int]int)
	nameServerMatches := make(map[int]int)

	for _, report := range character.RecentReports.Data {
		for _, rankedCharacter := range report.RankedCharacters {
			if rankedCharacter.ClassID == 0 {
				continue
			}
			if rankedCharacter.CanonicalID != 0 && rankedCharacter.CanonicalID == character.CanonicalID {
				canonicalMatches[rankedCharacter.ClassID]++
				continue
			}
			if strings.EqualFold(strings.TrimSpace(rankedCharacter.Name), strings.TrimSpace(character.Name)) &&
				strings.EqualFold(strings.TrimSpace(rankedCharacter.Server.Name), strings.TrimSpace(character.Server.Name)) {
				nameServerMatches[rankedCharacter.ClassID]++
			}
		}
	}

	if classID, ok := mostFrequentClassID(canonicalMatches); ok {
		return classID, true
	}
	if classID, ok := mostFrequentClassID(nameServerMatches); ok {
		return classID, true
	}

	return 0, false
}

func mostFrequentClassID(counts map[int]int) (int, bool) {
	bestClassID := 0
	bestCount := 0

	for classID, count := range counts {
		if count > bestCount {
			bestClassID = classID
			bestCount = count
		}
	}

	if bestClassID == 0 || bestCount == 0 {
		return 0, false
	}

	return bestClassID, true
}

func isAllowedRaidReportTitle(title string) bool {
	normalized := strings.ToUpper(strings.TrimSpace(title))
	if normalized == "" {
		return false
	}

	return strings.Contains(normalized, "VS / DR / MQD")
}

func absoluteReportTimestamp(reportStartTime, value int64) int64 {
	if reportStartTime > 0 && value >= 0 && value < reportStartTime {
		return reportStartTime + value
	}
	return value
}

func parseFightCharacters(raw json.RawMessage, actors []WCLActor, serverByID map[int]string) ([]types.CharacterOption, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	var payload interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}

	characters := make([]types.CharacterOption, 0)
	seen := make(map[int]struct{})
	collectFightCharacters(payload, "", actors, serverByID, seen, &characters)

	return characters, nil
}

func collectFightCharacters(node interface{}, role string, actors []WCLActor, serverByID map[int]string, seen map[int]struct{}, characters *[]types.CharacterOption) {
	switch typed := node.(type) {
	case map[string]interface{}:
		nextRole := role
		for key, value := range typed {
			switch strings.ToLower(strings.TrimSpace(key)) {
			case "tanks":
				collectFightCharacters(value, "Tank", actors, serverByID, seen, characters)
				continue
			case "healers":
				collectFightCharacters(value, "Healer", actors, serverByID, seen, characters)
				continue
			case "dps":
				collectFightCharacters(value, "DPS", actors, serverByID, seen, characters)
				continue
			}
		}

		if character, ok := buildCharacterOption(typed, nextRole, actors, serverByID); ok {
			if _, exists := seen[character.ID]; !exists {
				seen[character.ID] = struct{}{}
				*characters = append(*characters, character)
			}
		}

		for _, value := range typed {
			collectFightCharacters(value, nextRole, actors, serverByID, seen, characters)
		}
	case []interface{}:
		for _, value := range typed {
			collectFightCharacters(value, role, actors, serverByID, seen, characters)
		}
	}
}

func buildCharacterOption(payload map[string]interface{}, role string, actors []WCLActor, serverByID map[int]string) (types.CharacterOption, bool) {
	id := parseInt(payload["id"])
	name := parseString(payload["name"])
	className := normalizeClass(parseString(payload["type"]))

	if id == 0 || strings.TrimSpace(name) == "" || strings.TrimSpace(className) == "" {
		return types.CharacterOption{}, false
	}
	if strings.EqualFold(className, "Pet") {
		return types.CharacterOption{}, false
	}

	spec := extractSpecName(payload)
	if role == "" {
		role = inferRole(spec)
	}

	actorID := resolveActorID(id, name, className, actors)
	if actorID == 0 {
		actorID = id
	}

	return types.CharacterOption{
		ID:         actorID,
		Name:       name,
		Class:      className,
		Spec:       spec,
		Role:       role,
		ServerName: strings.TrimSpace(serverByID[actorID]),
	}, true
}

func resolveActorID(playerDetailID int, name, className string, actors []WCLActor) int {
	for _, actor := range actors {
		if actor.ID == playerDetailID {
			return actor.ID
		}
	}

	trimmedName := strings.TrimSpace(name)
	normalizedClass := normalizeClass(className)

	for _, actor := range actors {
		if strings.TrimSpace(actor.Name) != trimmedName {
			continue
		}
		if normalizedClass != "" && normalizeClass(actor.Type) != normalizedClass {
			continue
		}
		return actor.ID
	}

	for _, actor := range actors {
		if strings.TrimSpace(actor.Name) == trimmedName {
			return actor.ID
		}
	}

	return 0
}

func extractSpecName(payload map[string]interface{}) string {
	if value, ok := payload["spec"]; ok {
		if spec := normalizeSpec(parseString(value)); spec != "" {
			return spec
		}
	}

	specsValue, ok := payload["specs"]
	if !ok {
		return ""
	}

	switch typed := specsValue.(type) {
	case []interface{}:
		for _, entry := range typed {
			switch specEntry := entry.(type) {
			case string:
				if spec := normalizeSpec(specEntry); spec != "" {
					return spec
				}
			case map[string]interface{}:
				if spec := normalizeSpec(parseString(specEntry["spec"])); spec != "" {
					return spec
				}
				if spec := normalizeSpec(parseString(specEntry["name"])); spec != "" {
					return spec
				}
			}
		}
	}

	return ""
}

func extractKilledBossNames(fights []WCLFight) []string {
	if len(fights) == 0 {
		return nil
	}

	bossNames := make([]string, 0, len(fights))
	seen := make(map[string]struct{}, len(fights))
	for _, fight := range fights {
		name := strings.TrimSpace(fight.Name)
		if !isRelevantKilledBossFight(fight) || name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		bossNames = append(bossNames, name)
	}

	return bossNames
}

func hasSuccessfulBossKill(fights []WCLFight) bool {
	for _, fight := range fights {
		if isRelevantKilledBossFight(fight) && strings.TrimSpace(fight.Name) != "" {
			return true
		}
	}

	return false
}

func isRelevantKilledBossFight(fight WCLFight) bool {
	return fight.Kill &&
		fight.EncounterID != 0 &&
		fight.Difficulty != 0
}

func parseCharacterReportsCursor(cursor string) (int, int) {
	page := 1
	offset := 0
	trimmed := strings.TrimSpace(cursor)
	if trimmed == "" {
		return page, offset
	}

	parts := strings.Split(trimmed, ":")
	if len(parts) >= 1 {
		if parsedPage, err := strconv.Atoi(parts[0]); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}
	if len(parts) >= 2 {
		if parsedOffset, err := strconv.Atoi(parts[1]); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	return page, offset
}

func formatCharacterReportsCursor(page, offset int) string {
	return fmt.Sprintf("%d:%d", page, offset)
}
