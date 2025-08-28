package main

import (
	"context"
	"errors"
	"github.com/csherida/api-key-manager-service/internal/service/di"
	"log"
	"os"
)

func main() {
	application, err := di.SetupApplication()
	if err != nil {
		log.Fatalf("failed to setup application: %v", err)
	}

	exitCode := 0
	defer func() { os.Exit(exitCode) }()
	//defer func() { application.ShutdownAndCleanup() }()
	//defer cleanupDependencies()

	if err := application.Run(); err != nil {
		if !errors.Is(err, context.Canceled) {
			exitCode = 1
		}
	}
}
