package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/a-berahman/shopping-cart/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*Repository, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := NewRepository(db)
	return repo, mock
}

func TestCreateItem(t *testing.T) {
	repo, mock := setupTestDB(t)
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name    string
		item    *domain.Item
		setup   func(sqlmock.Sqlmock, *domain.Item)
		wantErr bool
	}{
		{
			name: "successful creation",
			item: &domain.Item{
				Name:     "Test Item",
				Quantity: 1,
				Status:   domain.StatusReservationPending,
			},
			setup: func(mock sqlmock.Sqlmock, item *domain.Item) {
				columns := []string{"id", "name", "quantity", "reservation_id", "status", "created_at", "updated_at"}
				mock.ExpectQuery(`INSERT INTO items (.+) RETURNING *`).
					WithArgs(item.Name, int32(item.Quantity), string(item.Status)).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(1, item.Name, int32(item.Quantity), sql.NullString{}, string(item.Status), now, now),
					)
			},
			wantErr: false,
		},
		{
			name: "database error",
			item: &domain.Item{
				Name:     "Test Item",
				Quantity: 1,
				Status:   domain.StatusReservationPending,
			},
			setup: func(mock sqlmock.Sqlmock, item *domain.Item) {
				mock.ExpectQuery(`INSERT INTO items (.+) RETURNING *`).
					WithArgs(item.Name, int32(item.Quantity), string(item.Status)).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(mock, tt.item)

			err := repo.CreateItem(ctx, tt.item)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotZero(t, tt.item.ID)
			assert.NotZero(t, tt.item.CreatedAt)
			assert.NotZero(t, tt.item.UpdatedAt)
		})
	}
}

func TestListItems(t *testing.T) {
	repo, mock := setupTestDB(t)
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name    string
		setup   func(sqlmock.Sqlmock)
		want    []domain.Item
		wantErr bool
	}{
		{
			name: "successful list",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "quantity", "reservation_id", "status", "created_at", "updated_at"}).
					AddRow(1, "Item 1", 1, sql.NullString{String: "res1", Valid: true}, "PENDING", now, now).
					AddRow(2, "Item 2", 2, sql.NullString{String: "res2", Valid: true}, "RESERVED", now, now)
				mock.ExpectQuery("SELECT (.+) FROM items").WillReturnRows(rows)
			},
			want: []domain.Item{
				{
					ID:            1,
					Name:          "Item 1",
					Quantity:      1,
					ReservationID: stringPtr("res1"),
					Status:        domain.StatusReservationPending,
					CreatedAt:     now,
					UpdatedAt:     now,
				},
				{
					ID:            2,
					Name:          "Item 2",
					Quantity:      2,
					ReservationID: stringPtr("res2"),
					Status:        domain.StatusReservationReserved,
					CreatedAt:     now,
					UpdatedAt:     now,
				},
			},
			wantErr: false,
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM items").WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(mock)

			got, err := repo.ListItems(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUpdateItemReservation(t *testing.T) {
	repo, mock := setupTestDB(t)
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name          string
		id            int64
		reservationID string
		setup         func(sqlmock.Sqlmock, int64, string)
		wantErr       bool
	}{
		{
			name:          "successful update",
			id:            1,
			reservationID: "res1",
			setup: func(mock sqlmock.Sqlmock, id int64, resID string) {
				columns := []string{"id", "name", "quantity", "reservation_id", "status", "created_at", "updated_at"}
				mock.ExpectQuery(`UPDATE items SET (.+) RETURNING (.+)`).
					WithArgs(int32(id), resID, string(domain.StatusReservationReserved)).
					WillReturnRows(sqlmock.NewRows(columns).
						AddRow(id, "Test Item", 1, sql.NullString{String: resID, Valid: true},
							string(domain.StatusReservationReserved), now, now))
			},
			wantErr: false,
		},
		{
			name:          "database error",
			id:            1,
			reservationID: "res1",
			setup: func(mock sqlmock.Sqlmock, id int64, resID string) {
				mock.ExpectQuery(`UPDATE items SET (.+) RETURNING (.+)`).
					WithArgs(int32(id), resID, string(domain.StatusReservationReserved)).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(mock, tt.id, tt.reservationID)

			err := repo.UpdateItemReservation(ctx, tt.id, tt.reservationID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
