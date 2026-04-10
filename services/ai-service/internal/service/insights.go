package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"wow-log-analyzer/services/ai-service/internal/config"
	"wow-log-analyzer/services/ai-service/internal/types"
)

type ModelClient interface {
	Generate(ctx context.Context, prompt string, req types.InsightGenerationRequest) (*types.InsightGenerationResponse, error)
}

type InsightService struct {
	modelName   string
	modelClient ModelClient
}

func NewInsightService(cfg config.Config) *InsightService {
	return &InsightService{
		modelName:   cfg.Model,
		modelClient: newModelClient(cfg),
	}
}

func (s *InsightService) GenerateInsights(ctx context.Context, req types.InsightGenerationRequest) (types.InsightGenerationResponse, error) {
	if err := validateInsightRequest(req); err != nil {
		return types.InsightGenerationResponse{}, err
	}

	prompt := buildPrompt(req)
	if s.modelClient != nil {
		response, err := s.modelClient.Generate(ctx, prompt, req)
		if err == nil && response != nil {
			response.FallbackUsed = false
			if response.Model == "" {
				response.Model = s.modelName
			}
			return *response, nil
		}
		if err != nil {
			fmt.Printf("ai-service model generation failed: %v\n", err)
		}
	}

	fallback := formatFallbackInsights(req)
	fallback.FallbackUsed = true
	fallback.Model = "deterministic-fallback"
	return fallback, nil
}

func validateInsightRequest(req types.InsightGenerationRequest) error {
	if strings.TrimSpace(req.Context.EncounterName) == "" {
		return errors.New("context.encounterName is required")
	}
	if strings.TrimSpace(req.Context.CharacterName) == "" {
		return errors.New("context.characterName is required")
	}
	if req.Context.CohortSize <= 0 {
		return errors.New("context.cohortSize must be greater than zero")
	}
	if len(req.Metrics) == 0 && len(req.AbilityHighlights) == 0 && len(req.BuffHighlights) == 0 {
		return errors.New("at least one metric or highlight is required")
	}

	for _, metric := range req.Metrics {
		if strings.TrimSpace(metric.Key) == "" {
			return errors.New("metric key is required")
		}
		if strings.TrimSpace(metric.Label) == "" {
			return fmt.Errorf("metric %s label is required", metric.Key)
		}
		if metric.Percentile < 0 || metric.Percentile > 100 {
			return fmt.Errorf("metric %s percentile must be between 0 and 100", metric.Key)
		}
		switch metric.Confidence {
		case "high", "medium", "low":
		default:
			return fmt.Errorf("metric %s confidence must be high, medium, or low", metric.Key)
		}
	}

	for _, highlight := range append(append([]types.InsightHighlight(nil), req.AbilityHighlights...), req.BuffHighlights...) {
		if strings.TrimSpace(highlight.Name) == "" {
			return errors.New("highlight name is required")
		}
		if highlight.Category != "" && highlight.Category != "offensive" && highlight.Category != "defensive" {
			return fmt.Errorf("highlight %s category must be offensive or defensive when provided", highlight.Name)
		}
	}

	return nil
}

type rankedMetric struct {
	metric        types.InsightMetric
	concernScore  float64
	positiveScore float64
	isConcern     bool
}

type rankedHighlight struct {
	highlight     types.InsightHighlight
	section       string
	concernScore  float64
	positiveScore float64
	isConcern     bool
	key           string
}

func formatFallbackInsights(req types.InsightGenerationRequest) types.InsightGenerationResponse {
	if len(req.AbilityHighlights) > 0 || len(req.BuffHighlights) > 0 {
		return formatTimelineDrivenFallback(req)
	}

	ranked := rankMetrics(req.Metrics)
	insights := make([]types.AIInsight, 0, 3)

	for _, item := range ranked {
		if item.concernScore <= 0 && len(insights) > 0 {
			continue
		}
		insights = append(insights, buildMetricInsight(item.metric, item.isConcern))
		if len(insights) == 3 {
			break
		}
	}

	if len(insights) < 3 {
		for _, item := range ranked {
			if containsMetric(insights, item.metric.Key) {
				continue
			}
			insights = append(insights, buildMetricInsight(item.metric, item.isConcern))
			if len(insights) == 3 {
				break
			}
		}
	}

	return types.InsightGenerationResponse{
		Insights:            insights,
		FocusRecommendation: buildMetricFocusRecommendation(req.Context, ranked),
	}
}

func formatTimelineDrivenFallback(req types.InsightGenerationRequest) types.InsightGenerationResponse {
	ranked := rankHighlights(req)
	insights := make([]types.AIInsight, 0, 3)

	for _, item := range ranked {
		if item.concernScore <= 0 && len(insights) > 0 {
			continue
		}
		insights = append(insights, buildHighlightInsight(item))
		if len(insights) == 3 {
			break
		}
	}

	if len(insights) < 3 {
		for _, item := range ranked {
			if containsMetric(insights, item.key) {
				continue
			}
			insights = append(insights, buildHighlightInsight(item))
			if len(insights) == 3 {
				break
			}
		}
	}

	return types.InsightGenerationResponse{
		Insights:            insights,
		FocusRecommendation: buildHighlightFocusRecommendation(req.Context, ranked),
	}
}

func rankMetrics(metrics []types.InsightMetric) []rankedMetric {
	ranked := make([]rankedMetric, 0, len(metrics))
	for _, metric := range metrics {
		concernDelta := concernMagnitude(metric)
		positiveDelta := positiveMagnitude(metric)
		ranked = append(ranked, rankedMetric{
			metric:        metric,
			concernScore:  concernDelta*confidenceWeight(metric.Confidence) + cautionBoost(metric.Caution),
			positiveScore: positiveDelta * confidenceWeight(metric.Confidence),
			isConcern:     concernDelta > 0,
		})
	}

	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].concernScore == ranked[j].concernScore {
			if ranked[i].positiveScore == ranked[j].positiveScore {
				return ranked[i].metric.Label < ranked[j].metric.Label
			}
			return ranked[i].positiveScore > ranked[j].positiveScore
		}
		return ranked[i].concernScore > ranked[j].concernScore
	})

	return ranked
}

func rankHighlights(req types.InsightGenerationRequest) []rankedHighlight {
	ranked := make([]rankedHighlight, 0, len(req.AbilityHighlights)+len(req.BuffHighlights))
	for _, highlight := range req.AbilityHighlights {
		ranked = append(ranked, buildRankedHighlight(highlight, "ability"))
	}
	for _, highlight := range req.BuffHighlights {
		ranked = append(ranked, buildRankedHighlight(highlight, "buff"))
	}

	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].concernScore == ranked[j].concernScore {
			if ranked[i].positiveScore == ranked[j].positiveScore {
				return ranked[i].highlight.Name < ranked[j].highlight.Name
			}
			return ranked[i].positiveScore > ranked[j].positiveScore
		}
		return ranked[i].concernScore > ranked[j].concernScore
	})

	return ranked
}

func buildRankedHighlight(highlight types.InsightHighlight, section string) rankedHighlight {
	concern := 0.0
	positive := 0.0
	isConcern := highlight.Difference < 0
	timingWeight := math.Abs(highlight.TimingDeltaSeconds) / 10
	if len(highlight.PlayerUseTimesSeconds) > 0 || len(highlight.EliteUseTimesSeconds) > 0 {
		timingWeight += compareUseTimeSequences(highlight.PlayerUseTimesSeconds, highlight.EliteUseTimesSeconds) / 10
	}
	if highlight.PlayerLargestGapSec > 0 || highlight.EliteLargestGapSec > 0 {
		timingWeight += math.Abs(highlight.PlayerLargestGapSec-highlight.EliteLargestGapSec) / 10
	}
	if isConcern {
		concern = math.Abs(highlight.Difference) + timingWeight
	} else {
		positive = math.Abs(highlight.Difference) + timingWeight
	}

	return rankedHighlight{
		highlight:     highlight,
		section:       section,
		concernScore:  concern,
		positiveScore: positive,
		isConcern:     isConcern,
		key:           section + ":" + strings.ToLower(highlight.Name),
	}
}

func buildMetricInsight(metric types.InsightMetric, isConcern bool) types.AIInsight {
	direction := compareDirection(metric)
	deltaText := formatSigned(metric.Difference, metric.Unit)
	playerText := formatValue(metric.PlayerValue, metric.Unit)
	cohortText := formatValue(metric.CohortValue, metric.Unit)

	var summary string
	if isConcern {
		summary = fmt.Sprintf("%s was %s the cohort reference (%s vs %s). This may point to a meaningful gap, although it should be read with the rest of the fight context.", metric.Label, direction, playerText, cohortText)
	} else {
		summary = fmt.Sprintf("%s was %s the cohort reference (%s vs %s). This looks like a relative strength, though it may still vary by fight pattern.", metric.Label, direction, playerText, cohortText)
	}

	if metric.Caution != "" {
		summary = fmt.Sprintf("%s %s.", summary, metric.Caution)
	}

	return types.AIInsight{
		MetricKey:  metric.Key,
		Title:      fmt.Sprintf("%s (%s)", metric.Label, deltaText),
		Summary:    summary,
		Confidence: metric.Confidence,
		Caution:    metric.Caution,
	}
}

func buildHighlightInsight(item rankedHighlight) types.AIInsight {
	sectionLabel := "Ability usage"
	if item.section == "buff" {
		sectionLabel = "Buff uptime"
	}

	deltaText := formatSigned(item.highlight.Difference, item.highlight.Unit)
	playerText := formatValue(item.highlight.PlayerValue, item.highlight.Unit)
	eliteText := formatValue(item.highlight.EliteValue, item.highlight.Unit)
	timingText := buildHighlightTimingText(item.highlight, item.section)

	summary := fmt.Sprintf(
		"%s for %s was %s versus elite (%s vs %s).%s",
		sectionLabel,
		item.highlight.Name,
		compareHighlightDirection(item.highlight.Difference),
		playerText,
		eliteText,
		timingText,
	)

	if item.isConcern {
		summary += " This is one of the clearer gaps to review in the timeline."
	} else {
		summary += " This looks like a relative strength compared with the elite sample."
	}

	return types.AIInsight{
		MetricKey:  item.key,
		Title:      fmt.Sprintf("%s (%s)", item.highlight.Name, deltaText),
		Summary:    summary,
		Confidence: "high",
	}
}

func buildMetricFocusRecommendation(context types.InsightContext, ranked []rankedMetric) types.FocusRecommendation {
	for _, item := range ranked {
		if item.concernScore <= 0 {
			continue
		}

		return types.FocusRecommendation{
			MetricKey: item.metric.Key,
			Title:     fmt.Sprintf("Focus on %s", item.metric.Label),
			Recommendation: fmt.Sprintf(
				"Prioritize %s in a similar %s %s pull. The current result trails the cohort benchmark, so this is the cleanest next area to review.",
				strings.ToLower(item.metric.Label),
				strings.ToLower(context.Difficulty),
				context.EncounterName,
			),
			Reasoning: fmt.Sprintf("%s is %s the cohort reference at the %s confidence level.", item.metric.Label, compareDirection(item.metric), item.metric.Confidence),
		}
	}

	best := ranked[0].metric
	return types.FocusRecommendation{
		MetricKey: best.Key,
		Title:     fmt.Sprintf("Preserve %s", best.Label),
		Recommendation: fmt.Sprintf(
			"Keep %s stable while reviewing smaller gaps. It currently compares well against the cohort for %s.",
			strings.ToLower(best.Label),
			context.CharacterName,
		),
		Reasoning: fmt.Sprintf("%s is not trailing the cohort and appears to be one of the steadier parts of the fight.", best.Label),
	}
}

func buildHighlightFocusRecommendation(context types.InsightContext, ranked []rankedHighlight) types.FocusRecommendation {
	for _, item := range ranked {
		if item.concernScore <= 0 {
			continue
		}

		area := "ability usage"
		if item.section == "buff" {
			area = "buff timing and uptime"
		}

		return types.FocusRecommendation{
			MetricKey: item.key,
			Title:     fmt.Sprintf("Focus on %s", item.highlight.Name),
			Recommendation: fmt.Sprintf(
				"Review %s on a similar %s %s pull. It is one of the clearest gaps against the elite %s comparisons.",
				strings.ToLower(item.highlight.Name),
				strings.ToLower(context.Difficulty),
				context.EncounterName,
				area,
			),
			Reasoning: fmt.Sprintf(
				"%s trails the elite sample at %s versus %s. %s",
				item.highlight.Name,
				formatValue(item.highlight.PlayerValue, item.highlight.Unit),
				formatValue(item.highlight.EliteValue, item.highlight.Unit),
				focusTimingReason(item.highlight, item.section),
			),
		}
	}

	best := ranked[0]
	return types.FocusRecommendation{
		MetricKey: best.key,
		Title:     fmt.Sprintf("Preserve %s", best.highlight.Name),
		Recommendation: fmt.Sprintf(
			"Keep %s steady while reviewing smaller gaps. It currently compares well against the elite sample for %s.",
			strings.ToLower(best.highlight.Name),
			context.CharacterName,
		),
		Reasoning: fmt.Sprintf("%s is one of the steadier timeline comparisons in this fight.", best.highlight.Name),
	}
}

func containsMetric(insights []types.AIInsight, key string) bool {
	for _, insight := range insights {
		if insight.MetricKey == key {
			return true
		}
	}
	return false
}

func concernMagnitude(metric types.InsightMetric) float64 {
	if metricIsConcern(metric) {
		return math.Abs(metric.Difference) + math.Abs(50-metric.Percentile)/10
	}
	return 0
}

func positiveMagnitude(metric types.InsightMetric) float64 {
	if metricIsConcern(metric) {
		return 0
	}
	return math.Abs(metric.Difference) + math.Abs(metric.Percentile-50)/10
}

func metricIsConcern(metric types.InsightMetric) bool {
	if metric.HigherIsBetter {
		return metric.Difference < 0
	}
	return metric.Difference > 0
}

func compareDirection(metric types.InsightMetric) string {
	if metric.Difference == 0 {
		return "in line with"
	}

	if metricIsConcern(metric) {
		if metric.HigherIsBetter {
			return "below"
		}
		return "above"
	}

	if metric.HigherIsBetter {
		return "above"
	}
	return "below"
}

func compareHighlightDirection(difference float64) string {
	if difference == 0 {
		return "in line with"
	}
	if difference > 0 {
		return "above"
	}
	return "below"
}

func confidenceWeight(confidence string) float64 {
	switch confidence {
	case "high":
		return 1.0
	case "medium":
		return 0.8
	default:
		return 0.6
	}
}

func cautionBoost(caution string) float64 {
	if strings.TrimSpace(caution) == "" {
		return 0
	}
	return 0.2
}

func formatSigned(value float64, unit string) string {
	sign := ""
	if value > 0 {
		sign = "+"
	}
	return sign + formatValue(value, unit)
}

func formatValue(value float64, unit string) string {
	if math.Abs(value-math.Round(value)) < 0.01 {
		return fmt.Sprintf("%.0f%s", value, unit)
	}
	return fmt.Sprintf("%.1f%s", value, unit)
}

func buildPrompt(req types.InsightGenerationRequest) string {
	var metricLines []string
	for _, metric := range req.Metrics {
		metricLines = append(
			metricLines,
			fmt.Sprintf(
				"- %s: player=%s, cohort=%s, delta=%s, percentile=%.1f, confidence=%s, caution=%s",
				metric.Label,
				formatValue(metric.PlayerValue, metric.Unit),
				formatValue(metric.CohortValue, metric.Unit),
				formatSigned(metric.Difference, metric.Unit),
				metric.Percentile,
				metric.Confidence,
				emptyFallback(metric.Caution, "none"),
			),
		)
	}

	var abilityLines []string
	for _, highlight := range req.AbilityHighlights {
		abilityLines = append(abilityLines, formatAbilityHighlightLine(highlight))
	}

	var buffLines []string
	for _, highlight := range req.BuffHighlights {
		buffLines = append(buffLines, formatBuffHighlightLine(highlight))
	}

	sections := []string{
		fmt.Sprintf(
			"Generate exactly 3 concise cautious insights and 1 focus recommendation for %s %s %s on %s %s. Use only these deterministic comparison outputs, especially the ability and buff comparisons against elite logs on the same boss. Prioritize timing, usage gaps, late windows, missing buttons, and clear strengths. Do not mention raw logs or invent causality.",
			req.Context.CharacterSpec,
			req.Context.CharacterClass,
			req.Context.CharacterName,
			req.Context.Difficulty,
			req.Context.EncounterName,
		),
		fmt.Sprintf(
			"Context: fightDuration=%ds, eliteSample=%d",
			req.Context.FightDurationSec,
			req.Context.CohortSize,
		),
	}

	if len(abilityLines) > 0 {
		sections = append(sections, "Top ability usage comparisons:\n"+strings.Join(abilityLines, "\n"))
	}
	if len(buffLines) > 0 {
		sections = append(sections, "Top buff uptime comparisons:\n"+strings.Join(buffLines, "\n"))
	}
	if len(metricLines) > 0 {
		sections = append(sections, "Legacy metric context:\n"+strings.Join(metricLines, "\n"))
	}

	return strings.Join(sections, "\n\n")
}

func formatAbilityHighlightLine(highlight types.InsightHighlight) string {
	label := highlight.Name
	if highlight.Category != "" {
		label = fmt.Sprintf("[%s] %s", highlight.Category, highlight.Name)
	}

	timingPart := ""
	if len(highlight.PlayerUseTimesSeconds) > 0 || len(highlight.EliteUseTimesSeconds) > 0 {
		timingPart = fmt.Sprintf(
			", player uses: %s, elite median uses: %s",
			formatTimelineList(highlight.PlayerUseTimesSeconds),
			formatTimelineList(highlight.EliteUseTimesSeconds),
		)
	} else if highlight.TimingLabel != "" && (highlight.PlayerTimingSeconds > 0 || highlight.EliteTimingSeconds > 0) {
		timingPart = fmt.Sprintf(
			", %s player=%s, elite=%s, delta=%s",
			highlight.TimingLabel,
			formatValue(highlight.PlayerTimingSeconds, "s"),
			formatValue(highlight.EliteTimingSeconds, "s"),
			formatSigned(highlight.TimingDeltaSeconds, "s"),
		)
	}

	countPart := ""
	if highlight.PlayerRawCount > 0 || highlight.EliteRawCount > 0 {
		countPart = fmt.Sprintf(
			", player count=%s, elite median count=%s",
			formatValue(highlight.PlayerRawCount, "casts"),
			formatValue(highlight.EliteRawCount, "casts"),
		)
	}

	return fmt.Sprintf(
		"- %s: player=%s, elite=%s, delta=%s%s%s",
		label,
		formatValue(highlight.PlayerValue, highlight.Unit),
		formatValue(highlight.EliteValue, highlight.Unit),
		formatSigned(highlight.Difference, highlight.Unit),
		countPart,
		timingPart,
	)
}

func formatBuffHighlightLine(highlight types.InsightHighlight) string {
	label := highlight.Name
	if highlight.Category != "" {
		label = fmt.Sprintf("[%s] %s", highlight.Category, highlight.Name)
	}

	gapPart := ""
	if highlight.PlayerLargestGapSec > 0 || highlight.EliteLargestGapSec > 0 {
		gapPart = fmt.Sprintf(
			", player largest gap=%s, elite largest gap=%s",
			formatValue(highlight.PlayerLargestGapSec, "s"),
			formatValue(highlight.EliteLargestGapSec, "s"),
		)
	}

	return fmt.Sprintf(
		"- %s: player=%s, elite=%s, delta=%s%s",
		label,
		formatValue(highlight.PlayerValue, highlight.Unit),
		formatValue(highlight.EliteValue, highlight.Unit),
		formatSigned(highlight.Difference, highlight.Unit),
		gapPart,
	)
}

func formatTimelineList(values []float64) string {
	if len(values) == 0 {
		return "none"
	}

	parts := make([]string, 0, len(values))
	for _, value := range values {
		totalSeconds := int(math.Round(value))
		minutes := totalSeconds / 60
		seconds := totalSeconds % 60
		parts = append(parts, fmt.Sprintf("%d:%02d", minutes, seconds))
	}
	return strings.Join(parts, ", ")
}

func compareUseTimeSequences(player, elite []float64) float64 {
	limit := len(player)
	if len(elite) < limit {
		limit = len(elite)
	}
	if limit == 0 {
		return 0
	}

	total := 0.0
	for index := 0; index < limit; index++ {
		total += math.Abs(player[index] - elite[index])
	}
	return total / float64(limit)
}

func buildHighlightTimingText(highlight types.InsightHighlight, section string) string {
	if section == "ability" && (len(highlight.PlayerUseTimesSeconds) > 0 || len(highlight.EliteUseTimesSeconds) > 0) {
		return fmt.Sprintf(
			" Player uses were %s versus elite median %s.",
			formatTimelineList(highlight.PlayerUseTimesSeconds),
			formatTimelineList(highlight.EliteUseTimesSeconds),
		)
	}
	if section == "buff" && (highlight.PlayerLargestGapSec > 0 || highlight.EliteLargestGapSec > 0) {
		return fmt.Sprintf(
			" Largest buff gap was %s versus elite %s.",
			formatValue(highlight.PlayerLargestGapSec, "s"),
			formatValue(highlight.EliteLargestGapSec, "s"),
		)
	}
	if highlight.TimingLabel != "" && (highlight.PlayerTimingSeconds > 0 || highlight.EliteTimingSeconds > 0) {
		return fmt.Sprintf(
			" %s was %s versus elite %s (%s).",
			strings.Title(highlight.TimingLabel),
			formatValue(highlight.PlayerTimingSeconds, "s"),
			formatValue(highlight.EliteTimingSeconds, "s"),
			formatSigned(highlight.TimingDeltaSeconds, "s"),
		)
	}
	return ""
}

func focusTimingReason(highlight types.InsightHighlight, section string) string {
	if section == "ability" && (len(highlight.PlayerUseTimesSeconds) > 0 || len(highlight.EliteUseTimesSeconds) > 0) {
		return fmt.Sprintf(
			"Player uses were %s versus elite median %s.",
			formatTimelineList(highlight.PlayerUseTimesSeconds),
			formatTimelineList(highlight.EliteUseTimesSeconds),
		)
	}
	if section == "buff" && (highlight.PlayerLargestGapSec > 0 || highlight.EliteLargestGapSec > 0) {
		return fmt.Sprintf(
			"Largest buff gap was %s versus elite %s.",
			formatValue(highlight.PlayerLargestGapSec, "s"),
			formatValue(highlight.EliteLargestGapSec, "s"),
		)
	}
	return fmt.Sprintf(
		"%s was %s versus %s.",
		emptyFallback(highlight.TimingLabel, "timing"),
		formatValue(highlight.PlayerTimingSeconds, "s"),
		formatValue(highlight.EliteTimingSeconds, "s"),
	)
}

func emptyFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
