package application

import (
	"context"

	billingdomain "github.com/eduaccess/eduaccess-api/internal/billing/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

type GetPaymentQuery struct {
	RequesterRole     string
	RequesterSchoolID *uuid.UUID
	SchoolID          uuid.UUID
	PaymentID         uuid.UUID
}

type GetPaymentHandler struct {
	payments billingdomain.PaymentRepository
	schools  SchoolAccessRepository
	students ActiveStudentCounter
	gateway  PaymentGateway
}

func NewGetPaymentHandler(
	payments billingdomain.PaymentRepository,
	schools SchoolAccessRepository,
	students ActiveStudentCounter,
	gateway PaymentGateway,
) *GetPaymentHandler {
	return &GetPaymentHandler{
		payments: payments,
		schools:  schools,
		students: students,
		gateway:  gateway,
	}
}

func (h *GetPaymentHandler) Handle(ctx context.Context, q GetPaymentQuery) (*billingdomain.PaymentTransaction, error) {
	if err := guardBillingAccess(q.RequesterRole, q.RequesterSchoolID, q.SchoolID); err != nil {
		return nil, err
	}

	payment, err := h.payments.FindByID(ctx, q.PaymentID)
	if err != nil {
		return nil, err
	}
	if payment.SchoolID != q.SchoolID {
		return nil, apperror.New(apperror.ErrForbidden, "payment transaction does not belong to this school")
	}

	if payment.Provider == billingdomain.ProviderMidtrans && payment.Status == billingdomain.PaymentStatusPending {
		status, err := h.gateway.GetTransactionStatus(ctx, payment.ProviderOrderID)
		if err == nil {
			payment, err = syncPaymentWithGatewayStatus(ctx, payment, status, h.payments, h.schools, h.students)
			if err != nil {
				return nil, err
			}
		}
	}
	return payment, nil
}
