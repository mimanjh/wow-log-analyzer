package types

// CastsPerMinuteMetric represents casts per minute calculation
type CastsPerMinuteMetric struct {
	Value         float64 `json:"value"`
	TotalCasts    int     `json:"totalCasts"`
	FightDuration float64 `json:"fightDuration"` // in minutes
	Confidence    string  `json:"confidence"`    // "high", "medium", "low"
	Caution       string  `json:"caution,omitempty"`
}

// MajorCDCountMetric represents major cooldown usage count
type MajorCDCountMetric struct {
	Value         int     `json:"value"`
	TotalCooldowns int    `json:"totalCooldowns"`
	Confidence    string  `json:"confidence"`    // "high", "medium", "low"
	Caution       string  `json:"caution,omitempty"`
}

// MajorCDDriftMetric represents timing drift of major cooldowns
type MajorCDDriftMetric struct {
	Value         float64 `json:"value"`         // average drift in seconds
	TotalDrift    float64 `json:"totalDrift"`    // total drift across all CDs
	CooldownCount int     `json:"cooldownCount"`
	Confidence    string  `json:"confidence"`    // "high", "medium", "low"
	Caution       string  `json:"caution,omitempty"`
}

// BuffUptimeMetric represents buff uptime percentage
type BuffUptimeMetric struct {
	Value         float64 `json:"value"`         // percentage (0-100)
	TotalUptime   float64 `json:"totalUptime"`   // total uptime in seconds
	FightDuration float64 `json:"fightDuration"` // fight duration in seconds
	Confidence    string  `json:"confidence"`    // "high", "medium", "low"
	Caution       string  `json:"caution,omitempty"`
}

// DowntimePercentageMetric represents percentage of fight spent not doing damage/healing
type DowntimePercentageMetric struct {
	Value         float64 `json:"value"`         // percentage (0-100)
	TotalDowntime float64 `json:"totalDowntime"` // total downtime in seconds
	FightDuration float64 `json:"fightDuration"` // fight duration in seconds
	Confidence    string  `json:"confidence"`    // "high", "medium", "low"
	Caution       string  `json:"caution,omitempty"`
}