package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/school/domain"
)

// ListPlansHandler returns the active subscription plans available to schools.
type ListPlansHandler struct {
	repo domain.SchoolRepository
}

func NewListPlansHandler(repo domain.SchoolRepository) *ListPlansHandler {
	return &ListPlansHandler{repo: repo}
}

func (h *ListPlansHandler) Handle(ctx context.Context) ([]*domain.Plan, error) {
	return h.repo.ListPlans(ctx)
}
