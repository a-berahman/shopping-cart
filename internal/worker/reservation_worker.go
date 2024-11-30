package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/a-berahman/shopping-cart/internal/core/domain"
	"github.com/a-berahman/shopping-cart/internal/core/ports"
	"github.com/google/uuid"
)

// ReservationWorker is responsible for processing reservation jobs from the queue
type ReservationWorker struct {
	queue          ports.Queue
	reservationSvc ports.ReservationService
	repository     ports.Repository
	maxRetries     int
	retryDelay     time.Duration
	shutdownCh     chan struct{}
}

// NewReservationWorker creates a new instance of ReservationWorker
func NewReservationWorker(
	queue ports.Queue,
	reservationSvc ports.ReservationService,
	repository ports.Repository,
) *ReservationWorker {
	return &ReservationWorker{
		queue:          queue,
		reservationSvc: reservationSvc,
		repository:     repository,
		maxRetries:     3,               // TODO: we nee to make this configurable
		retryDelay:     time.Second * 5, // TODO: we nee to make this configurable
		shutdownCh:     make(chan struct{}),
	}
}

// Start begins processing jobs from the queue in a separate goroutine
// Why a separate goroutine?
// - So that we can handle shutdown gracefully
// - So that we can log errors
func (w *ReservationWorker) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done(): // Stop processing if context is cancelled
				return
			case <-w.shutdownCh: // stop processing if shutdown signal is received
				return
			default:
				if err := w.processNextJob(ctx); err != nil {
					log.Printf("Error processing job: %v", err)
				}
			}
		}
	}()
}

// Stop signals the worker to stop processing jobs
func (w *ReservationWorker) Stop() {
	close(w.shutdownCh)
}

// processNextJob processes the next job from the queue with retry logic
func (w *ReservationWorker) processNextJob(ctx context.Context) error {
	job, err := w.queue.DequeueReservation(ctx)
	if err != nil {
		return err
	}

	// check if the job can be retried based on domain rules
	if !job.CanRetry(w.maxRetries) {
		log.Printf("Job %s exceeded maximum retry attempts", job.ID)
		return w.queue.FailJob(ctx, job)
	}

	// attempt to process the job
	err = w.processJob(ctx, job)
	if err != nil {
		job.Attempts++
		job.LastAttempted = time.Now()

		// if job can still be retried, requeue it
		if job.CanRetry(w.maxRetries) {
			log.Printf("Retrying job %s, attempt %d of %d", job.ID, job.Attempts, w.maxRetries)
			return w.queue.EnqueueReservation(ctx, job)
		}

		// job has exhausted all retries
		return w.queue.FailJob(ctx, job)
	}

	return w.queue.CompleteJob(ctx, job)
}

func (w *ReservationWorker) processJob(ctx context.Context, job *domain.ReservationJob) error {
	switch job.JobType {
	case domain.JobTypeAvailabilityCheck:
		return w.processAvailabilityCheck(ctx, job)
	case domain.JobTypeReservation:
		return w.processReservation(ctx, job)
	default:
		return fmt.Errorf("unknown job type: %s", job.JobType)
	}
}

func (w *ReservationWorker) processAvailabilityCheck(ctx context.Context, job *domain.ReservationJob) error {
	available, err := w.reservationSvc.CheckAvailability(ctx, job.ItemName, job.Quantity)
	if err != nil {
		_ = w.repository.UpdateItemStatus(ctx, job.ItemID, domain.StatusReservationFailed)
		return err
	}

	if !available {
		return w.repository.UpdateItemStatus(ctx, job.ItemID, domain.StatusReservationUnavailable)
	}

	reservationJob := &domain.ReservationJob{
		ID:       uuid.New().String(),
		ItemID:   job.ItemID,
		ItemName: job.ItemName,
		Quantity: job.Quantity,
		JobType:  domain.JobTypeReservation,
		Status:   domain.JobStatusPending,
	}

	if err := w.repository.UpdateItemStatus(ctx, job.ItemID, domain.StatusReservationAvailable); err != nil {
		return err
	}

	return w.queue.EnqueueReservation(ctx, reservationJob)
}

func (w *ReservationWorker) processReservation(ctx context.Context, job *domain.ReservationJob) error {
	reservationID, err := w.reservationSvc.ReserveItem(ctx, job.ItemName, job.Quantity)
	if err != nil {
		_ = w.repository.UpdateItemStatus(ctx, job.ItemID, domain.StatusReservationFailed)
		return err
	}

	return w.repository.UpdateItemReservation(ctx, job.ItemID, reservationID)
}
