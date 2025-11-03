package worker

import (
	"context"
	"ez2boot/internal/provider"
	"time"
)

func StartScrapeRoutine(w Worker, ctx context.Context, scraper provider.Scraper) {
	w.Logger.Debug("Running scraper")
	go func() {
		ticker := time.NewTicker(w.Config.ScrapeInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				w.Logger.Debug("Exiting scraper")
				// Break out of Go Routine
				return
			case <-ticker.C:
				err := scraper.Scrape()
				if err != nil {
					w.Logger.Error("An error occured during routine scape:", "error", err)
				}
			}
		}
	}()
}
