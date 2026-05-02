package auth

import (
	"context"
	"time"
)

// UserReader fetches user data needed by auth.
type UserReader interface {
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByID(ctx context.Context, id int64) (User, error)
}

// UserWriter persists user data needed by auth.
type UserWriter interface {
	Create(ctx context.Context, in CreateUserInput) (User, error)
}

// TokenRepository stores refresh tokens for auth.
type TokenRepository interface {
	Store(ctx context.Context, in StoreRefreshTokenInput) error
	GetByHash(ctx context.Context, tokenHash []byte) (RefreshToken, error)
	DeleteByHash(ctx context.Context, tokenHash []byte) error
}

// User is the auth view of a user.
type User struct {
	ID           int64
	Email        string
	PasswordHash []byte
	Role         string
}

// CreateUserInput contains the data auth needs to create a user.
type CreateUserInput struct {
	Email        string
	PasswordHash []byte
	Role         string
}

// StoreRefreshTokenInput contains the refresh token fields to persist.
type StoreRefreshTokenInput struct {
	UserID    int64
	TokenHash []byte
	ExpiresAt time.Time
}

// RefreshToken is the auth view of a stored refresh token.
type RefreshToken struct {
	UserID    int64
	TokenHash []byte
	ExpiresAt time.Time
	CreatedAt time.Time
}
