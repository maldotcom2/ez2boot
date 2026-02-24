package worker

import (
	"context"
	"ez2boot/internal/provider"
	"time"
)

func StartScrapeRoutine(w Worker, ctx context.Context, scraper provider.Scraper) {
	go func() {
		ticker := time.NewTicker(w.Config.ScrapeInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				w.Logger.Debug("Exiting scraper", "domain", "worker")
				// Break out of Go Routine
				return
			case <-ticker.C:
				w.Logger.Debug("Running scraper", "domain", "worker")
				err := scraper.Scrape()
				if err != nil {
					w.Logger.Error("Failed to scrape", "domain", "worker", "error", err)
				}
			}
		}
	}()
}
