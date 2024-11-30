package ports

import (
	"context"

	"github.com/a-berahman/shopping-cart/internal/core/domain"
)

// ReservationService is the interface for the reservation service
type ReservationService interface {
	CheckAvailability(ctx context.Context, itemName string, quantity int) (bool, error)
	ReserveItem(ctx context.Context, itemName string, quantity int) (string, error)
}

// CartService is the interface for the cart service
type CartService interface {
	AddItemToCart(ctx context.Context, name string, quantity int) (*domain.Item, error)
	ListCartItems(ctx context.Context) ([]domain.Item, error)
}
