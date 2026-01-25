package app

import (
	"fmt"

	"github.com/Rusich90/gophermart.git/config"
)

func InitializeApplication(cfg *config.Config) (*Application, error) {
	app, err := Initialize(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize application: %w", err)
	}

	return app, nil
}
