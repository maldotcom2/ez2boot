package worker

import (
	"context"
	"ez2boot/internal/provider"
	"time"
)

func StartManageRoutine(w Worker, ctx context.Context, manager provider.Manager) {
	go func() {
		ticker := time.NewTicker(w.Config.InternalClock)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				w.Logger.Debug("Exiting manager", "domain", "worker")
				// Break out of Go Routine
				return
			case <-ticker.C:
				w.Logger.Debug("Running manager", "domain", "worker")
				err := manager.Start()
				if err != nil {
					w.Logger.Error("Failed during managed start", "domain", "worker", "error", err)
				}
				err = manager.Stop()
				if err != nil {
					w.Logger.Error("Failed during managed stop", "domain", "worker", "error", err)
				}
			}
		}
	}()
}
