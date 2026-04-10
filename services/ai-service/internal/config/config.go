package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Port             string
	Env              string
	Provider         string
	Model            string
	ModelAPIKey      string
	LiveModelEnabled bool
}

func Load() Config {
	loadDotEnv()

	return Config{
		Port:             firstEnv("PORT", "AI_SERVICE_PORT", "8083"),
		Env:              getEnv("ENV", "development"),
		Provider:         getEnv("AI_PROVIDER", "disabled"),
		Model:            firstEnv("AI_MODEL", "OPENAI_MODEL", "fallback-only"),
		ModelAPIKey:      firstEnv("AI_MODEL_API_KEY", "OPENAI_API_KEY", ""),
		LiveModelEnabled: parseBoolEnv("AI_LIVE_MODEL_ENABLED", false),
	}
}

func loadDotEnv() {
	path, err := findDotEnvPath()
	if err != nil || path == "" {
		return
	}

	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if os.Getenv(key) != "" {
			continue
		}

		value = strings.TrimSpace(value)
		value = strings.Trim(value, "\"'")
		_ = os.Setenv(key, value)
	}
}

func findDotEnvPath() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := currentDir
	for {
		candidate := filepath.Join(dir, ".env")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", nil
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

func parseBoolEnv(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	case "":
		return fallback
	default:
		return fallback
	}
}
