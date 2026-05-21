package main

import (
	"log"

	"wow-log-analyzer/services/api-gateway/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatalf("api-gateway server failed: %v", err)
	}
}
