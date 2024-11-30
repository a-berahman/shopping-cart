package ports

import (
	"context"

	"github.com/a-berahman/shopping-cart/internal/core/domain"
)

// Queue is the interface for the queue
type Queue interface {
	EnqueueReservation(ctx context.Context, job *domain.ReservationJob) error
	DequeueReservation(ctx context.Context) (*domain.ReservationJob, error)
	CompleteJob(ctx context.Context, job *domain.ReservationJob) error
	FailJob(ctx context.Context, job *domain.ReservationJob) error
	RetryFailedJobs(ctx context.Context) error
}
