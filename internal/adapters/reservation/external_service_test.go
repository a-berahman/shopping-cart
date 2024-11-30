package reservation

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T, handler http.HandlerFunc) (*Service, *httptest.Server) {
	server := httptest.NewServer(handler)
	service := NewService(server.URL)
	service.httpClient.Timeout = 1 * time.Second
	return service, server
}

func TestCheckAvailability(t *testing.T) {
	tests := []struct {
		name       string
		itemName   string
		quantity   int
		mockServer func() (http.HandlerFunc, *reservationResponse)
		want       bool
		wantErr    bool
	}{
		{
			name:     "item is available",
			itemName: "Test Item",
			quantity: 1,
			mockServer: func() (http.HandlerFunc, *reservationResponse) {
				response := &reservationResponse{Available: true}
				return func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodGet, r.Method)
					assert.Equal(t, "/availability", r.URL.Path)

					var req reservationRequest
					err := json.NewDecoder(r.Body).Decode(&req)
					require.NoError(t, err)
					assert.Equal(t, "Test Item", req.Item)
					assert.Equal(t, 1, req.Quantity)

					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(response)
				}, response
			},
			want:    true,
			wantErr: false,
		},
		{
			name:     "item is not available",
			itemName: "Test Item",
			quantity: 100,
			mockServer: func() (http.HandlerFunc, *reservationResponse) {
				response := &reservationResponse{Available: false}
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(response)
				}, response
			},
			want:    false,
			wantErr: false,
		},
		{
			name:     "server error",
			itemName: "Test Item",
			quantity: 1,
			mockServer: func() (http.HandlerFunc, *reservationResponse) {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}, nil
			},
			want:    false,
			wantErr: true,
		},
		{
			name:     "invalid response",
			itemName: "Test Item",
			quantity: 1,
			mockServer: func() (http.HandlerFunc, *reservationResponse) {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("invalid json"))
				}, nil
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _ := tt.mockServer()
			service, server := setupTestServer(t, handler)
			defer server.Close()

			got, err := service.CheckAvailability(context.Background(), tt.itemName, tt.quantity)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReserveItem(t *testing.T) {
	tests := []struct {
		name       string
		itemName   string
		quantity   int
		mockServer func() (http.HandlerFunc, *reservationResponse)
		want       string
		wantErr    bool
	}{
		{
			name:     "successful reservation",
			itemName: "Test Item",
			quantity: 1,
			mockServer: func() (http.HandlerFunc, *reservationResponse) {
				response := &reservationResponse{ReservationID: "res123"}
				return func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodPost, r.Method)
					assert.Equal(t, "/reserve", r.URL.Path)

					var req reservationRequest
					err := json.NewDecoder(r.Body).Decode(&req)
					require.NoError(t, err)
					assert.Equal(t, "Test Item", req.Item)
					assert.Equal(t, 1, req.Quantity)

					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(response)
				}, response
			},
			want:    "res123",
			wantErr: false,
		},
		{
			name:     "reservation failed",
			itemName: "Test Item",
			quantity: 100,
			mockServer: func() (http.HandlerFunc, *reservationResponse) {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
				}, nil
			},
			want:    "",
			wantErr: true,
		},
		{
			name:     "server error",
			itemName: "Test Item",
			quantity: 1,
			mockServer: func() (http.HandlerFunc, *reservationResponse) {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}, nil
			},
			want:    "",
			wantErr: true,
		},
		{
			name:     "timeout error",
			itemName: "Test Item",
			quantity: 1,
			mockServer: func() (http.HandlerFunc, *reservationResponse) {
				return func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(2 * time.Second) // Will trigger timeout
				}, nil
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _ := tt.mockServer()
			service, server := setupTestServer(t, handler)
			defer server.Close()

			got, err := service.ReserveItem(context.Background(), tt.itemName, tt.quantity)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
