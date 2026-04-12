package application

import (
	"context"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ChangePasswordCommand changes a user's password.
type ChangePasswordCommand struct {
	RequesterID   uuid.UUID
	RequesterRole string
	UserID        uuid.UUID
	OldPassword   string // required when changing own password
	NewPassword   string
}

// ChangePasswordHandler handles the ChangePasswordCommand.
type ChangePasswordHandler struct {
	users UserWriteRepository
}

func NewChangePasswordHandler(users UserWriteRepository) *ChangePasswordHandler {
	return &ChangePasswordHandler{users: users}
}

func (h *ChangePasswordHandler) Handle(ctx context.Context, cmd ChangePasswordCommand) error {
	user, err := h.users.FindByID(ctx, cmd.UserID)
	if err != nil {
		return apperror.New(apperror.ErrNotFound, "user not found")
	}

	isSelf := user.ID == cmd.RequesterID
	isSuperadmin := cmd.RequesterRole == domain.RoleSuperadmin

	if !isSelf && !isSuperadmin {
		return apperror.New(apperror.ErrForbidden, "can only change your own password")
	}

	// Verify current password when the user is changing their own
	if isSelf {
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(cmd.OldPassword)); err != nil {
			return apperror.New(apperror.ErrWrongPassword, "current password is incorrect")
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cmd.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperror.New(apperror.ErrInternal, "failed to hash password")
	}

	user.Password = string(hash)
	user.UpdatedAt = time.Now()

	if err := h.users.Update(ctx, user); err != nil {
		return apperror.New(apperror.ErrInternal, "failed to update password")
	}

	return nil
}
