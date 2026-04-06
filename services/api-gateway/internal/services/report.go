package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GenerateReportRequest struct {
	ReportID  string           `json:"reportId"`
	Fight     FightSummary     `json:"fight"`
	Character CharacterSummary `json:"character"`
}

type GenerateReportResponse struct {
	Fight      FightSummary     `json:"fight"`
	Character  CharacterSummary `json:"character"`
	Comparison ComparisonResult `json:"comparison"`
	AI         AIReportSection  `json:"ai"`
}

type AIReportSection struct {
	Available           bool                `json:"available"`
	FallbackUsed        bool                `json:"fallbackUsed"`
	Model               string              `json:"model,omitempty"`
	Warning             string              `json:"warning,omitempty"`
	Insights            []AIInsight         `json:"insights"`
	FocusRecommendation FocusRecommendation `json:"focusRecommendation"`
}

type ComparisonResult struct {
	PlayerMetrics PlayerFightMetrics `json:"playerMetrics"`
	CohortStats   CohortStatistics   `json:"cohortStats"`
	Deltas        MetricDeltas       `json:"deltas"`
	Rankings      MetricRankings     `json:"rankings"`
}

type PlayerFightMetrics struct {
	PlayerID     int                      `json:"playerId"`
	FightID      int                      `json:"fightId"`
	FightStart   time.Time                `json:"fightStart"`
	FightEnd     time.Time                `json:"fightEnd"`
	Duration     int64                    `json:"duration"`
	CastsPerMin  CastsPerMinuteMetric     `json:"castsPerMin"`
	MajorCDCount MajorCDCountMetric       `json:"majorCdCount"`
	MajorCDDrift MajorCDDriftMetric       `json:"majorCdDrift"`
	BuffUptime   BuffUptimeMetric         `json:"buffUptime"`
	DowntimePct  DowntimePercentageMetric `json:"downtimePct"`
}

type CastsPerMinuteMetric struct {
	Value         float64 `json:"value"`
	TotalCasts    int     `json:"totalCasts"`
	FightDuration float64 `json:"fightDuration"`
	Confidence    string  `json:"confidence"`
	Caution       string  `json:"caution,omitempty"`
}

type MajorCDCountMetric struct {
	Value          int    `json:"value"`
	TotalCooldowns int    `json:"totalCooldowns"`
	Confidence     string `json:"confidence"`
	Caution        string `json:"caution,omitempty"`
}

type MajorCDDriftMetric struct {
	Value         float64 `json:"value"`
	TotalDrift    float64 `json:"totalDrift"`
	CooldownCount int     `json:"cooldownCount"`
	Confidence    string  `json:"confidence"`
	Caution       string  `json:"caution,omitempty"`
}

type BuffUptimeMetric struct {
	Value         float64 `json:"value"`
	TotalUptime   float64 `json:"totalUptime"`
	FightDuration float64 `json:"fightDuration"`
	Confidence    string  `json:"confidence"`
	Caution       string  `json:"caution,omitempty"`
}

type DowntimePercentageMetric struct {
	Value         float64 `json:"value"`
	TotalDowntime float64 `json:"totalDowntime"`
	FightDuration float64 `json:"fightDuration"`
	Confidence    string  `json:"confidence"`
	Caution       string  `json:"caution,omitempty"`
}

type CohortStatistics struct {
	SampleSize   int               `json:"sampleSize"`
	CastsPerMin  CohortMetricStats `json:"castsPerMin"`
	MajorCDCount CohortMetricStats `json:"majorCdCount"`
	MajorCDDrift CohortMetricStats `json:"majorCdDrift"`
	BuffUptime   CohortMetricStats `json:"buffUptime"`
	DowntimePct  CohortMetricStats `json:"downtimePct"`
}

type CohortMetricStats struct {
	Mean   float64 `json:"mean"`
	Median float64 `json:"median"`
	StdDev float64 `json:"stdDev"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	P25    float64 `json:"p25"`
	P75    float64 `json:"p75"`
	P95    float64 `json:"p95"`
}

type MetricDeltas struct {
	CastsPerMin  MetricDelta `json:"castsPerMin"`
	MajorCDCount MetricDelta `json:"majorCdCount"`
	MajorCDDrift MetricDelta `json:"majorCdDrift"`
	BuffUptime   MetricDelta `json:"buffUptime"`
	DowntimePct  MetricDelta `json:"downtimePct"`
}

type MetricDelta struct {
	PlayerValue float64 `json:"playerValue"`
	CohortValue float64 `json:"cohortValue"`
	Difference  float64 `json:"difference"`
	Percentile  float64 `json:"percentile"`
	Confidence  string  `json:"confidence"`
	Caution     string  `json:"caution,omitempty"`
}

type MetricRankings struct {
	CastsPerMin  float64 `json:"castsPerMin"`
	MajorCDCount float64 `json:"majorCdCount"`
	MajorCDDrift float64 `json:"majorCdDrift"`
	BuffUptime   float64 `json:"buffUptime"`
	DowntimePct  float64 `json:"downtimePct"`
}

type logComparisonRequest struct {
	FightID     int `json:"fightId"`
	CharacterID int `json:"characterId"`
}

type logComparisonResponse struct {
	ReportID   string            `json:"reportId"`
	Fight      FightSummary      `json:"fight"`
	PlayerData json.RawMessage   `json:"playerData"`
	CohortData []json.RawMessage `json:"cohortData"`
}

type analysisCompareRequest struct {
	PlayerData json.RawMessage   `json:"playerData"`
	CohortData []json.RawMessage `json:"cohortData"`
}

type insightGenerationRequest struct {
	Context insightContext  `json:"context"`
	Metrics []insightMetric `json:"metrics"`
}

type insightContext struct {
	EncounterName    string `json:"encounterName"`
	Difficulty       string `json:"difficulty"`
	CharacterName    string `json:"characterName"`
	CharacterClass   string `json:"characterClass"`
	CharacterSpec    string `json:"characterSpec"`
	FightDurationSec int    `json:"fightDurationSec"`
	CohortSize       int    `json:"cohortSize"`
}

type insightMetric struct {
	Key            string  `json:"key"`
	Label          string  `json:"label"`
	Unit           string  `json:"unit,omitempty"`
	HigherIsBetter bool    `json:"higherIsBetter"`
	PlayerValue    float64 `json:"playerValue"`
	CohortValue    float64 `json:"cohortValue"`
	Difference     float64 `json:"difference"`
	Percentile     float64 `json:"percentile"`
	Confidence     string  `json:"confidence"`
	Caution        string  `json:"caution,omitempty"`
}

type AIInsight struct {
	MetricKey  string `json:"metricKey"`
	Title      string `json:"title"`
	Summary    string `json:"summary"`
	Confidence string `json:"confidence"`
	Caution    string `json:"caution,omitempty"`
}

type FocusRecommendation struct {
	MetricKey      string `json:"metricKey"`
	Title          string `json:"title"`
	Recommendation string `json:"recommendation"`
	Reasoning      string `json:"reasoning"`
}

type insightGenerationResponse struct {
	Insights            []AIInsight         `json:"insights"`
	FocusRecommendation FocusRecommendation `json:"focusRecommendation"`
	FallbackUsed        bool                `json:"fallbackUsed"`
	Model               string              `json:"model"`
}

type ReportService struct {
	logClient      *http.Client
	analysisClient *http.Client
	aiClient       *http.Client
	logURL         string
	analysisURL    string
	aiURL          string
}

func NewReportService(logURL, analysisURL, aiURL string) *ReportService {
	return &ReportService{
		logClient:      &http.Client{Timeout: 30 * time.Second},
		analysisClient: &http.Client{Timeout: 30 * time.Second},
		aiClient:       &http.Client{Timeout: 30 * time.Second},
		logURL:         logURL,
		analysisURL:    analysisURL,
		aiURL:          aiURL,
	}
}

func (s *ReportService) Generate(ctx context.Context, req GenerateReportRequest) (GenerateReportResponse, error) {
	if req.ReportID == "" {
		return GenerateReportResponse{}, fmt.Errorf("reportId is required")
	}
	if req.Fight.ID == 0 || req.Character.ID == 0 {
		return GenerateReportResponse{}, fmt.Errorf("fight and character selections are required")
	}

	logData, err := s.fetchNormalizedComparisonData(ctx, req)
	if err != nil {
		return GenerateReportResponse{}, fmt.Errorf("failed to retrieve normalized comparison data: %w", err)
	}

	comparison, err := s.fetchComparison(ctx, logData)
	if err != nil {
		return GenerateReportResponse{}, fmt.Errorf("failed to retrieve comparison results: %w", err)
	}

	response := GenerateReportResponse{
		Fight:      req.Fight,
		Character:  req.Character,
		Comparison: comparison,
		AI: AIReportSection{
			Available: false,
			Insights:  []AIInsight{},
		},
	}

	insights, err := s.fetchInsights(ctx, req, response.Comparison)
	if err != nil {
		response.AI.Warning = "AI insights were unavailable. Deterministic metrics are still shown."
		return response, nil
	}

	response.AI = AIReportSection{
		Available:           true,
		FallbackUsed:        insights.FallbackUsed,
		Model:               insights.Model,
		Insights:            insights.Insights,
		FocusRecommendation: insights.FocusRecommendation,
	}

	if insights.FallbackUsed {
		response.AI.Warning = "AI used the deterministic fallback formatter for this report."
	}

	return response, nil
}

func (s *ReportService) fetchNormalizedComparisonData(ctx context.Context, req GenerateReportRequest) (logComparisonResponse, error) {
	body, err := json.Marshal(logComparisonRequest{
		FightID:     req.Fight.ID,
		CharacterID: req.Character.ID,
	})
	if err != nil {
		return logComparisonResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/reports/%s/comparison-data", s.logURL, req.ReportID), bytes.NewReader(body))
	if err != nil {
		return logComparisonResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.logClient.Do(httpReq)
	if err != nil {
		return logComparisonResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return logComparisonResponse{}, fmt.Errorf("log-service returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var payload logComparisonResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return logComparisonResponse{}, err
	}

	return payload, nil
}

func (s *ReportService) fetchComparison(ctx context.Context, payload logComparisonResponse) (ComparisonResult, error) {
	body, err := json.Marshal(analysisCompareRequest{
		PlayerData: payload.PlayerData,
		CohortData: payload.CohortData,
	})
	if err != nil {
		return ComparisonResult{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.analysisURL+"/analyze/compare", bytes.NewReader(body))
	if err != nil {
		return ComparisonResult{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.analysisClient.Do(httpReq)
	if err != nil {
		return ComparisonResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return ComparisonResult{}, fmt.Errorf("analysis-service returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var comparison ComparisonResult
	if err := json.NewDecoder(resp.Body).Decode(&comparison); err != nil {
		return ComparisonResult{}, err
	}

	return comparison, nil
}

func (s *ReportService) fetchInsights(ctx context.Context, req GenerateReportRequest, comparison ComparisonResult) (insightGenerationResponse, error) {
	body, err := json.Marshal(buildInsightRequest(req, comparison))
	if err != nil {
		return insightGenerationResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.aiURL+"/insights/generate", bytes.NewReader(body))
	if err != nil {
		return insightGenerationResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.aiClient.Do(httpReq)
	if err != nil {
		return insightGenerationResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return insightGenerationResponse{}, fmt.Errorf("ai-service returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var insights insightGenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&insights); err != nil {
		return insightGenerationResponse{}, err
	}

	return insights, nil
}

func buildInsightRequest(req GenerateReportRequest, comparison ComparisonResult) insightGenerationRequest {
	return insightGenerationRequest{
		Context: insightContext{
			EncounterName:    req.Fight.Name,
			Difficulty:       req.Fight.Difficulty,
			CharacterName:    req.Character.Name,
			CharacterClass:   req.Character.Class,
			CharacterSpec:    req.Character.Spec,
			FightDurationSec: req.Fight.KillTime,
			CohortSize:       comparison.CohortStats.SampleSize,
		},
		Metrics: []insightMetric{
			{Key: "castsPerMin", Label: "Casts per Minute", HigherIsBetter: true, PlayerValue: comparison.Deltas.CastsPerMin.PlayerValue, CohortValue: comparison.Deltas.CastsPerMin.CohortValue, Difference: comparison.Deltas.CastsPerMin.Difference, Percentile: comparison.Deltas.CastsPerMin.Percentile, Confidence: comparison.Deltas.CastsPerMin.Confidence, Caution: comparison.Deltas.CastsPerMin.Caution},
			{Key: "majorCdCount", Label: "Major Cooldown Count", HigherIsBetter: true, PlayerValue: comparison.Deltas.MajorCDCount.PlayerValue, CohortValue: comparison.Deltas.MajorCDCount.CohortValue, Difference: comparison.Deltas.MajorCDCount.Difference, Percentile: comparison.Deltas.MajorCDCount.Percentile, Confidence: comparison.Deltas.MajorCDCount.Confidence, Caution: comparison.Deltas.MajorCDCount.Caution},
			{Key: "majorCdDrift", Label: "Major Cooldown Timing Drift", Unit: "s", HigherIsBetter: false, PlayerValue: comparison.Deltas.MajorCDDrift.PlayerValue, CohortValue: comparison.Deltas.MajorCDDrift.CohortValue, Difference: comparison.Deltas.MajorCDDrift.Difference, Percentile: comparison.Deltas.MajorCDDrift.Percentile, Confidence: comparison.Deltas.MajorCDDrift.Confidence, Caution: comparison.Deltas.MajorCDDrift.Caution},
			{Key: "buffUptime", Label: "Buff Uptime", Unit: "%", HigherIsBetter: true, PlayerValue: comparison.Deltas.BuffUptime.PlayerValue, CohortValue: comparison.Deltas.BuffUptime.CohortValue, Difference: comparison.Deltas.BuffUptime.Difference, Percentile: comparison.Deltas.BuffUptime.Percentile, Confidence: comparison.Deltas.BuffUptime.Confidence, Caution: comparison.Deltas.BuffUptime.Caution},
			{Key: "downtimePct", Label: "Downtime Percentage", Unit: "%", HigherIsBetter: false, PlayerValue: comparison.Deltas.DowntimePct.PlayerValue, CohortValue: comparison.Deltas.DowntimePct.CohortValue, Difference: comparison.Deltas.DowntimePct.Difference, Percentile: comparison.Deltas.DowntimePct.Percentile, Confidence: comparison.Deltas.DowntimePct.Confidence, Caution: comparison.Deltas.DowntimePct.Caution},
		},
	}
}
