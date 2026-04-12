package application

import (
	"context"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

// UpdateUserCommand is the input for updating a user's profile fields.
type UpdateUserCommand struct {
	RequesterID       uuid.UUID
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	UserID            uuid.UUID
	Name              *string
	Avatar            *string
}

// UpdateUserHandler handles the UpdateUserCommand.
type UpdateUserHandler struct {
	users UserWriteRepository
}

func NewUpdateUserHandler(users UserWriteRepository) *UpdateUserHandler {
	return &UpdateUserHandler{users: users}
}

func (h *UpdateUserHandler) Handle(ctx context.Context, cmd UpdateUserCommand) (*domain.User, error) {
	user, err := h.users.FindByID(ctx, cmd.UserID)
	if err != nil {
		return nil, apperror.New(apperror.ErrNotFound, "user not found")
	}

	// Tenant guard: non-superadmin can only update users in their own school
	if cmd.RequesterRole != domain.RoleSuperadmin {
		if user.SchoolID == nil || cmd.RequesterSchoolID == nil || *user.SchoolID != *cmd.RequesterSchoolID {
			return nil, apperror.New(apperror.ErrForbidden, "access denied")
		}
	}

	if cmd.Name != nil {
		user.Name = *cmd.Name
	}
	if cmd.Avatar != nil {
		user.Avatar = *cmd.Avatar
	}
	user.UpdatedAt = time.Now()

	if err := h.users.Update(ctx, user); err != nil {
		return nil, apperror.New(apperror.ErrInternal, "failed to update user")
	}

	return user, nil
}
