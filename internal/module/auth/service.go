package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// Sentinel errors for HTTP mapping in handlers.
var (
	ErrEmailTaken          = errors.New("email already registered")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrWeakPassword        = errors.New("password does not meet policy")
)

// Service defines the auth business operations.
type Service interface {
	Register(ctx context.Context, in RegisterInput) (AuthOutput, error)
	Login(ctx context.Context, in LoginInput) (AuthOutput, error)
	Refresh(ctx context.Context, in RefreshInput) (AuthOutput, error)
	Logout(ctx context.Context, in LogoutInput) error
}

// Config contains auth settings.
type Config struct {
	SigningKey []byte
	Issuer     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// EmailVerifier optionally sends verification links after register.
type EmailVerifier interface {
	SendVerification(ctx context.Context, userID int64, email string) error
}

// RegisterInput is the service input for registering a user.
type RegisterInput struct {
	Email    string
	Password string
}

// LoginInput is the service input for logging in a user.
type LoginInput struct {
	Email    string
	Password string
}

// RefreshInput is the service input for refreshing tokens.
type RefreshInput struct {
	RefreshToken string
}

// LogoutInput is the service input for logging out.
type LogoutInput struct {
	RefreshToken string
}

// AuthOutput is the service output containing tokens.
type AuthOutput struct {
	AccessToken  string
	RefreshToken string
}

type service struct {
	users    UserReader
	writers  UserWriter
	tokens   TokenRepository
	cfg      Config
	verifier EmailVerifier
}

// NewService constructs a Service.
func NewService(users UserReader, writers UserWriter, tokens TokenRepository, cfg Config, verifier EmailVerifier) Service {
	return &service{
		users:    users,
		writers:  writers,
		tokens:   tokens,
		cfg:      cfg,
		verifier: verifier,
	}
}

func (s *service) Register(ctx context.Context, in RegisterInput) (AuthOutput, error) {
	if err := validatePassword(in.Password); err != nil {
		return AuthOutput{}, err
	}
	_, err := s.users.GetByEmail(ctx, in.Email)
	if err == nil {
		return AuthOutput{}, ErrEmailTaken
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return AuthOutput{}, err
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return AuthOutput{}, err
	}
	user, err := s.writers.Create(ctx, CreateUserInput{
		Email:        in.Email,
		PasswordHash: passwordHash,
		Role:         "user",
	})
	if err != nil {
		return AuthOutput{}, err
	}
	out, err := s.issueTokens(ctx, user.ID)
	if err != nil {
		return AuthOutput{}, err
	}
	if s.verifier != nil {
		if err := s.verifier.SendVerification(ctx, user.ID, user.Email); err != nil {
			return AuthOutput{}, err
		}
	}
	return out, nil
}

func (s *service) Login(ctx context.Context, in LoginInput) (AuthOutput, error) {
	user, err := s.users.GetByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AuthOutput{}, ErrInvalidCredentials
		}
		return AuthOutput{}, err
	}
	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(in.Password)); err != nil {
		return AuthOutput{}, ErrInvalidCredentials
	}
	return s.issueTokens(ctx, user.ID)
}

func (s *service) Refresh(ctx context.Context, in RefreshInput) (AuthOutput, error) {
	hash := hashToken(in.RefreshToken)
	stored, err := s.tokens.GetByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AuthOutput{}, ErrInvalidRefreshToken
		}
		return AuthOutput{}, err
	}
	if time.Now().After(stored.ExpiresAt) {
		_ = s.tokens.DeleteByHash(ctx, hash)
		return AuthOutput{}, ErrInvalidRefreshToken
	}
	if err := s.tokens.DeleteByHash(ctx, hash); err != nil {
		return AuthOutput{}, err
	}
	return s.issueTokens(ctx, stored.UserID)
}

func (s *service) Logout(ctx context.Context, in LogoutInput) error {
	return s.tokens.DeleteByHash(ctx, hashToken(in.RefreshToken))
}

func (s *service) issueTokens(ctx context.Context, userID int64) (AuthOutput, error) {
	accessToken, err := s.newAccessToken(userID)
	if err != nil {
		return AuthOutput{}, err
	}
	refreshToken, refreshHash, err := newRefreshToken()
	if err != nil {
		return AuthOutput{}, err
	}
	expiresAt := time.Now().Add(s.cfg.RefreshTTL)
	if err := s.tokens.Store(ctx, StoreRefreshTokenInput{
		UserID:    userID,
		TokenHash: refreshHash,
		ExpiresAt: expiresAt,
	}); err != nil {
		return AuthOutput{}, err
	}
	return AuthOutput{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *service) newAccessToken(userID int64) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(userID, 10),
		Issuer:    s.cfg.Issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.AccessTTL)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.cfg.SigningKey)
}

func newRefreshToken() (string, []byte, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", nil, err
	}
	token := base64.RawURLEncoding.EncodeToString(buf)
	hash := sha256.Sum256([]byte(token))
	return token, hash[:], nil
}

func hashToken(token string) []byte {
	hash := sha256.Sum256([]byte(token))
	return hash[:]
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return ErrWeakPassword
	}
	return nil
}
