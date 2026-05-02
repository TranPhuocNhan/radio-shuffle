package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, url string) (*pgxpool.Pool, func(), error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, nil, fmt.Errorf("pgx ParseConfig: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("pgxpool: %w", err)
	}
	return pool, func() { pool.Close() }, nil
}
