package domain

import (
	"context"

	"github.com/google/uuid"
)

// UserRepository defines persistence operations for the User aggregate.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	Create(ctx context.Context, user *User) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}

// RefreshTokenRepository defines persistence operations for refresh tokens.
type RefreshTokenRepository interface {
	Create(ctx context.Context, rt *RefreshToken) error
	FindByToken(ctx context.Context, token string) (*RefreshToken, error)
	DeleteByToken(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}
