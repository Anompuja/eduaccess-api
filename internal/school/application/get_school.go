package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/school/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

// GetSchoolQuery holds parameters for a single-school fetch.
type GetSchoolQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	SchoolID          uuid.UUID
}

// GetSchoolHandler retrieves a school by ID.
// Superadmin can fetch any school; others can only fetch their own.
type GetSchoolHandler struct {
	repo domain.SchoolRepository
}

func NewGetSchoolHandler(repo domain.SchoolRepository) *GetSchoolHandler {
	return &GetSchoolHandler{repo: repo}
}

func (h *GetSchoolHandler) Handle(ctx context.Context, q GetSchoolQuery) (*domain.School, error) {
	// Tenant guard
	if q.RequesterRole != "superadmin" {
		if q.RequesterSchoolID == nil || *q.RequesterSchoolID != q.SchoolID {
			return nil, apperror.New(apperror.ErrForbidden, "access denied to this school")
		}
	}

	school, err := h.repo.FindByID(ctx, q.SchoolID)
	if err != nil {
		return nil, err
	}
	return school, nil
}
