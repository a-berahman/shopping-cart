package config

import (
	"time"

	"github.com/spf13/viper"
)

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", time.Second*15)
	v.SetDefault("server.write_timeout", time.Second*15)
	v.SetDefault("server.idle_timeout", time.Second*60)

	// Database defaults
	v.SetDefault("database_url", "postgres://postgres:postgres@localhost:5432/shopping_cart?sslmode=disable")

	// Redis defaults
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 10)

	// Reservation service defaults
	v.SetDefault("reservation.timeout", time.Second*30)
	v.SetDefault("reservation.retry_attempts", 3)
	v.SetDefault("reservation.retry_delay", time.Second*5)
	v.SetDefault("reservation.mock_enabled", false)
	v.SetDefault("reservation.mock_latency", time.Second*2)
	v.SetDefault("reservation.mock_failure_rate", 0.1)

	// Logger defaults
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")
	v.SetDefault("logger.output_path", "stdout")
}
