package main

import (
	"ez2boot/internal/config"
	"ez2boot/internal/middleware"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func setupBackendRoutes(
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
	publicRouter.HandleFunc("/version", handlers.UtilHandler.GetVersion()).Methods("GET")
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
	uiRouter.HandleFunc("/users", handlers.UserHandler.GetUsers()).Methods("GET")
	uiRouter.HandleFunc("/user/session", handlers.UserHandler.CheckSession()).Methods("GET")
	uiRouter.HandleFunc("/user/auth", handlers.UserHandler.GetUserAuthorisation()).Methods("GET")
	uiRouter.HandleFunc("/user/auth/update", handlers.UserHandler.UpdateUserAuthorisation()).Methods("POST")
	uiRouter.HandleFunc("/user/new", handlers.UserHandler.CreateUser()).Methods("POST")
	uiRouter.HandleFunc("/user/delete", handlers.UserHandler.DeleteUser()).Methods("DELETE")
	uiRouter.HandleFunc("/user/changepassword", handlers.UserHandler.ChangePassword()).Methods("PUT")
	uiRouter.HandleFunc("/user/logout", handlers.UserHandler.Logout()).Methods("POST")
	/// Notification channels
	uiRouter.HandleFunc("/user/notification", handlers.NotificationHandler.GetUserNotificationSettings()).Methods("GET")
	uiRouter.HandleFunc("/user/notification", handlers.NotificationHandler.SetUserNotificationSettings()).Methods("POST")
	uiRouter.HandleFunc("/user/notification", handlers.NotificationHandler.DeleteUserNotificationSettings()).Methods("DELETE")
	uiRouter.HandleFunc("/notification/types", handlers.NotificationHandler.GetNotificationTypes()).Methods("GET")
}

func setupFrontendRoutes(router *mux.Router) {
	// Serve static frontend (Vue build output)
	staticDir := "./web"

	// File server for assets
	fileServer := http.FileServer(http.Dir(staticDir))

	// Serve static folders directly
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/", fileServer))
	router.PathPrefix("/css/").Handler(http.StripPrefix("/", fileServer))
	router.PathPrefix("/js/").Handler(http.StripPrefix("/", fileServer))

	// Catch-all route for SPA
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Attempt to serve a static file
		path := staticDir + r.URL.Path
		if _, err := os.Stat(path); err == nil {
			http.ServeFile(w, r, path)
			return
		}

		// Else serve the SPA entry point
		http.ServeFile(w, r, staticDir+"/index.html")
	})
}
