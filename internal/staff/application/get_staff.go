package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/staff/domain"
	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/google/uuid"
)

// GetStaffQuery represents a query to get a staff by ID.
type GetStaffQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	StaffID           uuid.UUID
}

// GetStaffHandler handles fetching a single staff by ID.
type GetStaffHandler struct {
	staffRepo domain.StaffRepository
}

// NewGetStaffHandler creates a new GetStaffHandler.
func NewGetStaffHandler(staffRepo domain.StaffRepository) *GetStaffHandler {
	return &GetStaffHandler{staffRepo: staffRepo}
}

// Handle retrieves a single staff by ID with authorization checks.
func (h *GetStaffHandler) Handle(ctx context.Context, q GetStaffQuery) (*domain.StaffProfile, error) {
	staff, err := h.staffRepo.FindStaffByID(ctx, q.StaffID)
	if err != nil {
		return nil, apperror.New(apperror.ErrNotFound, "staff not found")
	}

	// Authorization check
	if q.RequesterRole != authdomain.RoleSuperadmin {
		if q.RequesterSchoolID == nil || staff.SchoolID != *q.RequesterSchoolID {
			return nil, apperror.New(apperror.ErrForbidden, "not authorized to access this staff")
		}
	}

	return staff, nil
}
