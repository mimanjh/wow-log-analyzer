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
	if len(req.Metrics) == 0 {
		return errors.New("at least one metric is required")
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

	for _, highlight := range append(
		append([]types.InsightHighlight(nil), req.AbilityHighlights...),
		req.BuffHighlights...,
	) {
		if strings.TrimSpace(highlight.Name) == "" {
			return errors.New("highlight name is required")
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

func formatFallbackInsights(req types.InsightGenerationRequest) types.InsightGenerationResponse {
	ranked := rankMetrics(req.Metrics)
	insights := make([]types.AIInsight, 0, 3)

	for _, item := range ranked {
		if item.concernScore <= 0 && len(insights) > 0 {
			continue
		}
		insights = append(insights, buildInsight(item.metric, item.isConcern))
		if len(insights) == 3 {
			break
		}
	}

	if len(insights) < 3 {
		for _, item := range ranked {
			if containsMetric(insights, item.metric.Key) {
				continue
			}
			insights = append(insights, buildInsight(item.metric, item.isConcern))
			if len(insights) == 3 {
				break
			}
		}
	}

	return types.InsightGenerationResponse{
		Insights:            insights,
		FocusRecommendation: buildFocusRecommendation(req.Context, ranked),
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

func buildInsight(metric types.InsightMetric, isConcern bool) types.AIInsight {
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

func buildFocusRecommendation(context types.InsightContext, ranked []rankedMetric) types.FocusRecommendation {
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
		abilityLines = append(
			abilityLines,
			fmt.Sprintf(
				"- %s: player=%s, elite=%s, delta=%s",
				highlight.Name,
				formatValue(highlight.PlayerValue, highlight.Unit),
				formatValue(highlight.EliteValue, highlight.Unit),
				formatSigned(highlight.Difference, highlight.Unit),
			),
		)
	}

	var buffLines []string
	for _, highlight := range req.BuffHighlights {
		buffLines = append(
			buffLines,
			fmt.Sprintf(
				"- %s: player=%s, elite=%s, delta=%s",
				highlight.Name,
				formatValue(highlight.PlayerValue, highlight.Unit),
				formatValue(highlight.EliteValue, highlight.Unit),
				formatSigned(highlight.Difference, highlight.Unit),
			),
		)
	}

	sections := []string{
		fmt.Sprintf(
			"Generate exactly 3 concise cautious insights and 1 focus recommendation for %s %s on %s %s. Use only these deterministic comparison outputs. Do not mention raw logs or invent causality.",
			req.Context.CharacterName,
			req.Context.CharacterSpec,
			req.Context.Difficulty,
			req.Context.EncounterName,
		),
		strings.Join(metricLines, "\n"),
	}

	if len(abilityLines) > 0 {
		sections = append(sections, "Top ability usage comparisons:\n"+strings.Join(abilityLines, "\n"))
	}
	if len(buffLines) > 0 {
		sections = append(sections, "Top buff uptime comparisons:\n"+strings.Join(buffLines, "\n"))
	}

	return strings.Join(sections, "\n\n")
}

func emptyFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
