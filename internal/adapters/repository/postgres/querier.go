// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package db

import (
	"context"
)

type Querier interface {
	CreateItem(ctx context.Context, arg CreateItemParams) (Item, error)
	GetItem(ctx context.Context, id int32) (Item, error)
	ListItems(ctx context.Context) ([]Item, error)
	UpdateItemReservation(ctx context.Context, arg UpdateItemReservationParams) (Item, error)
	UpdateItemStatus(ctx context.Context, arg UpdateItemStatusParams) (Item, error)
}

var _ Querier = (*Queries)(nil)