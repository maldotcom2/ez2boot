package worker

import (
	"context"
	"ez2boot/internal/db"
	"ez2boot/internal/model"
	"ez2boot/internal/provider/aws"
	"log/slog"
)

func ScrapeAndPopulate(repo *db.Repository, ctx context.Context, cfg model.Config, isRoutine bool, logger *slog.Logger) error {
	// Switch provider specific scrape function
	var scrapeFunc func(*db.Repository, model.Config, *slog.Logger) error
	switch cfg.CloudProvider {
	case "aws":
		scrapeFunc = aws.GetEC2Instances
	default:
		logger.Error("Unsupported provider", "provider", cfg.CloudProvider)
	}

	logger.Info("Scraping targets", "provider", cfg.CloudProvider)

	if isRoutine {
		startScrapeRoutine(repo, ctx, cfg, scrapeFunc, logger) // Go Routine path does not return an error
		return nil
	} else {
		return scrapeFunc(repo, cfg, logger) // One shot path does return error
	}
}
