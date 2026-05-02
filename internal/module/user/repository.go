package user

import (
	"context"
	"time"
)

// Repository defines DB operations for user data.
type Repository interface {
	Create(ctx context.Context, in CreateUserInput) (UserRow, error)
	GetByEmail(ctx context.Context, email string) (UserRow, error)
	GetByID(ctx context.Context, id int64) (UserRow, error)
}

// CreateUserInput is the repository input for creating users.
type CreateUserInput struct {
	Email        string
	PasswordHash []byte
	Role         string
}

// UserRow is the persisted user data.
type UserRow struct {
	ID           int64
	Email        string
	PasswordHash []byte
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
