package service

import (
	"context"

	"github.com/a-berahman/shopping-cart/internal/core/domain"
	"github.com/a-berahman/shopping-cart/internal/core/ports"

	"github.com/google/uuid"
)

type CartService struct {
	repo           ports.Repository
	queue          ports.Queue
	reservationSvc ports.ReservationService
}

// NewCartService creates a new cart service
func NewCartService(
	repo ports.Repository,
	queue ports.Queue,
	reservationSvc ports.ReservationService,
) *CartService {
	return &CartService{
		repo:           repo,
		queue:          queue,
		reservationSvc: reservationSvc,
	}
}

func (s *CartService) AddItemToCart(ctx context.Context, name string, quantity int) (*domain.Item, error) {

	// first we create the item in pending state
	item := &domain.Item{
		Name:     name,
		Quantity: quantity,
		Status:   domain.StatusReservationPending,
	}

	if err := s.repo.CreateItem(ctx, item); err != nil {
		return nil, err
	}

	// create and enqueue reservation job
	job := &domain.ReservationJob{
		ID:       uuid.New().String(),
		ItemID:   item.ID,
		ItemName: name,
		Quantity: quantity,
		JobType:  domain.JobTypeAvailabilityCheck,
		Status:   domain.JobStatusPending,
	}

	// why enqueue? because we want to reserve the item in the background
	// so that we can return the item to the user immediately
	if err := s.queue.EnqueueReservation(ctx, job); err != nil {
		// If enqueueing fails, we should mark the item as failed
		item.Status = domain.StatusReservationFailed
		_ = s.repo.UpdateItemStatus(ctx, item.ID, domain.StatusReservationFailed)
		return nil, err
	}

	return item, nil
}

func (s *CartService) ListCartItems(ctx context.Context) ([]domain.Item, error) {
	return s.repo.ListItems(ctx)
}
