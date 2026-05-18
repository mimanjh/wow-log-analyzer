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
	"sort"
	"strings"
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
	Cohort     []CohortEntry    `json:"cohort"`
	Comparison ComparisonResult `json:"comparison"`
	Warnings   []ReportWarning  `json:"warnings,omitempty"`
	AI         AIReportSection  `json:"ai"`
}

type AbilityTimelineResponse struct {
	AbilityID       int                     `json:"abilityId"`
	AbilityName     string                  `json:"abilityName"`
	FightDurationMS int64                   `json:"fightDurationMs"`
	Player          AbilityTimelineSeries   `json:"player"`
	Elite           []AbilityTimelineSeries `json:"elite"`
}

type ResourceTimelineResponse struct {
	ResourceTypeID  int                      `json:"resourceTypeId"`
	ResourceType    string                   `json:"resourceType"`
	FightDurationMS int64                    `json:"fightDurationMs"`
	Player          ResourceTimelineSeries   `json:"player"`
	Elite           []ResourceTimelineSeries `json:"elite"`
}

type BuffTimelineResponse struct {
	AbilityID       int                  `json:"abilityId"`
	AbilityName     string               `json:"abilityName"`
	FightDurationMS int64                `json:"fightDurationMs"`
	Player          BuffTimelineSeries   `json:"player"`
	Elite           []BuffTimelineSeries `json:"elite"`
}

type AbilityTimelineSeries struct {
	Label     string  `json:"label"`
	Subtitle  string  `json:"subtitle,omitempty"`
	ReportURL string  `json:"reportUrl,omitempty"`
	CastsMS   []int64 `json:"castsMs"`
}

type ResourceTimelineSeries struct {
	Label        string                   `json:"label"`
	Subtitle     string                   `json:"subtitle,omitempty"`
	ReportURL    string                   `json:"reportUrl,omitempty"`
	DurationMS   int64                    `json:"durationMs"`
	Samples      []ResourceTimelineSample `json:"samples"`
	WasteMarkers []int64                  `json:"wasteMarkersMs,omitempty"`
}

type ResourceTimelineSample struct {
	TimestampMS int64   `json:"timestampMs"`
	Value       float64 `json:"value"`
	MaxValue    float64 `json:"maxValue,omitempty"`
	Waste       float64 `json:"waste,omitempty"`
}

type BuffTimelineSeries struct {
	Label     string               `json:"label"`
	Subtitle  string               `json:"subtitle,omitempty"`
	ReportURL string               `json:"reportUrl,omitempty"`
	Windows   []BuffTimelineWindow `json:"windows"`
}

type BuffTimelineWindow struct {
	StartMS int64 `json:"startMs"`
	EndMS   int64 `json:"endMs"`
}

type CohortEntry struct {
	Name         string  `json:"name"`
	Class        string  `json:"class"`
	Spec         string  `json:"spec"`
	Server       string  `json:"server,omitempty"`
	ServerRegion string  `json:"serverRegion,omitempty"`
	ReportID     string  `json:"reportId"`
	FightID      int     `json:"fightId"`
	RankValue    float64 `json:"rankValue"`
	DurationMS   int     `json:"durationMs"`
	ReportURL    string  `json:"reportUrl"`
}

type ReportWarning struct {
	Kind    string `json:"kind"`
	Title   string `json:"title"`
	Message string `json:"message"`
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
	PlayerMetrics PlayerFightMetrics        `json:"playerMetrics"`
	CohortStats   CohortStatistics          `json:"cohortStats"`
	Deltas        MetricDeltas              `json:"deltas"`
	Rankings      MetricRankings            `json:"rankings"`
	AbilityUsage  []AbilityUsageComparison  `json:"abilityUsage"`
	BuffUptimes   []BuffUptimeComparison    `json:"buffUptimes"`
	ResourceUsage []ResourceUsageComparison `json:"resourceUsage"`
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
	Caution       string  `json:"caution,omitempty"`
}

type MajorCDCountMetric struct {
	Value          int    `json:"value"`
	TotalCooldowns int    `json:"totalCooldowns"`
	Caution        string `json:"caution,omitempty"`
}

type MajorCDDriftMetric struct {
	Value         float64 `json:"value"`
	TotalDrift    float64 `json:"totalDrift"`
	CooldownCount int     `json:"cooldownCount"`
	Caution       string  `json:"caution,omitempty"`
}

type BuffUptimeMetric struct {
	Value         float64 `json:"value"`
	TotalUptime   float64 `json:"totalUptime"`
	FightDuration float64 `json:"fightDuration"`
	Caution       string  `json:"caution,omitempty"`
}

type DowntimePercentageMetric struct {
	Value         float64 `json:"value"`
	TotalDowntime float64 `json:"totalDowntime"`
	FightDuration float64 `json:"fightDuration"`
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
	Caution     string  `json:"caution,omitempty"`
}

type MetricRankings struct {
	CastsPerMin  float64 `json:"castsPerMin"`
	MajorCDCount float64 `json:"majorCdCount"`
	MajorCDDrift float64 `json:"majorCdDrift"`
	BuffUptime   float64 `json:"buffUptime"`
	DowntimePct  float64 `json:"downtimePct"`
}

type AbilityUsageComparison struct {
	AbilityID               int     `json:"abilityId"`
	AbilityName             string  `json:"abilityName"`
	PlayerCount             int     `json:"playerCount"`
	PlayerCastsPerMinute    float64 `json:"playerCastsPerMinute"`
	PlayerFirstUseSeconds   float64 `json:"playerFirstUseSeconds,omitempty"`
	CohortMedianCount       float64 `json:"cohortMedianCount"`
	CohortMedianPerMinute   float64 `json:"cohortMedianPerMinute"`
	CohortMedianFirstUseSec float64 `json:"cohortMedianFirstUseSeconds,omitempty"`
	CountDelta              float64 `json:"countDelta"`
	PerMinuteDelta          float64 `json:"perMinuteDelta"`
	FirstUseDeltaSeconds    float64 `json:"firstUseDeltaSeconds,omitempty"`
	Percentile              float64 `json:"percentile"`
	SampleSize              int     `json:"sampleSize"`
	Caution                 string  `json:"caution,omitempty"`
}

type BuffUptimeComparison struct {
	AbilityID               int     `json:"abilityId"`
	AbilityName             string  `json:"abilityName"`
	PlayerUptimePct         float64 `json:"playerUptimePct"`
	PlayerFirstApplySeconds float64 `json:"playerFirstApplySeconds,omitempty"`
	CohortMedianUptimePct   float64 `json:"cohortMedianUptimePct"`
	CohortMedianFirstApply  float64 `json:"cohortMedianFirstApplySeconds,omitempty"`
	UptimeDelta             float64 `json:"uptimeDelta"`
	FirstApplyDeltaSeconds  float64 `json:"firstApplyDeltaSeconds,omitempty"`
	Percentile              float64 `json:"percentile"`
	SampleSize              int     `json:"sampleSize"`
	Caution                 string  `json:"caution,omitempty"`
}

type ResourceUsageComparison struct {
	ResourceTypeID                 int     `json:"resourceTypeId"`
	ResourceType                   string  `json:"resourceType"`
	PlayerSampleCount              int     `json:"playerSampleCount"`
	CohortMedianSampleCount        float64 `json:"cohortMedianSampleCount"`
	SampleCountDelta               float64 `json:"sampleCountDelta"`
	PlayerFullMarkerCount          int     `json:"playerFullMarkerCount"`
	CohortMedianFullMarkerCount    float64 `json:"cohortMedianFullMarkerCount"`
	FullMarkerDelta                float64 `json:"fullMarkerDelta"`
	PlayerFullWindowSeconds        float64 `json:"playerFullWindowSeconds"`
	CohortMedianFullWindowSeconds  float64 `json:"cohortMedianFullWindowSeconds"`
	FullWindowDeltaSeconds         float64 `json:"fullWindowDeltaSeconds"`
	PlayerAveragePct               float64 `json:"playerAveragePct"`
	CohortMedianAveragePct         float64 `json:"cohortMedianAveragePct"`
	AveragePctDelta                float64 `json:"averagePctDelta"`
	PlayerTimeAtMaxSeconds         float64 `json:"playerTimeAtMaxSeconds"`
	CohortMedianTimeAtMaxSeconds   float64 `json:"cohortMedianTimeAtMaxSeconds"`
	TimeAtMaxDeltaSeconds          float64 `json:"timeAtMaxDeltaSeconds"`
	PlayerSpent                    float64 `json:"playerSpent"`
	CohortMedianSpent              float64 `json:"cohortMedianSpent"`
	SpentDelta                     float64 `json:"spentDelta"`
	PlayerGeneratedPerMinute       float64 `json:"playerGeneratedPerMinute"`
	CohortMedianGeneratedPerMinute float64 `json:"cohortMedianGeneratedPerMinute"`
	GeneratedDelta                 float64 `json:"generatedDelta"`
	PlayerWastePerMinute           float64 `json:"playerWastePerMinute"`
	CohortMedianWastePerMinute     float64 `json:"cohortMedianWastePerMinute"`
	WasteDelta                     float64 `json:"wasteDelta"`
	PlayerWastePct                 float64 `json:"playerWastePct"`
	CohortMedianWastePct           float64 `json:"cohortMedianWastePct"`
	WastePctDelta                  float64 `json:"wastePctDelta"`
	SampleSize                     int     `json:"sampleSize"`
	Caution                        string  `json:"caution,omitempty"`
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
	PlayerData     json.RawMessage   `json:"playerData"`
	CohortData     []json.RawMessage `json:"cohortData"`
	CharacterClass string            `json:"characterClass,omitempty"`
	CharacterSpec  string            `json:"characterSpec,omitempty"`
}

type insightGenerationRequest struct {
	Context            insightContext     `json:"context"`
	Metrics            []insightMetric    `json:"metrics"`
	AbilityHighlights  []insightHighlight `json:"abilityHighlights,omitempty"`
	BuffHighlights     []insightHighlight `json:"buffHighlights,omitempty"`
	ResourceHighlights []insightHighlight `json:"resourceHighlights,omitempty"`
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

type insightHighlight struct {
	Name                  string    `json:"name"`
	PlayerValue           float64   `json:"playerValue"`
	EliteValue            float64   `json:"eliteValue"`
	Difference            float64   `json:"difference"`
	Unit                  string    `json:"unit,omitempty"`
	PlayerRawCount        float64   `json:"playerRawCount,omitempty"`
	EliteRawCount         float64   `json:"eliteRawCount,omitempty"`
	PlayerTimingSeconds   float64   `json:"playerTimingSeconds,omitempty"`
	EliteTimingSeconds    float64   `json:"eliteTimingSeconds,omitempty"`
	TimingDeltaSeconds    float64   `json:"timingDeltaSeconds,omitempty"`
	TimingLabel           string    `json:"timingLabel,omitempty"`
	PlayerUseTimesSeconds []float64 `json:"playerUseTimesSeconds,omitempty"`
	EliteUseTimesSeconds  []float64 `json:"eliteUseTimesSeconds,omitempty"`
	PlayerLargestGapSec   float64   `json:"playerLargestGapSeconds,omitempty"`
	EliteLargestGapSec    float64   `json:"eliteLargestGapSeconds,omitempty"`
	Category              string    `json:"category,omitempty"`
}

type AIInsight struct {
	MetricKey string `json:"metricKey"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	Caution   string `json:"caution,omitempty"`
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

	timeline *reportTimelineData `json:"-"`
}

type reportTimelineData struct {
	Fight        FightSummary
	Character    CharacterSummary
	PlayerData   timelineFightData
	EliteData    []timelineFightData
	EliteEntries []CohortEntry
}

type timelineFightData struct {
	PlayerID            int                     `json:"playerId"`
	FightID             int                     `json:"fightId"`
	FightStart          time.Time               `json:"fightStart"`
	FightEnd            time.Time               `json:"fightEnd"`
	TalentImportCode    string                  `json:"talentImportCode,omitempty"`
	TalentCalculatorURL string                  `json:"talentCalculatorUrl,omitempty"`
	CastEvents          []timelineAbilityEvent  `json:"castEvents"`
	DamageEvents        []timelineAbilityEvent  `json:"damageEvents"`
	CooldownEvents      []timelineAbilityEvent  `json:"cooldownEvents"`
	BuffEvents          []timelineBuffEvent     `json:"buffEvents"`
	ResourceEvents      []timelineResourceEvent `json:"resourceEvents"`
}

type timelineAbilityEvent struct {
	Timestamp time.Time       `json:"timestamp"`
	Ability   timelineAbility `json:"ability"`
}

type timelineAbility struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type timelineBuffEvent struct {
	Timestamp time.Time       `json:"timestamp"`
	Ability   timelineAbility `json:"ability"`
	EventType string          `json:"eventType"`
}

type timelineResourceEvent struct {
	Timestamp      time.Time `json:"timestamp"`
	ResourceTypeID int       `json:"resourceTypeId"`
	ResourceType   string    `json:"resourceType"`
	Amount         float64   `json:"amount"`
	Change         float64   `json:"change"`
	Waste          float64   `json:"waste"`
	MaxAmount      float64   `json:"maxAmount"`
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

const (
	targetEliteCount         = 10
	rankingCandidateBatchCap = 25
	rankingCandidateMaxCap   = 100
)

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

func (s *ReportService) GetAbilityTimeline(jobID string, abilityID int) (AbilityTimelineResponse, error) {
	if abilityID == 0 {
		return AbilityTimelineResponse{}, fmt.Errorf("abilityId is required")
	}

	s.jobMu.RLock()
	job, ok := s.jobs[jobID]
	s.jobMu.RUnlock()
	if !ok {
		return AbilityTimelineResponse{}, fmt.Errorf("report job %s not found", jobID)
	}
	if job.timeline == nil {
		return AbilityTimelineResponse{}, fmt.Errorf("ability timeline is not available for this job yet")
	}

	playerSeries := buildAbilityTimelineSeries(
		job.timeline.PlayerData,
		abilityID,
		job.timeline.Character.Name,
		fmt.Sprintf("%s %s", job.timeline.Character.Spec, job.timeline.Character.Class),
		"",
	)
	if len(playerSeries.CastsMS) == 0 {
		return AbilityTimelineResponse{}, fmt.Errorf("no cast timeline was available for this ability")
	}

	eliteSeries := make([]AbilityTimelineSeries, 0, len(job.timeline.EliteData))
	for index, eliteData := range job.timeline.EliteData {
		entry := job.timeline.EliteEntries[index]
		subtitle := strings.TrimSpace(fmt.Sprintf("%s %s", entry.Spec, entry.Class))
		if entry.Server != "" {
			subtitle = strings.TrimSpace(fmt.Sprintf("%s • %s", subtitle, entry.Server))
		}
		series := buildAbilityTimelineSeries(
			eliteData,
			abilityID,
			entry.Name,
			subtitle,
			entry.ReportURL,
		)
		if len(series.CastsMS) > 0 {
			eliteSeries = append(eliteSeries, series)
		}
	}

	abilityName := findAbilityName(job.timeline.PlayerData, abilityID)
	if abilityName == "" {
		for _, eliteData := range job.timeline.EliteData {
			abilityName = findAbilityName(eliteData, abilityID)
			if abilityName != "" {
				break
			}
		}
	}
	if abilityName == "" {
		abilityName = "Selected Ability"
	}

	return AbilityTimelineResponse{
		AbilityID:       abilityID,
		AbilityName:     abilityName,
		FightDurationMS: job.timeline.PlayerData.FightEnd.Sub(job.timeline.PlayerData.FightStart).Milliseconds(),
		Player:          playerSeries,
		Elite:           eliteSeries,
	}, nil
}

func (s *ReportService) GetResourceTimeline(jobID string, resourceTypeID int) (ResourceTimelineResponse, error) {
	s.jobMu.RLock()
	job, ok := s.jobs[jobID]
	s.jobMu.RUnlock()
	if !ok {
		return ResourceTimelineResponse{}, fmt.Errorf("report job %s not found", jobID)
	}
	if job.timeline == nil {
		return ResourceTimelineResponse{}, fmt.Errorf("resource timeline is not available for this job yet")
	}

	resourceType := findResourceType(job.timeline.PlayerData, resourceTypeID)
	if resourceType == "" {
		for _, eliteData := range job.timeline.EliteData {
			resourceType = findResourceType(eliteData, resourceTypeID)
			if resourceType != "" {
				break
			}
		}
	}
	if resourceType == "" {
		resourceType = "Selected Resource"
	}

	playerSeries := buildResourceTimelineSeries(
		job.timeline.PlayerData,
		resourceTypeID,
		job.timeline.Character.Name,
		fmt.Sprintf("%s %s", job.timeline.Character.Spec, job.timeline.Character.Class),
		"",
	)
	if len(playerSeries.Samples) == 0 {
		return ResourceTimelineResponse{}, fmt.Errorf("no resource timeline was available for this resource")
	}

	eliteSeries := make([]ResourceTimelineSeries, 0, len(job.timeline.EliteData))
	for index, eliteData := range job.timeline.EliteData {
		entry := job.timeline.EliteEntries[index]
		subtitle := strings.TrimSpace(fmt.Sprintf("%s %s", entry.Spec, entry.Class))
		if entry.Server != "" {
			subtitle = strings.TrimSpace(fmt.Sprintf("%s - %s", subtitle, entry.Server))
		}
		series := buildResourceTimelineSeries(
			eliteData,
			resourceTypeID,
			entry.Name,
			subtitle,
			entry.ReportURL,
		)
		if len(series.Samples) > 0 {
			eliteSeries = append(eliteSeries, series)
		}
	}

	return ResourceTimelineResponse{
		ResourceTypeID:  resourceTypeID,
		ResourceType:    resourceType,
		FightDurationMS: job.timeline.PlayerData.FightEnd.Sub(job.timeline.PlayerData.FightStart).Milliseconds(),
		Player:          playerSeries,
		Elite:           eliteSeries,
	}, nil
}

func (s *ReportService) GetBuffTimeline(jobID string, abilityID int) (BuffTimelineResponse, error) {
	if abilityID == 0 {
		return BuffTimelineResponse{}, fmt.Errorf("abilityId is required")
	}

	s.jobMu.RLock()
	job, ok := s.jobs[jobID]
	s.jobMu.RUnlock()
	if !ok {
		return BuffTimelineResponse{}, fmt.Errorf("report job %s not found", jobID)
	}
	if job.timeline == nil {
		return BuffTimelineResponse{}, fmt.Errorf("buff timeline is not available for this job yet")
	}

	playerSeries := buildBuffTimelineSeries(
		job.timeline.PlayerData,
		abilityID,
		job.timeline.Character.Name,
		fmt.Sprintf("%s %s", job.timeline.Character.Spec, job.timeline.Character.Class),
		"",
	)

	eliteSeries := make([]BuffTimelineSeries, 0, len(job.timeline.EliteData))
	for index, eliteData := range job.timeline.EliteData {
		entry := job.timeline.EliteEntries[index]
		subtitle := strings.TrimSpace(fmt.Sprintf("%s %s", entry.Spec, entry.Class))
		if entry.Server != "" {
			subtitle = strings.TrimSpace(fmt.Sprintf("%s - %s", subtitle, entry.Server))
		}
		series := buildBuffTimelineSeries(
			eliteData,
			abilityID,
			entry.Name,
			subtitle,
			entry.ReportURL,
		)
		if len(series.Windows) > 0 {
			eliteSeries = append(eliteSeries, series)
		}
	}
	if len(playerSeries.Windows) == 0 && len(eliteSeries) == 0 {
		return BuffTimelineResponse{}, fmt.Errorf("no buff timeline was available for this ability")
	}

	abilityName := findBuffName(job.timeline.PlayerData, abilityID)
	if abilityName == "" {
		for _, eliteData := range job.timeline.EliteData {
			abilityName = findBuffName(eliteData, abilityID)
			if abilityName != "" {
				break
			}
		}
	}
	if abilityName == "" {
		abilityName = "Selected Buff"
	}

	return BuffTimelineResponse{
		AbilityID:       abilityID,
		AbilityName:     abilityName,
		FightDurationMS: job.timeline.PlayerData.FightEnd.Sub(job.timeline.PlayerData.FightStart).Milliseconds(),
		Player:          playerSeries,
		Elite:           eliteSeries,
	}, nil
}

func (s *ReportService) runJob(jobID string, req GenerateReportRequest) {
	ctx := context.Background()

	s.updateJob(jobID, ReportJobRunning, "player-data", "Fetching selected player fight data.", ReportJobProgress{Current: 1, Total: 5}, "", nil)
	playerData, err := s.fetchPlayerData(ctx, req)
	if err != nil {
		s.updateJob(jobID, ReportJobFailed, "player-data", "Failed to fetch selected player fight data.", ReportJobProgress{Current: 1, Total: 5}, err.Error(), nil)
		return
	}
	playerTimelineData, err := decodeTimelineFightData(playerData)
	if err != nil {
		s.updateJob(jobID, ReportJobFailed, "player-data", "Failed to decode selected player fight data.", ReportJobProgress{Current: 1, Total: 5}, err.Error(), nil)
		return
	}

	s.updateJob(jobID, ReportJobRunning, "rankings", "Fetching ranking candidates for the selected boss and spec.", ReportJobProgress{Current: 2, Total: 5}, "", nil)
	cohortData := make([]json.RawMessage, 0, targetEliteCount)
	cohortEntries := make([]CohortEntry, 0, targetEliteCount)
	cohortTimelineData := make([]timelineFightData, 0, targetEliteCount)
	processedCandidates := make(map[string]bool)

	for candidateLimit := rankingCandidateBatchCap; len(cohortData) < targetEliteCount && candidateLimit <= rankingCandidateMaxCap; candidateLimit += rankingCandidateBatchCap {
		candidates, err := s.fetchRankingCandidates(ctx, req, candidateLimit)
		if err != nil {
			s.updateJob(jobID, ReportJobFailed, "rankings", "Failed to fetch ranking candidates.", ReportJobProgress{Current: 2, Total: 5}, err.Error(), nil)
			return
		}
		if len(candidates) == 0 {
			break
		}

		newCandidates := 0
		for index, candidate := range candidates {
			key := rankingCandidateKey(candidate)
			if processedCandidates[key] {
				continue
			}
			processedCandidates[key] = true
			newCandidates++

			s.updateJob(jobID, ReportJobRunning, "cohort", fmt.Sprintf("Fetching cohort member %d of %d.", len(cohortData)+1, targetEliteCount), ReportJobProgress{Current: len(cohortData) + 1, Total: targetEliteCount}, "", nil)
			memberData, err := s.fetchCohortMember(ctx, candidate)
			if err != nil {
				continue
			}
			decodedTimeline, err := decodeTimelineFightData(memberData)
			if err != nil {
				continue
			}
			cohortData = append(cohortData, memberData)
			cohortEntries = append(cohortEntries, buildCohortEntry(candidate))
			cohortTimelineData = append(cohortTimelineData, decodedTimeline)
			if len(cohortData) == targetEliteCount {
				break
			}
			if index == len(candidates)-1 && len(candidates) < candidateLimit {
				break
			}
		}
		if newCandidates == 0 || len(candidates) < candidateLimit {
			break
		}
	}
	if len(cohortData) == 0 {
		s.updateJob(jobID, ReportJobFailed, "cohort", "Failed to collect any cohort members.", ReportJobProgress{Current: 0, Total: targetEliteCount}, "no cohort member data could be fetched", nil)
		return
	}
	s.setTimeline(jobID, &reportTimelineData{
		Fight:        req.Fight,
		Character:    req.Character,
		PlayerData:   playerTimelineData,
		EliteData:    cohortTimelineData,
		EliteEntries: cohortEntries,
	})

	s.updateJob(jobID, ReportJobRunning, "analyzing", "Running deterministic comparison analysis.", ReportJobProgress{Current: 4, Total: 5}, "", nil)
	comparison, err := s.fetchComparison(ctx, req, playerData, cohortData)
	if err != nil {
		s.updateJob(jobID, ReportJobFailed, "analyzing", "Failed to compute deterministic comparison metrics.", ReportJobProgress{Current: 4, Total: 5}, err.Error(), nil)
		return
	}

	response := GenerateReportResponse{
		Fight:      req.Fight,
		Character:  req.Character,
		Cohort:     cohortEntries,
		Comparison: comparison,
		Warnings:   buildReportWarnings(req, playerTimelineData, cohortTimelineData),
		AI: AIReportSection{
			Available: false,
			Insights:  []AIInsight{},
		},
	}

	s.updateJob(jobID, ReportJobRunning, "insights", "Generating AI insights.", ReportJobProgress{Current: 5, Total: 5}, "", &response)
	insights, err := s.fetchInsights(ctx, req, response.Comparison, playerTimelineData, cohortTimelineData)
	if err != nil {
		fmt.Printf("AI insights unavailable for job %s: %v\n", jobID, err)
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

func buildCohortEntry(candidate RankingCandidate) CohortEntry {
	return CohortEntry{
		Name:         candidate.Name,
		Class:        candidate.Class,
		Spec:         candidate.Spec,
		Server:       candidate.Server,
		ServerRegion: candidate.ServerRegion,
		ReportID:     candidate.ReportID,
		FightID:      candidate.FightID,
		RankValue:    candidate.RankValue,
		DurationMS:   candidate.DurationMS,
		ReportURL: fmt.Sprintf(
			"https://www.warcraftlogs.com/reports/%s#fight=%d",
			candidate.ReportID,
			candidate.FightID,
		),
	}
}

func buildReportWarnings(req GenerateReportRequest, playerData timelineFightData, cohortData []timelineFightData) []ReportWarning {
	warnings := make([]ReportWarning, 0, 1)

	playerTalentCode := strings.TrimSpace(req.Character.TalentImportCode)
	if playerTalentCode == "" {
		playerTalentCode = strings.TrimSpace(playerData.TalentImportCode)
	}
	if playerTalentCode == "" || len(cohortData) == 0 {
		return warnings
	}

	known := 0
	different := 0
	cohortTalentCounts := make(map[string]int)
	for _, cohort := range cohortData {
		cohortTalentCode := strings.TrimSpace(cohort.TalentImportCode)
		if cohortTalentCode == "" {
			continue
		}
		known++
		cohortTalentCounts[cohortTalentCode]++
		if cohortTalentCode != playerTalentCode {
			different++
		}
	}

	if known < 3 {
		return warnings
	}

	topEliteCount := 0
	for _, count := range cohortTalentCounts {
		if count > topEliteCount {
			topEliteCount = count
		}
	}

	if topEliteCount <= known/2 {
		warnings = append(warnings, ReportWarning{
			Kind:  "talents",
			Title: "Elite talent builds vary",
			Message: fmt.Sprintf(
				"The %d elite %s %s parses with available talent data did not share a clear majority talent build. Consider reviewing the talents of the elites before comparing rotation metrics.",
				known,
				req.Character.Spec,
				req.Character.Class,
			),
		})
		return warnings
	}

	if different <= known/2 {
		return warnings
	}

	warnings = append(warnings, ReportWarning{
		Kind:  "talents",
		Title: "Talent build differs from most elites",
		Message: fmt.Sprintf(
			"%d of %d elite %s %s parses with available talent data used the same talent build, and your selected player's talent build was different. Consider reviewing the talents of the elites before comparing rotation metrics.",
			topEliteCount,
			known,
			req.Character.Spec,
			req.Character.Class,
		),
	})

	return warnings
}

func (s *ReportService) setJob(job ReportJob) {
	s.jobMu.Lock()
	defer s.jobMu.Unlock()
	s.jobs[job.ID] = job
}

func (s *ReportService) setTimeline(jobID string, timeline *reportTimelineData) {
	s.jobMu.Lock()
	defer s.jobMu.Unlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return
	}
	job.timeline = timeline
	job.UpdatedAt = time.Now().UTC()
	s.jobs[jobID] = job
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
	return s.postForRaw(
		ctx,
		s.logClient,
		fmt.Sprintf("%s/reports/%s/player-data", s.logURL, req.ReportID),
		logComparisonRequest{
			Fight:       req.Fight,
			CharacterID: req.Character.ID,
		},
		"log-service",
	)
}

func (s *ReportService) fetchRankingCandidates(ctx context.Context, req GenerateReportRequest, limit int) ([]RankingCandidate, error) {
	var candidates []RankingCandidate
	err := s.postForJSON(
		ctx,
		s.logClient,
		s.logURL+"/rankings/candidates",
		logRankingCandidatesRequest{
			Fight:          req.Fight,
			CharacterClass: req.Character.Class,
			CharacterSpec:  req.Character.Spec,
			Limit:          limit,
		},
		&candidates,
		"log-service",
	)
	return candidates, err
}

func rankingCandidateKey(candidate RankingCandidate) string {
	return fmt.Sprintf("%s:%d:%s:%s", candidate.ReportID, candidate.FightID, candidate.Name, candidate.Server)
}

func (s *ReportService) fetchCohortMember(ctx context.Context, candidate RankingCandidate) (json.RawMessage, error) {
	return s.postForRaw(
		ctx,
		s.logClient,
		s.logURL+"/cohort/member-data",
		logCohortMemberRequest{Candidate: candidate},
		"log-service",
	)
}

func (s *ReportService) fetchComparison(ctx context.Context, req GenerateReportRequest, playerData json.RawMessage, cohortData []json.RawMessage) (ComparisonResult, error) {
	var comparison ComparisonResult
	err := s.postForJSON(
		ctx,
		s.analysisClient,
		s.analysisURL+"/analyze/compare",
		analysisCompareRequest{
			PlayerData:     playerData,
			CohortData:     cohortData,
			CharacterClass: req.Character.Class,
			CharacterSpec:  req.Character.Spec,
		},
		&comparison,
		"analysis-service",
	)
	return comparison, err
}

func (s *ReportService) fetchInsights(ctx context.Context, req GenerateReportRequest, comparison ComparisonResult, playerData timelineFightData, eliteData []timelineFightData) (insightGenerationResponse, error) {
	var insights insightGenerationResponse
	err := s.postForJSON(
		ctx,
		s.aiClient,
		s.aiURL+"/insights/generate",
		buildInsightRequest(req, comparison, playerData, eliteData),
		&insights,
		"ai-service",
	)
	return insights, err
}

func (s *ReportService) postForRaw(ctx context.Context, client *http.Client, endpoint string, payload interface{}, serviceName string) (json.RawMessage, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned status %d: %s", serviceName, resp.StatusCode, string(bodyBytes))
	}

	return json.RawMessage(bodyBytes), nil
}

func (s *ReportService) postForJSON(ctx context.Context, client *http.Client, endpoint string, payload, target interface{}, serviceName string) error {
	bodyBytes, err := s.postForRaw(ctx, client, endpoint, payload, serviceName)
	if err != nil {
		return err
	}

	return json.Unmarshal(bodyBytes, target)
}

func buildInsightRequest(req GenerateReportRequest, comparison ComparisonResult, playerData timelineFightData, eliteData []timelineFightData) insightGenerationRequest {
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
		Metrics:            nil,
		AbilityHighlights:  buildAbilityHighlights(comparison.AbilityUsage, 5, req.Character.Class, req.Character.Spec, playerData, eliteData),
		BuffHighlights:     buildBuffHighlights(comparison.BuffUptimes, 5, req.Character.Class, req.Character.Spec, playerData, eliteData),
		ResourceHighlights: buildResourceHighlights(comparison.ResourceUsage, 3),
	}
}

func decodeTimelineFightData(raw json.RawMessage) (timelineFightData, error) {
	var data timelineFightData
	if err := json.Unmarshal(raw, &data); err != nil {
		return timelineFightData{}, err
	}
	return data, nil
}

func buildAbilityTimelineSeries(data timelineFightData, abilityID int, label, subtitle, reportURL string) AbilityTimelineSeries {
	casts := make([]int64, 0)
	for _, event := range data.CastEvents {
		if event.Ability.ID != abilityID {
			continue
		}
		casts = append(casts, event.Timestamp.Sub(data.FightStart).Milliseconds())
	}
	if len(casts) == 0 {
		for _, event := range data.CooldownEvents {
			if event.Ability.ID != abilityID {
				continue
			}
			casts = append(casts, event.Timestamp.Sub(data.FightStart).Milliseconds())
		}
	}
	if len(casts) == 0 {
		for _, event := range data.DamageEvents {
			if event.Ability.ID != abilityID {
				continue
			}
			casts = append(casts, event.Timestamp.Sub(data.FightStart).Milliseconds())
		}
	}

	return AbilityTimelineSeries{
		Label:     label,
		Subtitle:  subtitle,
		ReportURL: reportURL,
		CastsMS:   casts,
	}
}

func findAbilityName(data timelineFightData, abilityID int) string {
	for _, event := range data.CastEvents {
		if event.Ability.ID == abilityID {
			return event.Ability.Name
		}
	}
	for _, event := range data.DamageEvents {
		if event.Ability.ID == abilityID {
			return event.Ability.Name
		}
	}
	for _, event := range data.CooldownEvents {
		if event.Ability.ID == abilityID {
			return event.Ability.Name
		}
	}
	return ""
}

func buildBuffTimelineSeries(data timelineFightData, abilityID int, label, subtitle, reportURL string) BuffTimelineSeries {
	windows := buffWindows(data, abilityID)
	timelineWindows := make([]BuffTimelineWindow, 0, len(windows))
	durationMS := data.FightEnd.Sub(data.FightStart).Milliseconds()
	if durationMS < 0 {
		durationMS = 0
	}

	for _, window := range windows {
		startMS := window.start.Sub(data.FightStart).Milliseconds()
		endMS := window.end.Sub(data.FightStart).Milliseconds()
		if startMS < 0 {
			startMS = 0
		}
		if endMS < 0 {
			endMS = 0
		}
		if durationMS > 0 {
			if startMS > durationMS {
				startMS = durationMS
			}
			if endMS > durationMS {
				endMS = durationMS
			}
		}
		if endMS <= startMS {
			continue
		}
		timelineWindows = append(timelineWindows, BuffTimelineWindow{
			StartMS: startMS,
			EndMS:   endMS,
		})
	}

	return BuffTimelineSeries{
		Label:     label,
		Subtitle:  subtitle,
		ReportURL: reportURL,
		Windows:   timelineWindows,
	}
}

func findBuffName(data timelineFightData, abilityID int) string {
	for _, event := range data.BuffEvents {
		if event.Ability.ID == abilityID {
			return event.Ability.Name
		}
	}
	return ""
}

func buildResourceTimelineSeries(data timelineFightData, resourceTypeID int, label, subtitle, reportURL string) ResourceTimelineSeries {
	samples := make([]ResourceTimelineSample, 0)
	wasteMarkers := make([]int64, 0)

	for _, event := range data.ResourceEvents {
		if event.ResourceTypeID != resourceTypeID {
			continue
		}

		timestampMS := event.Timestamp.Sub(data.FightStart).Milliseconds()
		if timestampMS < 0 {
			timestampMS = 0
		}
		sample := ResourceTimelineSample{
			TimestampMS: timestampMS,
			Value:       event.Amount,
			MaxValue:    event.MaxAmount,
			Waste:       event.Waste,
		}
		if sample.Value == 0 && event.Change > 0 {
			sample.Value = event.Change
		}
		maxValue := sample.MaxValue
		if maxValue <= 0 {
			maxValue = defaultResourceMaxValue(event.ResourceTypeID, event.ResourceType)
			sample.MaxValue = maxValue
		}
		if event.Waste > 0 && maxValue > 0 {
			sample.Value = maxValue
		}
		samples = append(samples, sample)

		if event.Waste > 0 || (maxValue > 0 && sample.Value >= maxValue) {
			wasteMarkers = append(wasteMarkers, timestampMS)
		}
	}

	sort.Slice(samples, func(i, j int) bool {
		return samples[i].TimestampMS < samples[j].TimestampMS
	})
	sort.Slice(wasteMarkers, func(i, j int) bool {
		return wasteMarkers[i] < wasteMarkers[j]
	})

	return ResourceTimelineSeries{
		Label:        label,
		Subtitle:     subtitle,
		ReportURL:    reportURL,
		DurationMS:   data.FightEnd.Sub(data.FightStart).Milliseconds(),
		Samples:      samples,
		WasteMarkers: wasteMarkers,
	}
}

func defaultResourceMaxValue(resourceTypeID int, resourceType string) float64 {
	switch resourceTypeID {
	case 1, 2, 3, 6, 8, 13, 18:
		return 100
	case 4, 7, 9:
		return 5
	case 5, 12, 19:
		return 6
	case 11:
		return 10
	case 16:
		return 4
	case 17:
		return 120
	}

	switch strings.ToLower(strings.TrimSpace(resourceType)) {
	case "rage", "focus", "energy", "runic power", "lunar power", "insanity", "pain":
		return 100
	case "fury":
		return 120
	case "combo points", "soul shards", "holy power":
		return 5
	case "runes", "chi", "essence":
		return 6
	case "maelstrom":
		return 10
	case "arcane charges":
		return 4
	default:
		return 0
	}
}

func buildResourceHighlights(values []ResourceUsageComparison, limit int) []insightHighlight {
	if len(values) == 0 || limit <= 0 {
		return nil
	}

	ordered := append([]ResourceUsageComparison(nil), values...)
	sort.SliceStable(ordered, func(i, j int) bool {
		leftConcern := ordered[i].TimeAtMaxDeltaSeconds + maxFloat64(0, -ordered[i].SpentDelta/10)
		rightConcern := ordered[j].TimeAtMaxDeltaSeconds + maxFloat64(0, -ordered[j].SpentDelta/10)
		if leftConcern == rightConcern {
			return ordered[i].ResourceType < ordered[j].ResourceType
		}
		return leftConcern > rightConcern
	})

	highlights := make([]insightHighlight, 0, limit)
	for _, value := range ordered {
		if value.PlayerAveragePct == 0 && value.CohortMedianAveragePct == 0 && value.PlayerSpent == 0 && value.CohortMedianSpent == 0 {
			continue
		}
		highlights = append(highlights, insightHighlight{
			Name:                value.ResourceType,
			PlayerValue:         value.PlayerAveragePct,
			EliteValue:          value.CohortMedianAveragePct,
			Difference:          value.AveragePctDelta,
			Unit:                "%",
			PlayerRawCount:      value.PlayerSpent,
			EliteRawCount:       value.CohortMedianSpent,
			PlayerTimingSeconds: value.PlayerTimeAtMaxSeconds,
			EliteTimingSeconds:  value.CohortMedianTimeAtMaxSeconds,
			TimingDeltaSeconds:  value.TimeAtMaxDeltaSeconds,
			TimingLabel:         "time at max",
			Category:            "resource",
		})
		if len(highlights) == limit {
			break
		}
	}

	return highlights
}

func maxFloat64(left, right float64) float64 {
	if left > right {
		return left
	}
	return right
}

func findResourceType(data timelineFightData, resourceTypeID int) string {
	for _, event := range data.ResourceEvents {
		if event.ResourceTypeID == resourceTypeID && event.ResourceType != "" {
			return event.ResourceType
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func clampPercentile(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func buildAbilityHighlights(values []AbilityUsageComparison, limit int, characterClass, characterSpec string, playerData timelineFightData, eliteData []timelineFightData) []insightHighlight {
	if len(values) == 0 || limit <= 0 {
		return nil
	}

	prioritySet := specPriorityFor(characterClass, characterSpec)
	orderedValues, categories := orderAbilityHighlights(values, characterClass, characterSpec)
	filteredValues := filterAbilityHighlightsForAI(orderedValues, prioritySet)
	if len(filteredValues) == 0 {
		filteredValues = orderedValues
	}

	highlights := make([]insightHighlight, 0, limit)
	for _, value := range filteredValues {
		playerUseTimes := abilityUseTimesSeconds(playerData, value.AbilityID, 3)
		eliteUseTimes := eliteMedianUseTimesSeconds(eliteData, value.AbilityID, 3)
		if isHighFrequencyAbility(value) {
			playerUseTimes = nil
			eliteUseTimes = nil
		}

		highlights = append(highlights, insightHighlight{
			Name:                  value.AbilityName,
			PlayerValue:           value.PlayerCastsPerMinute,
			EliteValue:            value.CohortMedianPerMinute,
			Difference:            value.PerMinuteDelta,
			Unit:                  "/min",
			PlayerRawCount:        float64(value.PlayerCount),
			EliteRawCount:         value.CohortMedianCount,
			PlayerTimingSeconds:   value.PlayerFirstUseSeconds,
			EliteTimingSeconds:    value.CohortMedianFirstUseSec,
			TimingDeltaSeconds:    value.FirstUseDeltaSeconds,
			TimingLabel:           "first use",
			PlayerUseTimesSeconds: playerUseTimes,
			EliteUseTimesSeconds:  eliteUseTimes,
			Category:              categories[normalizeTrackedCooldownName(value.AbilityName)],
		})
		if len(highlights) == limit {
			break
		}
	}

	return highlights
}

func buildBuffHighlights(values []BuffUptimeComparison, limit int, characterClass, characterSpec string, playerData timelineFightData, eliteData []timelineFightData) []insightHighlight {
	if len(values) == 0 || limit <= 0 {
		return nil
	}

	prioritySet := specPriorityFor(characterClass, characterSpec)
	orderedValues, categories := orderBuffHighlights(values, characterClass, characterSpec)
	filteredValues := filterBuffHighlightsForAI(orderedValues, prioritySet)
	if len(filteredValues) == 0 {
		filteredValues = orderedValues
	}

	highlights := make([]insightHighlight, 0, limit)
	for _, value := range filteredValues {
		highlights = append(highlights, insightHighlight{
			Name:                value.AbilityName,
			PlayerValue:         value.PlayerUptimePct,
			EliteValue:          value.CohortMedianUptimePct,
			Difference:          value.UptimeDelta,
			Unit:                "%",
			PlayerTimingSeconds: value.PlayerFirstApplySeconds,
			EliteTimingSeconds:  value.CohortMedianFirstApply,
			TimingDeltaSeconds:  value.FirstApplyDeltaSeconds,
			TimingLabel:         "first apply",
			PlayerLargestGapSec: largestBuffGapSeconds(playerData, value.AbilityID),
			EliteLargestGapSec:  medianLargestBuffGapSeconds(eliteData, value.AbilityID),
			Category:            categories[normalizeTrackedCooldownName(value.AbilityName)],
		})
		if len(highlights) == limit {
			break
		}
	}

	return highlights
}

func abilityUseTimesSeconds(data timelineFightData, abilityID int, limit int) []float64 {
	if limit <= 0 {
		return nil
	}

	casts := buildAbilityTimelineSeries(data, abilityID, "", "", "").CastsMS
	if len(casts) == 0 {
		return nil
	}
	if len(casts) > limit {
		casts = casts[:limit]
	}

	values := make([]float64, 0, len(casts))
	for _, cast := range casts {
		values = append(values, float64(cast)/1000)
	}
	return values
}

func eliteMedianUseTimesSeconds(values []timelineFightData, abilityID int, limit int) []float64 {
	if limit <= 0 {
		return nil
	}

	useTimesByElite := make([][]float64, 0, len(values))
	maxUses := 0
	for _, value := range values {
		useTimes := abilityUseTimesSeconds(value, abilityID, limit)
		if len(useTimes) == 0 {
			continue
		}
		useTimesByElite = append(useTimesByElite, useTimes)
		if len(useTimes) > maxUses {
			maxUses = len(useTimes)
		}
	}

	if len(useTimesByElite) == 0 {
		return nil
	}

	medians := make([]float64, 0, maxUses)
	for useIndex := 0; useIndex < maxUses; useIndex++ {
		samples := make([]float64, 0, len(useTimesByElite))
		for _, useTimes := range useTimesByElite {
			if useIndex < len(useTimes) {
				samples = append(samples, useTimes[useIndex])
			}
		}
		if len(samples) == 0 {
			continue
		}
		medians = append(medians, medianFloat64(samples))
	}

	return medians
}

func largestBuffGapSeconds(data timelineFightData, abilityID int) float64 {
	windows := buffWindows(data, abilityID)
	if len(windows) == 0 {
		return data.FightEnd.Sub(data.FightStart).Seconds()
	}

	largestGap := windows[0].start.Sub(data.FightStart).Seconds()
	previousEnd := data.FightStart
	for _, window := range windows {
		gap := window.start.Sub(previousEnd).Seconds()
		if gap > largestGap {
			largestGap = gap
		}
		previousEnd = window.end
	}
	finalGap := data.FightEnd.Sub(previousEnd).Seconds()
	if finalGap > largestGap {
		largestGap = finalGap
	}
	if largestGap < 0 {
		return 0
	}
	return largestGap
}

func medianLargestBuffGapSeconds(values []timelineFightData, abilityID int) float64 {
	samples := make([]float64, 0, len(values))
	for _, value := range values {
		samples = append(samples, largestBuffGapSeconds(value, abilityID))
	}
	if len(samples) == 0 {
		return 0
	}
	return medianFloat64(samples)
}

type buffWindow struct {
	start time.Time
	end   time.Time
}

func buffWindows(data timelineFightData, abilityID int) []buffWindow {
	if len(data.BuffEvents) == 0 {
		return nil
	}

	windows := make([]buffWindow, 0)
	active := false
	start := data.FightStart

	for _, event := range data.BuffEvents {
		if event.Ability.ID != abilityID {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(event.EventType)) {
		case "apply":
			if !active {
				active = true
				start = event.Timestamp
			}
		case "refresh":
			if !active {
				active = true
				start = event.Timestamp
			}
		case "remove":
			if active {
				windows = append(windows, buffWindow{start: start, end: event.Timestamp})
				active = false
			}
		}
	}

	if active {
		windows = append(windows, buffWindow{start: start, end: data.FightEnd})
	}

	return windows
}

func medianFloat64(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	ordered := append([]float64(nil), values...)
	sort.Float64s(ordered)
	middle := len(ordered) / 2
	if len(ordered)%2 == 1 {
		return ordered[middle]
	}
	return (ordered[middle-1] + ordered[middle]) / 2
}

func filterAbilityHighlightsForAI(values []AbilityUsageComparison, prioritySet specPrioritySet) []AbilityUsageComparison {
	if len(values) == 0 {
		return nil
	}

	tracked := make(map[string]struct{}, len(prioritySet.Offensives)+len(prioritySet.Defensives)+1)
	for _, value := range prioritySet.Offensives {
		tracked[normalizeTrackedCooldownName(value)] = struct{}{}
	}
	for _, value := range prioritySet.Defensives {
		tracked[normalizeTrackedCooldownName(value)] = struct{}{}
	}
	tracked["melee"] = struct{}{}

	filtered := make([]AbilityUsageComparison, 0, len(values))
	for _, value := range values {
		if _, ok := tracked[normalizeTrackedCooldownName(value.AbilityName)]; ok {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func filterBuffHighlightsForAI(values []BuffUptimeComparison, prioritySet specPrioritySet) []BuffUptimeComparison {
	if len(values) == 0 {
		return nil
	}

	tracked := make(map[string]struct{}, len(prioritySet.Offensives)+len(prioritySet.Defensives))
	for _, value := range prioritySet.Offensives {
		tracked[normalizeTrackedCooldownName(value)] = struct{}{}
	}
	for _, value := range prioritySet.Defensives {
		tracked[normalizeTrackedCooldownName(value)] = struct{}{}
	}

	filtered := make([]BuffUptimeComparison, 0, len(values))
	for _, value := range values {
		if _, ok := tracked[normalizeTrackedCooldownName(value.AbilityName)]; ok {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func isHighFrequencyAbility(value AbilityUsageComparison) bool {
	return normalizeTrackedCooldownName(value.AbilityName) == "melee" || value.PlayerCount > 300 || value.CohortMedianCount > 300
}

func orderAbilityHighlights(values []AbilityUsageComparison, characterClass, characterSpec string) ([]AbilityUsageComparison, map[string]string) {
	prioritySet := specPriorityFor(characterClass, characterSpec)
	return orderTrackedAbilities(values, prioritySet), buildCooldownCategoryMap(prioritySet)
}

func orderBuffHighlights(values []BuffUptimeComparison, characterClass, characterSpec string) ([]BuffUptimeComparison, map[string]string) {
	prioritySet := specPriorityFor(characterClass, characterSpec)
	return orderTrackedBuffs(values, prioritySet), buildCooldownCategoryMap(prioritySet)
}

func orderTrackedAbilities(values []AbilityUsageComparison, prioritySet specPrioritySet) []AbilityUsageComparison {
	order := append(append([]string{}, prioritySet.Offensives...), prioritySet.Defensives...)
	indexByName := make(map[string]int, len(order))
	for index, value := range order {
		indexByName[normalizeTrackedCooldownName(value)] = index
	}

	ordered := append([]AbilityUsageComparison(nil), values...)
	sort.SliceStable(ordered, func(i, j int) bool {
		leftIndex, leftTracked := indexByName[normalizeTrackedCooldownName(ordered[i].AbilityName)]
		rightIndex, rightTracked := indexByName[normalizeTrackedCooldownName(ordered[j].AbilityName)]
		if leftTracked && rightTracked {
			return leftIndex < rightIndex
		}
		if leftTracked != rightTracked {
			return leftTracked
		}
		return strings.ToLower(ordered[i].AbilityName) < strings.ToLower(ordered[j].AbilityName)
	})

	return ordered
}

func orderTrackedBuffs(values []BuffUptimeComparison, prioritySet specPrioritySet) []BuffUptimeComparison {
	order := append(append([]string{}, prioritySet.Offensives...), prioritySet.Defensives...)
	indexByName := make(map[string]int, len(order))
	for index, value := range order {
		indexByName[normalizeTrackedCooldownName(value)] = index
	}

	ordered := append([]BuffUptimeComparison(nil), values...)
	sort.SliceStable(ordered, func(i, j int) bool {
		leftIndex, leftTracked := indexByName[normalizeTrackedCooldownName(ordered[i].AbilityName)]
		rightIndex, rightTracked := indexByName[normalizeTrackedCooldownName(ordered[j].AbilityName)]
		if leftTracked && rightTracked {
			return leftIndex < rightIndex
		}
		if leftTracked != rightTracked {
			return leftTracked
		}
		return strings.ToLower(ordered[i].AbilityName) < strings.ToLower(ordered[j].AbilityName)
	})

	return ordered
}
