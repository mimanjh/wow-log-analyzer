package handlers

import (
	"encoding/json"
	"net/http"

	"wow-log-analyzer/services/api-gateway/internal/config"
	"wow-log-analyzer/services/api-gateway/internal/services"
)

type AuthHandler struct {
	authService    *services.AuthService
	browserService *services.BrowserService
	accountService *services.AccountService
	cfg            config.Config
}

func NewAuthHandler(authService *services.AuthService, browserService *services.BrowserService, accountService *services.AccountService, cfg config.Config) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		browserService: browserService,
		accountService: accountService,
		cfg:            cfg,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	loginURL, err := h.authService.BuildLoginURL()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, loginURL, http.StatusFound)
}

func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	session, err := h.authService.HandleCallback(code, state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.browserService.GetCurrentUser(session.AccessToken)
	if err == nil {
		h.authService.UpdateSessionUser(session.ID, user)

		account, accountErr := h.accountService.GetOrCreate(r.Context(), user.ID, user.Name, user.BattleTag)
		if accountErr == nil {
			h.authService.UpdateSessionTier(session.ID, account.Tier)
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "wowlog_session",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, h.cfg.FrontendURL+"/analyze", http.StatusFound)
}

func (h *AuthHandler) Status(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("wowlog_session")
	if err != nil {
		writeJSON(w, http.StatusOK, services.AuthStatusResponse{Authenticated: false})
		return
	}
	session, ok := h.authService.GetSession(cookie.Value)
	if !ok {
		writeJSON(w, http.StatusOK, services.AuthStatusResponse{Authenticated: false})
		return
	}

	user := session.User
	if user == nil {
		resolved, err := h.browserService.GetCurrentUser(session.AccessToken)
		if err == nil {
			h.authService.UpdateSessionUser(session.ID, resolved)
			user = resolved
		}
	}

	tier := session.AccountTier
	if tier == "" && user != nil {
		account, err := h.accountService.GetOrCreate(r.Context(), user.ID, user.Name, user.BattleTag)
		if err == nil {
			h.authService.UpdateSessionTier(session.ID, account.Tier)
			tier = account.Tier
		}
	}

	writeJSON(w, http.StatusOK, services.AuthStatusResponse{
		Authenticated: true,
		User:          user,
		Tier:          tier,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("wowlog_session"); err == nil {
		h.authService.DeleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "wowlog_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
	})

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
