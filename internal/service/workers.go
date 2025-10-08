package service

import (
	"context"
	"ez2boot/internal/aws"
	"fmt"
	"time"
)

func ScrapeAndPopulate(ctx context.Context, provider string, interval time.Duration, tagKey string) error {
	// Switch provider specific scrape function
	var scrapeFunc func(string) error
	switch provider {
	case "aws":
		scrapeFunc = aws.GetEC2Instances
	default:
		return fmt.Errorf("Provider %s is not supported", provider)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Break out of Go Routine
			return nil
		case <-ticker.C:
			scrapeFunc(tagKey)
		}
	}
}
