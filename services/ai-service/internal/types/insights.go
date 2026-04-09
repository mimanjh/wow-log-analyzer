package types

type InsightGenerationRequest struct {
	Context           InsightContext     `json:"context"`
	Metrics           []InsightMetric    `json:"metrics,omitempty"`
	AbilityHighlights []InsightHighlight `json:"abilityHighlights,omitempty"`
	BuffHighlights    []InsightHighlight `json:"buffHighlights,omitempty"`
}

type InsightContext struct {
	EncounterName    string `json:"encounterName"`
	Difficulty       string `json:"difficulty"`
	CharacterName    string `json:"characterName"`
	CharacterClass   string `json:"characterClass"`
	CharacterSpec    string `json:"characterSpec"`
	FightDurationSec int    `json:"fightDurationSec"`
	CohortSize       int    `json:"cohortSize"`
}

type InsightMetric struct {
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

type InsightHighlight struct {
	Name                  string    `json:"name"`
	PlayerValue           float64   `json:"playerValue"`
	EliteValue            float64   `json:"eliteValue"`
	Difference            float64   `json:"difference"`
	Unit                  string    `json:"unit,omitempty"`
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

type InsightGenerationResponse struct {
	Insights            []AIInsight         `json:"insights"`
	FocusRecommendation FocusRecommendation `json:"focusRecommendation"`
	FallbackUsed        bool                `json:"fallbackUsed"`
	Model               string              `json:"model"`
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
