package domain

import (
	"context"

	"github.com/google/uuid"
)

// UserRepository defines persistence operations for the User aggregate.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByEmailIncludingDeleted(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	Create(ctx context.Context, user *User) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}

