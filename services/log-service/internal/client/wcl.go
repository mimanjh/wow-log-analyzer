package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
}

// WCLHTTPClient implements WCLClient using HTTP requests
type WCLHTTPClient struct {
	config     config.WCLConfig
	httpClient httpDoer
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
		req.SetBasicAuth(c.config.ClientID, c.config.ClientSecret)
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
