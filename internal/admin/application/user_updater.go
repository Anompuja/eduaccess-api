package application

import (
	"context"

	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/google/uuid"
)

// UserUpdater abstracts user read/write operations required by admin update use-cases.
type UserUpdater interface {
	FindByID(ctx context.Context, id uuid.UUID) (*authdomain.User, error)
	Update(ctx context.Context, user *authdomain.User) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}
