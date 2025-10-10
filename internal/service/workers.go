package service

import (
	"context"
	"ez2boot/internal/aws"
	"ez2boot/internal/model"
	"ez2boot/internal/repository"
	"log/slog"
	"time"
)

func ScrapeAndPopulate(repo *repository.Repository, ctx context.Context, cfg model.Config, isRoutine bool, logger *slog.Logger) error {
	// Switch provider specific scrape function
	var scrapeFunc func(*repository.Repository, model.Config, *slog.Logger) error
	switch cfg.CloudProvider {
	case "aws":
		scrapeFunc = aws.GetEC2Instances
	default:
		logger.Error("Unsupported provider", "provider", cfg.CloudProvider)
	}

	if isRoutine {
		startScrapeRoutine(repo, ctx, cfg, scrapeFunc, logger) // Go Routine path does not return an error
		return nil
	} else {
		return scrapeFunc(repo, cfg, logger) // One shot path does return error
	}
}

func startScrapeRoutine(repo *repository.Repository, ctx context.Context, cfg model.Config, scrapeFunc func(*repository.Repository, model.Config, *slog.Logger) error, logger *slog.Logger) {
	go func() {
		ticker := time.NewTicker(cfg.ScrapeInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// Break out of Go Routine
				return
			case <-ticker.C:
				err := scrapeFunc(repo, cfg, logger)
				if err != nil {
					logger.Error("An error occured during routine scape:", "error", err)
				}
			}
		}
	}()
}
