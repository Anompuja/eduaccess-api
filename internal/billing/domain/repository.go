package domain

import (
	"context"

	"github.com/google/uuid"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *PaymentTransaction) error
	List(ctx context.Context, filter PaymentFilter) ([]*PaymentTransaction, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*PaymentTransaction, error)
	FindByProviderOrderID(ctx context.Context, orderID string) (*PaymentTransaction, error)
	FindLatestPendingBySchool(ctx context.Context, schoolID uuid.UUID) (*PaymentTransaction, error)
	Update(ctx context.Context, payment *PaymentTransaction) error
}

type PaymentFilter struct {
	SchoolID *uuid.UUID
	Status   string
	Search   string
	Offset   int
	Limit    int
}
