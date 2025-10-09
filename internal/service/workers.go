package service

import (
	"context"
	"ez2boot/internal/aws"
	"ez2boot/internal/repository"
	"log/slog"
	"time"
)

// TODO Requires new struct to reduce number of parameters here
func ScrapeAndPopulate(repo *repository.Repository, ctx context.Context, provider string, interval time.Duration, tagKey string, isRoutine bool, logger *slog.Logger) error {
	// Switch provider specific scrape function
	var scrapeFunc func(*repository.Repository, string, *slog.Logger) error
	switch provider {
	case "aws":
		scrapeFunc = aws.GetEC2Instances
	default:
		logger.Error("Unsupported provider", "provider", provider)
	}

	if isRoutine {
		startScrapeRoutine(repo, ctx, interval, tagKey, scrapeFunc, logger) // Go Routine path does not return an error
		return nil
	} else {
		return scrapeFunc(repo, tagKey, logger) // One shot path does return error
	}
}

func startScrapeRoutine(repo *repository.Repository, ctx context.Context, interval time.Duration, tagKey string, scrapeFunc func(*repository.Repository, string, *slog.Logger) error, logger *slog.Logger) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// Break out of Go Routine
				return
			case <-ticker.C:
				err := scrapeFunc(repo, tagKey, logger)
				if err != nil {
					logger.Error("An error occured during routine scape:", "error", err)
				}
			}
		}
	}()
}
