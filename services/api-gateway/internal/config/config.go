package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port               string
	Env                string
	LogServiceURL      string
	AnalysisServiceURL string
	AIServiceURL       string
	FrontendURL        string
	WCLClientID        string
	WCLClientSecret    string
	WCLAuthorizeURL    string
	WCLTokenURL        string
	WCLRedirectURL     string
	RedisURL             string
	DatabaseURL          string
	StripeSecretKey      string
	StripeWebhookSecret  string
	StripeProPriceID     string
}

func Load() Config {
	return Config{
		Port:               firstEnv("PORT", "API_GATEWAY_PORT", "8080"),
		Env:                getEnv("ENV", "development"),
		LogServiceURL:      getEnv("LOG_SERVICE_URL", "http://localhost:8081"),
		AnalysisServiceURL: getEnv("ANALYSIS_SERVICE_URL", "http://localhost:8082"),
		AIServiceURL:       getEnv("AI_SERVICE_URL", "http://localhost:8083"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:5173"),
		WCLClientID:        firstEnv("WCL_CLIENT_ID", "WARCRAFTLOGS_CLIENT_ID", ""),
		WCLClientSecret:    firstEnv("WCL_CLIENT_SECRET", "WARCRAFTLOGS_CLIENT_SECRET", ""),
		WCLAuthorizeURL:    getEnv("WCL_AUTHORIZE_URL", "https://www.warcraftlogs.com/oauth/authorize"),
		WCLTokenURL:        getEnv("WCL_TOKEN_URL", "https://www.warcraftlogs.com/oauth/token"),
		WCLRedirectURL:     getEnv("WCL_REDIRECT_URL", "http://localhost:8080/api/auth/callback"),
		RedisURL:            getEnv("REDIS_URL", ""),
		DatabaseURL:         buildDatabaseURL(),
		StripeSecretKey:     getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripeProPriceID:    getEnv("STRIPE_PRO_PRICE_ID", ""),
	}
}

func buildDatabaseURL() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		return ""
	}
	port := getEnv("POSTGRES_PORT", "5432")
	user := getEnv("POSTGRES_USER", "postgres")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := getEnv("POSTGRES_DB", "wowloganalyzer")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbname)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func firstEnv(primary, secondary, fallback string) string {
	if value := os.Getenv(primary); value != "" {
		return value
	}
	if value := os.Getenv(secondary); value != "" {
		return value
	}
	return fallback
}
