package application

import (
	"context"
	"strings"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/admin/domain"
	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

// UpdateAdminCommand holds mutable fields for admin profile and account.
type UpdateAdminCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	AdminID           uuid.UUID

	// User account fields
	Name     *string
	Email    *string
	Username *string

	// Profile fields
	PhoneNumber  *string
	Address      *string
	Gender       *string
	Religion     *string
	BirthPlace   *string
	BirthDate    *time.Time
	NIK          *string
	KTPImagePath *string
}

// UpdateAdminHandler updates admin profile fields.
type UpdateAdminHandler struct {
	users UserUpdater
	repo  domain.AdminRepository
}

func NewUpdateAdminHandler(users UserUpdater, repo domain.AdminRepository) *UpdateAdminHandler {
	return &UpdateAdminHandler{users: users, repo: repo}
}

func (h *UpdateAdminHandler) Handle(ctx context.Context, cmd UpdateAdminCommand) (*domain.AdminProfile, error) {
	if cmd.RequesterRole != authdomain.RoleSuperadmin && cmd.RequesterRole != authdomain.RoleAdminSekolah {
		return nil, apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can update admin")
	}

	profile, err := h.repo.FindAdminByID(ctx, cmd.AdminID)
	if err != nil {
		return nil, err
	}

	if cmd.RequesterRole != authdomain.RoleSuperadmin {
		if cmd.RequesterSchoolID == nil || profile.SchoolID != *cmd.RequesterSchoolID {
			return nil, apperror.New(apperror.ErrForbidden, "access denied to this admin")
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

	if cmd.Username != nil && user.Username != *cmd.Username {
		exists, err := h.users.ExistsByUsername(ctx, *cmd.Username)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperror.New(apperror.ErrConflict, "username already in use")
		}
		user.Username = *cmd.Username
		profile.Username = *cmd.Username
	}

	if cmd.Name != nil {
		user.Name = *cmd.Name
		profile.Name = *cmd.Name
	}
	if cmd.PhoneNumber != nil {
		profile.PhoneNumber = *cmd.PhoneNumber
	}
	if cmd.Address != nil {
		profile.Address = *cmd.Address
	}
	if cmd.Gender != nil {
		profile.Gender = *cmd.Gender
	}
	if cmd.Religion != nil {
		profile.Religion = *cmd.Religion
	}
	if cmd.BirthPlace != nil {
		profile.BirthPlace = *cmd.BirthPlace
	}
	if cmd.BirthDate != nil {
		profile.BirthDate = cmd.BirthDate
	}
	if cmd.NIK != nil {
		profile.NIK = *cmd.NIK
	}
	if cmd.KTPImagePath != nil {
		profile.KTPImagePath = *cmd.KTPImagePath
	}

	now := time.Now()
	user.UpdatedAt = now
	profile.UpdatedAt = now

	if err := h.users.Update(ctx, user); err != nil {
		return nil, err
	}
	if err := h.repo.UpdateAdminProfile(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}
