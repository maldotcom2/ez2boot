package service

import (
	"context"
	"ez2boot/internal/aws"
	"fmt"
	"log/slog"
	"time"
)

func ScrapeAndPopulate(ctx context.Context, provider string, interval time.Duration, tagKey string, isRoutine bool, logger *slog.Logger) error {
	// Switch provider specific scrape function
	var scrapeFunc func(string) error
	switch provider {
	case "aws":
		scrapeFunc = aws.GetEC2Instances
	default:
		return fmt.Errorf("Provider %s is not supported", provider)
	}

	if isRoutine {
		startScrapeRoutine(ctx, interval, tagKey, scrapeFunc, logger) // Go Routine path does not return an error
		return nil
	} else {
		return scrapeFunc(tagKey) // One shot path does return error
	}
}

func startScrapeRoutine(ctx context.Context, interval time.Duration, tagKey string, scrapeFunc func(string) error, logger *slog.Logger) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// Break out of Go Routine
				return
			case <-ticker.C:
				err := scrapeFunc(tagKey)
				if err != nil {
					logger.Error("An error occured during routine scape:", "error", err)
				}
			}
		}
	}()
}
