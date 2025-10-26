package worker

import (
	"context"
	"time"
)

func StartScrapeRoutine(w Worker, ctx context.Context, scrapeFunc func() error) {
	go func() {
		ticker := time.NewTicker(w.Config.ScrapeInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// Break out of Go Routine
				return
			case <-ticker.C:
				err := scrapeFunc()
				if err != nil {
					w.Logger.Error("An error occured during routine scape:", "error", err)
				}
			}
		}
	}()
}
