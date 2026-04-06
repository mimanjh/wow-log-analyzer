package config

import "os"

type Config struct {
	Port    string
	Env     string
	WCL     WCLConfig
}

type WCLConfig struct {
	ClientID     string
	ClientSecret string
	BaseURL      string
}

func Load() Config {
	return Config{
		Port: getEnv("PORT", "8081"),
		Env:  getEnv("ENV", "development"),
		WCL: WCLConfig{
			ClientID:     getEnv("WCL_CLIENT_ID", ""),
			ClientSecret: getEnv("WCL_CLIENT_SECRET", ""),
			BaseURL:      getEnv("WCL_BASE_URL", "https://www.warcraftlogs.com/api/v2"),
		},
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
