package main

import (
	"log"

	"github.com/Rusich90/gophermart.git/config"
	"github.com/Rusich90/gophermart.git/internal/app"
)

func main() {
	cfg := config.InitConfig()

	application, err := app.InitializeApplication(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer application.DB.Close()

	if err := application.Run(); err != nil {
		log.Fatalf("Server failed to start: %v\n", err)
	}
}
