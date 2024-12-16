package execreceiver

import (
	"fmt"
	"time"
)

// Config represents the receiver config settings within the collector's config.yaml
type Config struct {
	Interval    string `mapstructure:"interval"`
	Script string `mapstructure:"script"`
}

// Validate checks if the receiver configuration is valid
func (cfg *Config) Validate() error {
	interval, _ := time.ParseDuration(cfg.Interval)
	if interval.Minutes() < 1 {
		return fmt.Errorf("when defined, the interval has to be set to at least 1 minute (1m)")
	}

	if cfg.Script == "" {
		return fmt.Errorf("script path must be defined")
	}
	return nil
}