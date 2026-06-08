package application

import (
	"context"
	"time"

	billingdomain "github.com/eduaccess/eduaccess-api/internal/billing/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

type CreateCheckoutCommand struct {
	RequesterRole     string
	RequesterSchoolID *uuid.UUID
	RequesterUserID   uuid.UUID
	SchoolID          uuid.UUID
	PlanID            uuid.UUID
	Cycle             string
}

type CreateCheckoutHandler struct {
	payments billingdomain.PaymentRepository
	schools  SchoolAccessRepository
	students ActiveStudentCounter
	gateway  PaymentGateway
}

func NewCreateCheckoutHandler(
	payments billingdomain.PaymentRepository,
	schools SchoolAccessRepository,
	students ActiveStudentCounter,
	gateway PaymentGateway,
) *CreateCheckoutHandler {
	return &CreateCheckoutHandler{
		payments: payments,
		schools:  schools,
		students: students,
		gateway:  gateway,
	}
}

func (h *CreateCheckoutHandler) Handle(ctx context.Context, cmd CreateCheckoutCommand) (*billingdomain.PaymentTransaction, error) {
	if err := guardBillingAccess(cmd.RequesterRole, cmd.RequesterSchoolID, cmd.SchoolID); err != nil {
		return nil, err
	}

	school, err := h.schools.FindByID(ctx, cmd.SchoolID)
	if err != nil {
		return nil, err
	}

	plan, err := h.schools.FindPlanByID(ctx, cmd.PlanID)
	if err != nil {
		return nil, err
	}
	if err := validateCheckoutPlan(plan); err != nil {
		return nil, err
	}

	currentSub, err := h.schools.FindActiveSubscription(ctx, cmd.SchoolID)
	if err != nil && !apperror.Is(err, apperror.ErrNotFound) {
		return nil, err
	}
	if err := validateCheckoutPath(currentSub, plan); err != nil {
		return nil, err
	}

	if plan.MaxStudents > 0 {
		totalStudents, err := h.students.CountActiveStudents(ctx, cmd.SchoolID)
		if err != nil {
			return nil, err
		}
		if totalStudents > int64(plan.MaxStudents) {
			return nil, apperror.New(
				apperror.ErrBadRequest,
				"selected plan does not support the current number of active students in this school",
			)
		}
	}

	if pending, err := h.payments.FindLatestPendingBySchool(ctx, cmd.SchoolID); err == nil {
		if pending.ExpiresAt == nil || pending.ExpiresAt.After(time.Now()) {
			return nil, apperror.New(apperror.ErrConflict, "there is already a pending payment transaction for this school")
		}
		pending.Status = billingdomain.PaymentStatusExpired
		pending.UpdatedAt = time.Now()
		_ = h.payments.Update(ctx, pending)
	} else if !apperror.Is(err, apperror.ErrNotFound) {
		return nil, err
	}

	previewSub, err := buildPaidSubscription(cmd.SchoolID, plan, cmd.Cycle, time.Now())
	if err != nil {
		return nil, err
	}

	paymentID := uuid.New()
	orderID := "EA-" + paymentID.String()
	session, err := h.gateway.CreateCheckout(ctx, GatewayCreateCheckoutInput{
		OrderID:      orderID,
		Amount:       previewSub.Price,
		SchoolName:   school.Name,
		SchoolEmail:  school.Email,
		SchoolPhone:  school.Phone,
		PlanName:     plan.Name,
		Cycle:        cmd.Cycle,
		ExpiryMinute: 1440,
	})
	if err != nil {
		return nil, err
	}

	now := time.Now()
	payment := &billingdomain.PaymentTransaction{
		ID:                  paymentID,
		SchoolID:            cmd.SchoolID,
		SchoolName:          school.Name,
		PlanID:              plan.ID,
		PlanName:            plan.Name,
		CreatedByUserID:     cmd.RequesterUserID,
		Status:              billingdomain.PaymentStatusPending,
		Cycle:               cmd.Cycle,
		Amount:              previewSub.Price,
		Currency:            "IDR",
		Provider:            billingdomain.ProviderMidtrans,
		ProviderOrderID:     orderID,
		ProviderSnapToken:   session.Token,
		ProviderRedirectURL: session.RedirectURL,
		ExpiresAt:           session.ExpiresAt,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	if err := h.payments.Create(ctx, payment); err != nil {
		return nil, err
	}

	return payment, nil
}
