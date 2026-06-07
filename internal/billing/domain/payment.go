package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	ProviderMidtrans = "midtrans"

	PaymentStatusPending   = "pending"
	PaymentStatusPaid      = "paid"
	PaymentStatusFailed    = "failed"
	PaymentStatusExpired   = "expired"
	PaymentStatusCancelled = "cancelled"
)

// PaymentTransaction stores a gateway-backed purchase attempt for a school subscription.
type PaymentTransaction struct {
	ID                      uuid.UUID
	SchoolID                uuid.UUID
	PlanID                  uuid.UUID
	CreatedByUserID         uuid.UUID
	ActivatedSubscriptionID *uuid.UUID
	Status                  string
	Cycle                   string
	Amount                  int64
	Currency                string
	Provider                string
	ProviderOrderID         string
	ProviderTransactionID   string
	ProviderSnapToken       string
	ProviderRedirectURL     string
	PaymentType             string
	TransactionStatus       string
	FraudStatus             string
	RawNotification         string
	PaidAt                  *time.Time
	ExpiresAt               *time.Time
	CreatedAt               time.Time
	UpdatedAt               time.Time
}
