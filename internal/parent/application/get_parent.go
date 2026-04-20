package application

import (
	"context"

	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/parent/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

type GetParentQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ParentID          uuid.UUID
}

type GetParentHandler struct {
	repo domain.ParentRepository
}

func NewGetParentHandler(repo domain.ParentRepository) *GetParentHandler {
	return &GetParentHandler{repo: repo}
}

func (h *GetParentHandler) Handle(ctx context.Context, q GetParentQuery) (*domain.ParentProfile, error) {
	parent, err := h.repo.FindParentByID(ctx, q.ParentID)
	if err != nil {
		return nil, err
	}

	if q.RequesterRole != authdomain.RoleSuperadmin {
		if q.RequesterSchoolID == nil || parent.SchoolID != *q.RequesterSchoolID {
			return nil, apperror.New(apperror.ErrForbidden, "access denied to this parent")
		}
	}

	return parent, nil
}
