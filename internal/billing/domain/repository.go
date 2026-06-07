package domain

import (
	"context"

	"github.com/google/uuid"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *PaymentTransaction) error
	FindByID(ctx context.Context, id uuid.UUID) (*PaymentTransaction, error)
	FindByProviderOrderID(ctx context.Context, orderID string) (*PaymentTransaction, error)
	FindLatestPendingBySchool(ctx context.Context, schoolID uuid.UUID) (*PaymentTransaction, error)
	Update(ctx context.Context, payment *PaymentTransaction) error
}
