package handlers

import (
	"io"
	"log"
	"net/http"

	"wow-log-analyzer/services/api-gateway/internal/services"
)

type BillingHandler struct {
	authService    *services.AuthService
	accountService *services.AccountService
	billingService *services.BillingService
	returnURL      string
}

func NewBillingHandler(authService *services.AuthService, accountService *services.AccountService, billingService *services.BillingService, returnURL string) *BillingHandler {
	return &BillingHandler{
		authService:    authService,
		accountService: accountService,
		billingService: billingService,
		returnURL:      returnURL,
	}
}

// CreateCheckout POST /api/billing/checkout
// Returns {"url": "https://checkout.stripe.com/..."} for the frontend to redirect to.
func (h *BillingHandler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	session, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if session.User == nil {
		http.Error(w, "user profile not loaded", http.StatusUnauthorized)
		return
	}

	url, err := h.billingService.CreateCheckoutSession(r.Context(), session.User.ID)
	if err != nil {
		log.Printf("CreateCheckout: %v", err)
		http.Error(w, "failed to create checkout session", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}

// GetStatus GET /api/billing/status
// Always reads from DB — used by the frontend after returning from Stripe checkout.
func (h *BillingHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	session, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if session.User == nil {
		http.Error(w, "user profile not loaded", http.StatusUnauthorized)
		return
	}

	status, err := h.billingService.GetStatus(r.Context(), session.User.ID)
	if err != nil {
		log.Printf("GetBillingStatus user=%d: %v", session.User.ID, err)
		http.Error(w, "failed to load billing status", http.StatusInternalServerError)
		return
	}

	// Sync session tier if the DB has a newer value (e.g. after checkout).
	if status.Tier != session.AccountTier {
		h.authService.UpdateSessionTier(session.ID, status.Tier)
	}

	writeJSON(w, http.StatusOK, status)
}

// CreatePortal POST /api/billing/portal
// Returns {"url": "https://billing.stripe.com/..."} for managing/cancelling the subscription.
func (h *BillingHandler) CreatePortal(w http.ResponseWriter, r *http.Request) {
	session, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if session.User == nil {
		http.Error(w, "user profile not loaded", http.StatusUnauthorized)
		return
	}

	account, err := h.accountService.GetByWCLUserID(r.Context(), session.User.ID)
	if err != nil || account.StripeCustomerID == "" {
		http.Error(w, "no active subscription found", http.StatusBadRequest)
		return
	}

	url, err := h.billingService.CreatePortalSession(r.Context(), account.StripeCustomerID, h.returnURL)
	if err != nil {
		log.Printf("CreatePortal user=%d: %v", session.User.ID, err)
		http.Error(w, "failed to create portal session", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}

// Webhook POST /api/billing/webhook
// Receives Stripe events; no session auth — verified by Stripe-Signature header.
func (h *BillingHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	sig := r.Header.Get("Stripe-Signature")
	if err := h.billingService.HandleWebhook(r.Context(), payload, sig); err != nil {
		log.Printf("Webhook error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BillingHandler) requireAuth(w http.ResponseWriter, r *http.Request) (*services.SessionState, bool) {
	cookie, err := r.Cookie("wowlog_session")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return nil, false
	}
	session, ok := h.authService.GetSession(cookie.Value)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return nil, false
	}
	return session, true
}
