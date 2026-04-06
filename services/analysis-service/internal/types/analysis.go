package types

import "time"

// PlayerFightMetrics contains all calculated metrics for a single player's fight
type PlayerFightMetrics struct {
	PlayerID     int                    `json:"playerId"`
	FightID      int                    `json:"fightId"`
	FightStart   time.Time              `json:"fightStart"`
	FightEnd     time.Time              `json:"fightEnd"`
	Duration     time.Duration          `json:"duration"`
	CastsPerMin  CastsPerMinuteMetric   `json:"castsPerMin"`
	MajorCDCount MajorCDCountMetric     `json:"majorCdCount"`
	MajorCDDrift MajorCDDriftMetric     `json:"majorCdDrift"`
	BuffUptime   BuffUptimeMetric       `json:"buffUptime"`
	DowntimePct  DowntimePercentageMetric `json:"downtimePct"`
}

// ComparisonResult contains the comparison of a player against a cohort
type ComparisonResult struct {
	PlayerMetrics PlayerFightMetrics     `json:"playerMetrics"`
	CohortStats   CohortStatistics       `json:"cohortStats"`
	Deltas        MetricDeltas           `json:"deltas"`
	Rankings      MetricRankings         `json:"rankings"`
}

// CohortStatistics contains aggregated statistics for the comparison cohort
type CohortStatistics struct {
	SampleSize     int                    `json:"sampleSize"`
	CastsPerMin    CohortMetricStats      `json:"castsPerMin"`
	MajorCDCount   CohortMetricStats      `json:"majorCdCount"`
	MajorCDDrift   CohortMetricStats      `json:"majorCdDrift"`
	BuffUptime     CohortMetricStats      `json:"buffUptime"`
	DowntimePct    CohortMetricStats      `json:"downtimePct"`
}

// CohortMetricStats contains statistical measures for a metric across the cohort
type CohortMetricStats struct {
	Mean     float64 `json:"mean"`
	Median   float64 `json:"median"`
	StdDev   float64 `json:"stdDev"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	P25      float64 `json:"p25"`      // 25th percentile
	P75      float64 `json:"p75"`      // 75th percentile
	P95      float64 `json:"p95"`      // 95th percentile
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
	PlayerValue    float64 `json:"playerValue"`
	CohortValue    float64 `json:"cohortValue"`    // typically median or mean
	Difference     float64 `json:"difference"`     // player - cohort
	Percentile     float64 `json:"percentile"`     // player's percentile in cohort
	Confidence     string  `json:"confidence"`     // "high", "medium", "low"
	Caution        string  `json:"caution,omitempty"` // explanation of uncertainty
}

// MetricRankings contains percentile rankings for each metric
type MetricRankings struct {
	CastsPerMin  float64 `json:"castsPerMin"`
	MajorCDCount float64 `json:"majorCdCount"`
	MajorCDDrift float64 `json:"majorCdDrift"`
	BuffUptime   float64 `json:"buffUptime"`
	DowntimePct  float64 `json:"downtimePct"`
}