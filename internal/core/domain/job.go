package domain

import "time"

type JobType string

const (
	JobTypeAvailabilityCheck JobType = "AVAILABILITY_CHECK"
	JobTypeReservation       JobType = "RESERVATION"
)

// JobStatus is the status of a job
type JobStatus string

const (
	JobStatusPending    JobStatus = "PENDING"
	JobStatusProcessing JobStatus = "PROCESSING"
	JobStatusCompleted  JobStatus = "COMPLETED"
	JobStatusFailed     JobStatus = "FAILED"
)

// ReservationJob is the domain object for a reservation job
type ReservationJob struct {
	ID            string    `json:"id"`
	ItemID        int64     `json:"item_id"`
	ItemName      string    `json:"item_name"`
	Quantity      int       `json:"quantity"`
	Status        JobStatus `json:"status"`
	Attempts      int       `json:"attempts"`
	LastAttempted time.Time `json:"last_attempted"`
	CreatedAt     time.Time `json:"created_at"`
	JobType       JobType   `json:"job_type"`
}

// CanRetry is a business rules for jobs
func (j *ReservationJob) CanRetry(maxAttempts int) bool {
	return j.Status != JobStatusCompleted && j.Attempts < maxAttempts
}
