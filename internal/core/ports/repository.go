package ports

import (
	"context"

	"github.com/a-berahman/shopping-cart/internal/core/domain"
)

// Repository is the interface for the repository
type Repository interface {
	CreateItem(ctx context.Context, item *domain.Item) error
	ListItems(ctx context.Context) ([]domain.Item, error)
	UpdateItemReservation(ctx context.Context, id int64, reservationID string) error
	GetItem(ctx context.Context, id int64) (*domain.Item, error)
	UpdateItemStatus(ctx context.Context, id int64, status domain.ItemStatus) error
}
