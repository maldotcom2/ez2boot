package app

import (
	"ez2boot/internal/config"
	"ez2boot/internal/middleware"
	"net/http"
	"os"

	_ "embed"

	"github.com/gorilla/mux"
)

func SetupBackendRoutes(
	cfg *config.Config,
	router *mux.Router,
	mw *middleware.Middleware,
	handlers *Handlers,
) {

	/////////////////////////// Public routes, no auth ///////////////////////////////////

	publicRouter := router.PathPrefix("/ui").Subrouter()
	publicRouter.Use(mw.CORSMiddleware)
	publicRouter.Use(mw.LimitMiddleware)
	publicRouter.Use(mw.JsonContentTypeMiddleware)

	publicRouter.HandleFunc("/user/login", handlers.UserHandler.Login()).Methods("POST")
	publicRouter.HandleFunc("/mode", handlers.UserHandler.GetMode()).Methods("GET")
	if cfg.SetupMode {
		publicRouter.HandleFunc("/setup", handlers.UserHandler.CreateFirstTimeUser()).Methods("POST")
	}

	/////////////////////////// Admin UI subrouter and routes ////////////////////////////

	adminUIRouter := router.PathPrefix("/ui").Subrouter()
	adminUIRouter.Use(mw.CORSMiddleware)
	adminUIRouter.Use(mw.LimitMiddleware)
	adminUIRouter.Use(mw.JsonContentTypeMiddleware)
	adminUIRouter.Use(mw.SessionAuthMiddleware()) // This pattern allows passing in params, can be simplified.
	adminUIRouter.Use(mw.AdminMiddleware)

	// User
	adminUIRouter.HandleFunc("/user", handlers.UserHandler.CreateUser()).Methods("POST")
	adminUIRouter.HandleFunc("/user", handlers.UserHandler.DeleteUser()).Methods("DELETE")
	adminUIRouter.HandleFunc("/user/auth", handlers.UserHandler.UpdateUserAuthorisation()).Methods("PUT")
	// Notification
	adminUIRouter.HandleFunc("/notifications/passphrase", handlers.NotificationHandler.RotateEncryptionPhrase()).Methods("PUT")
	// Audit
	adminUIRouter.HandleFunc("/audit/events", handlers.AuditHandler.GetAuditEvents()).Methods("GET")

	/////////////////////////// UI subrouter and routes //////////////////////////////////

	uiRouter := router.PathPrefix("/ui").Subrouter()
	uiRouter.Use(mw.CORSMiddleware)
	uiRouter.Use(mw.LimitMiddleware)
	uiRouter.Use(mw.JsonContentTypeMiddleware)
	uiRouter.Use(mw.SessionAuthMiddleware()) // This pattern allows passing in params, can be simplified.

	//// Server Sessions
	uiRouter.HandleFunc("/sessions/summary", handlers.SessionHandler.GetServerSessionSummary()).Methods("GET")
	uiRouter.HandleFunc("/session", handlers.SessionHandler.NewServerSession()).Methods("POST")
	uiRouter.HandleFunc("/session", handlers.SessionHandler.UpdateServerSession()).Methods("PUT")
	//// Users
	uiRouter.HandleFunc("/users", handlers.UserHandler.GetUsers()).Methods("GET")
	uiRouter.HandleFunc("/user/session", handlers.UserHandler.CheckSession()).Methods("GET") // UI specific
	uiRouter.HandleFunc("/user/auth", handlers.UserHandler.GetUserAuthorisation()).Methods("GET")
	uiRouter.HandleFunc("/user/password", handlers.UserHandler.ChangePassword()).Methods("PUT")
	uiRouter.HandleFunc("/user/logout", handlers.UserHandler.Logout()).Methods("POST") // UI specific
	uiRouter.HandleFunc("/user/mfa/enrol", handlers.UserHandler.EnrolMFA()).Methods("POST")
	uiRouter.HandleFunc("/user/mfa/validate", handlers.UserHandler.ValidateMFA()).Methods("POST")
	/// Notification channels
	uiRouter.HandleFunc("/user/notification", handlers.NotificationHandler.GetUserNotificationSettings()).Methods("GET")
	uiRouter.HandleFunc("/user/notification", handlers.NotificationHandler.SetUserNotificationSettings()).Methods("POST")
	uiRouter.HandleFunc("/user/notification", handlers.NotificationHandler.DeleteUserNotificationSettings()).Methods("DELETE")
	uiRouter.HandleFunc("/notification/types", handlers.NotificationHandler.GetNotificationTypes()).Methods("GET")

	// Version
	uiRouter.HandleFunc("/version", handlers.UtilHandler.GetVersion()).Methods("GET")

	/////////////////////////// Admin API subrouter and routes ///////////////////////////

	adminAPIRouter := router.PathPrefix("/api/v1").Subrouter()
	adminAPIRouter.Use(mw.CORSMiddleware)
	adminAPIRouter.Use(mw.LimitMiddleware)
	adminAPIRouter.Use(mw.JsonContentTypeMiddleware)
	adminAPIRouter.Use(mw.BasicAuthMiddleware()) // This pattern allows passing in params, can be simplified.
	adminAPIRouter.Use(mw.AdminMiddleware)

	// User
	adminAPIRouter.HandleFunc("/user", handlers.UserHandler.CreateUser()).Methods("POST")
	adminAPIRouter.HandleFunc("/user", handlers.UserHandler.DeleteUser()).Methods("DELETE")
	adminAPIRouter.HandleFunc("/user/auth", handlers.UserHandler.UpdateUserAuthorisation()).Methods("PUT")
	// Notification
	adminAPIRouter.HandleFunc("/notifications/passphrase", handlers.NotificationHandler.RotateEncryptionPhrase()).Methods("PUT")
	// Audit
	adminAPIRouter.HandleFunc("/audit/events", handlers.AuditHandler.GetAuditEvents()).Methods("GET")

	/////////////////////////// API subrouter and routes /////////////////////////////////

	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	apiRouter.Use(mw.CORSMiddleware)
	apiRouter.Use(mw.LimitMiddleware)
	apiRouter.Use(mw.JsonContentTypeMiddleware)
	apiRouter.Use(mw.BasicAuthMiddleware()) // This pattern allows passing in params, can be simplified.

	//// Server sessions
	apiRouter.HandleFunc("/session", handlers.SessionHandler.NewServerSession()).Methods("POST")
	apiRouter.HandleFunc("/session", handlers.SessionHandler.UpdateServerSession()).Methods("PUT")
	//// Users
	apiRouter.HandleFunc("/users", handlers.UserHandler.GetUsers()).Methods("GET")
	apiRouter.HandleFunc("/user/auth", handlers.UserHandler.GetUserAuthorisation()).Methods("GET")
	apiRouter.HandleFunc("/user/password", handlers.UserHandler.ChangePassword()).Methods("PUT")
	/// Notification channels
	apiRouter.HandleFunc("/user/notification", handlers.NotificationHandler.GetUserNotificationSettings()).Methods("GET")
	apiRouter.HandleFunc("/user/notification", handlers.NotificationHandler.SetUserNotificationSettings()).Methods("POST")
	apiRouter.HandleFunc("/user/notification", handlers.NotificationHandler.DeleteUserNotificationSettings()).Methods("DELETE")
	apiRouter.HandleFunc("/notification/types", handlers.NotificationHandler.GetNotificationTypes()).Methods("GET")
	// Version
	apiRouter.HandleFunc("/version", handlers.UtilHandler.GetVersion()).Methods("GET")
}

func SetupFrontendRoutes(router *mux.Router) {
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
