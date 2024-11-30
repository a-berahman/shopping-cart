package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {

	tmpDir, err := os.MkdirTemp("", "config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	validConfig := []byte(`
server:
  host: "localhost"
  port: 8080
  read_timeout: "5s"
  write_timeout: "10s"
  idle_timeout: "15s"
database_url: "postgres://user:pass@localhost:5432/db?sslmode=disable"
redis_url: "redis://localhost:6379"
reservation:
  service_url: "http://reservation:8080"
  timeout: "5s"
  retry_attempts: 3
  retry_delay: "1s"
  mock_enabled: true
  mock_latency: "100ms"
  mock_failure_rate: 0.1
logger:
  level: "info"
  format: "json"
  output_path: "stdout"
env: "development"
`)

	validConfigPath := filepath.Join(tmpDir, "valid-config.yaml")
	err = os.WriteFile(validConfigPath, validConfig, 0644)
	require.NoError(t, err)

	invalidConfig := []byte(`
server:
  port: "not-a-number"
`)

	invalidConfigPath := filepath.Join(tmpDir, "invalid-config.yaml")
	err = os.WriteFile(invalidConfigPath, invalidConfig, 0644)
	require.NoError(t, err)

	tests := []struct {
		name       string
		configPath string
		envVars    map[string]string
		want       *Config
		wantErr    bool
	}{
		{
			name:       "valid config file",
			configPath: validConfigPath,
			want: &Config{
				Server: ServerConfig{
					Host:         "localhost",
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 10 * time.Second,
					IdleTimeout:  15 * time.Second,
				},
				DatabaseURL: "postgres://user:pass@localhost:5432/db?sslmode=disable",
				RedisURL:    "redis://localhost:6379",
				Reservation: ReservationConfig{
					ServiceURL:      "http://reservation:8080",
					Timeout:         5 * time.Second,
					RetryAttempts:   3,
					RetryDelay:      time.Second,
					MockEnabled:     true,
					MockLatency:     100 * time.Millisecond,
					MockFailureRate: 0.1,
				},
				Logger: LoggerConfig{
					Level:      "info",
					Format:     "json",
					OutputPath: "stdout",
				},
				Env: "development",
			},
			wantErr: false,
		},
		{
			name:       "invalid config file",
			configPath: invalidConfigPath,
			wantErr:    true,
		},
		{
			name:       "non-existent config file",
			configPath: "non-existent.yaml",
			wantErr:    true,
		},
		{
			name:       "environment variables override",
			configPath: validConfigPath,
			envVars: map[string]string{
				"SHOP_SERVER_PORT":         "9090",
				"SHOP_DATABASE_URL":        "postgres://other:pass@otherhost:5432/db",
				"SHOP_RESERVATION_TIMEOUT": "10s",
			},
			want: &Config{
				Server: ServerConfig{
					Host:         "localhost",
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 10 * time.Second,
					IdleTimeout:  15 * time.Second,
				},
				DatabaseURL: "postgres://other:pass@otherhost:5432/db",
				RedisURL:    "redis://localhost:6379",
				Reservation: ReservationConfig{
					ServiceURL:      "http://reservation:8080",
					Timeout:         5 * time.Second,
					RetryAttempts:   3,
					RetryDelay:      time.Second,
					MockEnabled:     true,
					MockLatency:     100 * time.Millisecond,
					MockFailureRate: 0.1,
				},
				Logger: LoggerConfig{
					Level:      "info",
					Format:     "json",
					OutputPath: "stdout",
				},
				Env: "development",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set environment variables for the test
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			got, err := Load(tt.configPath)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.want != nil {
				assert.Equal(t, tt.want.Server, got.Server)
				assert.Equal(t, tt.want.DatabaseURL, got.DatabaseURL)
				assert.Equal(t, tt.want.RedisURL, got.RedisURL)
				assert.Equal(t, tt.want.Reservation, got.Reservation)
				assert.Equal(t, tt.want.Logger, got.Logger)
				assert.Equal(t, tt.want.Env, got.Env)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				DatabaseURL: "postgres://user:pass@localhost:5432/db",
				RedisURL:    "redis://localhost:6379",
				Server: ServerConfig{
					Port: 8080,
				},
				Env: "development",
			},
			wantErr: false,
		},
		{
			name: "missing database URL",
			config: &Config{
				RedisURL: "redis://localhost:6379",
				Server: ServerConfig{
					Port: 8080,
				},
				Env: "development",
			},
			wantErr: true,
		},
		{
			name: "missing redis URL",
			config: &Config{
				DatabaseURL: "postgres://user:pass@localhost:5432/db",
				Server: ServerConfig{
					Port: 8080,
				},
				Env: "development",
			},
			wantErr: true,
		},
		{
			name: "invalid environment",
			config: &Config{
				DatabaseURL: "postgres://user:pass@localhost:5432/db",
				RedisURL:    "redis://localhost:6379",
				Server: ServerConfig{
					Port: 8080,
				},
				Env: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
