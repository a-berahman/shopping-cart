package config

import (
	"time"

	"github.com/spf13/viper"
)

func setDefaults(v *viper.Viper) {
	// server default
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)

	// satabase default
	v.SetDefault("database_url", "postgres://postgres:postgres@localhost:5432/shopping_cart?sslmode=disable")

	// redis default
	v.SetDefault("redis_url", "redis://localhost:6379")

	// reservation service default
	v.SetDefault("reservation.timeout", time.Second*30)
	v.SetDefault("reservation.retry_attempts", 3)
	v.SetDefault("reservation.retry_delay", time.Second*5)
	v.SetDefault("reservation.mock_enabled", false)
	v.SetDefault("reservation.mock_latency", time.Second*2)
	v.SetDefault("reservation.mock_failure_rate", 0.1)

}
