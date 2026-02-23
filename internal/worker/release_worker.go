package worker

import (
	"context"
	"time"
)

func StartReleaseWorker(w Worker, ctx context.Context) {
	go func() {
		ticker := time.NewTicker(12 * time.Hour)
		defer ticker.Stop()

		// Initial check at startup
		if err := w.UtilService.CheckRelease(); err != nil {
			w.Logger.Error("Failed to check for new releases", "domain", "worker", "error", err)
		}

		for {
			select {
			case <-ctx.Done():
				w.Logger.Debug("Exiting release worker", "domain", "worker")
				// Break out of Go Routine
				return
			case <-ticker.C:
				w.Logger.Debug("Running release worker", "domain", "worker")
				// Get new releases
				if err := w.UtilService.CheckRelease(); err != nil {
					w.Logger.Error("Failed to check for new releases", "domain", "worker", "error", err)
				}
			}
		}
	}()
}
