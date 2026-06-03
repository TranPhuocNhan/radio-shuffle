package syncer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tranphuocnhan/radio-shuffle/pkg/dbsqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type sqlcRepository struct {
	q *dbsqlc.Queries
}

// NewRepository returns a SQLC-backed Repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &sqlcRepository{q: dbsqlc.New(pool)}
}

func (r *sqlcRepository) UpsertBatch(ctx context.Context, stations []UpsertInput) (int64, error) {
	var upserted int64
	for _, s := range stations {
		affected, err := r.q.UpsertRadioBrowserStation(ctx, dbsqlc.UpsertRadioBrowserStationParams{
			Stationuuid: s.StationUUID,
			Name:        s.Name,
			Url:         s.URL,
			UrlResolved: s.URLResolved,
			Homepage:    s.Homepage,
			Favicon:     s.Favicon,
			Country:     s.Country,
			Countrycode: s.CountryCode,
			State:       s.State,
			Language:    s.Language,
			Codec:       s.Codec,
			Bitrate:     s.Bitrate,
			Votes:       s.Votes,
			Tags:        s.Tags,
			LastCheckOk: s.LastCheckOK,
		})
		if err != nil {
			return upserted, fmt.Errorf("upsert station %s: %w", s.StationUUID, err)
		}
		upserted += affected
	}
	return upserted, nil
}

func (r *sqlcRepository) CreateSyncJob(ctx context.Context, in CreateSyncJobInput) error {
	return r.q.CreateSyncJob(ctx, dbsqlc.CreateSyncJobParams{
		RequestID:   in.RequestID,
		Status:      string(in.Status),
		RequestedBy: in.RequestedBy,
		Scope:       in.Scope,
	})
}

func (r *sqlcRepository) GetActiveSyncJob(ctx context.Context) (SyncJob, error) {
	row, err := r.q.GetActiveSyncJob(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SyncJob{}, ErrRepoJobNotFound
		}
		return SyncJob{}, err
	}
	return syncJobFromDB(row), nil
}

func (r *sqlcRepository) GetSyncJob(ctx context.Context, requestID string) (SyncJob, error) {
	row, err := r.q.GetSyncJobByID(ctx, requestID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SyncJob{}, ErrRepoJobNotFound
		}
		return SyncJob{}, err
	}
	return syncJobFromDB(row), nil
}

func syncJobFromDB(row dbsqlc.SyncJobs) SyncJob {
	var startedAt *time.Time
	if row.StartedAt.Valid {
		startedAt = &row.StartedAt.Time
	}
	var finishedAt *time.Time
	if row.FinishedAt.Valid {
		finishedAt = &row.FinishedAt.Time
	}
	return SyncJob{
		RequestID:    row.RequestID,
		Status:       SyncJobStatus(row.Status),
		RequestedBy:  row.RequestedBy,
		Scope:        row.Scope,
		ErrorMessage: row.ErrorMessage,
		StartedAt:    startedAt,
		FinishedAt:   finishedAt,
		CreatedAt:    row.CreatedAt.Time,
	}
}

func (r *sqlcRepository) MarkSyncJobRunning(ctx context.Context, requestID string) error {
	return r.q.MarkSyncJobRunning(ctx, requestID)
}

func (r *sqlcRepository) MarkSyncJobCompleted(ctx context.Context, requestID string) error {
	return r.q.MarkSyncJobCompleted(ctx, requestID)
}

func (r *sqlcRepository) MarkSyncJobFailed(ctx context.Context, requestID string, message string) error {
	return r.q.MarkSyncJobFailed(ctx, dbsqlc.MarkSyncJobFailedParams{
		RequestID:    requestID,
		ErrorMessage: &message,
	})
}
