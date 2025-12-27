// Package main provides the entry point for the GophKeeper server application.
//
// The server provides a secure password and sensitive data management system
// with encryption, authentication, and item storage capabilities.
package main

import (
	"log"

	"github.com/Pro100x3mal/gophkeeper/internal/server/app"
	"github.com/Pro100x3mal/gophkeeper/internal/server/config"
)

var (
	// buildVersion contains the version of the application, set at build time.
	buildVersion = "dev"
	// buildDate contains the build timestamp, set at build time.
	buildDate = "unknown"
)

func main() {
	log.Printf("Starting GophKeeper Server %s (%s)", buildVersion, buildDate)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	application, err := app.NewApp(cfg, buildVersion, buildDate)
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}

	if err = application.Run(); err != nil {
		log.Fatalf("application failed: %v", err)
	}
}
