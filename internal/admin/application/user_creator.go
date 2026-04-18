package application

import (
	"context"

	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
)

// UserCreator abstracts user creation so admin use-cases don't depend
// directly on auth infrastructure.
type UserCreator interface {
	Create(ctx context.Context, user *authdomain.User) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}
