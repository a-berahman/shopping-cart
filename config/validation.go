package config

import (
	"fmt"
)

func validateConfig(cfg *Config) error {
	validators := []func(*Config) error{
		validateServer,
	}

	for _, validate := range validators {
		if err := validate(cfg); err != nil {
			return err
		}
	}

	if cfg.DatabaseURL == "" {
		return fmt.Errorf("database URL is required")
	}
	if cfg.RedisURL == "" {
		return fmt.Errorf("redis URL is required")
	}
	if cfg.Env != "development" && cfg.Env != "production" && cfg.Env != "test" {
		return fmt.Errorf("invalid environment: %s", cfg.Env)
	}

	return nil
}

func validateServer(cfg *Config) error {
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}
	return nil
}
