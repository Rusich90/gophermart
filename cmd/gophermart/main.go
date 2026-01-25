package main

import (
	"log"

	"github.com/Rusich90/gophermart.git/config"
	"github.com/Rusich90/gophermart.git/internal/http"
)

func main() {
	cfg := config.InitConfig()

	r, db, err := http.SetupServer(cfg)
	if err != nil {
		log.Fatalf("Failed to setup server: %v", err)
	}
	defer db.Close()

	log.Printf("Starting server on %s\n", cfg.RunAddress)
	if err := r.Run(cfg.RunAddress); err != nil {
		log.Fatalf("Server failed to start: %v\n", err)
	}
}
