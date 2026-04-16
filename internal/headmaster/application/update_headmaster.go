package application

import (
	"context"
	"time"

	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/headmaster/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

// UpdateHeadmasterCommand holds mutable profile fields.
type UpdateHeadmasterCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	HeadmasterID      uuid.UUID
	PhoneNumber       *string
	Address           *string
	Gender            *string
	Religion          *string
	BirthPlace        *string
	BirthDate         *time.Time
	NIK               *string
	KTPImagePath      *string
}

// UpdateHeadmasterHandler handles UpdateHeadmasterCommand.
type UpdateHeadmasterHandler struct {
	repo domain.HeadmasterRepository
}

func NewUpdateHeadmasterHandler(repo domain.HeadmasterRepository) *UpdateHeadmasterHandler {
	return &UpdateHeadmasterHandler{repo: repo}
}

func (h *UpdateHeadmasterHandler) Handle(ctx context.Context, cmd UpdateHeadmasterCommand) (*domain.HeadmasterProfile, error) {
	if cmd.RequesterRole != authdomain.RoleSuperadmin && cmd.RequesterRole != authdomain.RoleAdminSekolah {
		return nil, apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can update a headmaster profile")
	}

	profile, err := h.repo.FindHeadmasterByID(ctx, cmd.HeadmasterID)
	if err != nil {
		return nil, err
	}

	if cmd.RequesterRole != authdomain.RoleSuperadmin {
		if cmd.RequesterSchoolID == nil || profile.SchoolID != *cmd.RequesterSchoolID {
			return nil, apperror.New(apperror.ErrForbidden, "access denied to this headmaster")
		}
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

	profile.UpdatedAt = time.Now()
	if err := h.repo.UpdateHeadmasterProfile(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}
