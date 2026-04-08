package services

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"wow-log-analyzer/services/api-gateway/internal/config"
)

const sessionCookieName = "wowlog_session"

type AuthStatusResponse struct {
	Authenticated bool      `json:"authenticated"`
	User          *AuthUser `json:"user,omitempty"`
}

type AuthUser struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Avatar    string `json:"avatar,omitempty"`
	BattleTag string `json:"battleTag,omitempty"`
}

type sessionState struct {
	ID          string
	AccessToken string
	ExpiresAt   time.Time
	User        *AuthUser
}

type pendingOAuthState struct {
	ExpiresAt time.Time
}

type AuthService struct {
	cfg           config.Config
	httpClient    *http.Client
	mu            sync.RWMutex
	sessions      map[string]sessionState
	pendingStates map[string]pendingOAuthState
}

func NewAuthService(cfg config.Config) *AuthService {
	return &AuthService{
		cfg:           cfg,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		sessions:      make(map[string]sessionState),
		pendingStates: make(map[string]pendingOAuthState),
	}
}

func (s *AuthService) BuildLoginURL() (string, error) {
	state := randomToken(16)

	s.mu.Lock()
	s.pendingStates[state] = pendingOAuthState{ExpiresAt: time.Now().UTC().Add(10 * time.Minute)}
	s.mu.Unlock()

	query := url.Values{}
	query.Set("client_id", s.cfg.WCLClientID)
	query.Set("redirect_uri", s.cfg.WCLRedirectURL)
	query.Set("response_type", "code")
	query.Set("scope", "view-user-profile")
	query.Set("state", state)

	return s.cfg.WCLAuthorizeURL + "?" + query.Encode(), nil
}

func (s *AuthService) HandleCallback(code, state string) (*sessionState, error) {
	if code == "" || state == "" {
		return nil, fmt.Errorf("code and state are required")
	}

	if !s.consumePendingState(state) {
		return nil, fmt.Errorf("invalid or expired oauth state")
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", s.cfg.WCLRedirectURL)

	req, err := http.NewRequest(http.MethodPost, s.cfg.WCLTokenURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(s.cfg.WCLClientID, s.cfg.WCLClientSecret)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("oauth token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var payload struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.AccessToken == "" {
		return nil, fmt.Errorf("oauth token exchange returned an empty access token")
	}

	session := sessionState{
		ID:          randomToken(24),
		AccessToken: payload.AccessToken,
		ExpiresAt:   time.Now().UTC().Add(time.Duration(payload.ExpiresIn) * time.Second),
	}

	s.mu.Lock()
	s.sessions[session.ID] = session
	s.mu.Unlock()

	return &session, nil
}

func (s *AuthService) GetSession(sessionID string) (*sessionState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, false
	}
	if session.ExpiresAt.Before(time.Now().UTC()) {
		return nil, false
	}

	copy := session
	return &copy, true
}

func (s *AuthService) UpdateSessionUser(sessionID string, user *AuthUser) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return
	}
	session.User = user
	s.sessions[sessionID] = session
}

func (s *AuthService) DeleteSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
}

func (s *AuthService) consumePendingState(state string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	pending, ok := s.pendingStates[state]
	if !ok {
		return false
	}
	delete(s.pendingStates, state)
	return pending.ExpiresAt.After(time.Now().UTC())
}

func randomToken(size int) string {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}
