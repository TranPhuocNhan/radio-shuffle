package auth

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tranphuocnhan/radio-shuffle/pkg/dbsqlc"
)

// sqlcTokenRepository implements TokenRepository with SQLC queries.
type sqlcTokenRepository struct {
	q *dbsqlc.Queries
}

// NewTokenRepository returns a SQLC-backed TokenRepository.
func NewTokenRepository(pool *pgxpool.Pool) TokenRepository {
	return &sqlcTokenRepository{q: dbsqlc.New(pool)}
}

func (r *sqlcTokenRepository) Store(ctx context.Context, in StoreRefreshTokenInput) error {
	_, err := r.q.InsertRefreshToken(ctx, dbsqlc.InsertRefreshTokenParams{
		UserID:    in.UserID,
		TokenHash: in.TokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: in.ExpiresAt, Valid: true},
	})
	return err
}

func (r *sqlcTokenRepository) GetByHash(ctx context.Context, tokenHash []byte) (RefreshToken, error) {
	row, err := r.q.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return RefreshToken{}, err
	}
	return RefreshToken{
		UserID:    row.UserID,
		TokenHash: row.TokenHash,
		ExpiresAt: timeFromDB(row.ExpiresAt),
		CreatedAt: timeFromDB(row.CreatedAt),
	}, nil
}

func (r *sqlcTokenRepository) DeleteByHash(ctx context.Context, tokenHash []byte) error {
	return r.q.DeleteRefreshTokenByHash(ctx, tokenHash)
}

func timeFromDB(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}
