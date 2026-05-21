package main

import (
	"log"

	"wow-log-analyzer/services/log-service/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatalf("log-service server failed: %v", err)
	}
}
