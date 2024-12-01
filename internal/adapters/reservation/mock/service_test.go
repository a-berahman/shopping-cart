package mock

import (
	"context"
	"strings"
	"testing"
)

func TestMockReservationService_CheckAvailability(t *testing.T) {
	svc := NewMockReservationService(MockConfig{
		LatencyRange: 0,
		FailureRate:  0,
	})

	tests := []struct {
		name        string
		itemName    string
		quantity    int
		wantResult  bool
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid item with sufficient quantity",
			itemName:   "laptop",
			quantity:   5,
			wantResult: true,
			wantErr:    false,
		},
		{
			name:       "valid item with exact quantity",
			itemName:   "laptop",
			quantity:   10,
			wantResult: true,
			wantErr:    false,
		},
		{
			name:       "valid item with insufficient quantity",
			itemName:   "laptop",
			quantity:   11,
			wantResult: false,
			wantErr:    false,
		},
		{
			name:       "non-existent item",
			itemName:   "nonexistent",
			quantity:   1,
			wantResult: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.CheckAvailability(context.Background(), tt.itemName, tt.quantity)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckAvailability() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantResult {
				t.Errorf("CheckAvailability() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func TestMockReservationService_ReserveItem(t *testing.T) {
	svc := NewMockReservationService(MockConfig{
		LatencyRange: 0,
		FailureRate:  0,
	})

	tests := []struct {
		name        string
		itemName    string
		quantity    int
		wantErr     bool
		errContains string
	}{
		{
			name:     "successful reservation",
			itemName: "laptop",
			quantity: 5,
			wantErr:  false,
		},
		{
			name:        "insufficient quantity",
			itemName:    "laptop",
			quantity:    20,
			wantErr:     true,
			errContains: "insufficient inventory",
		},
		{
			name:        "non-existent item",
			itemName:    "nonexistent",
			quantity:    1,
			wantErr:     true,
			errContains: "item not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.ReserveItem(context.Background(), tt.itemName, tt.quantity)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReserveItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ReserveItem() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}
			if got == "" {
				t.Error("ReserveItem() returned empty reservation ID for successful reservation")
			}
		})
	}
}

func TestMockReservationService_WithFailureRate(t *testing.T) {
	svc := NewMockReservationService(MockConfig{
		LatencyRange: 0,
		FailureRate:  1.0, // with this configuration we have 100% failure rate
	})

	if _, err := svc.CheckAvailability(context.Background(), "laptop", 1); err == nil {
		t.Error("CheckAvailability() should fail when FailureRate is 1.0")
	}

	if _, err := svc.ReserveItem(context.Background(), "laptop", 1); err == nil {
		t.Error("ReserveItem() should fail when FailureRate is 1.0")
	}
}
