package main

import (
	"context"
	"ez2boot/internal/config"
	"ez2boot/internal/provider"
	"ez2boot/internal/worker"
	"log/slog"
)

func startWorkers(ctx context.Context, cfg *config.Config, wkr *worker.Worker, services *Services, logger *slog.Logger) {
	// Assign scrape implementation based off configured cloud provider
	var scraper provider.Scraper
	var manager provider.Manager

	switch cfg.CloudProvider {
	case "aws":
		scraper = services.AWSService
		manager = services.AWSService
	default:
		logger.Error("Unsupported provider", "provider", cfg.CloudProvider)
		return
	}

	// Start scraper
	worker.StartScrapeRoutine(*wkr, ctx, scraper)

	// Start manager
	worker.StartManageRoutine(*wkr, ctx, manager)

	// Start session worker
	worker.StartServerSessionWorker(*wkr, ctx)

	// Start user session cleanup
	worker.StartExpiredUserSessionCleanup(*wkr, ctx)

	// Start notification worker
	worker.StartNotificationWorker(*wkr, ctx)
}
