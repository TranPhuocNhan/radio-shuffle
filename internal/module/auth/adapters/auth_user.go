package adapters

import (
	"context"
	"errors"

	"github.com/tranphuocnhan/radio-shuffle/internal/module/auth"
	"github.com/tranphuocnhan/radio-shuffle/internal/module/user"

	"github.com/jackc/pgx/v5"
)

// AuthUserAdapter maps user.Repository to auth.UserReader/UserWriter.
type AuthUserAdapter struct {
	repo user.Repository
}

func NewAuthUserAdapter(repo user.Repository) AuthUserAdapter {
	return AuthUserAdapter{repo: repo}
}

func (a AuthUserAdapter) GetByEmail(ctx context.Context, email string) (auth.User, error) {
	row, err := a.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.User{}, auth.ErrUserNotFound
		}
		return auth.User{}, err
	}
	return auth.User{ID: row.ID, Email: row.Email, PasswordHash: row.PasswordHash, Role: row.Role}, nil
}

func (a AuthUserAdapter) GetByID(ctx context.Context, id int64) (auth.User, error) {
	row, err := a.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.User{}, auth.ErrUserNotFound
		}
		return auth.User{}, err
	}
	return auth.User{ID: row.ID, Email: row.Email, PasswordHash: row.PasswordHash, Role: row.Role}, nil
}

func (a AuthUserAdapter) Create(ctx context.Context, in auth.CreateUserInput) (auth.User, error) {
	row, err := a.repo.Create(ctx, user.CreateUserInput{
		Email:        in.Email,
		PasswordHash: in.PasswordHash,
		Role:         in.Role,
	})
	if err != nil {
		return auth.User{}, err
	}
	return auth.User{ID: row.ID, Email: row.Email, PasswordHash: row.PasswordHash, Role: row.Role}, nil
}
