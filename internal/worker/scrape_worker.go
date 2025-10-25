package worker

import (
	"context"
	"ez2boot/internal/db"
	"ez2boot/internal/model"
	"log/slog"
	"time"
)

func startScrapeRoutine(repo *db.Repository, ctx context.Context, cfg model.Config, scrapeFunc func(*db.Repository, model.Config, *slog.Logger) error, logger *slog.Logger) {
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
