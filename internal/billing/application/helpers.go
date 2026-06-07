package application

import (
	"context"
	"fmt"
	"time"

	billingdomain "github.com/eduaccess/eduaccess-api/internal/billing/domain"
	schooldomain "github.com/eduaccess/eduaccess-api/internal/school/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

func guardBillingAccess(role string, requesterSchoolID *uuid.UUID, targetSchoolID uuid.UUID) error {
	if role == "superadmin" {
		return nil
	}
	if role != "admin_sekolah" {
		return apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can manage subscription payments")
	}
	if requesterSchoolID == nil || *requesterSchoolID != targetSchoolID {
		return apperror.New(apperror.ErrForbidden, "access denied to this school")
	}
	return nil
}

func buildPaidSubscription(schoolID uuid.UUID, plan *schooldomain.Plan, cycle string, now time.Time) (*schooldomain.Subscription, error) {
	var (
		price  int64
		endsAt *time.Time
	)

	switch cycle {
	case "month":
		price = plan.MonthlyPrice
		end := now.AddDate(0, 1, 0)
		endsAt = &end
	case "year":
		price = plan.YearlyPrice
		end := now.AddDate(1, 0, 0)
		endsAt = &end
	default:
		return nil, apperror.New(apperror.ErrBadRequest, "cycle must be one of: month, year")
	}

	return &schooldomain.Subscription{
		ID:        uuid.New(),
		SchoolID:  schoolID,
		PlanID:    plan.ID,
		Status:    "active",
		Cycle:     cycle,
		Quantity:  1,
		Price:     price,
		EndsAt:    endsAt,
		CreatedAt: now,
		UpdatedAt: now,
		Plan:      plan,
	}, nil
}

func validateCheckoutPlan(plan *schooldomain.Plan) error {
	if plan.IsDefault || plan.MonthlyPrice == 0 && plan.YearlyPrice == 0 {
		return apperror.New(apperror.ErrBadRequest, "trial plan cannot be purchased through payment checkout")
	}
	return nil
}

func validateUpgradePath(current *schooldomain.Subscription, target *schooldomain.Plan) error {
	if current == nil || current.Plan == nil {
		return nil
	}
	if current.Plan.ID == target.ID {
		return apperror.New(apperror.ErrConflict, "selected plan is already active for this school")
	}
	if current.Plan.MaxStudents >= target.MaxStudents {
		return apperror.New(
			apperror.ErrBadRequest,
			fmt.Sprintf("checkout only supports upgrades to a higher plan than %s", current.Plan.Name),
		)
	}
	return nil
}

func mapGatewayStatus(txStatus, fraudStatus string) string {
	switch txStatus {
	case "settlement":
		return billingdomain.PaymentStatusPaid
	case "capture":
		if fraudStatus == "" || fraudStatus == "accept" {
			return billingdomain.PaymentStatusPaid
		}
		return billingdomain.PaymentStatusPending
	case "pending":
		return billingdomain.PaymentStatusPending
	case "deny":
		return billingdomain.PaymentStatusFailed
	case "expire":
		return billingdomain.PaymentStatusExpired
	case "cancel":
		return billingdomain.PaymentStatusCancelled
	default:
		return billingdomain.PaymentStatusPending
	}
}

func syncPaymentWithGatewayStatus(
	ctx context.Context,
	payment *billingdomain.PaymentTransaction,
	status *GatewayTransactionStatus,
	payments billingdomain.PaymentRepository,
	schools SchoolAccessRepository,
	students ActiveStudentCounter,
) (*billingdomain.PaymentTransaction, error) {
	now := time.Now()
	payment.ProviderTransactionID = status.TransactionID
	payment.PaymentType = status.PaymentType
	payment.TransactionStatus = status.TransactionStatus
	payment.FraudStatus = status.FraudStatus
	payment.RawNotification = status.RawResponse
	payment.UpdatedAt = now

	switch mapGatewayStatus(status.TransactionStatus, status.FraudStatus) {
	case billingdomain.PaymentStatusPaid:
		if payment.ActivatedSubscriptionID == nil {
			plan, err := schools.FindPlanByID(ctx, payment.PlanID)
			if err != nil {
				return nil, err
			}
			if plan.MaxStudents > 0 {
				totalStudents, err := students.CountActiveStudents(ctx, payment.SchoolID)
				if err != nil {
					return nil, err
				}
				if totalStudents > int64(plan.MaxStudents) {
					return nil, apperror.New(
						apperror.ErrConflict,
						"payment was received but the selected plan no longer fits the current student count; manual review is required",
					)
				}
			}

			sub, err := buildPaidSubscription(payment.SchoolID, plan, payment.Cycle, now)
			if err != nil {
				return nil, err
			}
			if err := schools.ReplaceSubscription(ctx, sub); err != nil {
				return nil, err
			}
			payment.ActivatedSubscriptionID = &sub.ID
		}
		payment.Status = billingdomain.PaymentStatusPaid
		if payment.PaidAt == nil {
			paidAt := now
			payment.PaidAt = &paidAt
		}
	case billingdomain.PaymentStatusFailed:
		payment.Status = billingdomain.PaymentStatusFailed
	case billingdomain.PaymentStatusExpired:
		payment.Status = billingdomain.PaymentStatusExpired
	case billingdomain.PaymentStatusCancelled:
		payment.Status = billingdomain.PaymentStatusCancelled
	default:
		payment.Status = billingdomain.PaymentStatusPending
	}

	if err := payments.Update(ctx, payment); err != nil {
		return nil, err
	}
	return payment, nil
}
