package services

import (
	"time"

	"wow-log-analyzer/services/api-gateway/internal/config"
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

type insightGenerationRequest struct {
	Context            insightContext     `json:"context"`
	Metrics            []insightMetric    `json:"metrics"`
	AbilityHighlights  []insightHighlight `json:"abilityHighlights,omitempty"`
	BuffHighlights     []insightHighlight `json:"buffHighlights,omitempty"`
	ResourceHighlights []insightHighlight `json:"resourceHighlights,omitempty"`
}

type insightContext struct {
	EncounterName    string             `json:"encounterName"`
	Difficulty       string             `json:"difficulty"`
	CharacterName    string             `json:"characterName"`
	CharacterClass   string             `json:"characterClass"`
	CharacterSpec    string             `json:"characterSpec"`
	FightDurationSec int                `json:"fightDurationSec"`
	CohortSize       int                `json:"cohortSize"`
	SpecProfile      config.SpecProfile `json:"specProfile,omitempty"`
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
	Warning             string              `json:"warning,omitempty"`
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

	// Owner fields scope job reads to the creating user. They round-trip
	// through Redis persistence; handlers must zero them before writing a job
	// to the HTTP response.
	OwnerUserID    int    `json:"ownerUserId,omitempty"`
	OwnerSessionID string `json:"ownerSessionId,omitempty"`

	timeline *reportTimelineData `json:"-"`
}

// JobOwner identifies who created a report job. UserID is the WCL user ID
// when known; SessionID is the fallback identity for sessions whose user
// lookup failed.
type JobOwner struct {
	SessionID string
	UserID    int
}

// CanAccess reports whether the given session may read this job.
func (j ReportJob) CanAccess(session *SessionState) bool {
	if session == nil {
		return false
	}
	if j.OwnerUserID != 0 && session.User != nil && session.User.ID == j.OwnerUserID {
		return true
	}
	return j.OwnerSessionID != "" && j.OwnerSessionID == session.ID
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
