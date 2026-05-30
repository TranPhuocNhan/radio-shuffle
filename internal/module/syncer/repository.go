package syncer

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var (
	// ErrRepoJobNotFound indicates that a sync job row does not exist.
	ErrRepoJobNotFound = errors.New("syncer repository: job not found")
)

// RadioBrowserStation is the domain representation of a synced station row.
type RadioBrowserStation struct {
	ID          int64
	StationUUID string
	Name        string
	URL         string
	URLResolved *string
	Homepage    *string
	Favicon     *string
	Country     *string
	CountryCode *string
	State       *string
	Language    *string
	Codec       *string
	Bitrate     int32
	Votes       int32
	Tags        *string
	LastCheckOK bool
	SyncedAt    time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UpsertInput is the data required to insert or update one synced station.
type UpsertInput struct {
	StationUUID string
	Name        string
	URL         string
	URLResolved *string
	Homepage    *string
	Favicon     *string
	Country     *string
	CountryCode *string
	State       *string
	Language    *string
	Codec       *string
	Bitrate     int32
	Votes       int32
	Tags        *string
	LastCheckOK bool
}

type SyncJobStatus string

const (
	SyncJobPending   SyncJobStatus = "pending"
	SyncJobRunning   SyncJobStatus = "running"
	SyncJobCompleted SyncJobStatus = "completed"
	SyncJobFailed    SyncJobStatus = "failed"
)

type SyncJob struct {
	RequestID    string
	Status       SyncJobStatus
	RequestedBy  string
	Scope        json.RawMessage
	ErrorMessage *string
	StartedAt    *time.Time
	FinishedAt   *time.Time
	CreatedAt    time.Time
}

type CreateSyncJobInput struct {
	RequestID   string
	Status      SyncJobStatus
	RequestedBy string
	Scope       json.RawMessage
}

// Repository persists radio-browser stations.
type Repository interface {
	UpsertBatch(ctx context.Context, stations []UpsertInput) (int64, error)

	CreateSyncJob(ctx context.Context, in CreateSyncJobInput) error
	GetSyncJob(ctx context.Context, requestID string) (SyncJob, error)
	MarkSyncJobRunning(ctx context.Context, requestID string) error
	MarkSyncJobCompleted(ctx context.Context, requestID string) error
	MarkSyncJobFailed(ctx context.Context, requestID string, message string) error
}
