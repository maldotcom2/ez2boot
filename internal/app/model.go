package app

import (
	"ez2boot/internal/audit"
	"ez2boot/internal/auth"
	"ez2boot/internal/auth/ldap"
	"ez2boot/internal/auth/oidc"
	"ez2boot/internal/encryption"
	"ez2boot/internal/notification"
	"ez2boot/internal/notification/email"
	"ez2boot/internal/notification/teams"
	"ez2boot/internal/notification/telegram"
	"ez2boot/internal/provider/aws"
	"ez2boot/internal/provider/azure"
	"ez2boot/internal/server"
	"ez2boot/internal/session"
	"ez2boot/internal/user"
	"ez2boot/internal/util"
)

type Services struct {
	AuthService         *auth.Service
	UserService         *user.Service
	LdapService         *ldap.Service
	OidcService         *oidc.Service
	ServerService       *server.Service
	SessionService      *session.Service
	NotificationService *notification.Service
	UtilService         *util.Service
	EmailService        *email.Service
	AWSService          *aws.Service
	AzureService        *azure.Service
}

type Handlers struct {
	AuthHandler         *auth.Handler
	UserHandler         *user.Handler
	LdapHandler         *ldap.Handler
	OidcHandler         *oidc.Handler
	AuditHandler        *audit.Handler
	ServerHandler       *server.Handler
	SessionHandler      *session.Handler
	NotificationHandler *notification.Handler
	UtilHandler         *util.Handler
	EncryptionHandler   *encryption.Handler
	EmailHandler        *email.Handler
	TeamsHandler        *teams.Handler
	TelegramHandler     *telegram.Handler
}
