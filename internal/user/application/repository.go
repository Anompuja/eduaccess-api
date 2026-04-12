package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/google/uuid"
)

// ListFilter is the parameter bag for User listing queries.
type ListFilter struct {
	SchoolID *uuid.UUID
	Role     string
	Search   string // matches against name, email, username
	Offset   int
	Limit    int
}

// UserReadRepository defines read operations needed by user management queries.
type UserReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	List(ctx context.Context, filter ListFilter) ([]*domain.User, int64, error)
}

// UserWriteRepository defines write operations needed by user management commands.
// It embeds UserReadRepository so handlers can load before writing.
type UserWriteRepository interface {
	UserReadRepository
	Update(ctx context.Context, user *domain.User) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}
