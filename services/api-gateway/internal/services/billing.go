package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/stripe/stripe-go/v81"
	checkoutsession "github.com/stripe/stripe-go/v81/checkout/session"
	portalsession "github.com/stripe/stripe-go/v81/billingportal/session"
	"github.com/stripe/stripe-go/v81/webhook"
)

type BillingService struct {
	accounts      *AccountService
	webhookSecret string
	priceID       string
	successURL    string
	cancelURL     string
}

type BillingStatusResponse struct {
	Tier         string              `json:"tier"`
	Subscription *SubscriptionStatus `json:"subscription,omitempty"`
}

type SubscriptionStatus struct {
	Status           string     `json:"status"`
	CurrentPeriodEnd *time.Time `json:"currentPeriodEnd,omitempty"`
}

func NewBillingService(accounts *AccountService, webhookSecret, priceID, successURL, cancelURL string) *BillingService {
	return &BillingService{
		accounts:      accounts,
		webhookSecret: webhookSecret,
		priceID:       priceID,
		successURL:    successURL,
		cancelURL:     cancelURL,
	}
}

// CreateCheckoutSession returns a Stripe-hosted checkout URL for the Pro subscription.
func (s *BillingService) CreateCheckoutSession(ctx context.Context, wclUserID int) (string, error) {
	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(s.priceID),
				Quantity: stripe.Int64(1),
			},
		},
		ClientReferenceID: stripe.String(strconv.Itoa(wclUserID)),
		SuccessURL:        stripe.String(s.successURL + "?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:         stripe.String(s.cancelURL),
	}
	cs, err := checkoutsession.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe checkout session: %w", err)
	}
	return cs.URL, nil
}

// CreatePortalSession returns a Stripe Customer Portal URL so the user can manage or cancel.
func (s *BillingService) CreatePortalSession(ctx context.Context, stripeCustomerID, returnURL string) (string, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(stripeCustomerID),
		ReturnURL: stripe.String(returnURL),
	}
	ps, err := portalsession.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe portal session: %w", err)
	}
	return ps.URL, nil
}

// HandleWebhook verifies the Stripe signature and processes subscription lifecycle events.
func (s *BillingService) HandleWebhook(ctx context.Context, payload []byte, sig string) error {
	event, err := webhook.ConstructEvent(payload, sig, s.webhookSecret)
	if err != nil {
		return fmt.Errorf("webhook signature: %w", err)
	}

	switch event.Type {
	case "checkout.session.completed":
		return s.handleCheckoutCompleted(ctx, event.Data.Raw)
	case "customer.subscription.updated":
		return s.handleSubscriptionUpdated(ctx, event.Data.Raw)
	case "customer.subscription.deleted":
		return s.handleSubscriptionDeleted(ctx, event.Data.Raw)
	}
	return nil
}

// GetStatus reads tier and subscription state directly from the DB (bypasses session cache).
func (s *BillingService) GetStatus(ctx context.Context, wclUserID int) (*BillingStatusResponse, error) {
	account, err := s.accounts.GetByWCLUserID(ctx, wclUserID)
	if err != nil {
		return nil, err
	}
	resp := &BillingStatusResponse{Tier: account.Tier}

	sub, err := s.accounts.GetSubscriptionByAccountID(ctx, account.ID)
	if err != nil {
		return nil, err
	}
	if sub != nil {
		resp.Subscription = &SubscriptionStatus{
			Status:           sub.Status,
			CurrentPeriodEnd: sub.CurrentPeriodEnd,
		}
	}
	return resp, nil
}

func (s *BillingService) handleCheckoutCompleted(ctx context.Context, raw json.RawMessage) error {
	var cs stripe.CheckoutSession
	if err := json.Unmarshal(raw, &cs); err != nil {
		return err
	}

	wclUserID, err := strconv.Atoi(cs.ClientReferenceID)
	if err != nil {
		return fmt.Errorf("invalid client_reference_id %q: %w", cs.ClientReferenceID, err)
	}

	account, err := s.accounts.GetByWCLUserID(ctx, wclUserID)
	if err != nil {
		return fmt.Errorf("account not found for wcl_user_id=%d: %w", wclUserID, err)
	}

	customerID := cs.Customer.ID
	if err := s.accounts.SetStripeCustomerID(ctx, account.ID, customerID); err != nil {
		return err
	}

	subID := cs.Subscription.ID
	if err := s.accounts.UpsertSubscription(ctx, account.ID, customerID, subID, s.priceID, "active", nil); err != nil {
		return err
	}

	return s.accounts.SetTier(ctx, account.ID, TierPro)
}

func (s *BillingService) handleSubscriptionUpdated(ctx context.Context, raw json.RawMessage) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(raw, &sub); err != nil {
		return err
	}
	periodEnd := time.Unix(sub.CurrentPeriodEnd, 0).UTC()
	return s.accounts.UpdateSubscriptionStatus(ctx, sub.ID, string(sub.Status), &periodEnd)
}

func (s *BillingService) handleSubscriptionDeleted(ctx context.Context, raw json.RawMessage) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(raw, &sub); err != nil {
		return err
	}

	account, err := s.accounts.GetByStripeCustomerID(ctx, sub.Customer.ID)
	if err != nil {
		return err
	}

	if err := s.accounts.UpdateSubscriptionStatus(ctx, sub.ID, "canceled", nil); err != nil {
		return err
	}
	return s.accounts.SetTier(ctx, account.ID, TierFree)
}
