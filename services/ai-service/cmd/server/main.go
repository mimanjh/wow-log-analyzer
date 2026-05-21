package main

import (
	"log"

	"wow-log-analyzer/services/ai-service/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatalf("ai-service server failed: %v", err)
	}
}
