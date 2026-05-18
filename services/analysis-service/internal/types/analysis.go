package types

import "time"

// PlayerFightMetrics contains all calculated metrics for a single player's fight
type PlayerFightMetrics struct {
	PlayerID     int                      `json:"playerId"`
	FightID      int                      `json:"fightId"`
	FightStart   time.Time                `json:"fightStart"`
	FightEnd     time.Time                `json:"fightEnd"`
	Duration     time.Duration            `json:"duration"`
	CastsPerMin  CastsPerMinuteMetric     `json:"castsPerMin"`
	MajorCDCount MajorCDCountMetric       `json:"majorCdCount"`
	MajorCDDrift MajorCDDriftMetric       `json:"majorCdDrift"`
	BuffUptime   BuffUptimeMetric         `json:"buffUptime"`
	DowntimePct  DowntimePercentageMetric `json:"downtimePct"`
}

// ComparisonResult contains the comparison of a player against a cohort
type ComparisonResult struct {
	PlayerMetrics PlayerFightMetrics        `json:"playerMetrics"`
	CohortStats   CohortStatistics          `json:"cohortStats"`
	Deltas        MetricDeltas              `json:"deltas"`
	Rankings      MetricRankings            `json:"rankings"`
	AbilityUsage  []AbilityUsageComparison  `json:"abilityUsage"`
	BuffUptimes   []BuffUptimeComparison    `json:"buffUptimes"`
	ResourceUsage []ResourceUsageComparison `json:"resourceUsage"`
}

// CohortStatistics contains aggregated statistics for the comparison cohort
type CohortStatistics struct {
	SampleSize   int               `json:"sampleSize"`
	CastsPerMin  CohortMetricStats `json:"castsPerMin"`
	MajorCDCount CohortMetricStats `json:"majorCdCount"`
	MajorCDDrift CohortMetricStats `json:"majorCdDrift"`
	BuffUptime   CohortMetricStats `json:"buffUptime"`
	DowntimePct  CohortMetricStats `json:"downtimePct"`
}

// CohortMetricStats contains statistical measures for a metric across the cohort
type CohortMetricStats struct {
	Mean   float64 `json:"mean"`
	Median float64 `json:"median"`
	StdDev float64 `json:"stdDev"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	P25    float64 `json:"p25"` // 25th percentile
	P75    float64 `json:"p75"` // 75th percentile
	P95    float64 `json:"p95"` // 95th percentile
}

// MetricDeltas contains the differences between player and cohort
type MetricDeltas struct {
	CastsPerMin  MetricDelta `json:"castsPerMin"`
	MajorCDCount MetricDelta `json:"majorCdCount"`
	MajorCDDrift MetricDelta `json:"majorCdDrift"`
	BuffUptime   MetricDelta `json:"buffUptime"`
	DowntimePct  MetricDelta `json:"downtimePct"`
}

// MetricDelta represents the difference between player value and cohort reference
type MetricDelta struct {
	PlayerValue float64 `json:"playerValue"`
	CohortValue float64 `json:"cohortValue"`       // typically median or mean
	Difference  float64 `json:"difference"`        // player - cohort
	Percentile  float64 `json:"percentile"`        // player's percentile in cohort
	Confidence  string  `json:"confidence"`        // "high", "medium", "low"
	Caution     string  `json:"caution,omitempty"` // explanation of uncertainty
}

// MetricRankings contains percentile rankings for each metric
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
	Confidence              string  `json:"confidence"`
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
	Confidence              string  `json:"confidence"`
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
	Confidence                     string  `json:"confidence"`
	Caution                        string  `json:"caution,omitempty"`
}
