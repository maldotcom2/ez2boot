package main

import (
	"ez2boot/internal/db"
	"ez2boot/internal/middleware"
	"ez2boot/internal/server"
	"ez2boot/internal/session"
	"ez2boot/internal/user"
	"log/slog"

	"github.com/gorilla/mux"
)

func SetupRoutes(router *mux.Router, repo *db.Repository, logger *slog.Logger) {

	// API subrouter and routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(middleware.BasicAuthMiddleware(repo, logger))
	apiRouter.Use(middleware.JsonContentTypeMiddleware)
	apiRouter.Use(middleware.CORSMiddleware)

	apiRouter.HandleFunc("/servers", server.GetServers(repo, logger)).Methods("GET")
	apiRouter.HandleFunc("/sessions", session.GetSessions(repo, logger)).Methods("GET")
	apiRouter.HandleFunc("/sessions", session.NewSession(repo, logger)).Methods("POST")
	apiRouter.HandleFunc("/sessions", session.UpdateSession(repo, logger)).Methods("PUT")
	apiRouter.HandleFunc("/register", user.RegisterUser(repo, logger)).Methods("POST")
	apiRouter.HandleFunc("/changepassword", user.ChangePassword(repo, logger)).Methods("PUT")

	// UI subrouter and routes
	uiRouter := router.PathPrefix("/ui").Subrouter()
	uiRouter.Use(middleware.SessionAuthMiddleware(repo, logger))
	uiRouter.Use(middleware.JsonContentTypeMiddleware)
	uiRouter.Use(middleware.CORSMiddleware)

	uiRouter.HandleFunc("/servers", server.GetServers(repo, logger)).Methods("GET")
	uiRouter.HandleFunc("/sessions", session.GetSessions(repo, logger)).Methods("GET")
	uiRouter.HandleFunc("/sessions", session.NewSession(repo, logger)).Methods("POST")
	uiRouter.HandleFunc("/sessions", session.UpdateSession(repo, logger)).Methods("PUT")
	uiRouter.HandleFunc("/register", user.RegisterUser(repo, logger)).Methods("POST")
	uiRouter.HandleFunc("/changepassword", user.ChangePassword(repo, logger)).Methods("PUT")
}
