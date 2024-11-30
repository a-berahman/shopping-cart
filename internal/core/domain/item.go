package domain

import (
	"time"
)

// ItemStatus is the status of an item
type ItemStatus string

const (
	StatusReservationPending           ItemStatus = "PENDING"
	StatusReservationAvailabilityCheck ItemStatus = "AVAILABILITY_CHECK"
	StatusReservationAvailable         ItemStatus = "AVAILABLE"
	StatusReservationUnavailable       ItemStatus = "UNAVAILABLE"
	StatusReservationReserved          ItemStatus = "RESERVED"
	StatusReservationFailed            ItemStatus = "FAILED"
)

// Add method to check if item can be shown as potentially available
func (i *Item) IsAvailable() bool {
	return i.Status == StatusReservationAvailable ||
		i.Status == StatusReservationPending ||
		i.Status == StatusReservationAvailabilityCheck
}

// Item is the domain object for an item
type Item struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	Quantity      int        `json:"quantity"`
	ReservationID *string    `json:"reservation_id,omitempty"`
	Status        ItemStatus `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
