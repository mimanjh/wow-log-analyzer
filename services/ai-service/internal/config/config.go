package config

import "os"

type Config struct {
	Port        string
	Env         string
	Provider    string
	Model       string
	ModelAPIKey string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8083"),
		Env:         getEnv("ENV", "development"),
		Provider:    getEnv("AI_PROVIDER", "disabled"),
		Model:       getEnv("AI_MODEL", "fallback-only"),
		ModelAPIKey: getEnv("AI_MODEL_API_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
