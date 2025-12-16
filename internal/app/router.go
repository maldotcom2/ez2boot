package app

import (
	"ez2boot/internal/config"
	"ez2boot/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func BuildRouter(cfg *config.Config, mw *middleware.Middleware, handlers *Handlers) http.Handler {
	router := mux.NewRouter()

	// Setup backend routes
	SetupBackendRoutes(cfg, router, mw, handlers)
	// Setup frontend routes
	SetupFrontendRoutes(router)

	return router
}
