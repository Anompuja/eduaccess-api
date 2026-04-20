package application

import (
	"context"
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
	FatherName        *string
	MotherName        *string
	FatherReligion    *string
	MotherReligion    *string
	PhoneNumber       *string
	Address           *string
}

type UpdateParentHandler struct {
	repo domain.ParentRepository
}

func NewUpdateParentHandler(repo domain.ParentRepository) *UpdateParentHandler {
	return &UpdateParentHandler{repo: repo}
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

	if cmd.FatherName != nil {
		profile.FatherName = *cmd.FatherName
	}
	if cmd.MotherName != nil {
		profile.MotherName = *cmd.MotherName
	}
	if cmd.FatherReligion != nil {
		profile.FatherReligion = *cmd.FatherReligion
	}
	if cmd.MotherReligion != nil {
		profile.MotherReligion = *cmd.MotherReligion
	}
	if cmd.PhoneNumber != nil {
		profile.PhoneNumber = *cmd.PhoneNumber
	}
	if cmd.Address != nil {
		profile.Address = *cmd.Address
	}

	profile.UpdatedAt = time.Now()
	if err := h.repo.UpdateParentProfile(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}
