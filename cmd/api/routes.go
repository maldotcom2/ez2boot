package main

import (
	"ez2boot/internal/middleware"
	"ez2boot/internal/server"
	"ez2boot/internal/session"
	"ez2boot/internal/user"

	"github.com/gorilla/mux"
)

func SetupRoutes(router *mux.Router, mw *middleware.Middleware, server *server.Handler, session *session.Handler, user *user.Handler) {

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
