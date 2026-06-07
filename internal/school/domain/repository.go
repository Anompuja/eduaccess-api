package domain

import (
	"context"

	"github.com/google/uuid"
)

// SchoolRepository combines read and write operations for schools.
type SchoolRepository interface {
	Create(ctx context.Context, school *School) error
	CreateWithDefaultSubscription(ctx context.Context, school *School) (*Subscription, error)
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
	ListPlans(ctx context.Context) ([]*Plan, error)
	FindPlanByID(ctx context.Context, id uuid.UUID) (*Plan, error)
	FindActiveSubscription(ctx context.Context, schoolID uuid.UUID) (*Subscription, error)
	ReplaceSubscription(ctx context.Context, sub *Subscription) error

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
