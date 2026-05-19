package config

import "os"

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
	RedisAddr          string
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
		RedisAddr:          buildRedisAddr(),
	}
}

func buildRedisAddr() string {
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		return addr
	}
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		return ""
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	return host + ":" + port
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
