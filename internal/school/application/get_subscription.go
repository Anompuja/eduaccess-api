package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/school/domain"
	"github.com/google/uuid"
)

// GetSubscriptionQuery holds parameters for fetching a school's subscription.
type GetSubscriptionQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	SchoolID          uuid.UUID
}

// GetSubscriptionHandler returns the active subscription for a school.
type GetSubscriptionHandler struct {
	repo domain.SchoolRepository
}

func NewGetSubscriptionHandler(repo domain.SchoolRepository) *GetSubscriptionHandler {
	return &GetSubscriptionHandler{repo: repo}
}

func (h *GetSubscriptionHandler) Handle(ctx context.Context, q GetSubscriptionQuery) (*domain.Subscription, error) {
	if err := guardSchoolAccess(q.RequesterRole, q.RequesterSchoolID, q.SchoolID); err != nil {
		return nil, err
	}
	return h.repo.FindActiveSubscription(ctx, q.SchoolID)
}
