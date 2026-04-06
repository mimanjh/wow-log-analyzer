package config

import "os"

type Config struct {
	Port string
	Env  string
	WCL  WCLConfig
}

type WCLConfig struct {
	ClientID     string
	ClientSecret string
	BaseURL      string
}

func Load() Config {
	return Config{
		Port: firstEnv("PORT", "LOG_SERVICE_PORT", "8081"),
		Env:  getEnv("ENV", "development"),
		WCL: WCLConfig{
			ClientID:     firstEnv("WCL_CLIENT_ID", "WARCRAFTLOGS_CLIENT_ID", ""),
			ClientSecret: firstEnv("WCL_CLIENT_SECRET", "WARCRAFTLOGS_CLIENT_SECRET", ""),
			BaseURL:      firstEnv("WCL_BASE_URL", "WARCRAFTLOGS_BASE_URL", "https://www.warcraftlogs.com/api/v2"),
		},
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
