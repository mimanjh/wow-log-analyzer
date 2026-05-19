package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type ReportService struct {
	logClient      *http.Client
	analysisClient *http.Client
	aiClient       *http.Client
	logURL         string
	analysisURL    string
	aiURL          string

	jobMu sync.RWMutex
	jobs  map[string]ReportJob
}

const (
	targetEliteCount         = 10
	rankingCandidateBatchCap = 25
	rankingCandidateMaxCap   = 100
)

func NewReportService(logURL, analysisURL, aiURL string) *ReportService {
	const serviceTimeout = 120 * time.Second

	return &ReportService{
		logClient:      &http.Client{Timeout: serviceTimeout},
		analysisClient: &http.Client{Timeout: serviceTimeout},
		aiClient:       &http.Client{Timeout: serviceTimeout},
		logURL:         logURL,
		analysisURL:    analysisURL,
		aiURL:          aiURL,
		jobs:           make(map[string]ReportJob),
	}
}

func (s *ReportService) CreateJob(req GenerateReportRequest) (ReportJob, error) {
	if req.ReportID == "" {
		return ReportJob{}, fmt.Errorf("reportId is required")
	}
	if req.Fight.ID == 0 || req.Character.ID == 0 {
		return ReportJob{}, fmt.Errorf("fight and character selections are required")
	}

	job := ReportJob{
		ID:        newJobID(),
		Status:    ReportJobQueued,
		Stage:     "queued",
		Message:   "Queued for report generation.",
		Fight:     req.Fight,
		Character: req.Character,
		Progress: ReportJobProgress{
			Current: 0,
			Total:   5,
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	s.setJob(job)

	go s.runJob(job.ID, req)

	return job, nil
}

func (s *ReportService) GetJob(jobID string) (ReportJob, error) {
	s.jobMu.RLock()
	defer s.jobMu.RUnlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return ReportJob{}, fmt.Errorf("report job %s not found", jobID)
	}

	return job, nil
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
	defer s.jobMu.Unlock()

	job, ok := s.jobs[jobID]
	if !ok {
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
}

func newJobID() string {
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return fmt.Sprintf("job-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(randomBytes)
}
