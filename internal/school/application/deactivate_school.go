package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/school/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

// DeactivateSchoolCommand holds soft-delete parameters.
type DeactivateSchoolCommand struct {
	RequesterRole string
	SchoolID      uuid.UUID
}

// DeactivateSchoolHandler soft-deletes a school. Superadmin only.
type DeactivateSchoolHandler struct {
	repo domain.SchoolRepository
}

func NewDeactivateSchoolHandler(repo domain.SchoolRepository) *DeactivateSchoolHandler {
	return &DeactivateSchoolHandler{repo: repo}
}

func (h *DeactivateSchoolHandler) Handle(ctx context.Context, cmd DeactivateSchoolCommand) error {
	if cmd.RequesterRole != "superadmin" {
		return apperror.New(apperror.ErrForbidden, "only superadmin can delete schools")
	}

	_, err := h.repo.FindByID(ctx, cmd.SchoolID)
	if err != nil {
		return err
	}

	return h.repo.SoftDelete(ctx, cmd.SchoolID)
}
