package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	supabasePkg "github.com/eduaccess/eduaccess-api/pkg/supabase"
	"github.com/google/uuid"
)

// ChangePasswordCommand changes a user's password via Supabase Auth Admin API.
type ChangePasswordCommand struct {
	RequesterID   uuid.UUID
	RequesterRole string
	UserID        uuid.UUID
	OldPassword   string // required when the user changes their own password
	NewPassword   string
}

// ChangePasswordHandler handles the ChangePasswordCommand.
type ChangePasswordHandler struct {
	users    UserWriteRepository
	supabase *supabasePkg.Client
}

func NewChangePasswordHandler(users UserWriteRepository, supabase *supabasePkg.Client) *ChangePasswordHandler {
	return &ChangePasswordHandler{users: users, supabase: supabase}
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

	// When a user changes their own password, verify the current one first
	// by attempting a Supabase sign-in — we no longer store bcrypt hashes.
	if isSelf {
		if err := h.supabase.VerifyPassword(ctx, user.Email, cmd.OldPassword); err != nil {
			return err // already ErrWrongPassword AppError
		}
	}

	return h.supabase.UpdateUserPassword(ctx, cmd.UserID, cmd.NewPassword)
}
