package app

import (
	"ez2boot/internal/audit"
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

func InitServices(version string, buildDate string, cfg *config.Config, repo *db.Repository, logger *slog.Logger) (*middleware.Middleware, *worker.Worker, *Handlers, *Services) {

	// Audit
	auditRepo := audit.NewRepository(repo)
	auditService := audit.NewService(auditRepo, logger)

	// Notification
	notificationRepo := notification.NewRepository(repo)
	notificationService := notification.NewService(notificationRepo, auditService, logger)
	notificationHandler := notification.NewHandler(notificationService, logger)

	// Server
	serverRepo := server.NewRepository(repo)
	serverService := server.NewService(serverRepo, logger)
	serverHandler := server.NewHandler(serverService, logger)

	// User
	userRepo := user.NewRepository(repo, logger)
	userService := user.NewService(userRepo, cfg, auditService, logger)
	userHandler := user.NewHandler(userService, logger)

	// Session
	sessionRepo := session.NewRepository(repo)
	sessionService := session.NewService(sessionRepo, notificationService, userService, auditService, logger)
	sessionHandler := session.NewHandler(sessionService, logger)

	// Util
	utilHandler := util.NewHandler(version, buildDate)

	// Email
	emailRepo := email.NewRepository(repo)
	emailService := email.NewService(emailRepo, logger)
	emailHandler := email.NewHandler(emailService, logger)

	// Telegram
	telegramRepo := telegram.NewRepository(repo)
	telegramService := telegram.NewService(telegramRepo, logger)
	telegramHandler := telegram.NewHandler(telegramService, logger)

	// aws
	awsRepo := aws.NewRepository(repo)
	awsService := aws.NewService(awsRepo, cfg, serverService, logger)

	// Middlware
	mw := middleware.NewMiddleware(userService, cfg, auditService, logger)

	// Worker
	wkr := worker.NewWorker(serverService, sessionService, userService, notificationService, cfg, logger)

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
