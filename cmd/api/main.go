package main

import (
	"context"
	"ez2boot/internal/app"
	"ez2boot/internal/config"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	var version = "dev"
	var buildDate = "unknown"

	// Load env vars
	cfg, err := config.GetEnvVars()
	if err != nil {
		log.Print("Error getting env vars", "error", err)
	}

	// Create logger
	logger := initLogger(cfg)

	logger.Info(fmt.Sprintf("ez2boot version %s date %s", version, buildDate))

	// Connect to db and hold connection open
	conn, repo := initDatabase(logger)
	defer conn.Close()

	// Setup DB tables
	if err := repo.SetupDB(); err != nil {
		logger.Error("Failed to setup tables in database", "error", err)
		os.Exit(1)
	}

	router, services, wkr, err := app.NewApp(version, buildDate, cfg, repo, logger)

	// Set Go routine context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background workers
	startWorkers(ctx, cfg, wkr, services, logger)

	//Start server
	logger.Info("Server is ready and listening", "port", cfg.Port)
	err = http.ListenAndServe("0.0.0.0:"+cfg.Port, router)
	if err != nil {
		logger.Error("Failed to start http server", "error", err)
		os.Exit(1)
	}
}
