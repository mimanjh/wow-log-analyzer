package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	jobRedisKeyPrefix = "report:job:"
	jobRedisTTL       = 30 * 24 * time.Hour

	resultCacheKeyPrefix              = "report:result:"
	resultCacheTTL                    = 30 * 24 * time.Hour
	cohortMemberContextCacheKeyPrefix = "cohort:member-ctx:"
	cohortMemberContextCacheTTL       = 30 * 24 * time.Hour

	usageRedisKeyPrefix = "usage:daily:"
	DailyAnalysisLimit  = 5
)

type ReportService struct {
	logClient      *http.Client
	analysisClient *http.Client
	aiClient       *http.Client
	logURL         string
	analysisURL    string
	aiURL          string

	jobMu       sync.RWMutex
	jobs        map[string]ReportJob
	redisClient *redis.Client
	keyPrefix   string
}

const (
	targetEliteCount         = 10
	rankingCandidateBatchCap = 25
	rankingCandidateMaxCap   = 100
)

func NewReportService(logURL, analysisURL, aiURL string, redisClient *redis.Client, keyPrefix string) *ReportService {
	const serviceTimeout = 120 * time.Second

	return &ReportService{
		logClient:      &http.Client{Timeout: serviceTimeout},
		analysisClient: &http.Client{Timeout: serviceTimeout},
		aiClient:       &http.Client{Timeout: serviceTimeout},
		logURL:         logURL,
		analysisURL:    analysisURL,
		aiURL:          aiURL,
		jobs:           make(map[string]ReportJob),
		redisClient:    redisClient,
		keyPrefix:      keyPrefix,
	}
}

func (s *ReportService) key(suffix string) string { return s.keyPrefix + suffix }

func (s *ReportService) CreateJob(req GenerateReportRequest, owner JobOwner) (ReportJob, error) {
	if req.ReportID == "" {
		return ReportJob{}, fmt.Errorf("reportId is required")
	}
	if req.Fight.ID == 0 || req.Character.ID == 0 {
		return ReportJob{}, fmt.Errorf("fight and character selections are required")
	}

	jobID, err := newJobID()
	if err != nil {
		return ReportJob{}, err
	}

	if cached, ok := s.getCachedResult(req); ok {
		job := ReportJob{
			ID:             jobID,
			Status:         ReportJobCompleted,
			Stage:          "completed",
			Message:        "Report loaded from cache.",
			Fight:          req.Fight,
			Character:      req.Character,
			Progress:       ReportJobProgress{Current: 5, Total: 5},
			Result:         &cached,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
			OwnerUserID:    owner.UserID,
			OwnerSessionID: owner.SessionID,
		}
		s.setJob(job)
		return job, nil
	}

	job := ReportJob{
		ID:        jobID,
		Status:    ReportJobQueued,
		Stage:     "queued",
		Message:   "Queued for report generation.",
		Fight:     req.Fight,
		Character: req.Character,
		Progress: ReportJobProgress{
			Current: 0,
			Total:   5,
		},
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		OwnerUserID:    owner.UserID,
		OwnerSessionID: owner.SessionID,
	}

	s.setJob(job)

	go s.runJob(job.ID, req)

	return job, nil
}

func (s *ReportService) GetJob(jobID string) (ReportJob, error) {
	s.jobMu.RLock()
	job, ok := s.jobs[jobID]
	s.jobMu.RUnlock()
	if ok {
		return job, nil
	}

	if s.redisClient != nil {
		data, err := s.redisClient.Get(context.Background(), s.key(jobRedisKeyPrefix+jobID)).Bytes()
		if err == nil {
			var cached ReportJob
			if err := json.Unmarshal(data, &cached); err == nil {
				return cached, nil
			}
		}
	}

	return ReportJob{}, fmt.Errorf("report job %s not found", jobID)
}

func (s *ReportService) persistJobToRedis(job ReportJob) {
	if s.redisClient == nil {
		return
	}
	data, err := json.Marshal(job)
	if err != nil {
		log.Printf("Failed to marshal job %s for Redis: %v", job.ID, err)
		return
	}
	if err := s.redisClient.Set(context.Background(), s.key(jobRedisKeyPrefix+job.ID), data, jobRedisTTL).Err(); err != nil {
		log.Printf("Failed to persist job %s to Redis: %v", job.ID, err)
	}
}

func resultCacheKey(reportID string, fightID, characterID int) string {
	return fmt.Sprintf("%s%s:%d:%d", resultCacheKeyPrefix, reportID, fightID, characterID)
}

func (s *ReportService) getCachedResult(req GenerateReportRequest) (GenerateReportResponse, bool) {
	if s.redisClient == nil {
		return GenerateReportResponse{}, false
	}
	data, err := s.redisClient.Get(context.Background(), s.key(resultCacheKey(req.ReportID, req.Fight.ID, req.Character.ID))).Bytes()
	if err != nil {
		return GenerateReportResponse{}, false
	}
	var result GenerateReportResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return GenerateReportResponse{}, false
	}
	return result, true
}

func (s *ReportService) setCachedResult(req GenerateReportRequest, result GenerateReportResponse) {
	if s.redisClient == nil {
		return
	}
	data, err := json.Marshal(result)
	if err != nil {
		log.Printf("Failed to marshal result for caching: %v", err)
		return
	}
	if err := s.redisClient.Set(context.Background(), s.key(resultCacheKey(req.ReportID, req.Fight.ID, req.Character.ID)), data, resultCacheTTL).Err(); err != nil {
		log.Printf("Failed to cache result for %s fight=%d char=%d: %v", req.ReportID, req.Fight.ID, req.Character.ID, err)
	}
}

func (s *ReportService) HasCachedResult(req GenerateReportRequest) bool {
	_, ok := s.getCachedResult(req)
	return ok
}

// HasCachedResultForKey returns true if a cached report result exists for the
// (reportID, fightID, characterID) tuple. Cheaper than HasCachedResult — uses
// EXISTS instead of GET + Unmarshal.
func (s *ReportService) HasCachedResultForKey(ctx context.Context, reportID string, fightID, characterID int) bool {
	if s.redisClient == nil {
		return false
	}
	n, err := s.redisClient.Exists(ctx, s.key(resultCacheKey(reportID, fightID, characterID))).Result()
	return err == nil && n > 0
}

func (s *ReportService) CheckAndIncrementDailyUsage(ctx context.Context, userID, limit int) (allowed bool, used int, err error) {
	if s.redisClient == nil {
		return true, 0, nil
	}
	today := time.Now().UTC().Format("2006-01-02")
	key := s.key(fmt.Sprintf("%s%d:%s", usageRedisKeyPrefix, userID, today))

	count, err := s.redisClient.Incr(ctx, key).Result()
	if err != nil {
		return true, 0, err
	}
	if count == 1 {
		s.redisClient.Expire(ctx, key, 48*time.Hour)
	}
	return count <= int64(limit), int(count), nil
}

func (s *ReportService) getCachedRaw(ctx context.Context, key string) (json.RawMessage, bool) {
	if s.redisClient == nil {
		return nil, false
	}
	data, err := s.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}
	return json.RawMessage(data), true
}

func (s *ReportService) setCachedRaw(ctx context.Context, key string, data json.RawMessage, ttl time.Duration) {
	if s.redisClient == nil {
		return
	}
	if err := s.redisClient.Set(ctx, key, []byte(data), ttl).Err(); err != nil {
		log.Printf("Failed to cache raw data at %s: %v", key, err)
	}
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
	job, ok := s.jobs[jobID]
	if !ok {
		s.jobMu.Unlock()
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
	s.jobMu.Unlock()

	if status == ReportJobCompleted || status == ReportJobFailed {
		go s.persistJobToRedis(job)
	}
}

func newJobID() (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate job id: %w", err)
	}
	return hex.EncodeToString(randomBytes), nil
}
