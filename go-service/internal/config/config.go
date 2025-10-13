package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all application configuration
type Config struct {
	Port        string `envconfig:"PORT" default:"8080"`
	BackendURL  string `envconfig:"BACKEND_URL" default:"http://localhost:5007"`
	APIKey      string `envconfig:"INTERNAL_API_KEY" required:"true"`
	Environment string `envconfig:"ENV" default:"development"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	return &cfg, nil
}
