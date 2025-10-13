package main

import (
	"context"
	"ez2boot/internal/config"
	"ez2boot/internal/handler"
	"ez2boot/internal/middleware"
	"ez2boot/internal/repository"
	"ez2boot/internal/service"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	// Load env vars
	cfg, err := config.GetEnvVars()
	if err != nil {
		log.Print("No .env file present")
	}

	// Create log handler
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     cfg.LogLevel,
		AddSource: true,
	})

	// create logger
	logger := slog.New(logHandler)
	logger.Info("Start app")
	logger.Info("Log Level", "level", cfg.LogLevel)

	// connect to DB
	db, err := repository.Connect()
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	defer db.Close()

	// Wrap DB pointer to get data access layer
	repo := repository.NewRepository(db, logger)

	// Setup DB
	err = repo.SetupDB()
	if err != nil {
		logger.Error("Failed to setup tables in database", "error", err)
		os.Exit(1)
	}

	// Create router
	router := mux.NewRouter()

	// Setup routes
	handler.SetupRoutes(router, repo, logger)

	// Chain middleware
	router.Use(middleware.AuthMiddleware(logger))
	router.Use(middleware.JsonContentTypeMiddleware)
	router.Use(middleware.CORSMiddleware)

	// Set Go routine context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start scraper
	isRoutine := true
	service.ScrapeAndPopulate(repo, ctx, cfg, isRoutine, logger)

	// Start session worker
	service.StartSessionWorker(repo, ctx, cfg, logger)

	//Start server
	logger.Info("Server is ready and listening", "port", cfg.Port)
	err = http.ListenAndServe(":"+cfg.Port, router)
	if err != nil {
		logger.Error("Failed to start http server", "error", err)
		os.Exit(1)
	}
}
