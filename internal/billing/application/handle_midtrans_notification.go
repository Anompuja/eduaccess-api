package application

import (
	"context"

	billingdomain "github.com/eduaccess/eduaccess-api/internal/billing/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
)

type HandleMidtransNotificationCommand struct {
	OrderID           string
	StatusCode        string
	GrossAmount       string
	SignatureKey      string
	TransactionID     string
	TransactionStatus string
	PaymentType       string
	FraudStatus       string
	RawNotification   string
}

type HandleMidtransNotificationHandler struct {
	payments billingdomain.PaymentRepository
	schools  SchoolAccessRepository
	students ActiveStudentCounter
	gateway  PaymentGateway
}

func NewHandleMidtransNotificationHandler(
	payments billingdomain.PaymentRepository,
	schools SchoolAccessRepository,
	students ActiveStudentCounter,
	gateway PaymentGateway,
) *HandleMidtransNotificationHandler {
	return &HandleMidtransNotificationHandler{
		payments: payments,
		schools:  schools,
		students: students,
		gateway:  gateway,
	}
}

func (h *HandleMidtransNotificationHandler) Handle(ctx context.Context, cmd HandleMidtransNotificationCommand) (*billingdomain.PaymentTransaction, error) {
	if !h.gateway.VerifySignature(cmd.OrderID, cmd.StatusCode, cmd.GrossAmount, cmd.SignatureKey) {
		return nil, apperror.New(apperror.ErrUnauthorized, "invalid midtrans signature")
	}

	payment, err := h.payments.FindByProviderOrderID(ctx, cmd.OrderID)
	if err != nil {
		return nil, err
	}

	status, err := h.gateway.GetTransactionStatus(ctx, cmd.OrderID)
	if err != nil {
		if cmd.TransactionStatus == "" {
			return nil, err
		}
		status = &GatewayTransactionStatus{
			OrderID:           cmd.OrderID,
			TransactionID:     cmd.TransactionID,
			TransactionStatus: cmd.TransactionStatus,
			StatusCode:        cmd.StatusCode,
			GrossAmount:       cmd.GrossAmount,
			PaymentType:       cmd.PaymentType,
			FraudStatus:       cmd.FraudStatus,
			RawResponse:       cmd.RawNotification,
		}
	}
	return syncPaymentWithGatewayStatus(ctx, payment, status, h.payments, h.schools, h.students)
}
