package application

import (
	"context"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/staff/domain"
	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/google/uuid"
)

// UpdateStaffCommand represents the command to update a staff profile.
type UpdateStaffCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	StaffID           uuid.UUID
	Name              *string
	Email             *string
	Username          *string
	PhoneNumber       *string
	Address           *string
	Gender            *string
	Religion          *string
	BirthPlace        *string
	BirthDate         *string
	NIK               *string
	KTPImagePath      *string
}

// UpdateStaffHandler handles updating a staff profile.
type UpdateStaffHandler struct {
	staffRepo   domain.StaffRepository
	userUpdater UserUpdater
}

// NewUpdateStaffHandler creates a new UpdateStaffHandler.
func NewUpdateStaffHandler(
	staffRepo domain.StaffRepository,
	userUpdater UserUpdater,
) *UpdateStaffHandler {
	return &UpdateStaffHandler{
		staffRepo:   staffRepo,
		userUpdater: userUpdater,
	}
}

// Handle updates a staff profile with authorization checks.
func (h *UpdateStaffHandler) Handle(ctx context.Context, cmd UpdateStaffCommand) (*domain.StaffProfile, error) {
	if cmd.RequesterRole != authdomain.RoleSuperadmin && cmd.RequesterRole != authdomain.RoleAdminSekolah {
		return nil, apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can update staff")
	}

	staff, err := h.staffRepo.FindStaffByID(ctx, cmd.StaffID)
	if err != nil {
		return nil, apperror.New(apperror.ErrNotFound, "staff not found")
	}

	if cmd.RequesterRole != authdomain.RoleSuperadmin {
		if cmd.RequesterSchoolID == nil || staff.SchoolID != *cmd.RequesterSchoolID {
			return nil, apperror.New(apperror.ErrForbidden, "access denied to this staff")
		}
	}

	user, err := h.userUpdater.FindByID(ctx, staff.UserID)
	if err != nil {
		return nil, apperror.New(apperror.ErrNotFound, "user not found")
	}

	// Check email uniqueness if updating
	if cmd.Email != nil && user.Email != *cmd.Email {
		exists, err := h.userUpdater.ExistsByEmail(ctx, *cmd.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperror.New(apperror.ErrConflict, "email already in use")
		}
		user.Email = *cmd.Email
		staff.Email = *cmd.Email
	}

	// Check username uniqueness if updating
	if cmd.Username != nil && user.Username != *cmd.Username {
		exists, err := h.userUpdater.ExistsByUsername(ctx, *cmd.Username)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperror.New(apperror.ErrConflict, "username already in use")
		}
		user.Username = *cmd.Username
		staff.Username = *cmd.Username
	}

	// Update user account fields
	if cmd.Name != nil {
		user.Name = *cmd.Name
		staff.Name = *cmd.Name
	}

	// Update staff profile fields
	if cmd.PhoneNumber != nil {
		staff.PhoneNumber = cmd.PhoneNumber
	}
	if cmd.Address != nil {
		staff.Address = cmd.Address
	}
	if cmd.Gender != nil {
		staff.Gender = cmd.Gender
	}
	if cmd.Religion != nil {
		staff.Religion = cmd.Religion
	}
	if cmd.BirthPlace != nil {
		staff.BirthPlace = cmd.BirthPlace
	}
	if cmd.BirthDate != nil && *cmd.BirthDate != "" {
		t, err := time.Parse("2006-01-02", *cmd.BirthDate)
		if err != nil {
			return nil, apperror.New(apperror.ErrBadRequest, "birth_date must be YYYY-MM-DD")
		}
		staff.BirthDate = &t
	}
	if cmd.NIK != nil {
		staff.NIK = cmd.NIK
	}
	if cmd.KTPImagePath != nil {
		staff.KTPImagePath = cmd.KTPImagePath
	}

	now := time.Now()
	user.UpdatedAt = now
	staff.UpdatedAt = now

	if err := h.userUpdater.Update(ctx, user); err != nil {
		return nil, err
	}
	if err := h.staffRepo.UpdateStaffProfile(ctx, staff); err != nil {
		return nil, err
	}

	return staff, nil
}
