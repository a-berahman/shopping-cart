package queue

import (
	"context"
	"testing"

	"github.com/a-berahman/shopping-cart/internal/core/domain"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupTestRedis(t *testing.T) (*RedisQueue, *redis.Client) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	queue := &RedisQueue{client: client}
	return queue, client
}

func TestEnqueueReservation(t *testing.T) {
	queue, client := setupTestRedis(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		job     *domain.ReservationJob
		wantErr bool
		setup   func(*redis.Client, *domain.ReservationJob)
	}{
		{
			name: "successful enqueue",
			job: &domain.ReservationJob{
				ItemID:   123,
				ItemName: "Test Item",
				Quantity: 1,
			},
			setup: func(c *redis.Client, job *domain.ReservationJob) {
				c.LPush(ctx, ReservationQueueKey, []byte(job.ItemName))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(client, tt.job)
			}

			err := queue.EnqueueReservation(ctx, tt.job)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
