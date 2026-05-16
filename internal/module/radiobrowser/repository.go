package radiobrowser

import (
	"context"
	"time"
)

// Station is the domain representation of a radio browser station.
type Station struct {
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

// Repository loads radio browser stations from the database.
type Repository interface {
	List(ctx context.Context, name, country, language string, limit, offset int64) ([]Station, error)
	Count(ctx context.Context, name, country, language string) (int64, error)
}
