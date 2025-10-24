package handler

import (
	"ez2boot/internal/middleware"
	"ez2boot/internal/repository"
	"log/slog"

	"github.com/gorilla/mux"
)

func SetupRoutes(router *mux.Router, repo *repository.Repository, logger *slog.Logger) {

	// API subrouter and routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(middleware.BasicAuthMiddleware(repo, logger))
	apiRouter.Use(middleware.JsonContentTypeMiddleware)
	apiRouter.Use(middleware.CORSMiddleware)

	apiRouter.HandleFunc("/servers", GetServers(repo, logger)).Methods("GET")
	apiRouter.HandleFunc("/sessions", GetSessions(repo, logger)).Methods("GET")
	apiRouter.HandleFunc("/sessions", NewSession(repo, logger)).Methods("POST")
	apiRouter.HandleFunc("/sessions", UpdateSession(repo, logger)).Methods("PUT")
	apiRouter.HandleFunc("/register", RegisterUser(repo, logger)).Methods("POST")
	apiRouter.HandleFunc("/changepassword", ChangePassword(repo, logger)).Methods("PUT")

	// UI subrouter and routes
	uiRouter := router.PathPrefix("/ui").Subrouter()
	uiRouter.Use(middleware.SessionAuthMiddleware(repo, logger))
	uiRouter.Use(middleware.JsonContentTypeMiddleware)
	uiRouter.Use(middleware.CORSMiddleware)

	uiRouter.HandleFunc("/servers", GetServers(repo, logger)).Methods("GET")
	uiRouter.HandleFunc("/sessions", GetSessions(repo, logger)).Methods("GET")
	uiRouter.HandleFunc("/sessions", NewSession(repo, logger)).Methods("POST")
	uiRouter.HandleFunc("/sessions", UpdateSession(repo, logger)).Methods("PUT")
	uiRouter.HandleFunc("/register", RegisterUser(repo, logger)).Methods("POST")
	uiRouter.HandleFunc("/changepassword", ChangePassword(repo, logger)).Methods("PUT")
}
