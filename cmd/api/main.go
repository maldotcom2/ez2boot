package main

import (
	"context"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	// Load env vars
	cfg := loadVars()

	// Create logger
	logger := initLogger(cfg)

	// Connect to db and hold connection open
	conn, repo := initDatabase(logger)
	defer conn.Close()

	// Setup domain/service structs
	mw, wkr, handlers, services := initServices(cfg, repo, logger)

	// Setup DB tables
	if err := repo.SetupDB(); err != nil {
		logger.Error("Failed to setup tables in database", "error", err)
		os.Exit(1)
	}

	// Set the runtime mode, if no users then SetupMode is on
	setupMode, err := services.UserService.HasUsers()
	if err != nil {
		logger.Error("Failed to check database for existing users", "error", err)
		os.Exit(1)
	}

	// No users = setup mode
	cfg.SetupMode = !setupMode

	// Create router
	router := mux.NewRouter()

	// Setup routes
	setupRoutes(cfg, router, mw, handlers)

	// Set Go routine context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background workers
	startWorkers(ctx, cfg, wkr, services, logger)

	//Start server
	logger.Info("Server is ready and listening", "port", cfg.Port)
	err = http.ListenAndServe(":"+cfg.Port, router)
	if err != nil {
		logger.Error("Failed to start http server", "error", err)
		os.Exit(1)
	}
}
