package services

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	TierFree = "free"
	TierPro  = "pro"
)

type Account struct {
	ID               int
	WCLUserID        int
	Name             string
	BattleTag        string
	Tier             string
	StripeCustomerID string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Subscription struct {
	ID                   int
	AccountID            int
	StripeCustomerID     string
	StripeSubscriptionID string
	StripePriceID        string
	Status               string
	CurrentPeriodEnd     *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type AccountService struct {
	db *pgxpool.Pool
}

func NewAccountService(db *pgxpool.Pool) *AccountService {
	return &AccountService{db: db}
}

// Migrate creates or evolves the accounts and subscriptions schema. Idempotent.
func (s *AccountService) Migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS accounts (
			id                 SERIAL       PRIMARY KEY,
			wcl_user_id        INTEGER      NOT NULL UNIQUE,
			name               TEXT         NOT NULL,
			battle_tag         TEXT         NOT NULL DEFAULT '',
			tier               TEXT         NOT NULL DEFAULT 'free',
			stripe_customer_id TEXT,
			created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			updated_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		)`,
		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS stripe_customer_id TEXT`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_stripe_customer_id
			ON accounts(stripe_customer_id) WHERE stripe_customer_id IS NOT NULL`,
		`CREATE TABLE IF NOT EXISTS subscriptions (
			id                     SERIAL       PRIMARY KEY,
			account_id             INTEGER      NOT NULL REFERENCES accounts(id),
			stripe_customer_id     TEXT         NOT NULL,
			stripe_subscription_id TEXT         NOT NULL UNIQUE,
			stripe_price_id        TEXT         NOT NULL DEFAULT '',
			status                 TEXT         NOT NULL DEFAULT 'inactive',
			current_period_end     TIMESTAMPTZ,
			created_at             TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			updated_at             TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_subscriptions_account_id ON subscriptions(account_id)`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}
	return nil
}

// GetOrCreate upserts the account on every login; tier and stripe fields are never overwritten.
func (s *AccountService) GetOrCreate(ctx context.Context, wclUserID int, name, battleTag string) (*Account, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not configured")
	}
	const q = `
		INSERT INTO accounts (wcl_user_id, name, battle_tag)
		VALUES ($1, $2, $3)
		ON CONFLICT (wcl_user_id) DO UPDATE
			SET name       = EXCLUDED.name,
			    battle_tag = CASE WHEN EXCLUDED.battle_tag = '' THEN accounts.battle_tag ELSE EXCLUDED.battle_tag END,
			    updated_at = NOW()
		RETURNING id, wcl_user_id, name, battle_tag, tier, COALESCE(stripe_customer_id, ''), created_at, updated_at
	`
	var a Account
	err := s.db.QueryRow(ctx, q, wclUserID, name, battleTag).Scan(
		&a.ID, &a.WCLUserID, &a.Name, &a.BattleTag, &a.Tier, &a.StripeCustomerID, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("account upsert for wcl_user_id=%d: %w", wclUserID, err)
	}
	return &a, nil
}

func (s *AccountService) GetByWCLUserID(ctx context.Context, wclUserID int) (*Account, error) {
	const q = `
		SELECT id, wcl_user_id, name, battle_tag, tier, COALESCE(stripe_customer_id, ''), created_at, updated_at
		FROM accounts WHERE wcl_user_id = $1
	`
	var a Account
	err := s.db.QueryRow(ctx, q, wclUserID).Scan(
		&a.ID, &a.WCLUserID, &a.Name, &a.BattleTag, &a.Tier, &a.StripeCustomerID, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("account lookup wcl_user_id=%d: %w", wclUserID, err)
	}
	return &a, nil
}

func (s *AccountService) GetByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*Account, error) {
	const q = `
		SELECT id, wcl_user_id, name, battle_tag, tier, COALESCE(stripe_customer_id, ''), created_at, updated_at
		FROM accounts WHERE stripe_customer_id = $1
	`
	var a Account
	err := s.db.QueryRow(ctx, q, stripeCustomerID).Scan(
		&a.ID, &a.WCLUserID, &a.Name, &a.BattleTag, &a.Tier, &a.StripeCustomerID, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("account lookup stripe_customer_id=%s: %w", stripeCustomerID, err)
	}
	return &a, nil
}

func (s *AccountService) SetStripeCustomerID(ctx context.Context, accountID int, stripeCustomerID string) error {
	_, err := s.db.Exec(ctx,
		`UPDATE accounts SET stripe_customer_id = $1, updated_at = NOW() WHERE id = $2`,
		stripeCustomerID, accountID,
	)
	return err
}

func (s *AccountService) SetTier(ctx context.Context, accountID int, tier string) error {
	_, err := s.db.Exec(ctx,
		`UPDATE accounts SET tier = $1, updated_at = NOW() WHERE id = $2`,
		tier, accountID,
	)
	return err
}

func (s *AccountService) UpsertSubscription(ctx context.Context, accountID int, stripeCustomerID, subscriptionID, priceID, status string, periodEnd *time.Time) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO subscriptions (account_id, stripe_customer_id, stripe_subscription_id, stripe_price_id, status, current_period_end)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (stripe_subscription_id) DO UPDATE
			SET status             = EXCLUDED.status,
			    stripe_price_id    = EXCLUDED.stripe_price_id,
			    current_period_end = EXCLUDED.current_period_end,
			    updated_at         = NOW()
	`, accountID, stripeCustomerID, subscriptionID, priceID, status, periodEnd)
	return err
}

func (s *AccountService) UpdateSubscriptionStatus(ctx context.Context, subscriptionID, status string, periodEnd *time.Time) error {
	_, err := s.db.Exec(ctx,
		`UPDATE subscriptions SET status = $1, current_period_end = $2, updated_at = NOW()
		 WHERE stripe_subscription_id = $3`,
		status, periodEnd, subscriptionID,
	)
	return err
}

func (s *AccountService) GetSubscriptionByAccountID(ctx context.Context, accountID int) (*Subscription, error) {
	const q = `
		SELECT id, account_id, stripe_customer_id, stripe_subscription_id, stripe_price_id,
		       status, current_period_end, created_at, updated_at
		FROM subscriptions WHERE account_id = $1
		ORDER BY created_at DESC LIMIT 1
	`
	var sub Subscription
	err := s.db.QueryRow(ctx, q, accountID).Scan(
		&sub.ID, &sub.AccountID, &sub.StripeCustomerID, &sub.StripeSubscriptionID,
		&sub.StripePriceID, &sub.Status, &sub.CurrentPeriodEnd, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// TierDailyLimit returns the analyses-per-day allowed for a given tier.
func TierDailyLimit(tier string) int {
	if tier == TierPro {
		return 50
	}
	return DailyAnalysisLimit // free tier — defined in report_job.go
}
