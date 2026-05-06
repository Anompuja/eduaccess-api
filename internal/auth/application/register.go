package application

import (
	"context"
	"strings"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

// RegisterCommand is the input for the register use-case.
type RegisterCommand struct {
	SchoolID *uuid.UUID
	Role     string
	Name     string
	Username string  // optional — derived from email prefix if blank
	Email    string
	Password string  // plaintext; Supabase Auth will hash it
}

// RegisterResult is returned on success.
type RegisterResult struct {
	UserID uuid.UUID
}

// RegisterHandler handles the RegisterCommand.
type RegisterHandler struct {
	users domain.UserRepository
}

func NewRegisterHandler(users domain.UserRepository) *RegisterHandler {
	return &RegisterHandler{users: users}
}

func (h *RegisterHandler) Handle(ctx context.Context, cmd RegisterCommand) (*RegisterResult, error) {
	username := cmd.Username
	if username == "" {
		parts := strings.SplitN(cmd.Email, "@", 2)
		username = parts[0]
	}

	emailExists, err := h.users.ExistsByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, apperror.New(apperror.ErrInternal, "failed to check email")
	}
	if emailExists {
		return nil, apperror.New(apperror.ErrConflict, "email already registered")
	}

	usernameExists, err := h.users.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, apperror.New(apperror.ErrInternal, "failed to check username")
	}
	if usernameExists {
		return nil, apperror.New(apperror.ErrConflict, "username already taken")
	}

	now := time.Now()
	user := &domain.User{
		SchoolID:  cmd.SchoolID,
		Role:      cmd.Role,
		Name:      cmd.Name,
		Username:  username,
		Email:     cmd.Email,
		Password:  cmd.Password, // passed to Supabase Admin API, never stored in our DB
		Avatar:    "default.png",
		Verified:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.users.Create(ctx, user); err != nil {
		return nil, err
	}

	return &RegisterResult{UserID: user.ID}, nil
}
