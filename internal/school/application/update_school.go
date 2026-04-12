package application

import (
	"context"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/school/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

// UpdateSchoolCommand holds the fields that may be changed.
type UpdateSchoolCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	SchoolID          uuid.UUID
	Name              *string
	Address           *string
	Phone             *string
	Email             *string
	Description       *string
	ImagePath         *string
	TimeZone          *string
	Status            *string // superadmin only
}

// UpdateSchoolHandler updates mutable school fields.
// Superadmin can update any school and set Status.
// admin_sekolah can update only their own school (no Status change).
type UpdateSchoolHandler struct {
	repo domain.SchoolRepository
}

func NewUpdateSchoolHandler(repo domain.SchoolRepository) *UpdateSchoolHandler {
	return &UpdateSchoolHandler{repo: repo}
}

func (h *UpdateSchoolHandler) Handle(ctx context.Context, cmd UpdateSchoolCommand) (*domain.School, error) {
	if cmd.RequesterRole != "superadmin" {
		if cmd.RequesterSchoolID == nil || *cmd.RequesterSchoolID != cmd.SchoolID {
			return nil, apperror.New(apperror.ErrForbidden, "access denied to this school")
		}
		if cmd.RequesterRole != "admin_sekolah" {
			return nil, apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can update school info")
		}
	}

	school, err := h.repo.FindByID(ctx, cmd.SchoolID)
	if err != nil {
		return nil, err
	}

	if cmd.Name != nil {
		school.Name = *cmd.Name
	}
	if cmd.Address != nil {
		school.Address = *cmd.Address
	}
	if cmd.Phone != nil {
		school.Phone = *cmd.Phone
	}
	if cmd.Email != nil {
		school.Email = *cmd.Email
	}
	if cmd.Description != nil {
		school.Description = *cmd.Description
	}
	if cmd.ImagePath != nil {
		school.ImagePath = *cmd.ImagePath
	}
	if cmd.TimeZone != nil {
		school.TimeZone = *cmd.TimeZone
	}
	if cmd.Status != nil && cmd.RequesterRole == "superadmin" {
		if *cmd.Status != domain.StatusActive && *cmd.Status != domain.StatusNonactive {
			return nil, apperror.New(apperror.ErrBadRequest, "status must be 'active' or 'nonactive'")
		}
		school.Status = *cmd.Status
	}

	school.UpdatedAt = time.Now()
	if err := h.repo.Update(ctx, school); err != nil {
		return nil, err
	}
	return school, nil
}
