package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/staff/domain"
	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/google/uuid"
)

// DeactivateStaffCommand represents the command to deactivate a staff.
type DeactivateStaffCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	StaffID           uuid.UUID
}

// DeactivateStaffHandler handles deactivating a staff (soft delete).
type DeactivateStaffHandler struct {
	staffRepo domain.StaffRepository
}

// NewDeactivateStaffHandler creates a new DeactivateStaffHandler.
func NewDeactivateStaffHandler(staffRepo domain.StaffRepository) *DeactivateStaffHandler {
	return &DeactivateStaffHandler{staffRepo: staffRepo}
}

// Handle deactivates a staff with authorization checks.
func (h *DeactivateStaffHandler) Handle(ctx context.Context, cmd DeactivateStaffCommand) error {
	// Fetch the staff first to check authorization
	staff, err := h.staffRepo.FindStaffByID(ctx, cmd.StaffID)
	if err != nil {
		return apperror.New(apperror.ErrNotFound, "staff not found")
	}

	// Authorization check
	if cmd.RequesterRole != authdomain.RoleSuperadmin {
		if cmd.RequesterSchoolID == nil || staff.SchoolID != *cmd.RequesterSchoolID {
			return apperror.New(apperror.ErrForbidden, "access denied to this staff")
		}
	}

	return h.staffRepo.SoftDeleteStaff(ctx, cmd.StaffID)
}
