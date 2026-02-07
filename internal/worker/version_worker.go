package worker

import (
	"context"
	"time"
)

func StartVersionWorker(w Worker, ctx context.Context) {
	go func() {
		ticker := time.NewTicker(12 * time.Hour)
		defer ticker.Stop()

		// Initial check at startup
		if err := w.UtilService.UpdateVersion(); err != nil {
			w.Logger.Error("Error while checking new version", "error", err)
		}

		for {
			select {
			case <-ctx.Done():
				// Break out of Go Routine
				return
			case <-ticker.C:
				// Get new versions
				if err := w.UtilService.UpdateVersion(); err != nil {
					w.Logger.Error("Error while getting checking new version", "error", err)
				}
			}
		}
	}()
}
