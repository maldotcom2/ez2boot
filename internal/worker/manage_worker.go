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
				w.Logger.Debug("Exiting manager")
				// Break out of Go Routine
				return
			case <-ticker.C:
				w.Logger.Debug("Running manager")
				err := manager.Start()
				if err != nil {
					w.Logger.Error("An error occured during managed start:", "error", err)
				}
				err = manager.Stop()
				if err != nil {
					w.Logger.Error("An error occured during managed stop:", "error", err)
				}
			}
		}
	}()
}
