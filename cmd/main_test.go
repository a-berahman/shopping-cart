package main

import (
	"context"
	"log/slog"
	"testing"

	"github.com/a-berahman/shopping-cart/config"
	"github.com/go-playground/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializeApplication(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name: "invalid database URL",
			cfg: &config.Config{
				Env:         "development",
				DatabaseURL: "invalid-url",
				RedisURL:    "redis://localhost:6379",
			},
			wantErr: true,
		},
		{
			name: "invalid redis URL",
			cfg: &config.Config{
				Env:         "development",
				DatabaseURL: "postgres://localhost:5432/test_db?sslmode=disable",
				RedisURL:    "invalid-url",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()
			ctx := context.Background()

			app, err := initializeApplication(ctx, logger, tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, app)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, app)
			require.NotNil(t, app.server)
			require.NotNil(t, app.worker)
		})
	}
}

func TestCustomValidator(t *testing.T) {
	type TestStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
	}

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name: "valid struct",
			input: TestStruct{
				Name:  "TEST TEST",
				Email: "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			input: TestStruct{
				Name:  "TEST TEST",
				Email: "invalid-email",
			},
			wantErr: true,
		},
		{
			name: "missing required fields",
			input: TestStruct{
				Name:  "",
				Email: "",
			},
			wantErr: true,
		},
	}

	validator := &CustomValidator{Validator: validator.New()}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
