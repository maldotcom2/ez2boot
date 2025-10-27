package main

import (
	"context"
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"ez2boot/internal/middleware"
	"ez2boot/internal/provider/aws"
	"ez2boot/internal/server"
	"ez2boot/internal/session"
	"ez2boot/internal/user"
	"ez2boot/internal/worker"
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
		log.Print("Could not load environment variables, check that .env file is present or that env variables have been configured")
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
	conn, err := db.Connect()
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	defer conn.Close()

	// Shared base repo
	repo := db.NewRepository(conn, logger)

	// Server repo
	serverRepo := &server.Repository{
		Base: repo,
	}

	// Server service
	serverService := &server.Service{
		Repo:   serverRepo,
		Logger: logger,
	}

	// Server handler
	serverHandler := &server.Handler{
		Service: serverService,
	}

	// Session repo
	sessionRepo := &session.Repository{
		Base: repo,
	}

	// Session service
	sessionService := &session.Service{
		Repo:   sessionRepo,
		Logger: logger,
	}

	// Session handler
	sessionHandler := &session.Handler{
		Service: sessionService,
	}

	// User repo
	userRepo := &user.Repository{
		Base: repo,
	}

	// User service
	userService := &user.Service{
		Repo:   userRepo,
		Config: cfg,
		Logger: logger,
	}

	// User handler
	userHandler := &user.Handler{
		Service: userService,
		Logger:  logger,
	}

	// Middlware
	mw := &middleware.Middleware{
		Service: userService,
		Logger:  logger,
	}

	// aws repository
	awsRepo := &aws.Repository{
		Base: repo,
	}

	// aws service
	awsService := &aws.Service{
		Repo:          awsRepo,
		Config:        cfg,
		ServerService: serverService,
		Logger:        logger,
	}

	// Setup DB
	err = repo.SetupDB()
	if err != nil {
		logger.Error("Failed to setup tables in database", "error", err)
		os.Exit(1)
	}

	// Create router
	router := mux.NewRouter()

	// Setup routes
	SetupRoutes(router, mw, serverHandler, sessionHandler, userHandler)

	// Set Go routine context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var scrapeFunc func() error
	switch cfg.CloudProvider {
	case "aws":
		scrapeFunc = awsService.GetEC2Instances
	default:
		logger.Error("Unsupported provider", "provider", cfg.CloudProvider)
	}

	// worker
	w := &worker.Worker{
		ServerService:  serverService,
		SessionService: sessionService,
		Config:         cfg,
		Logger:         logger,
	}

	// Start scraper
	worker.StartScrapeRoutine(*w, ctx, scrapeFunc)

	// Start session worker
	worker.StartSessionWorker(*w, ctx)

	//Start server
	logger.Info("Server is ready and listening", "port", cfg.Port)
	err = http.ListenAndServe(":"+cfg.Port, router)
	if err != nil {
		logger.Error("Failed to start http server", "error", err)
		os.Exit(1)
	}
}

func SetupRoutes(router *mux.Router, mw *middleware.Middleware, server *server.Handler, session *session.Handler, user *user.Handler) {

	// Public routes, no auth
	publicRouter := router.PathPrefix("/ui").Subrouter()
	publicRouter.Use(middleware.JsonContentTypeMiddleware)
	publicRouter.Use(middleware.CORSMiddleware)
	publicRouter.HandleFunc("/login", user.Login()).Methods("POST")

	// API subrouter and routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(mw.BasicAuthMiddleware())
	apiRouter.Use(middleware.JsonContentTypeMiddleware)
	apiRouter.Use(middleware.CORSMiddleware)
	apiRouter.HandleFunc("/servers", server.GetServers()).Methods("GET")
	apiRouter.HandleFunc("/sessions", session.GetSessions()).Methods("GET")
	apiRouter.HandleFunc("/sessions", session.NewSession()).Methods("POST")
	apiRouter.HandleFunc("/sessions", session.UpdateSession()).Methods("PUT")
	apiRouter.HandleFunc("/register", user.RegisterUser()).Methods("POST")
	apiRouter.HandleFunc("/changepassword", user.ChangePassword()).Methods("PUT")

	// UI subrouter and routes
	uiRouter := router.PathPrefix("/ui").Subrouter()
	uiRouter.Use(mw.SessionAuthMiddleware())
	uiRouter.Use(middleware.JsonContentTypeMiddleware)
	uiRouter.Use(middleware.CORSMiddleware)
	uiRouter.HandleFunc("/servers", server.GetServers()).Methods("GET")
	uiRouter.HandleFunc("/sessions", session.GetSessions()).Methods("GET")
	uiRouter.HandleFunc("/sessions", session.NewSession()).Methods("POST")
	uiRouter.HandleFunc("/sessions", session.UpdateSession()).Methods("PUT")
	uiRouter.HandleFunc("/register", user.RegisterUser()).Methods("POST")
	uiRouter.HandleFunc("/changepassword", user.ChangePassword()).Methods("PUT")
}
