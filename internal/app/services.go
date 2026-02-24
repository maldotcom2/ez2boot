package app

import (
	"ez2boot/internal/audit"
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"ez2boot/internal/encryption"
	"ez2boot/internal/middleware"
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
	"ez2boot/internal/worker"
	"fmt"
	"log/slog"
)

// TODO this is a mess
func InitServices(version string, buildDate string, cfg *config.Config, repo *db.Repository, logger *slog.Logger) (*middleware.Middleware, *worker.Worker, *Handlers, *Services, error) {
	buildInfo := util.BuildInfo{
		Version:   version,
		BuildDate: buildDate,
	}

	// Create encryptor for user notification settings
	encryptor, err := encryption.NewAESGCMEncryptor(cfg.EncryptionPhrase)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to init encryptor: %w", err)
	}

	// Audit
	auditRepo := audit.NewRepository(repo)
	auditService := audit.NewService(auditRepo, logger)

	// Server
	serverRepo := server.NewRepository(repo)
	serverService := server.NewService(serverRepo, logger)
	serverHandler := server.NewHandler(serverService, logger)

	// User
	userRepo := user.NewRepository(repo, logger)
	userService := user.NewService(userRepo, cfg, auditService, logger)
	userHandler := user.NewHandler(userService, cfg, logger)

	// Notification
	notificationRepo := notification.NewRepository(repo)
	notificationService := notification.NewService(notificationRepo, auditService, encryptor, logger)
	notificationHandler := notification.NewHandler(notificationService, userHandler, logger)

	auditHandler := audit.NewHandler(auditService, userHandler, logger) //TODO reorganise this

	// Session
	sessionRepo := session.NewRepository(repo)
	sessionService := session.NewService(sessionRepo, notificationService, userService, auditService, logger)
	sessionHandler := session.NewHandler(sessionService, logger)

	// Util
	utilRepo := util.NewRepository(repo)
	utilService := util.NewService(utilRepo, cfg, buildInfo, logger)
	utilHandler := util.NewHandler(utilService, logger)

	// Email
	emailRepo := email.NewRepository(repo)
	emailService := email.NewService(emailRepo, logger)
	emailHandler := email.NewHandler(emailService, logger)

	// Teams
	teamsRepo := teams.NewRepository(repo)
	teamsService := teams.NewService(teamsRepo, logger)
	teamsHandler := teams.NewHandler(teamsService, logger)

	// Telegram
	telegramRepo := telegram.NewRepository(repo)
	telegramService := telegram.NewService(telegramRepo, logger)
	telegramHandler := telegram.NewHandler(telegramService, logger)

	// AWS
	awsRepo := aws.NewRepository(repo)
	awsService, err := aws.NewService(awsRepo, cfg, serverService, logger)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Azure
	azureRepo := azure.NewRepository(repo)
	azureService, err := azure.NewService(azureRepo, cfg, serverService, logger)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Middlware
	mw := middleware.NewMiddleware(userService, cfg, auditService, logger)

	// Worker
	wkr := worker.NewWorker(serverService, sessionService, userService, notificationService, utilService, cfg, logger)

	handlers := &Handlers{
		AuditHandler:        auditHandler,
		UserHandler:         userHandler,
		ServerHandler:       serverHandler,
		SessionHandler:      sessionHandler,
		NotificationHandler: notificationHandler,
		UtilHandler:         utilHandler,
		TeamsHandler:        teamsHandler,
		EmailHandler:        emailHandler,
		TelegramHandler:     telegramHandler,
	}

	services := &Services{
		UserService:         userService,
		ServerService:       serverService,
		SessionService:      sessionService,
		NotificationService: notificationService,
		UtilService:         utilService,
		EmailService:        emailService,
		AWSService:          awsService,
		AzureService:        azureService,
	}

	return mw, wkr, handlers, services, nil
}
