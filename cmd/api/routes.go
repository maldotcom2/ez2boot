package main

import (
	"ez2boot/internal/config"
	"ez2boot/internal/middleware"

	"github.com/gorilla/mux"
)

func setupRoutes(
	cfg *config.Config,
	router *mux.Router,
	mw *middleware.Middleware,
	handlers *Handlers,
) {

	// Public routes, no auth
	publicRouter := router.PathPrefix("/ui").Subrouter()
	publicRouter.Use(middleware.CORSMiddleware)
	publicRouter.Use(mw.LimitMiddleware)
	publicRouter.Use(middleware.JsonContentTypeMiddleware)

	publicRouter.HandleFunc("/user/login", handlers.UserHandler.Login()).Methods("POST")
	publicRouter.HandleFunc("/mode", handlers.UserHandler.GetMode()).Methods("GET")
	if cfg.SetupMode {
		publicRouter.HandleFunc("/setup", handlers.UserHandler.CreateFirstTimeUser()).Methods("POST")
	}

	// API subrouter and routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(middleware.CORSMiddleware)
	apiRouter.Use(mw.LimitMiddleware)
	apiRouter.Use(middleware.JsonContentTypeMiddleware)
	apiRouter.Use(mw.BasicAuthMiddleware())

	//// Servers
	apiRouter.HandleFunc("/servers", handlers.ServerHandler.GetServers()).Methods("GET")
	//// Server sessions
	apiRouter.HandleFunc("/sessions", handlers.SessionHandler.GetServerSessions()).Methods("GET")
	apiRouter.HandleFunc("/sessions", handlers.SessionHandler.NewServerSession()).Methods("POST")
	apiRouter.HandleFunc("/sessions", handlers.SessionHandler.UpdateServerSession()).Methods("PUT")
	//// Users
	apiRouter.HandleFunc("/user/new", handlers.UserHandler.CreateUser()).Methods("POST")
	apiRouter.HandleFunc("/user/changepassword", handlers.UserHandler.ChangePassword()).Methods("PUT")

	// UI subrouter and routes
	uiRouter := router.PathPrefix("/ui").Subrouter()
	uiRouter.Use(middleware.CORSMiddleware)
	uiRouter.Use(mw.LimitMiddleware)
	uiRouter.Use(middleware.JsonContentTypeMiddleware)
	uiRouter.Use(mw.SessionAuthMiddleware())

	//// Servers
	uiRouter.HandleFunc("/servers", handlers.ServerHandler.GetServers()).Methods("GET")
	//// Server Sessions
	uiRouter.HandleFunc("/sessions", handlers.SessionHandler.GetServerSessions()).Methods("GET")
	uiRouter.HandleFunc("/session/summary", handlers.SessionHandler.GetServerSessionSummary()).Methods("GET")
	uiRouter.HandleFunc("/session/new", handlers.SessionHandler.NewServerSession()).Methods("POST")
	uiRouter.HandleFunc("/session/update", handlers.SessionHandler.UpdateServerSession()).Methods("PUT")
	//// Users
	uiRouter.HandleFunc("/user/session", handlers.UserHandler.CheckSession()).Methods("GET")
	uiRouter.HandleFunc("/user/new", handlers.UserHandler.CreateUser()).Methods("POST")
	uiRouter.HandleFunc("/user/changepassword", handlers.UserHandler.ChangePassword()).Methods("PUT")
	uiRouter.HandleFunc("/user/logout", handlers.UserHandler.Logout()).Methods("POST")

	/// Notification channels
	uiRouter.HandleFunc("/notification/types", handlers.NotificationHandler.GetNotificationTypes()).Methods("GET")
	uiRouter.HandleFunc("/notification/email/update", handlers.EmailHandler.AddOrUpdate()).Methods("POST")
}
