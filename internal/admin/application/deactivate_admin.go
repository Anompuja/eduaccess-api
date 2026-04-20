package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/admin/domain"
	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

// DeactivateAdminCommand holds soft-delete parameters.
type DeactivateAdminCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	AdminID           uuid.UUID
}

// DeactivateAdminHandler soft-deletes an admin profile.
type DeactivateAdminHandler struct {
	repo domain.AdminRepository
}

func NewDeactivateAdminHandler(repo domain.AdminRepository) *DeactivateAdminHandler {
	return &DeactivateAdminHandler{repo: repo}
}

func (h *DeactivateAdminHandler) Handle(ctx context.Context, cmd DeactivateAdminCommand) error {
	if cmd.RequesterRole != authdomain.RoleSuperadmin && cmd.RequesterRole != authdomain.RoleAdminSekolah {
		return apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can deactivate admin")
	}

	admin, err := h.repo.FindAdminByID(ctx, cmd.AdminID)
	if err != nil {
		return err
	}

	if cmd.RequesterRole != authdomain.RoleSuperadmin {
		if cmd.RequesterSchoolID == nil || admin.SchoolID != *cmd.RequesterSchoolID {
			return apperror.New(apperror.ErrForbidden, "access denied to this admin")
		}
	}

	return h.repo.SoftDeleteAdmin(ctx, cmd.AdminID)
}
