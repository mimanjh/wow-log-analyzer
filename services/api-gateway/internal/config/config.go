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
		Port:               getEnv("PORT", "8080"),
		Env:                getEnv("ENV", "development"),
		LogServiceURL:      getEnv("LOG_SERVICE_URL", "http://log-service:8081"),
		AnalysisServiceURL: getEnv("ANALYSIS_SERVICE_URL", "http://analysis-service:8082"),
		AIServiceURL:       getEnv("AI_SERVICE_URL", "http://ai-service:8083"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
