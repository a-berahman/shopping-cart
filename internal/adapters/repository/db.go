package repository

import (
	"context"
	"database/sql"
	"fmt"

	db "github.com/a-berahman/shopping-cart/internal/adapters/repository/postgres"
	"github.com/a-berahman/shopping-cart/internal/core/domain"

	_ "github.com/lib/pq"
)

type Repository struct {
	db *db.Queries
}

// NewRepository creates a new instance of Repository
func NewRepository(in *sql.DB) *Repository {
	return &Repository{
		db: db.New(in),
	}
}
func NewPostgresDB(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	return db, nil
}

func (r *Repository) CreateItem(ctx context.Context, item *domain.Item) error {
	dbItem, err := r.db.CreateItem(ctx, db.CreateItemParams{
		Name:     item.Name,
		Quantity: int32(item.Quantity),
		Status:   db.ItemStatus(item.Status),
	})
	if err != nil {
		return fmt.Errorf("error creating item: %w", err)
	}

	item.ID = int64(dbItem.ID)
	item.CreatedAt = dbItem.CreatedAt
	item.UpdatedAt = dbItem.UpdatedAt
	return nil
}

func (r *Repository) ListItems(ctx context.Context) ([]domain.Item, error) {
	dbItems, err := r.db.ListItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing items: %w", err)
	}

	items := make([]domain.Item, len(dbItems))
	for i, dbItem := range dbItems {
		items[i] = domain.Item{
			ID:            int64(dbItem.ID),
			Name:          dbItem.Name,
			Quantity:      int(dbItem.Quantity),
			ReservationID: &dbItem.ReservationID.String,
			Status:        domain.ItemStatus(dbItem.Status),
			CreatedAt:     dbItem.CreatedAt,
			UpdatedAt:     dbItem.UpdatedAt,
		}
	}
	return items, nil
}

func (r *Repository) UpdateItemReservation(ctx context.Context, id int64, reservationID string) error {
	if _, err := r.db.UpdateItemReservation(ctx, db.UpdateItemReservationParams{
		ID:            int32(id),
		ReservationID: sql.NullString{String: reservationID, Valid: true},
		Status:        db.ItemStatus(domain.StatusReservationReserved),
	}); err != nil {
		return fmt.Errorf("error updating item reservation: %w", err)
	}
	return nil
}

func (r *Repository) GetItem(ctx context.Context, id int64) (*domain.Item, error) {
	dbItem, err := r.db.GetItem(ctx, int32(id))
	if err != nil {
		return nil, fmt.Errorf("error getting item: %w", err)
	}

	return &domain.Item{
		ID:            int64(dbItem.ID),
		Name:          dbItem.Name,
		Quantity:      int(dbItem.Quantity),
		ReservationID: &dbItem.ReservationID.String,
		Status:        domain.ItemStatus(dbItem.Status),
		CreatedAt:     dbItem.CreatedAt,
		UpdatedAt:     dbItem.UpdatedAt,
	}, nil
}

func (r *Repository) UpdateItemStatus(ctx context.Context, id int64, status domain.ItemStatus) error {
	_, err := r.db.UpdateItemStatus(ctx, db.UpdateItemStatusParams{
		ID:     int32(id),
		Status: db.ItemStatus(status),
	})
	if err != nil {
		return fmt.Errorf("error updating item status: %w", err)
	}
	return nil
}
