package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/a-berahman/shopping-cart/internal/core/domain"
	"github.com/redis/go-redis/v9"
)

const (
	ReservationQueueKey = "reservation:queue"
	ProcessingSetKey    = "reservation:processing"
	FailedSetKey        = "reservation:failed"
	CompletedSetKey     = "reservation:completed"
)

type ReservationJob struct {
	ID            string           `json:"id"`
	ItemID        int64            `json:"item_id"`
	ItemName      string           `json:"item_name"`
	Quantity      int              `json:"quantity"`
	Status        domain.JobStatus `json:"status"`
	Attempts      int              `json:"attempts"`
	LastAttempted time.Time        `json:"last_attempted"`
	CreatedAt     time.Time        `json:"created_at"`
}

type RedisQueue struct {
	client *redis.Client
}

func NewRedisQueue(redisURL string) (*RedisQueue, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)
	return &RedisQueue{client: client}, nil
}

func (q *RedisQueue) EnqueueReservation(ctx context.Context, job *domain.ReservationJob) error {
	job.CreatedAt = time.Now()
	job.Status = domain.JobStatusPending

	jobData, err := json.Marshal(job)
	if err != nil {
		return err
	}

	// Add to pending queue
	return q.client.LPush(ctx, ReservationQueueKey, jobData).Err()
}

func (q *RedisQueue) DequeueReservation(ctx context.Context) (*domain.ReservationJob, error) {
	// Move job from queue to processing set with BRPOPLPUSH
	result, err := q.client.BRPopLPush(ctx, ReservationQueueKey, ProcessingSetKey, 0).Result()
	if err != nil {
		return nil, err
	}

	var job domain.ReservationJob
	if err := json.Unmarshal([]byte(result), &job); err != nil {
		return nil, err
	}

	job.Status = domain.JobStatusProcessing
	job.LastAttempted = time.Now()

	return &job, nil
}

func (q *RedisQueue) CompleteJob(ctx context.Context, job *domain.ReservationJob) error {
	job.Status = domain.JobStatusCompleted

	jobData, err := json.Marshal(job)
	if err != nil {
		return err
	}

	pipe := q.client.Pipeline()
	// remove from processing set
	pipe.LRem(ctx, ProcessingSetKey, 0, jobData)
	// add to completed set
	pipe.LPush(ctx, CompletedSetKey, jobData)

	_, err = pipe.Exec(ctx)
	return err
}

func (q *RedisQueue) FailJob(ctx context.Context, job *domain.ReservationJob) error {
	job.Status = domain.JobStatusFailed

	jobData, err := json.Marshal(job)
	if err != nil {
		return err
	}

	pipe := q.client.Pipeline()
	// removing from processing set
	pipe.LRem(ctx, ProcessingSetKey, 0, jobData)
	// adding to failed set
	pipe.LPush(ctx, FailedSetKey, jobData)

	_, err = pipe.Exec(ctx)
	return err
}

// RetryFailedJobs moves failed jobs back to the pending queue
func (q *RedisQueue) RetryFailedJobs(ctx context.Context) error {
	pipe := q.client.Pipeline()

	// we move all failed jobs back to pending queue
	pipe.LRange(ctx, FailedSetKey, 0, -1).Result()
	pipe.Del(ctx, FailedSetKey)

	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	jobs := cmds[0].(*redis.StringSliceCmd).Val()
	if len(jobs) == 0 {
		return nil
	}

	return q.client.LPush(ctx, ReservationQueueKey, jobs).Err()
}
