package user

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tranphuocnhan/radio-shuffle/pkg/dbsqlc"
)

type sqlcRepository struct {
	q *dbsqlc.Queries
}

// NewRepository returns a SQLC-backed Repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &sqlcRepository{q: dbsqlc.New(pool)}
}

func (r *sqlcRepository) Create(ctx context.Context, in CreateUserInput) (UserRow, error) {
	row, err := r.q.CreateUser(ctx, dbsqlc.CreateUserParams{
		Email:        in.Email,
		PasswordHash: in.PasswordHash,
		Role:         in.Role,
	})
	if err != nil {
		return UserRow{}, err
	}
	return rowFromDB(row), nil
}

func (r *sqlcRepository) GetByEmail(ctx context.Context, email string) (UserRow, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return UserRow{}, err
	}
	return rowFromDB(row), nil
}

func (r *sqlcRepository) GetByID(ctx context.Context, id int64) (UserRow, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return UserRow{}, err
	}
	return rowFromDB(row), nil
}

func rowFromDB(u dbsqlc.Users) UserRow {
	return UserRow{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         u.Role,
		CreatedAt:    timeFromDB(u.CreatedAt),
		UpdatedAt:    timeFromDB(u.UpdatedAt),
	}
}

func timeFromDB(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}
