package main

import (
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"ez2boot/internal/middleware"
	"ez2boot/internal/notification"
	"ez2boot/internal/notification/email"
	"ez2boot/internal/notification/telegram"
	"ez2boot/internal/provider/aws"
	"ez2boot/internal/server"
	"ez2boot/internal/session"
	"ez2boot/internal/user"
	"ez2boot/internal/util"
	"ez2boot/internal/worker"
	"log/slog"
)

func initServices(version string, buildDate string, cfg *config.Config, repo *db.Repository, logger *slog.Logger) (*middleware.Middleware, *worker.Worker, *Handlers, *Services) {
	// Notification
	notificationRepo := &notification.Repository{Base: repo}
	notificationService := &notification.Service{Repo: notificationRepo, Logger: logger}
	notificationHandler := &notification.Handler{Service: notificationService, Logger: logger}

	// Server
	serverRepo := &server.Repository{Base: repo}
	serverService := &server.Service{Repo: serverRepo, Logger: logger}
	serverHandler := &server.Handler{Service: serverService, Logger: logger}

	// User
	userRepo := &user.Repository{Base: repo, Logger: logger}
	userService := &user.Service{Repo: userRepo, Config: cfg, Logger: logger}
	userHandler := &user.Handler{Service: userService, Logger: logger}

	// Session
	sessionRepo := &session.Repository{Base: repo}
	sessionService := &session.Service{Repo: sessionRepo, NotificationService: notificationService, UserService: userService, Logger: logger}
	sessionHandler := &session.Handler{Service: sessionService, Logger: logger}

	// Util
	utilHandler := &util.Handler{Version: version, BuildDate: buildDate}

	// Email
	emailRepo := &email.Repository{Base: repo}
	emailService := &email.Service{Repo: emailRepo, Logger: logger}
	emailHandler := &email.Handler{Service: emailService, Logger: logger}

	// Telegram
	telegramRepo := &telegram.Repository{Base: repo}
	telegramService := &telegram.Service{Repo: telegramRepo, Logger: logger}
	telegramHandler := &telegram.Handler{Service: telegramService, Logger: logger}

	// aws
	awsRepo := &aws.Repository{Base: repo}
	awsService := &aws.Service{Repo: awsRepo, Config: cfg, ServerService: serverService, Logger: logger}

	// Middlware
	mw := &middleware.Middleware{UserService: userService, Config: cfg, Logger: logger}

	// Worker
	wkr := &worker.Worker{ServerService: serverService, SessionService: sessionService, UserService: userService, NotificationService: notificationService, Config: cfg, Logger: logger}

	handlers := &Handlers{
		UserHandler:         userHandler,
		ServerHandler:       serverHandler,
		SessionHandler:      sessionHandler,
		NotificationHandler: notificationHandler,
		UtilHandler:         utilHandler,
		EmailHandler:        emailHandler,
		TelegramHandler:     telegramHandler,
	}

	services := &Services{
		UserService:         userService,
		ServerService:       serverService,
		SessionService:      sessionService,
		NotificationService: notificationService,
		EmailService:        emailService,
		AWSService:          awsService,
	}

	return mw, wkr, handlers, services
}
