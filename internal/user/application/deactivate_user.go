package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

// DeactivateUserCommand soft-deletes a user.
type DeactivateUserCommand struct {
	RequesterID       uuid.UUID
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	UserID            uuid.UUID
}

// DeactivateUserHandler handles the DeactivateUserCommand.
type DeactivateUserHandler struct {
	users UserWriteRepository
}

func NewDeactivateUserHandler(users UserWriteRepository) *DeactivateUserHandler {
	return &DeactivateUserHandler{users: users}
}

func (h *DeactivateUserHandler) Handle(ctx context.Context, cmd DeactivateUserCommand) error {
	user, err := h.users.FindByID(ctx, cmd.UserID)
	if err != nil {
		return apperror.New(apperror.ErrNotFound, "user not found")
	}

	// Cannot deactivate yourself
	if user.ID == cmd.RequesterID {
		return apperror.New(apperror.ErrBadRequest, "cannot deactivate your own account")
	}

	// Tenant guard
	if cmd.RequesterRole != domain.RoleSuperadmin {
		if user.SchoolID == nil || cmd.RequesterSchoolID == nil || *user.SchoolID != *cmd.RequesterSchoolID {
			return apperror.New(apperror.ErrForbidden, "access denied")
		}
		// admin_sekolah cannot deactivate another admin_sekolah
		if user.Role == domain.RoleAdminSekolah {
			return apperror.New(apperror.ErrForbidden, "cannot deactivate another admin")
		}
	}

	return h.users.SoftDelete(ctx, cmd.UserID)
}
