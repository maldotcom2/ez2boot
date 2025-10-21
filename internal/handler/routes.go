package handler

import (
	"ez2boot/internal/repository"
	"log/slog"

	"github.com/gorilla/mux"
)

func SetupRoutes(router *mux.Router, repo *repository.Repository, logger *slog.Logger) {
	router.HandleFunc("/servers", GetServers(repo, logger)).Methods("GET")
	router.HandleFunc("/sessions", GetSessions(repo, logger)).Methods("GET")
	router.HandleFunc("/sessions", NewSession(repo, logger)).Methods("POST")
	router.HandleFunc("/sessions", UpdateSession(repo, logger)).Methods("PUT")

	router.HandleFunc("/register", RegisterUser(repo, logger)).Methods("POST")
	router.HandleFunc("/changepassword", ChangePassword(repo, logger)).Methods("PUT")
}
