package mock

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// MockReservationService implements the ReservationService interface for demonstration
type MockReservationService struct {
	inventory    map[string]int
	reservations map[string]bool
	latencyRange time.Duration
	failureRate  float64
	mu           sync.RWMutex
}

// NewMockReservationService creates a new mock service with configurable behavior
func NewMockReservationService(config MockConfig) *MockReservationService {
	return &MockReservationService{
		inventory: map[string]int{
			"laptop":     10,
			"phone":      20,
			"tablet":     15,
			"headphones": 30,
		},
		reservations: make(map[string]bool),
		latencyRange: config.LatencyRange,
		failureRate:  config.FailureRate,
	}
}

type MockConfig struct {
	LatencyRange time.Duration // Maximum artificial delay
	FailureRate  float64       // Rate of simulated failures (0.0 to 1.0)
}

func (m *MockReservationService) CheckAvailability(ctx context.Context, itemName string, quantity int) (bool, error) {
	// Simulate network latency
	m.simulateLatency()

	// Simulate potential failures
	if m.shouldFail() {
		return false, fmt.Errorf("service temporarily unavailable")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	available, exists := m.inventory[itemName]
	if !exists {
		return false, nil
	}

	return available >= quantity, nil
}

func (m *MockReservationService) ReserveItem(ctx context.Context, itemName string, quantity int) (string, error) {
	// Simulate network latency
	m.simulateLatency()

	// Simulate potential failures
	if m.shouldFail() {
		return "", fmt.Errorf("reservation failed: service temporarily unavailable")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	available, exists := m.inventory[itemName]
	if !exists {
		return "", fmt.Errorf("item not found")
	}

	if available < quantity {
		return "", fmt.Errorf("insufficient inventory")
	}

	// Generate a reservation ID
	reservationID := fmt.Sprintf("RSV-%s-%d", itemName, time.Now().Unix())

	// Update inventory
	m.inventory[itemName] = available - quantity
	m.reservations[reservationID] = true

	return reservationID, nil
}

func (m *MockReservationService) simulateLatency() {
	if m.latencyRange > 0 {
		delay := time.Duration(rand.Int63n(int64(m.latencyRange)))
		time.Sleep(delay)
	}
}

func (m *MockReservationService) shouldFail() bool {
	return rand.Float64() < m.failureRate
}

// Additional methods for demonstration control
func (m *MockReservationService) SetInventory(itemName string, quantity int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inventory[itemName] = quantity
}

func (m *MockReservationService) GetInventory(itemName string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.inventory[itemName]
}
