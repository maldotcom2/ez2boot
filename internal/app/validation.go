package app

import (
	"errors"
	"ez2boot/internal/config"
	"fmt"
)

// Validates required env vars have been provided
func validateProviderConfig(cfg *config.Config) error {
	switch cfg.CloudProvider {
	case "aws":
		if cfg.AWSRegion == "" {
			return errors.New("AWS_REGION environment variable is required")
		}
	case "azure":
		if cfg.AzureSubscriptionID == "" {
			return errors.New("AZURE_SUBSCRIPTION_ID is required")
		}
	default:
		return fmt.Errorf("unsupported value for CLOUD_PROVIDER (supported aws, azure): %s", cfg.CloudProvider)
	}

	return nil
}
