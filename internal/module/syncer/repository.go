package syncer

import (
	"context"
	"time"
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

// Repository persists radio-browser stations.
type Repository interface {
	UpsertBatch(ctx context.Context, stations []UpsertInput) (int64, error)
}
