package config

import "os"

type Config struct {
	Port               string
	Env                string
	LogServiceURL      string
	AnalysisServiceURL string
	AIServiceURL       string
}

func Load() Config {
	return Config{
		Port:               firstEnv("PORT", "API_GATEWAY_PORT", "8080"),
		Env:                getEnv("ENV", "development"),
		LogServiceURL:      getEnv("LOG_SERVICE_URL", "http://localhost:8081"),
		AnalysisServiceURL: getEnv("ANALYSIS_SERVICE_URL", "http://localhost:8082"),
		AIServiceURL:       getEnv("AI_SERVICE_URL", "http://localhost:8083"),
	}
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
