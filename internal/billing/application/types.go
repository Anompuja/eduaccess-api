package application

import (
	"context"
	"time"

	schooldomain "github.com/eduaccess/eduaccess-api/internal/school/domain"
	"github.com/google/uuid"
)

type PaymentGateway interface {
	CreateCheckout(ctx context.Context, input GatewayCreateCheckoutInput) (*GatewayCheckoutSession, error)
	GetTransactionStatus(ctx context.Context, orderID string) (*GatewayTransactionStatus, error)
	VerifySignature(orderID, statusCode, grossAmount, signature string) bool
}

type SchoolAccessRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*schooldomain.School, error)
	FindPlanByID(ctx context.Context, id uuid.UUID) (*schooldomain.Plan, error)
	FindActiveSubscription(ctx context.Context, schoolID uuid.UUID) (*schooldomain.Subscription, error)
	ReplaceSubscription(ctx context.Context, sub *schooldomain.Subscription) error
}

type ActiveStudentCounter interface {
	CountActiveStudents(ctx context.Context, schoolID uuid.UUID) (int64, error)
}

type GatewayCreateCheckoutInput struct {
	OrderID      string
	Amount       int64
	SchoolName   string
	SchoolEmail  string
	SchoolPhone  string
	PlanName     string
	Cycle        string
	ExpiryMinute int
}

type GatewayCheckoutSession struct {
	Token       string
	RedirectURL string
	ExpiresAt   *time.Time
}

type GatewayTransactionStatus struct {
	OrderID           string
	TransactionID     string
	TransactionStatus string
	StatusCode        string
	GrossAmount       string
	PaymentType       string
	FraudStatus       string
	SignatureKey      string
	TransactionTime   *time.Time
	SettlementTime    *time.Time
	RawResponse       string
}
