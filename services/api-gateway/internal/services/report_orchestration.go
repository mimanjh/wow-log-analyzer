package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	analysistypes "wow-log-analyzer/services/analysis-service/types"
	"wow-log-analyzer/services/api-gateway/internal/config"
)

// getOrFetchCohortMemberContext returns the precomputed MemberContext for a
// cohort candidate, falling back to a raw fetch + analysis-service compute on
// cache miss. The MemberContext is what we cache now (~5-20KB) instead of the
// ~1MB raw fight JSON — the cache-hit path skips both the WCL call AND the
// analysis-service computation.
//
// The second return is a decoded timelineFightData populated ONLY on the cache
// miss path (we already have the raw JSON in hand, so we hand it back so the
// timeline UI can render this elite's per-ability bars). On cache hit it's
// nil — that elite is absent from the timeline UI but still contributes to
// the deterministic comparison and AI insights via MemberContext.
func (s *ReportService) getOrFetchCohortMemberContext(ctx context.Context, candidate RankingCandidate, characterClass, characterSpec string) (*analysistypes.MemberContext, *timelineFightData, error) {
	key := s.key(cohortMemberContextCacheKeyPrefix + rankingCandidateKey(candidate))
	if data, ok := s.getCachedRaw(ctx, key); ok {
		var memberCtx analysistypes.MemberContext
		if err := json.Unmarshal(data, &memberCtx); err == nil {
			return &memberCtx, nil, nil
		}
	}
	raw, err := s.fetchCohortMember(ctx, candidate)
	if err != nil {
		return nil, nil, err
	}
	memberCtx, err := s.computeMemberContext(ctx, raw, characterClass, characterSpec)
	if err != nil {
		return nil, nil, err
	}
	if data, err := json.Marshal(memberCtx); err == nil {
		s.setCachedRaw(ctx, key, data, cohortMemberContextCacheTTL)
	}
	decoded, decodeErr := decodeTimelineFightData(raw)
	if decodeErr != nil {
		return memberCtx, nil, nil
	}
	return memberCtx, &decoded, nil
}

// computeMemberContext POSTs raw fight JSON to analysis-service /analyze/member
// and returns the small computed snapshot. Used for both the player's own data
// and for cohort members on cache miss.
func (s *ReportService) computeMemberContext(ctx context.Context, raw json.RawMessage, characterClass, characterSpec string) (*analysistypes.MemberContext, error) {
	var memberCtx analysistypes.MemberContext
	err := s.postForJSON(
		ctx,
		s.analysisClient,
		s.analysisURL+"/analyze/member",
		analyzeMemberRequest{
			PlayerData:     raw,
			CharacterClass: characterClass,
			CharacterSpec:  characterSpec,
		},
		&memberCtx,
		"analysis-service",
	)
	if err != nil {
		return nil, err
	}
	return &memberCtx, nil
}

type analyzeMemberRequest struct {
	PlayerData     json.RawMessage `json:"playerData"`
	CharacterClass string          `json:"characterClass"`
	CharacterSpec  string          `json:"characterSpec"`
}

type compareContextsRequest struct {
	Player         analysistypes.MemberContext   `json:"player"`
	Cohort         []analysistypes.MemberContext `json:"cohort"`
	CharacterClass string                        `json:"characterClass"`
	CharacterSpec  string                        `json:"characterSpec"`
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
	playerCtx, err := s.computeMemberContext(ctx, playerData, req.Character.Class, req.Character.Spec)
	if err != nil {
		s.updateJob(jobID, ReportJobFailed, "player-data", "Failed to compute selected player metrics.", ReportJobProgress{Current: 1, Total: 5}, err.Error(), nil)
		return
	}

	s.updateJob(jobID, ReportJobRunning, "rankings", "Fetching ranking candidates for the selected boss and spec.", ReportJobProgress{Current: 2, Total: 5}, "", nil)
	cohortEntries := make([]CohortEntry, 0, targetEliteCount)
	cohortContexts := make([]analysistypes.MemberContext, 0, targetEliteCount)
	// EliteData/EliteEntries on the timeline are populated only for members
	// fetched fresh from log-service this run. Cache-hit members (no raw kept)
	// contribute to comparison + AI insights but not the per-elite timeline UI.
	cohortTimelineData := make([]timelineFightData, 0, targetEliteCount)
	cohortTimelineEntries := make([]CohortEntry, 0, targetEliteCount)
	processedCandidates := make(map[string]bool)

	for candidateLimit := rankingCandidateBatchCap; len(cohortContexts) < targetEliteCount && candidateLimit <= rankingCandidateMaxCap; candidateLimit += rankingCandidateBatchCap {
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

			s.updateJob(jobID, ReportJobRunning, "cohort", fmt.Sprintf("Fetching cohort member %d of %d.", len(cohortContexts)+1, targetEliteCount), ReportJobProgress{Current: len(cohortContexts) + 1, Total: targetEliteCount}, "", nil)
			memberCtx, memberTimeline, err := s.getOrFetchCohortMemberContext(ctx, candidate, req.Character.Class, req.Character.Spec)
			if err != nil {
				log.Printf("cohort member fetch failed for %s: %v", key, err)
				continue
			}
			entry := buildCohortEntry(candidate)
			cohortEntries = append(cohortEntries, entry)
			cohortContexts = append(cohortContexts, *memberCtx)
			if memberTimeline != nil {
				cohortTimelineData = append(cohortTimelineData, *memberTimeline)
				cohortTimelineEntries = append(cohortTimelineEntries, entry)
			}
			if len(cohortContexts) == targetEliteCount {
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
	if len(cohortContexts) == 0 {
		s.updateJob(jobID, ReportJobFailed, "cohort", "Failed to collect any cohort members.", ReportJobProgress{Current: 0, Total: targetEliteCount}, "no cohort member data could be fetched", nil)
		return
	}
	s.setTimeline(jobID, &reportTimelineData{
		Fight:        req.Fight,
		Character:    req.Character,
		PlayerData:   playerTimelineData,
		EliteData:    cohortTimelineData,
		EliteEntries: cohortTimelineEntries,
	})

	s.updateJob(jobID, ReportJobRunning, "analyzing", "Running deterministic comparison analysis.", ReportJobProgress{Current: 4, Total: 5}, "", nil)
	comparison, err := s.fetchComparisonFromContexts(ctx, req, *playerCtx, cohortContexts)
	if err != nil {
		s.updateJob(jobID, ReportJobFailed, "analyzing", "Failed to compute deterministic comparison metrics.", ReportJobProgress{Current: 4, Total: 5}, err.Error(), nil)
		return
	}

	response := GenerateReportResponse{
		Fight:      req.Fight,
		Character:  req.Character,
		Cohort:     cohortEntries,
		Comparison: comparison,
		Warnings:   buildReportWarnings(req, *playerCtx, cohortContexts),
		AI: AIReportSection{
			Available: false,
			Insights:  []AIInsight{},
		},
	}

	// Hold the result off the job while AI runs — otherwise the frontend
	// polling will see a result without AI populated, render the report
	// immediately, and have the AI section appear seconds later. By leaving
	// Result nil here the progress view stays up until the job completes.
	s.updateJob(jobID, ReportJobRunning, "insights", "Contacting LLM for coaching insights — this can take a few seconds.", ReportJobProgress{Current: 5, Total: 5}, "", nil)
	insights, err := s.fetchInsights(ctx, req, response.Comparison, *playerCtx, cohortContexts)
	if err != nil {
		fmt.Printf("AI insights unavailable for job %s: %v\n", jobID, err)
		response.AI.Warning = "AI insights were unavailable. Deterministic metrics are still shown."
		go s.setCachedResult(req, response)
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

	go s.setCachedResult(req, response)
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

func buildReportWarnings(req GenerateReportRequest, playerCtx analysistypes.MemberContext, cohortContexts []analysistypes.MemberContext) []ReportWarning {
	warnings := make([]ReportWarning, 0, 1)

	playerTalentCode := strings.TrimSpace(req.Character.TalentImportCode)
	if playerTalentCode == "" {
		playerTalentCode = strings.TrimSpace(playerCtx.TalentImportCode)
	}
	if playerTalentCode == "" || len(cohortContexts) == 0 {
		return warnings
	}

	known := 0
	different := 0
	cohortTalentCounts := make(map[string]int)
	for _, cohort := range cohortContexts {
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

func (s *ReportService) fetchComparisonFromContexts(ctx context.Context, req GenerateReportRequest, playerCtx analysistypes.MemberContext, cohortContexts []analysistypes.MemberContext) (ComparisonResult, error) {
	var comparison ComparisonResult
	err := s.postForJSON(
		ctx,
		s.analysisClient,
		s.analysisURL+"/analyze/compare-contexts",
		compareContextsRequest{
			Player:         playerCtx,
			Cohort:         cohortContexts,
			CharacterClass: req.Character.Class,
			CharacterSpec:  req.Character.Spec,
		},
		&comparison,
		"analysis-service",
	)
	return comparison, err
}

func (s *ReportService) fetchInsights(ctx context.Context, req GenerateReportRequest, comparison ComparisonResult, playerCtx analysistypes.MemberContext, cohortContexts []analysistypes.MemberContext) (insightGenerationResponse, error) {
	var insights insightGenerationResponse
	err := s.postForJSON(
		ctx,
		s.aiClient,
		s.aiURL+"/insights/generate",
		buildInsightRequest(req, comparison, playerCtx, cohortContexts),
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

	// Event-heavy payloads run to a few MB; 64MB bounds a runaway response.
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
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

func buildInsightRequest(req GenerateReportRequest, comparison ComparisonResult, playerCtx analysistypes.MemberContext, cohortContexts []analysistypes.MemberContext) insightGenerationRequest {
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
		AbilityHighlights:  buildAbilityHighlights(comparison.AbilityUsage, 5, req.Character.Class, req.Character.Spec, playerCtx, cohortContexts),
		BuffHighlights:     buildBuffHighlights(comparison.BuffUptimes, 5, req.Character.Class, req.Character.Spec, playerCtx, cohortContexts),
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
