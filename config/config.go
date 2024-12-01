package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config is the configuration for the application
type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	DatabaseURL string            `mapstructure:"database_url"`
	RedisURL    string            `mapstructure:"redis_url"`
	Reservation ReservationConfig `mapstructure:"reservation"`
	Env         string            `mapstructure:"env"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type ReservationConfig struct {
	ServiceURL      string        `mapstructure:"service_url"`
	Timeout         time.Duration `mapstructure:"timeout"`
	RetryAttempts   int           `mapstructure:"retry_attempts"`
	RetryDelay      time.Duration `mapstructure:"retry_delay"`
	MockEnabled     bool          `mapstructure:"mock_enabled"`
	MockLatency     time.Duration `mapstructure:"mock_latency"`
	MockFailureRate float64       `mapstructure:"mock_failure_rate"`
}

type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	OutputPath string `mapstructure:"output_path"`
}

func Load(configPath string) (*Config, error) {
	v := viper.New()

	// we set default values for the configuration
	setDefaults(v)

	// read the configuration file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("/etc/shopping-cart/")
	}

	// read environment variables
	v.AutomaticEnv()
	v.SetEnvPrefix("SHOP")

	// load the configuration
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// validate the configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}
