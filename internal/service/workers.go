package service

import (
	"context"
	"time"
)

// Import to DB
func ScrapeAndPopulate(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Break out of Go Routine
			return
		case <-ticker.C:
			// Run the downstream function determined by env var
		}
	}
}
