package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port        string `envconfig:"PORT" default:"8080"`
	Environment string `envconfig:"ENV" default:"development"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`

	// Backend Client
	BackendURL    string `envconfig:"BACKEND_URL" default:"http://localhost:5007"`
	APIKey        string `envconfig:"INTERNAL_API_KEY" required:"true"`
	RetryAttempts int    `envconfig:"RETRY_ATTEMPTS" default:"3"`

	// Rate Limiting (per minute, per IP address)
	EnableRateLimit    bool `envconfig:"ENABLE_RATE_LIMIT" default:"true"`
	RateLimitPerMinute int  `envconfig:"RATE_LIMIT_PER_MINUTE" default:"100"`

	// Cache Configuration
	EnableCache bool          `envconfig:"ENABLE_CACHE" default:"true"`
	CachePath   string        `envconfig:"CACHE_PATH" default:"./cache/pdf-reports"`
	CacheTTL    time.Duration `envconfig:"CACHE_TTL" default:"1h"`
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	return &cfg, nil
}
