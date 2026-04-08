package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
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
	Fight       FightSummary `json:"fight"`
	CharacterID int          `json:"characterId"`
}

type logRankingCandidatesRequest struct {
	Fight          FightSummary `json:"fight"`
	CharacterClass string       `json:"characterClass"`
	CharacterSpec  string       `json:"characterSpec"`
	Limit          int          `json:"limit"`
}

type logCohortMemberRequest struct {
	Candidate RankingCandidate `json:"candidate"`
}

type RankingCandidate struct {
	Name         string  `json:"name"`
	Class        string  `json:"class"`
	Spec         string  `json:"spec"`
	Server       string  `json:"server,omitempty"`
	ServerRegion string  `json:"serverRegion,omitempty"`
	ReportID     string  `json:"reportId"`
	FightID      int     `json:"fightId"`
	RankValue    float64 `json:"rankValue"`
	DurationMS   int     `json:"durationMs"`
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

type ReportJobStatus string

const (
	ReportJobQueued    ReportJobStatus = "queued"
	ReportJobRunning   ReportJobStatus = "running"
	ReportJobCompleted ReportJobStatus = "completed"
	ReportJobFailed    ReportJobStatus = "failed"
)

type ReportJob struct {
	ID        string                  `json:"jobId"`
	Status    ReportJobStatus         `json:"status"`
	Stage     string                  `json:"stage"`
	Message   string                  `json:"message"`
	Fight     FightSummary            `json:"fight"`
	Character CharacterSummary        `json:"character"`
	Progress  ReportJobProgress       `json:"progress"`
	Error     string                  `json:"error,omitempty"`
	Result    *GenerateReportResponse `json:"result,omitempty"`
	CreatedAt time.Time               `json:"createdAt"`
	UpdatedAt time.Time               `json:"updatedAt"`
}

type ReportJobProgress struct {
	Current int `json:"current"`
	Total   int `json:"total"`
}

type ReportService struct {
	logClient      *http.Client
	analysisClient *http.Client
	aiClient       *http.Client
	logURL         string
	analysisURL    string
	aiURL          string

	jobMu sync.RWMutex
	jobs  map[string]ReportJob
}

func NewReportService(logURL, analysisURL, aiURL string) *ReportService {
	const serviceTimeout = 120 * time.Second

	return &ReportService{
		logClient:      &http.Client{Timeout: serviceTimeout},
		analysisClient: &http.Client{Timeout: serviceTimeout},
		aiClient:       &http.Client{Timeout: serviceTimeout},
		logURL:         logURL,
		analysisURL:    analysisURL,
		aiURL:          aiURL,
		jobs:           make(map[string]ReportJob),
	}
}

func (s *ReportService) CreateJob(req GenerateReportRequest) (ReportJob, error) {
	if req.ReportID == "" {
		return ReportJob{}, fmt.Errorf("reportId is required")
	}
	if req.Fight.ID == 0 || req.Character.ID == 0 {
		return ReportJob{}, fmt.Errorf("fight and character selections are required")
	}

	job := ReportJob{
		ID:        newJobID(),
		Status:    ReportJobQueued,
		Stage:     "queued",
		Message:   "Queued for report generation.",
		Fight:     req.Fight,
		Character: req.Character,
		Progress: ReportJobProgress{
			Current: 0,
			Total:   5,
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	s.setJob(job)

	go s.runJob(job.ID, req)

	return job, nil
}

func (s *ReportService) GetJob(jobID string) (ReportJob, error) {
	s.jobMu.RLock()
	defer s.jobMu.RUnlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return ReportJob{}, fmt.Errorf("report job %s not found", jobID)
	}

	return job, nil
}

func (s *ReportService) runJob(jobID string, req GenerateReportRequest) {
	ctx := context.Background()

	s.updateJob(jobID, ReportJobRunning, "player-data", "Fetching selected player fight data.", ReportJobProgress{Current: 1, Total: 5}, "", nil)
	playerData, err := s.fetchPlayerData(ctx, req)
	if err != nil {
		s.updateJob(jobID, ReportJobFailed, "player-data", "Failed to fetch selected player fight data.", ReportJobProgress{Current: 1, Total: 5}, err.Error(), nil)
		return
	}

	s.updateJob(jobID, ReportJobRunning, "rankings", "Fetching ranking candidates for the selected boss and spec.", ReportJobProgress{Current: 2, Total: 5}, "", nil)
	candidates, err := s.fetchRankingCandidates(ctx, req)
	if err != nil {
		s.updateJob(jobID, ReportJobFailed, "rankings", "Failed to fetch ranking candidates.", ReportJobProgress{Current: 2, Total: 5}, err.Error(), nil)
		return
	}
	if len(candidates) == 0 {
		s.updateJob(jobID, ReportJobFailed, "rankings", "No ranking candidates were available for this fight.", ReportJobProgress{Current: 2, Total: 5}, "no ranking candidates returned", nil)
		return
	}

	cohortData := make([]json.RawMessage, 0, len(candidates))
	for index, candidate := range candidates {
		s.updateJob(jobID, ReportJobRunning, "cohort", fmt.Sprintf("Fetching cohort member %d of %d.", index+1, len(candidates)), ReportJobProgress{Current: index + 1, Total: len(candidates)}, "", nil)
		memberData, err := s.fetchCohortMember(ctx, candidate)
		if err != nil {
			continue
		}
		cohortData = append(cohortData, memberData)
	}
	if len(cohortData) == 0 {
		s.updateJob(jobID, ReportJobFailed, "cohort", "Failed to collect any cohort members.", ReportJobProgress{Current: len(candidates), Total: len(candidates)}, "no cohort member data could be fetched", nil)
		return
	}

	s.updateJob(jobID, ReportJobRunning, "analyzing", "Running deterministic comparison analysis.", ReportJobProgress{Current: 4, Total: 5}, "", nil)
	comparison, err := s.fetchComparison(ctx, playerData, cohortData)
	if err != nil {
		s.updateJob(jobID, ReportJobFailed, "analyzing", "Failed to compute deterministic comparison metrics.", ReportJobProgress{Current: 4, Total: 5}, err.Error(), nil)
		return
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

	s.updateJob(jobID, ReportJobRunning, "insights", "Generating AI insights.", ReportJobProgress{Current: 5, Total: 5}, "", &response)
	insights, err := s.fetchInsights(ctx, req, response.Comparison)
	if err != nil {
		response.AI.Warning = "AI insights were unavailable. Deterministic metrics are still shown."
		s.updateJob(jobID, ReportJobCompleted, "completed", "Report completed without AI insights.", ReportJobProgress{Current: 5, Total: 5}, "", &response)
		return
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

	s.updateJob(jobID, ReportJobCompleted, "completed", "Report completed.", ReportJobProgress{Current: 5, Total: 5}, "", &response)
}

func (s *ReportService) setJob(job ReportJob) {
	s.jobMu.Lock()
	defer s.jobMu.Unlock()
	s.jobs[job.ID] = job
}

func (s *ReportService) updateJob(jobID string, status ReportJobStatus, stage, message string, progress ReportJobProgress, errText string, result *GenerateReportResponse) {
	s.jobMu.Lock()
	defer s.jobMu.Unlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return
	}

	job.Status = status
	job.Stage = stage
	job.Message = message
	job.Progress = progress
	job.Error = errText
	job.Result = result
	job.UpdatedAt = time.Now().UTC()
	s.jobs[jobID] = job
}

func newJobID() string {
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return fmt.Sprintf("job-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(randomBytes)
}

func (s *ReportService) fetchPlayerData(ctx context.Context, req GenerateReportRequest) (json.RawMessage, error) {
	body, err := json.Marshal(logComparisonRequest{
		Fight:       req.Fight,
		CharacterID: req.Character.ID,
	})
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/reports/%s/player-data", s.logURL, req.ReportID), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.logClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("log-service returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(bodyBytes), nil
}

func (s *ReportService) fetchRankingCandidates(ctx context.Context, req GenerateReportRequest) ([]RankingCandidate, error) {
	body, err := json.Marshal(logRankingCandidatesRequest{
		Fight:          req.Fight,
		CharacterClass: req.Character.Class,
		CharacterSpec:  req.Character.Spec,
		Limit:          10,
	})
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.logURL+"/rankings/candidates", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.logClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("log-service returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var candidates []RankingCandidate
	if err := json.NewDecoder(resp.Body).Decode(&candidates); err != nil {
		return nil, err
	}

	return candidates, nil
}

func (s *ReportService) fetchCohortMember(ctx context.Context, candidate RankingCandidate) (json.RawMessage, error) {
	body, err := json.Marshal(logCohortMemberRequest{Candidate: candidate})
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.logURL+"/cohort/member-data", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.logClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("log-service returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(bodyBytes), nil
}

func (s *ReportService) fetchComparison(ctx context.Context, playerData json.RawMessage, cohortData []json.RawMessage) (ComparisonResult, error) {
	body, err := json.Marshal(analysisCompareRequest{
		PlayerData: playerData,
		CohortData: cohortData,
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
