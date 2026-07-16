package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"

	aiserver "wow-log-analyzer/services/ai-service/server"
	analysisserver "wow-log-analyzer/services/analysis-service/server"
	gatewayserver "wow-log-analyzer/services/api-gateway/server"
	logserver "wow-log-analyzer/services/log-service/server"
)

func main() {
	// Prime env vars from .env BEFORE any service's config.Load() runs.
	// Without this, the goroutine that starts api-gateway first may read its
	// config before ai-service's own dotenv loader fires, silently dropping
	// REDIS_URL / DATABASE_URL / etc.
	loadDotEnv()

	log.Println("wow-log-analyzer: starting all services in one process")

	errs := make(chan error, 4)
	go func() { errs <- logserver.Run() }()
	go func() { errs <- analysisserver.Run() }()
	go func() { errs <- aiserver.Run() }()
	go func() { errs <- gatewayserver.Run() }()

	// First service to exit takes the whole process down — they're all
	// required, so a single failure should fail fast and surface in logs.
	// A nil error means a graceful signal-driven shutdown.
	if err := <-errs; err != nil {
		log.Fatalf("service exited: %v", err)
	}
	log.Println("wow-log-analyzer: shut down cleanly")
}

func loadDotEnv() {
	path := findDotEnv()
	if path == "" {
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
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		_ = os.Setenv(key, value)
	}
}

func findDotEnv() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		candidate := filepath.Join(dir, ".env")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
