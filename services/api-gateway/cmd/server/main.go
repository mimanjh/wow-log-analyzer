package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stripe/stripe-go/v81"
	"wow-log-analyzer/services/api-gateway/internal/config"
	"wow-log-analyzer/services/api-gateway/internal/handlers"
	"wow-log-analyzer/services/api-gateway/internal/services"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()
	mux := http.NewServeMux()

	redisClient, redisDesc, err := buildRedisClient(cfg)
	if err != nil {
		log.Fatalf("Failed to configure Redis: %v", err)
	}
	if redisClient != nil {
		if pingErr := redisClient.Ping(ctx).Err(); pingErr != nil {
			log.Printf("Redis ping failed (%s): %v — continuing without cache", redisDesc, pingErr)
			redisClient = nil
		} else {
			log.Printf("Redis connected (%s)", redisDesc)
		}
	}

	var accountService *services.AccountService
	if cfg.DatabaseURL != "" {
		db, err := pgxpool.New(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("Failed to connect to PostgreSQL: %v", err)
		}
		accountService = services.NewAccountService(db)
		if err := accountService.Migrate(ctx); err != nil {
			log.Fatalf("Failed to run account migrations: %v", err)
		}
		log.Printf("PostgreSQL connected and migrated")
	}

	if cfg.StripeSecretKey != "" {
		stripe.Key = cfg.StripeSecretKey
	}

	analyzeService := services.NewAnalyzeService(cfg.LogServiceURL, redisClient)
	reportService := services.NewReportService(cfg.LogServiceURL, cfg.AnalysisServiceURL, cfg.AIServiceURL, redisClient)
	authService := services.NewAuthService(cfg, redisClient)
	browserService := services.NewBrowserService(cfg.LogServiceURL, redisClient)
	billingService := services.NewBillingService(
		accountService,
		cfg.StripeWebhookSecret,
		cfg.StripeProPriceID,
		cfg.FrontendURL+"/billing/success",
		cfg.FrontendURL+"/analyze",
	)
	handlers.RegisterRoutes(mux, cfg, analyzeService, reportService, authService, browserService, accountService, billingService)

	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("api-gateway starting on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalf("api-gateway server failed: %v", err)
	}
}

// buildRedisClient parses REDIS_URL (redis:// or rediss://) so the same
// config drives local Docker and Redis Cloud. Empty URL = caching disabled.
func buildRedisClient(cfg config.Config) (*redis.Client, string, error) {
	if cfg.RedisURL == "" {
		return nil, "", nil
	}
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, "", fmt.Errorf("parse REDIS_URL: %w", err)
	}
	return redis.NewClient(opts), opts.Addr, nil
}
