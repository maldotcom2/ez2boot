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
	publicRouter.Use(mw.PublicLimitMiddleware)
	publicRouter.Use(mw.JsonContentTypeMiddleware)

	publicRouter.HandleFunc("/auth/login", handlers.AuthHandler.Login()).Methods("POST")
	publicRouter.HandleFunc("/auth/oidc/login", handlers.OidcHandler.Login()).Methods("GET")
	publicRouter.HandleFunc("/auth/oidc/callback", handlers.OidcHandler.Callback()).Methods("GET")
	publicRouter.HandleFunc("/auth/oidc/status", handlers.OidcHandler.HasOidc()).Methods("GET")
	publicRouter.HandleFunc("/user/mfa/verify", handlers.UserHandler.VerifyMFA()).Methods("POST")
	publicRouter.HandleFunc("/mode", handlers.UserHandler.GetMode()).Methods("GET")

	if cfg.SetupMode {
		publicRouter.HandleFunc("/setup", handlers.UserHandler.CreateFirstTimeUser()).Methods("POST")
	}

	/////////////////////////// Admin UI subrouter and routes ////////////////////////////

	adminUIRouter := router.PathPrefix("/ui").Subrouter()
	adminUIRouter.Use(mw.CORSMiddleware)
	adminUIRouter.Use(mw.PrivateLimitMiddleware)
	adminUIRouter.Use(mw.JsonContentTypeMiddleware)
	adminUIRouter.Use(mw.SessionAuthMiddleware()) // This pattern allows passing in params, can be simplified.
	adminUIRouter.Use(mw.AdminMiddleware)

	//// Server Sessions
	adminUIRouter.HandleFunc("/admin/session", handlers.SessionHandler.UpdateServerSessionAdmin()).Methods("PUT")
	// User
	adminUIRouter.HandleFunc("/users", handlers.UserHandler.GetUsers()).Methods("GET")
	adminUIRouter.HandleFunc("/user", handlers.UserHandler.CreateUser()).Methods("POST")
	adminUIRouter.HandleFunc("/user/ldap", handlers.LdapHandler.CreateLdapUser()).Methods("POST")
	adminUIRouter.HandleFunc("/user", handlers.UserHandler.DeleteUser()).Methods("DELETE")
	adminUIRouter.HandleFunc("/user/auth", handlers.UserHandler.UpdateUserAuthorisation()).Methods("PUT")
	// Encryption
	adminUIRouter.HandleFunc("/encryption/passphrase", handlers.EncryptionHandler.RotateEncryptionPhrase()).Methods("PUT")
	// Audit
	adminUIRouter.HandleFunc("/audit/events", handlers.AuditHandler.GetAuditEvents()).Methods("GET")
	/// Ldap
	adminUIRouter.HandleFunc("/auth/ldap", handlers.LdapHandler.GetLdapConfig()).Methods("GET")
	adminUIRouter.HandleFunc("/auth/ldap", handlers.LdapHandler.SetLdapConfig()).Methods("POST")
	adminUIRouter.HandleFunc("/auth/ldap", handlers.LdapHandler.DeleteLdapConfig()).Methods("DELETE")
	adminUIRouter.HandleFunc("/auth/ldap/users/search", handlers.LdapHandler.SearchUser()).Methods("POST")
	// Oidc
	adminUIRouter.HandleFunc("/auth/oidc", handlers.OidcHandler.GetOidcConfig()).Methods("GET")
	adminUIRouter.HandleFunc("/auth/oidc", handlers.OidcHandler.SetOidcConfig()).Methods("POST")
	adminUIRouter.HandleFunc("/auth/oidc", handlers.OidcHandler.DeleteOidcConfig()).Methods("DELETE")
	adminUIRouter.HandleFunc("/auth/oidc/test", handlers.OidcHandler.TestOidcConnection()).Methods("POST")

	/////////////////////////// UI subrouter and routes //////////////////////////////////

	uiRouter := router.PathPrefix("/ui").Subrouter()
	uiRouter.Use(mw.CORSMiddleware)
	uiRouter.Use(mw.PrivateLimitMiddleware)
	uiRouter.Use(mw.JsonContentTypeMiddleware)
	uiRouter.Use(mw.SessionAuthMiddleware()) // This pattern allows passing in params, can be simplified.

	//// Server Sessions
	uiRouter.HandleFunc("/sessions/summary", handlers.SessionHandler.GetServerSessionSummary()).Methods("GET")
	uiRouter.HandleFunc("/session", handlers.SessionHandler.NewServerSession()).Methods("POST")
	uiRouter.HandleFunc("/session", handlers.SessionHandler.UpdateServerSession()).Methods("PUT")
	//// Users
	uiRouter.HandleFunc("/user/session", handlers.UserHandler.CheckSession()).Methods("GET") // UI specific
	uiRouter.HandleFunc("/user/auth", handlers.UserHandler.GetUserAuthorisation()).Methods("GET")
	uiRouter.HandleFunc("/user/password", handlers.UserHandler.ChangePassword()).Methods("PUT")
	uiRouter.HandleFunc("/user/logout", handlers.AuthHandler.Logout()).Methods("POST")          // UI specific
	uiRouter.HandleFunc("/user/mfa", handlers.UserHandler.EnrolMFA()).Methods("POST")           // UI specific
	uiRouter.HandleFunc("/user/mfa/confirm", handlers.UserHandler.ConfirmMFA()).Methods("POST") // UI specific
	uiRouter.HandleFunc("/user/mfa/delete", handlers.UserHandler.DeleteMFA()).Methods("POST")   // UI specific - Post to allow body
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
	adminAPIRouter.Use(mw.PrivateLimitMiddleware)
	adminAPIRouter.Use(mw.JsonContentTypeMiddleware)
	adminAPIRouter.Use(mw.BasicAuthMiddleware()) // This pattern allows passing in params, can be simplified.
	adminAPIRouter.Use(mw.AdminMiddleware)

	//// Server Sessions
	adminAPIRouter.HandleFunc("/admin/session", handlers.SessionHandler.UpdateServerSessionAdmin()).Methods("PUT")
	// User
	adminAPIRouter.HandleFunc("/users", handlers.UserHandler.GetUsers()).Methods("GET")
	adminAPIRouter.HandleFunc("/user", handlers.UserHandler.CreateUser()).Methods("POST")
	adminAPIRouter.HandleFunc("/user/ldap", handlers.LdapHandler.CreateLdapUser()).Methods("POST")
	adminAPIRouter.HandleFunc("/user", handlers.UserHandler.DeleteUser()).Methods("DELETE")
	adminAPIRouter.HandleFunc("/user/auth", handlers.UserHandler.UpdateUserAuthorisation()).Methods("PUT")
	// Encryption
	adminUIRouter.HandleFunc("/encryption/passphrase", handlers.EncryptionHandler.RotateEncryptionPhrase()).Methods("PUT")
	// Audit
	adminAPIRouter.HandleFunc("/audit/events", handlers.AuditHandler.GetAuditEvents()).Methods("GET")
	/// Ldap
	adminAPIRouter.HandleFunc("/auth/ldap", handlers.LdapHandler.GetLdapConfig()).Methods("GET")
	adminAPIRouter.HandleFunc("/auth/ldap", handlers.LdapHandler.SetLdapConfig()).Methods("POST")
	adminAPIRouter.HandleFunc("/auth/ldap", handlers.LdapHandler.DeleteLdapConfig()).Methods("DELETE")
	adminAPIRouter.HandleFunc("/auth/ldap/users/search", handlers.LdapHandler.SearchUser()).Methods("POST")
	// Oidc
	adminAPIRouter.HandleFunc("/auth/oidc", handlers.OidcHandler.GetOidcConfig()).Methods("GET")
	adminAPIRouter.HandleFunc("/auth/oidc", handlers.OidcHandler.SetOidcConfig()).Methods("POST")
	adminAPIRouter.HandleFunc("/auth/oidc", handlers.OidcHandler.DeleteOidcConfig()).Methods("DELETE")
	adminAPIRouter.HandleFunc("/auth/oidc/test", handlers.OidcHandler.TestOidcConnection()).Methods("POST")

	/////////////////////////// API subrouter and routes /////////////////////////////////

	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	apiRouter.Use(mw.CORSMiddleware)
	apiRouter.Use(mw.PrivateLimitMiddleware)
	apiRouter.Use(mw.JsonContentTypeMiddleware)
	apiRouter.Use(mw.BasicAuthMiddleware()) // This pattern allows passing in params, can be simplified.

	//// Server sessions
	apiRouter.HandleFunc("/session", handlers.SessionHandler.NewServerSession()).Methods("POST")
	apiRouter.HandleFunc("/session", handlers.SessionHandler.UpdateServerSession()).Methods("PUT")
	//// Users
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
