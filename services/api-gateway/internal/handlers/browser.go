package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"wow-log-analyzer/services/api-gateway/internal/services"
)

type BrowserHandler struct {
	authService    *services.AuthService
	browserService *services.BrowserService
}

func NewBrowserHandler(authService *services.AuthService, browserService *services.BrowserService) *BrowserHandler {
	return &BrowserHandler{
		authService:    authService,
		browserService: browserService,
	}
}

func (h *BrowserHandler) GetCharacters(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("wowlog_session")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	session, ok := h.authService.GetSession(cookie.Value)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	characters, err := h.browserService.GetCharacters(session.AccessToken, cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	writeJSON(w, http.StatusOK, characters)
}

func (h *BrowserHandler) GetCharacterReports(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("wowlog_session")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	session, ok := h.authService.GetSession(cookie.Value)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/browser/characters/")
	path = strings.TrimSuffix(path, "/reports")
	characterID, err := strconv.Atoi(path)
	if err != nil || characterID <= 0 {
		http.Error(w, "invalid character id", http.StatusBadRequest)
		return
	}

	page, err := h.browserService.GetCharacterReports(
		session.AccessToken,
		cookie.Value,
		characterID,
		r.URL.Query().Get("cursor"),
		parseBrowserLimit(r.URL.Query().Get("limit"), 10),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	writeJSON(w, http.StatusOK, page)
}

func parseBrowserLimit(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
