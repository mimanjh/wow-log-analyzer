package main

import (
	"log"

	"wow-log-analyzer/services/analysis-service/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatalf("analysis-service server failed: %v", err)
	}
}
