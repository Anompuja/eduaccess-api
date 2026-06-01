package domain

import (
	"context"

	"github.com/google/uuid"
)

// SchoolRepository combines read and write operations for schools.
type SchoolRepository interface {
	Create(ctx context.Context, school *School) error
	FindByID(ctx context.Context, id uuid.UUID) (*School, error)
	List(ctx context.Context, filter SchoolFilter) ([]*School, int64, error)
	Update(ctx context.Context, school *School) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	ExistsByName(ctx context.Context, name string) (bool, error)

	// Rules
	ListRules(ctx context.Context, schoolID uuid.UUID) ([]*SchoolRule, error)
	UpsertRule(ctx context.Context, rule *SchoolRule) error
	DeleteRule(ctx context.Context, schoolID uuid.UUID, key string) error

	// Subscription (read-only)
	FindActiveSubscription(ctx context.Context, schoolID uuid.UUID) (*Subscription, error)

	// SetHeadmasterID updates the schools.headmaster_id column.
	SetHeadmasterID(ctx context.Context, schoolID, headmasterUserID uuid.UUID) error
}

// SchoolFilter holds list query parameters.
type SchoolFilter struct {
	Search string
	Status string
	Offset int
	Limit  int
}
