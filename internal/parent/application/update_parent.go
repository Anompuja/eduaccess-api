package application

import (
	"context"
	"strings"
	"time"

	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/parent/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

type UpdateParentCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ParentID          uuid.UUID
	Name              *string
	Email             *string
	Religion          *string
	PhoneNumber       *string
	Address           *string
}

type UpdateParentHandler struct {
	users UserUpdater
	repo  domain.ParentRepository
}

func NewUpdateParentHandler(users UserUpdater, repo domain.ParentRepository) *UpdateParentHandler {
	return &UpdateParentHandler{users: users, repo: repo}
}

func (h *UpdateParentHandler) Handle(ctx context.Context, cmd UpdateParentCommand) (*domain.ParentProfile, error) {
	if cmd.RequesterRole != authdomain.RoleSuperadmin && cmd.RequesterRole != authdomain.RoleAdminSekolah {
		return nil, apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can update parent profiles")
	}

	profile, err := h.repo.FindParentByID(ctx, cmd.ParentID)
	if err != nil {
		return nil, err
	}
	if cmd.RequesterRole != authdomain.RoleSuperadmin {
		if cmd.RequesterSchoolID == nil || profile.SchoolID != *cmd.RequesterSchoolID {
			return nil, apperror.New(apperror.ErrForbidden, "access denied to this parent")
		}
	}

	user, err := h.users.FindByID(ctx, profile.UserID)
	if err != nil {
		return nil, apperror.New(apperror.ErrNotFound, "user not found")
	}

	if cmd.Email != nil && !strings.EqualFold(user.Email, *cmd.Email) {
		exists, err := h.users.ExistsByEmail(ctx, *cmd.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperror.New(apperror.ErrConflict, "email already in use")
		}
		user.Email = *cmd.Email
		profile.Email = *cmd.Email
	}

	if cmd.Name != nil {
		user.Name = *cmd.Name
		profile.Name = *cmd.Name
	}

	if cmd.Religion != nil {
		profile.Religion = *cmd.Religion
	}
	if cmd.PhoneNumber != nil {
		profile.PhoneNumber = *cmd.PhoneNumber
	}
	if cmd.Address != nil {
		profile.Address = *cmd.Address
	}

	now := time.Now()
	user.UpdatedAt = now
	profile.UpdatedAt = now

	if err := h.users.Update(ctx, user); err != nil {
		return nil, err
	}
	if err := h.repo.UpdateParentProfile(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}
