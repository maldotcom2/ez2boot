package service

import (
	"context"
	"ez2boot/internal/model"
	"ez2boot/internal/repository"
	"log/slog"
	"time"
)

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
