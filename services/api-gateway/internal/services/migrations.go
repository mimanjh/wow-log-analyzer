package services

import (
	"context"
	"fmt"
	"log"
)

// migrationLockKey is an arbitrary advisory-lock key that serializes
// concurrent instances running migrations against the same database.
const migrationLockKey = 727215001

// migrations is the append-only, ordered schema history. A migration's
// version is its 1-based position in this slice and is recorded in
// schema_migrations once applied. NEVER edit or reorder an entry that has
// shipped — add a new entry instead.
var migrations = []struct {
	name  string
	stmts []string
}{
	{
		name: "initial_schema",
		stmts: []string{
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
			`CREATE TABLE IF NOT EXISTS user_reports (
				id              SERIAL      PRIMARY KEY,
				account_id      INTEGER     NOT NULL REFERENCES accounts(id),
				report_id       TEXT        NOT NULL,
				fight_id        INTEGER     NOT NULL,
				character_id    INTEGER     NOT NULL,
				encounter_name  TEXT        NOT NULL DEFAULT '',
				difficulty      TEXT        NOT NULL DEFAULT '',
				character_name  TEXT        NOT NULL DEFAULT '',
				character_class TEXT        NOT NULL DEFAULT '',
				character_spec  TEXT        NOT NULL DEFAULT '',
				analyzed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
			)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS idx_user_reports_unique
				ON user_reports(account_id, report_id, fight_id, character_id)`,
			`CREATE INDEX IF NOT EXISTS idx_user_reports_account_recent
				ON user_reports(account_id, analyzed_at DESC)`,
		},
	},
}

// Migrate applies any unapplied migrations in order, recording each in
// schema_migrations. Each migration runs in its own transaction under an
// advisory lock, so concurrent instances can't double-apply.
func (s *AccountService) Migrate(ctx context.Context) error {
	if _, err := s.db.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER     PRIMARY KEY,
		name       TEXT        NOT NULL,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for i, migration := range migrations {
		version := i + 1
		if err := s.applyMigration(ctx, version, migration.name, migration.stmts); err != nil {
			return fmt.Errorf("migration %d (%s): %w", version, migration.name, err)
		}
	}
	return nil
}

func (s *AccountService) applyMigration(ctx context.Context, version int, name string, stmts []string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, migrationLockKey); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}

	var applied bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&applied); err != nil {
		return fmt.Errorf("check applied state: %w", err)
	}
	if applied {
		return nil
	}

	for _, stmt := range stmts {
		if _, err := tx.Exec(ctx, stmt); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`, version, name); err != nil {
		return fmt.Errorf("record migration: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	log.Printf("applied database migration %d (%s)", version, name)
	return nil
}
