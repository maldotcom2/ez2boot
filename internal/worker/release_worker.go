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
			w.Logger.Error("Error while checking for new releases", "error", err)
		}

		for {
			select {
			case <-ctx.Done():
				// Break out of Go Routine
				return
			case <-ticker.C:
				// Get new releases
				if err := w.UtilService.CheckRelease(); err != nil {
					w.Logger.Error("Error while checking for new releases", "error", err)
				}
			}
		}
	}()
}
