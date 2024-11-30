package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/a-berahman/shopping-cart/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateItem(ctx context.Context, item *domain.Item) error {
	args := m.Called(ctx, item)
	if args.Get(0) != nil {
		item.ID = 1
		item.CreatedAt = time.Now()
		item.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *MockRepository) ListItems(ctx context.Context) ([]domain.Item, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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

type MockQueue struct {
	mock.Mock
}

func (m *MockQueue) EnqueueReservation(ctx context.Context, job *domain.ReservationJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
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

func (m *MockQueue) DequeueReservation(ctx context.Context) (*domain.ReservationJob, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ReservationJob), args.Error(1)
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

func TestAddItemToCart(t *testing.T) {
	tests := []struct {
		name          string
		itemName      string
		quantity      int
		setupMocks    func(*MockRepository, *MockQueue, *MockReservationService)
		expectedItem  *domain.Item
		expectedError error
	}{
		{
			name:     "successful item addition",
			itemName: "Test Item",
			quantity: 1,
			setupMocks: func(repo *MockRepository, queue *MockQueue, resSvc *MockReservationService) {
				repo.On("CreateItem", mock.Anything, mock.MatchedBy(func(item *domain.Item) bool {
					return item.Name == "Test Item" && item.Quantity == 1 &&
						item.Status == domain.StatusReservationPending
				})).Run(func(args mock.Arguments) {
					item := args.Get(1).(*domain.Item)
					item.ID = 1
					item.CreatedAt = time.Now()
					item.UpdatedAt = time.Now()
				}).Return(nil)

				queue.On("EnqueueReservation", mock.Anything, mock.MatchedBy(func(job *domain.ReservationJob) bool {
					return job.ItemName == "Test Item" && job.Quantity == 1 &&
						job.Status == domain.JobStatusPending
				})).Return(nil)
			},
			expectedItem: &domain.Item{
				ID:       1,
				Name:     "Test Item",
				Quantity: 1,
				Status:   domain.StatusReservationPending,
			},
			expectedError: nil,
		},
		{
			name:     "repository error",
			itemName: "Test Item",
			quantity: 1,
			setupMocks: func(repo *MockRepository, queue *MockQueue, resSvc *MockReservationService) {
				repo.On("CreateItem", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedItem:  nil,
			expectedError: errors.New("db error"),
		},
		{
			name:     "queue error",
			itemName: "Test Item",
			quantity: 1,
			setupMocks: func(repo *MockRepository, queue *MockQueue, resSvc *MockReservationService) {
				repo.On("CreateItem", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					item := args.Get(1).(*domain.Item)
					item.ID = 1
				}).Return(nil)
				queue.On("EnqueueReservation", mock.Anything, mock.Anything).Return(errors.New("queue error"))
				repo.On("UpdateItemStatus", mock.Anything, int64(1), domain.StatusReservationFailed).Return(nil)
			},
			expectedItem:  nil,
			expectedError: errors.New("queue error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			repo := new(MockRepository)
			queue := new(MockQueue)
			resSvc := new(MockReservationService)
			tt.setupMocks(repo, queue, resSvc)

			service := NewCartService(repo, queue, resSvc)
			item, err := service.AddItemToCart(context.Background(), tt.itemName, tt.quantity)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, item)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, item)
				assert.Equal(t, tt.expectedItem.Name, item.Name)
				assert.Equal(t, tt.expectedItem.Quantity, item.Quantity)
				assert.Equal(t, tt.expectedItem.Status, item.Status)
				assert.NotZero(t, item.ID)
				assert.NotZero(t, item.CreatedAt)
				assert.NotZero(t, item.UpdatedAt)
			}

			repo.AssertExpectations(t)
			queue.AssertExpectations(t)
			resSvc.AssertExpectations(t)
		})
	}
}

func TestListCartItems(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*MockRepository)
		expectedItems []domain.Item
		expectedError error
	}{
		{
			name: "successful items listing",
			setupMocks: func(repo *MockRepository) {
				items := []domain.Item{
					{
						ID:       1,
						Name:     "Item 1",
						Quantity: 1,
						Status:   domain.StatusReservationPending,
					},
					{
						ID:       2,
						Name:     "Item 2",
						Quantity: 2,
						Status:   domain.StatusReservationReserved,
					},
				}
				repo.On("ListItems", mock.Anything).Return(items, nil)
			},
			expectedItems: []domain.Item{
				{
					ID:       1,
					Name:     "Item 1",
					Quantity: 1,
					Status:   domain.StatusReservationPending,
				},
				{
					ID:       2,
					Name:     "Item 2",
					Quantity: 2,
					Status:   domain.StatusReservationReserved,
				},
			},
			expectedError: nil,
		},
		{
			name: "repository error",
			setupMocks: func(repo *MockRepository) {
				repo.On("ListItems", mock.Anything).Return(nil, errors.New("db error"))
			},
			expectedItems: nil,
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			repo := new(MockRepository)
			tt.setupMocks(repo)

			service := NewCartService(repo, nil, nil)

			items, err := service.ListCartItems(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, items)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedItems, items)
			}

			repo.AssertExpectations(t)
		})
	}
}
