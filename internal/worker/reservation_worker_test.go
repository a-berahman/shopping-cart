package worker

import (
	"context"
	"testing"

	"github.com/a-berahman/shopping-cart/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockQueue struct {
	mock.Mock
}

func (m *MockQueue) EnqueueReservation(ctx context.Context, job *domain.ReservationJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *MockQueue) DequeueReservation(ctx context.Context) (*domain.ReservationJob, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ReservationJob), args.Error(1)
}

func (m *MockQueue) CompleteJob(ctx context.Context, job *domain.ReservationJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *MockQueue) FailJob(ctx context.Context, job *domain.ReservationJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *MockQueue) RetryFailedJobs(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockReservationService struct {
	mock.Mock
}

func (m *MockReservationService) CheckAvailability(ctx context.Context, itemName string, quantity int) (bool, error) {
	args := m.Called(ctx, itemName, quantity)
	return args.Bool(0), args.Error(1)
}

func (m *MockReservationService) ReserveItem(ctx context.Context, itemName string, quantity int) (string, error) {
	args := m.Called(ctx, itemName, quantity)
	return args.String(0), args.Error(1)
}

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateItem(ctx context.Context, item *domain.Item) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockRepository) ListItems(ctx context.Context) ([]domain.Item, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Item), args.Error(1)
}

func (m *MockRepository) UpdateItemStatus(ctx context.Context, id int64, status domain.ItemStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockRepository) GetItem(ctx context.Context, id int64) (*domain.Item, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Item), args.Error(1)
}

func (m *MockRepository) UpdateItemReservation(ctx context.Context, id int64, reservationID string) error {
	args := m.Called(ctx, id, reservationID)
	return args.Error(0)
}

func TestProcessNextJob(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockQueue, *MockReservationService, *MockRepository)
		wantErr    bool
	}{
		{
			name: "successful availability check",
			setupMocks: func(queue *MockQueue, resSvc *MockReservationService, repo *MockRepository) {
				job := &domain.ReservationJob{
					ID:       "test-job",
					ItemID:   1,
					ItemName: "Test Item",
					Quantity: 1,
					JobType:  domain.JobTypeAvailabilityCheck,
					Status:   domain.JobStatusPending,
				}

				queue.On("DequeueReservation", mock.Anything).Return(job, nil)
				resSvc.On("CheckAvailability", mock.Anything, "Test Item", 1).Return(true, nil)
				repo.On("UpdateItemStatus", mock.Anything, int64(1), domain.StatusReservationAvailable).Return(nil)
				queue.On("EnqueueReservation", mock.Anything, mock.MatchedBy(func(j *domain.ReservationJob) bool {
					return j.JobType == domain.JobTypeReservation
				})).Return(nil)
				queue.On("CompleteJob", mock.Anything, job).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "successful reservation",
			setupMocks: func(queue *MockQueue, resSvc *MockReservationService, repo *MockRepository) {
				job := &domain.ReservationJob{
					ID:       "test-job",
					ItemID:   1,
					ItemName: "Test Item",
					Quantity: 1,
					JobType:  domain.JobTypeReservation,
					Status:   domain.JobStatusPending,
				}

				queue.On("DequeueReservation", mock.Anything).Return(job, nil)
				resSvc.On("ReserveItem", mock.Anything, "Test Item", 1).Return("res123", nil)
				repo.On("UpdateItemReservation", mock.Anything, int64(1), "res123").Return(nil)
				queue.On("CompleteJob", mock.Anything, job).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "item not available",
			setupMocks: func(queue *MockQueue, resSvc *MockReservationService, repo *MockRepository) {
				job := &domain.ReservationJob{
					ID:       "test-job",
					ItemID:   1,
					ItemName: "Test Item",
					Quantity: 1,
					JobType:  domain.JobTypeAvailabilityCheck,
					Status:   domain.JobStatusPending,
				}

				queue.On("DequeueReservation", mock.Anything).Return(job, nil)
				resSvc.On("CheckAvailability", mock.Anything, "Test Item", 1).Return(false, nil)
				repo.On("UpdateItemStatus", mock.Anything, int64(1), domain.StatusReservationUnavailable).Return(nil)
				queue.On("CompleteJob", mock.Anything, job).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "retry exceeded",
			setupMocks: func(queue *MockQueue, resSvc *MockReservationService, repo *MockRepository) {
				job := &domain.ReservationJob{
					ID:       "test-job",
					ItemID:   1,
					ItemName: "Test Item",
					Quantity: 1,
					JobType:  domain.JobTypeAvailabilityCheck,
					Status:   domain.JobStatusPending,
					Attempts: 3,
				}

				queue.On("DequeueReservation", mock.Anything).Return(job, nil)
				queue.On("FailJob", mock.Anything, job).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := new(MockQueue)
			resSvc := new(MockReservationService)
			repo := new(MockRepository)
			tt.setupMocks(queue, resSvc, repo)

			worker := NewReservationWorker(queue, resSvc, repo)
			err := worker.processNextJob(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			queue.AssertExpectations(t)
			resSvc.AssertExpectations(t)
			repo.AssertExpectations(t)
		})
	}
}
