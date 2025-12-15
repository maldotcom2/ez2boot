package app

import (
	"ez2boot/internal/config"
	"ez2boot/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func BuildHTTPApp(
	cfg *config.Config,
	mw *middleware.Middleware,
	handlers *Handlers,
) http.Handler {
	router := mux.NewRouter()
	SetupBackendRoutes(cfg, router, mw, handlers)
	return router
}
