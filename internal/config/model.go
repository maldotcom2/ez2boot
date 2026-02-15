package config

import (
	"log/slog"
	"time"
)

type Config struct {
	SetupMode           bool          // Mode which allows initial user bootstrap, not manually setable
	TrustProxyHeaders   bool          // Affects source IP address recognition within middleware
	CloudProvider       string        // Cloud provider eg aws, azure
	Port                string        // Listener port for this application
	ScrapeInterval      time.Duration // Interval for scraping cloud provider
	InternalClock       time.Duration // Interval for all other background workers
	TagKey              string        // Tag Key used to itentify target servers, where the values are the server groups
	AWSRegion           string        // AWS Region, AWS scrape specific
	UserSessionDuration time.Duration // Duration for user UI authenticated session, not related to server session duration
	LogLevel            slog.Level    // Logging level, use info unless debugging
	EncryptionPhrase    string        // Implementation specific encryption phrase used to derive an encryption key to encrypt sensitive credentials within the app
	RateLimit           int           // Max number of requests per second allowed by each user of this application
	ShowBetaVersions    bool          // UI will show alert for beta releases and not just full releases
	// Add more fields as needed
}
