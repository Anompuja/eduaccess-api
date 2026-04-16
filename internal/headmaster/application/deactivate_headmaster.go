package application

import (
	"context"

	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/headmaster/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

// DeactivateHeadmasterCommand identifies the profile to soft-delete.
type DeactivateHeadmasterCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	HeadmasterID      uuid.UUID
}

// DeactivateHeadmasterHandler handles DeactivateHeadmasterCommand.
type DeactivateHeadmasterHandler struct {
	repo domain.HeadmasterRepository
}

func NewDeactivateHeadmasterHandler(repo domain.HeadmasterRepository) *DeactivateHeadmasterHandler {
	return &DeactivateHeadmasterHandler{repo: repo}
}

func (h *DeactivateHeadmasterHandler) Handle(ctx context.Context, cmd DeactivateHeadmasterCommand) error {
	if cmd.RequesterRole != authdomain.RoleSuperadmin && cmd.RequesterRole != authdomain.RoleAdminSekolah {
		return apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can deactivate a headmaster")
	}

	profile, err := h.repo.FindHeadmasterByID(ctx, cmd.HeadmasterID)
	if err != nil {
		return err
	}

	if cmd.RequesterRole != authdomain.RoleSuperadmin {
		if cmd.RequesterSchoolID == nil || profile.SchoolID != *cmd.RequesterSchoolID {
			return apperror.New(apperror.ErrForbidden, "access denied to this headmaster")
		}
	}

	return h.repo.SoftDeleteHeadmaster(ctx, cmd.HeadmasterID)
}
