package main

import (
	"context"
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"ez2boot/internal/middleware"
	"ez2boot/internal/notification"
	"ez2boot/internal/notification/email"
	"ez2boot/internal/provider"
	"ez2boot/internal/provider/aws"
	"ez2boot/internal/server"
	"ez2boot/internal/session"
	"ez2boot/internal/user"
	"ez2boot/internal/worker"
	"fmt"
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

	// Notification repo
	notificationRepo := &notification.Repository{
		Base: repo,
	}

	// Notification service
	notificationService := &notification.Service{
		Repo:   notificationRepo,
		Logger: logger,
	}

	// Notification handler
	notificationHandler := &notification.Handler{
		Service: notificationService,
		Logger:  logger,
	}

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
		Logger:  logger,
	}

	// User repo
	userRepo := &user.Repository{
		Base:   repo,
		Logger: logger,
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

	// Session repo
	sessionRepo := &session.Repository{
		Base: repo,
	}

	// Session service
	sessionService := &session.Service{
		Repo:                sessionRepo,
		NotificationService: notificationService,
		UserService:         userService,
		Logger:              logger,
	}

	// Session handler
	sessionHandler := &session.Handler{
		Service: sessionService,
		Logger:  logger,
	}

	// Middlware
	mw := &middleware.Middleware{
		UserService: userService,
		Logger:      logger,
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

	// Email repo
	emailRepo := &email.Repository{
		Base: repo,
	}

	// Email service
	emailService := &email.Service{
		Repo:   emailRepo,
		Logger: logger,
	}

	// Email handler
	emailHandler := &email.Handler{
		Service: emailService,
		Logger:  logger,
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
	SetupRoutes(router, mw, serverHandler, sessionHandler, userHandler, notificationHandler, emailHandler)

	// Set Go routine context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Assign scrape implementation based off configured cloud provider
	var scraper provider.Scraper
	switch cfg.CloudProvider {
	case "aws":
		scraper = awsService
	default:
		logger.Error("Unsupported provider", "provider", cfg.CloudProvider)
	}

	logger.Info("Scraper type", "type", fmt.Sprintf("%T", scraper))

	// worker
	w := &worker.Worker{
		ServerService:       serverService,
		SessionService:      sessionService,
		UserService:         userService,
		NotificationService: notificationService,
		Config:              cfg,
		Logger:              logger,
	}

	// Start scraper
	worker.StartScrapeRoutine(*w, ctx, scraper)

	// Start session worker
	worker.StartServerSessionWorker(*w, ctx)

	// Start user session cleanup
	worker.StartExpiredUserSessionCleanup(*w, ctx)

	// Start notification worker
	worker.StartNotificationWorker(*w, ctx)

	//Start server
	logger.Info("Server is ready and listening", "port", cfg.Port)
	err = http.ListenAndServe(":"+cfg.Port, router)
	if err != nil {
		logger.Error("Failed to start http server", "error", err)
		os.Exit(1)
	}
}

func SetupRoutes(
	router *mux.Router,
	mw *middleware.Middleware,
	server *server.Handler,
	session *session.Handler,
	user *user.Handler,
	notification *notification.Handler,
	email *email.Handler,
) {

	// Public routes, no auth
	publicRouter := router.PathPrefix("/ui").Subrouter()
	publicRouter.Use(middleware.JsonContentTypeMiddleware)
	publicRouter.Use(middleware.CORSMiddleware)
	publicRouter.HandleFunc("user/login", user.Login()).Methods("POST")
	publicRouter.HandleFunc("user/new", user.CreateUser()).Methods("POST") //TODO remove

	// API subrouter and routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(mw.BasicAuthMiddleware())
	apiRouter.Use(middleware.JsonContentTypeMiddleware)
	apiRouter.Use(middleware.CORSMiddleware)
	//// Servers
	apiRouter.HandleFunc("/servers", server.GetServers()).Methods("GET")
	//// Server sessions
	apiRouter.HandleFunc("/sessions", session.GetServerSessions()).Methods("GET")
	apiRouter.HandleFunc("/sessions", session.NewServerSession()).Methods("POST")
	apiRouter.HandleFunc("/sessions", session.UpdateServerSession()).Methods("PUT")
	//// Users
	apiRouter.HandleFunc("/user/new", user.CreateUser()).Methods("POST")
	apiRouter.HandleFunc("/user/changepassword", user.ChangePassword()).Methods("PUT")

	// UI subrouter and routes
	uiRouter := router.PathPrefix("/ui").Subrouter()
	uiRouter.Use(mw.SessionAuthMiddleware())
	uiRouter.Use(middleware.JsonContentTypeMiddleware)
	uiRouter.Use(middleware.CORSMiddleware)
	//// Servers
	uiRouter.HandleFunc("/servers", server.GetServers()).Methods("GET")
	//// Server Sessions
	uiRouter.HandleFunc("/sessions", session.GetServerSessions()).Methods("GET")
	uiRouter.HandleFunc("/sessions", session.NewServerSession()).Methods("POST")
	uiRouter.HandleFunc("/sessions", session.UpdateServerSession()).Methods("PUT")
	//// Users
	uiRouter.HandleFunc("/user/new", user.CreateUser()).Methods("POST")
	uiRouter.HandleFunc("/user/changepassword", user.ChangePassword()).Methods("PUT")
	uiRouter.HandleFunc("/user/logout", user.Logout()).Methods("POST")
	//uiRouter.HandleFunc("/notification/sender", notification.GetNotificationTypes()).Methods("GET")
	/// Notification channels
	uiRouter.HandleFunc("/email/update", email.AddOrUpdate()).Methods("POST")
}
