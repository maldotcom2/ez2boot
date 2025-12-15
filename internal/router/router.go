package router

import (
	"ez2boot/internal/app"
	"ez2boot/internal/config"
	"ez2boot/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func BuildRouter(cfg *config.Config, mw *middleware.Middleware, handlers *app.Handlers) http.Handler {
	router := mux.NewRouter()

	// Setup backend routes
	app.SetupBackendRoutes(cfg, router, mw, handlers)
	// Setup frontend routes
	app.SetupFrontendRoutes(router)

	return router
}
