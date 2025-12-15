package app

import (
	"ez2boot/internal/notification"
	"ez2boot/internal/notification/email"
	"ez2boot/internal/notification/telegram"
	"ez2boot/internal/provider/aws"
	"ez2boot/internal/server"
	"ez2boot/internal/session"
	"ez2boot/internal/user"
	"ez2boot/internal/util"
)

type Services struct {
	UserService         *user.Service
	ServerService       *server.Service
	SessionService      *session.Service
	NotificationService *notification.Service
	EmailService        *email.Service
	AWSService          *aws.Service
}

type Handlers struct {
	UserHandler         *user.Handler
	ServerHandler       *server.Handler
	SessionHandler      *session.Handler
	NotificationHandler *notification.Handler
	UtilHandler         *util.Handler
	EmailHandler        *email.Handler
	TelegramHandler     *telegram.Handler
}
