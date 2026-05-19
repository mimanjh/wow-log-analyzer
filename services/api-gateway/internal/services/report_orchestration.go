package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"wow-log-analyzer/services/api-gateway/internal/config"
)

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
		response.AI.Warning = firstNonEmpty(insights.Warning, "AI used the deterministic fallback formatter for this report.")
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
	specProfile, _ := config.SpecProfileFor(req.Character.Class, req.Character.Spec)
	return insightGenerationRequest{
		Context: insightContext{
			EncounterName:    req.Fight.Name,
			Difficulty:       req.Fight.Difficulty,
			CharacterName:    req.Character.Name,
			CharacterClass:   req.Character.Class,
			CharacterSpec:    req.Character.Spec,
			FightDurationSec: req.Fight.KillTime,
			CohortSize:       comparison.CohortStats.SampleSize,
			SpecProfile:      specProfile,
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
