package main

import (
	"log"

	"github.com/Pro100x3mal/gophkeeper/internal/client/app"
	"github.com/Pro100x3mal/gophkeeper/internal/client/config"
)

var (
	buildVersion = "dev"
	buildDate    = "unknown"
)

func main() {
	log.Printf("Starting GophKeeper Client %s (%s)", buildVersion, buildDate)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	application, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}

	defer func() {
		cerr := application.Close()
		if cerr != nil {
			log.Printf("failed to save cache: %v", cerr)
		}
	}()
	if err = application.Run(); err != nil {
		log.Fatalf("application failed: %v", err)
	}
}
