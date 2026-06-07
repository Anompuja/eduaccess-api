package application

import (
	"context"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/auth/domain"
	schooldomain "github.com/eduaccess/eduaccess-api/internal/school/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

// UpdateSubscriptionCommand changes the active subscription for a school.
type UpdateSubscriptionCommand struct {
	RequesterRole string
	SchoolID      uuid.UUID
	PlanID        uuid.UUID
	Cycle         string
}

// UpdateSubscriptionHandler replaces the active subscription for a school.
type UpdateSubscriptionHandler struct {
	repo          schooldomain.SchoolRepository
	studentCounts ActiveStudentCounter
}

type ActiveStudentCounter interface {
	CountActiveStudents(ctx context.Context, schoolID uuid.UUID) (int64, error)
}

func NewUpdateSubscriptionHandler(repo schooldomain.SchoolRepository, studentCounts ...ActiveStudentCounter) *UpdateSubscriptionHandler {
	var counter ActiveStudentCounter
	if len(studentCounts) > 0 {
		counter = studentCounts[0]
	}
	return &UpdateSubscriptionHandler{repo: repo, studentCounts: counter}
}

func (h *UpdateSubscriptionHandler) Handle(ctx context.Context, cmd UpdateSubscriptionCommand) (*schooldomain.Subscription, error) {
	if cmd.RequesterRole != domain.RoleSuperadmin {
		return nil, apperror.New(apperror.ErrForbidden, "only superadmin can change school subscriptions")
	}

	if _, err := h.repo.FindByID(ctx, cmd.SchoolID); err != nil {
		return nil, err
	}

	plan, err := h.repo.FindPlanByID(ctx, cmd.PlanID)
	if err != nil {
		return nil, err
	}
	if h.studentCounts != nil && plan.MaxStudents > 0 {
		totalStudents, err := h.studentCounts.CountActiveStudents(ctx, cmd.SchoolID)
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

	price, endsAt, err := buildSubscriptionBilling(plan, cmd.Cycle)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	sub := &schooldomain.Subscription{
		ID:        uuid.New(),
		SchoolID:  cmd.SchoolID,
		PlanID:    plan.ID,
		Status:    "active",
		Cycle:     cmd.Cycle,
		Quantity:  1,
		Price:     price,
		EndsAt:    endsAt,
		CreatedAt: now,
		UpdatedAt: now,
		Plan:      plan,
	}

	if err := h.repo.ReplaceSubscription(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

func buildSubscriptionBilling(plan *schooldomain.Plan, cycle string) (int64, *time.Time, error) {
	now := time.Now()

	switch cycle {
	case "month":
		endsAt := now.AddDate(0, 1, 0)
		return plan.MonthlyPrice, &endsAt, nil
	case "year":
		endsAt := now.AddDate(1, 0, 0)
		return plan.YearlyPrice, &endsAt, nil
	case "onetime":
		if plan.OnetimePrice == nil {
			return 0, nil, apperror.New(apperror.ErrBadRequest, "selected plan does not support one-time billing")
		}
		return *plan.OnetimePrice, nil, nil
	default:
		return 0, nil, apperror.New(apperror.ErrBadRequest, "cycle must be one of: month, year, onetime")
	}
}
