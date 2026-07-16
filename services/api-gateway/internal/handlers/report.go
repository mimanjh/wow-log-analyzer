package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"wow-log-analyzer/services/api-gateway/internal/services"
)

type ReportHandler struct {
	reportService  *services.ReportService
	authService    *services.AuthService
	accountService *services.AccountService
}

func NewReportHandler(reportService *services.ReportService, authService *services.AuthService, accountService *services.AccountService) *ReportHandler {
	return &ReportHandler{reportService: reportService, authService: authService, accountService: accountService}
}

func (h *ReportHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req services.GenerateReportRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		log.Printf("Failed to decode report request: %v", err)
		http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	session, ok := h.requireSession(w, r)
	if !ok {
		return
	}

	// Enforce daily usage limit for authenticated users (skip if cached), and
	// record this report against the user's account so it shows up on /reports.
	if session.User != nil {
		if !h.reportService.HasCachedResult(req) {
			limit := services.TierDailyLimit(session.AccountTier)
			allowed, _, usageErr := h.reportService.CheckAndIncrementDailyUsage(r.Context(), session.User.ID, limit)
			if usageErr != nil {
				log.Printf("Usage check failed for user %d: %v", session.User.ID, usageErr)
			} else if !allowed {
				http.Error(w, fmt.Sprintf("Daily analysis limit of %d reached. Try again tomorrow.", limit), http.StatusTooManyRequests)
				return
			}
		}
		if h.accountService != nil {
			if account, accErr := h.accountService.GetByWCLUserID(r.Context(), session.User.ID); accErr == nil {
				recordErr := h.accountService.RecordReport(r.Context(), account.ID, services.UserReport{
					ReportID:       req.ReportID,
					FightID:        req.Fight.ID,
					CharacterID:    req.Character.ID,
					EncounterName:  req.Fight.Name,
					Difficulty:     req.Fight.Difficulty,
					CharacterName:  req.Character.Name,
					CharacterClass: req.Character.Class,
					CharacterSpec:  req.Character.Spec,
				})
				if recordErr != nil {
					log.Printf("Failed to record user report for account %d: %v", account.ID, recordErr)
				}
			}
		}
	}

	owner := services.JobOwner{SessionID: session.ID}
	if session.User != nil {
		owner.UserID = session.User.ID
	}

	response, err := h.reportService.CreateJob(req, owner)
	if err != nil {
		log.Printf("Failed to create report job: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJob(w, response)
}

// requireSession resolves the session cookie, writing a 401 when it is
// missing or expired.
func (h *ReportHandler) requireSession(w http.ResponseWriter, r *http.Request) (*services.SessionState, bool) {
	cookie, err := r.Cookie("wowlog_session")
	if err != nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return nil, false
	}
	session, ok := h.authService.GetSession(cookie.Value)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return nil, false
	}
	return session, true
}

// loadOwnedJob fetches a job and verifies the session owns it. A job that
// exists but belongs to someone else is reported as 404 so job IDs cannot be
// probed.
func (h *ReportHandler) loadOwnedJob(w http.ResponseWriter, r *http.Request, jobID string) (services.ReportJob, *services.SessionState, bool) {
	session, ok := h.requireSession(w, r)
	if !ok {
		return services.ReportJob{}, nil, false
	}

	job, err := h.reportService.GetJob(jobID)
	if err != nil || !job.CanAccess(session) {
		http.Error(w, fmt.Sprintf("report job %s not found", jobID), http.StatusNotFound)
		return services.ReportJob{}, nil, false
	}
	return job, session, true
}

// writeJob encodes a job response with owner fields stripped.
func writeJob(w http.ResponseWriter, job services.ReportJob) {
	job.OwnerUserID = 0
	job.OwnerSessionID = ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(job); err != nil {
		log.Printf("Failed to encode report job response: %v", err)
	}
}

// List returns the authenticated user's saved reports, newest first, each
// flagged with whether the cached result is still warm in Redis. Anonymous
// callers get 401 — the listing is meaningless without an account.
type userReportListEntry struct {
	ReportID       string `json:"reportId"`
	FightID        int    `json:"fightId"`
	CharacterID    int    `json:"characterId"`
	EncounterName  string `json:"encounterName"`
	Difficulty     string `json:"difficulty"`
	CharacterName  string `json:"characterName"`
	CharacterClass string `json:"characterClass"`
	CharacterSpec  string `json:"characterSpec"`
	AnalyzedAt     string `json:"analyzedAt"`
	Cached         bool   `json:"cached"`
}

func (h *ReportHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("wowlog_session")
	if err != nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	session, ok := h.authService.GetSession(cookie.Value)
	if !ok || session.User == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	if h.accountService == nil {
		http.Error(w, "Account storage not configured", http.StatusServiceUnavailable)
		return
	}

	account, err := h.accountService.GetByWCLUserID(r.Context(), session.User.ID)
	if err != nil {
		log.Printf("Account lookup failed for wcl_user_id=%d: %v", session.User.ID, err)
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	reports, err := h.accountService.ListReports(r.Context(), account.ID, 50)
	if err != nil {
		log.Printf("List reports failed for account %d: %v", account.ID, err)
		http.Error(w, "Failed to list reports", http.StatusInternalServerError)
		return
	}

	entries := make([]userReportListEntry, 0, len(reports))
	for _, rep := range reports {
		entries = append(entries, userReportListEntry{
			ReportID:       rep.ReportID,
			FightID:        rep.FightID,
			CharacterID:    rep.CharacterID,
			EncounterName:  rep.EncounterName,
			Difficulty:     rep.Difficulty,
			CharacterName:  rep.CharacterName,
			CharacterClass: rep.CharacterClass,
			CharacterSpec:  rep.CharacterSpec,
			AnalyzedAt:     rep.AnalyzedAt.UTC().Format("2006-01-02T15:04:05Z"),
			Cached:         h.reportService.HasCachedResultForKey(r.Context(), rep.ReportID, rep.FightID, rep.CharacterID),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(entries); err != nil {
		log.Printf("Failed to encode reports list: %v", err)
	}
}

func (h *ReportHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobID := strings.TrimPrefix(r.URL.Path, "/api/report/jobs/")
	if strings.TrimSpace(jobID) == "" {
		http.Error(w, "jobId is required", http.StatusBadRequest)
		return
	}

	job, _, ok := h.loadOwnedJob(w, r, jobID)
	if !ok {
		return
	}

	writeJob(w, job)
}

func (h *ReportHandler) GetAbilityTimeline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobID := strings.TrimPrefix(r.URL.Path, "/api/report/jobs/")
	jobID = strings.TrimSuffix(jobID, "/ability-timeline")
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		http.Error(w, "jobId is required", http.StatusBadRequest)
		return
	}

	abilityID, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("abilityId")))
	if err != nil || abilityID == 0 {
		http.Error(w, "abilityId is required", http.StatusBadRequest)
		return
	}

	if _, _, ok := h.loadOwnedJob(w, r, jobID); !ok {
		return
	}

	response, err := h.reportService.GetAbilityTimeline(jobID, abilityID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode ability timeline response: %v", err)
	}
}

func (h *ReportHandler) GetBuffTimeline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobID := strings.TrimPrefix(r.URL.Path, "/api/report/jobs/")
	jobID = strings.TrimSuffix(jobID, "/buff-timeline")
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		http.Error(w, "jobId is required", http.StatusBadRequest)
		return
	}

	abilityID, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("abilityId")))
	if err != nil || abilityID == 0 {
		http.Error(w, "abilityId is required", http.StatusBadRequest)
		return
	}

	if _, _, ok := h.loadOwnedJob(w, r, jobID); !ok {
		return
	}

	response, err := h.reportService.GetBuffTimeline(jobID, abilityID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode buff timeline response: %v", err)
	}
}

func (h *ReportHandler) GetResourceTimeline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobID := strings.TrimPrefix(r.URL.Path, "/api/report/jobs/")
	jobID = strings.TrimSuffix(jobID, "/resource-timeline")
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		http.Error(w, "jobId is required", http.StatusBadRequest)
		return
	}

	resourceTypeID, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("resourceTypeId")))
	if err != nil {
		http.Error(w, "resourceTypeId is required", http.StatusBadRequest)
		return
	}

	if _, _, ok := h.loadOwnedJob(w, r, jobID); !ok {
		return
	}

	response, err := h.reportService.GetResourceTimeline(jobID, resourceTypeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode resource timeline response: %v", err)
	}
}
