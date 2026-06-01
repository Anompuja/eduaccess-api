package application

import (
	"context"

	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/parent/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

type DeactivateParentCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ParentID          uuid.UUID
}

type DeactivateParentHandler struct {
	repo domain.ParentRepository
}

func NewDeactivateParentHandler(repo domain.ParentRepository) *DeactivateParentHandler {
	return &DeactivateParentHandler{repo: repo}
}

func (h *DeactivateParentHandler) Handle(ctx context.Context, cmd DeactivateParentCommand) error {
	if cmd.RequesterRole != authdomain.RoleSuperadmin && cmd.RequesterRole != authdomain.RoleAdminSekolah {
		return apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can deactivate parents")
	}

	profile, err := h.repo.FindParentByID(ctx, cmd.ParentID)
	if err != nil {
		return err
	}
	if cmd.RequesterRole != authdomain.RoleSuperadmin {
		if cmd.RequesterSchoolID == nil || profile.SchoolID != *cmd.RequesterSchoolID {
			return apperror.New(apperror.ErrForbidden, "access denied to this parent")
		}
	}

	return h.repo.SoftDeleteParent(ctx, cmd.ParentID)
}
