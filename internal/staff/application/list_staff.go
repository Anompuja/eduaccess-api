package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/staff/domain"
	"github.com/google/uuid"
)

// ListStaffQuery represents a query to list staff with pagination and filtering.
type ListStaffQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	Search            string
	Page              int
	PerPage           int
}

// ListStaffResult contains the result of a list operation.
type ListStaffResult struct {
	Staff   []*domain.StaffProfile
	Page    int
	PerPage int
	Total   int64
}

// ListStaffHandler handles fetching a list of staff with pagination.
type ListStaffHandler struct {
	staffRepo domain.StaffRepository
}

// NewListStaffHandler creates a new ListStaffHandler.
func NewListStaffHandler(staffRepo domain.StaffRepository) *ListStaffHandler {
	return &ListStaffHandler{staffRepo: staffRepo}
}

// Handle retrieves a paginated list of staff with authorization checks.
func (h *ListStaffHandler) Handle(ctx context.Context, q ListStaffQuery) (*ListStaffResult, error) {
	// Set defaults
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PerPage < 1 {
		q.PerPage = 20
	}
	if q.PerPage > 100 {
		q.PerPage = 100
	}

	// Check authorization: requester must have a school assigned
	if q.RequesterSchoolID == nil {
		return nil, apperror.New(apperror.ErrForbidden, "user not assigned to a school")
	}

	// Build filter
	filter := domain.StaffFilter{
		SchoolID: *q.RequesterSchoolID,
		Search:   q.Search,
		Offset:   (q.Page - 1) * q.PerPage,
		Limit:    q.PerPage,
	}

	staff, total, err := h.staffRepo.ListStaff(ctx, filter)
	if err != nil {
		return nil, apperror.New(apperror.ErrBadRequest, "failed to list staff")
	}

	return &ListStaffResult{
		Staff:   staff,
		Page:    q.Page,
		PerPage: q.PerPage,
		Total:   total,
	}, nil
}
