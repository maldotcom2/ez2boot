package main

import (
	"context"
	"ez2boot/internal/app"
	"ez2boot/internal/config"
	"ez2boot/internal/provider"
	"ez2boot/internal/worker"
	"fmt"
	"log/slog"
)

func startWorkers(ctx context.Context, cfg *config.Config, wkr *worker.Worker, services *app.Services, logger *slog.Logger) error {
	// Assign scrape implementation based off configured cloud provider
	var scraper provider.Scraper
	var manager provider.Manager

	switch cfg.CloudProvider {
	case "aws":
		scraper = services.AWSService
		manager = services.AWSService
	default:
		return fmt.Errorf("unsupported provider: %s", cfg.CloudProvider)
	}

	// Start scraper
	worker.StartScrapeRoutine(*wkr, ctx, scraper)

	// Start manager
	worker.StartManageRoutine(*wkr, ctx, manager)

	// Start notification worker
	worker.StartNotificationWorker(*wkr, ctx)

	// Start session worker
	worker.StartServerSessionWorker(*wkr, ctx)

	// Start user session cleanup worker
	worker.StartExpiredUserSessionCleanup(*wkr, ctx)

	// Start release check worker
	worker.StartReleaseWorker(*wkr, ctx)

	return nil
}
