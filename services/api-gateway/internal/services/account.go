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

// Migrate lives in migrations.go — it applies the ordered, recorded schema
// history in the package-level migrations slice.

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

// UserReport is one analysis the user has run — metadata only. The full result
// blob lives in Redis at report:result:{reportId}:{fightId}:{characterId} and
// expires on the standard 30-day cache TTL.
type UserReport struct {
	ReportID       string
	FightID        int
	CharacterID    int
	EncounterName  string
	Difficulty     string
	CharacterName  string
	CharacterClass string
	CharacterSpec  string
	AnalyzedAt     time.Time
}

// RecordReport upserts a row in user_reports keyed by (account, report, fight,
// character). Repeat analyses of the same report bump analyzed_at so the list
// stays sorted by most-recent.
func (s *AccountService) RecordReport(ctx context.Context, accountID int, report UserReport) error {
	if s.db == nil {
		return fmt.Errorf("database not configured")
	}
	_, err := s.db.Exec(ctx, `
		INSERT INTO user_reports (
			account_id, report_id, fight_id, character_id,
			encounter_name, difficulty, character_name, character_class, character_spec, analyzed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (account_id, report_id, fight_id, character_id) DO UPDATE
			SET encounter_name  = EXCLUDED.encounter_name,
			    difficulty      = EXCLUDED.difficulty,
			    character_name  = EXCLUDED.character_name,
			    character_class = EXCLUDED.character_class,
			    character_spec  = EXCLUDED.character_spec,
			    analyzed_at     = NOW()
	`, accountID, report.ReportID, report.FightID, report.CharacterID,
		report.EncounterName, report.Difficulty, report.CharacterName, report.CharacterClass, report.CharacterSpec)
	return err
}

// ListReports returns the user's saved reports, newest first. Capped at `limit`
// (server-side default 50) — no cursor yet since the MVP fits on one screen.
func (s *AccountService) ListReports(ctx context.Context, accountID, limit int) ([]UserReport, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not configured")
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	const q = `
		SELECT report_id, fight_id, character_id, encounter_name, difficulty,
		       character_name, character_class, character_spec, analyzed_at
		FROM user_reports
		WHERE account_id = $1
		ORDER BY analyzed_at DESC
		LIMIT $2
	`
	rows, err := s.db.Query(ctx, q, accountID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reports := make([]UserReport, 0)
	for rows.Next() {
		var r UserReport
		if err := rows.Scan(
			&r.ReportID, &r.FightID, &r.CharacterID, &r.EncounterName, &r.Difficulty,
			&r.CharacterName, &r.CharacterClass, &r.CharacterSpec, &r.AnalyzedAt,
		); err != nil {
			return nil, err
		}
		reports = append(reports, r)
	}
	return reports, rows.Err()
}

// TierDailyLimit returns the analyses-per-day allowed for a given tier.
func TierDailyLimit(tier string) int {
	if tier == TierPro {
		return 50
	}
	return DailyAnalysisLimit // free tier — defined in report_job.go
}
