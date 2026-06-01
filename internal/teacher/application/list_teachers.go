package application

import (
	"context"

	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/teacher/domain"
	"github.com/google/uuid"
)

// ListTeachersQuery represents a query to list teachers with pagination and filtering.
type ListTeachersQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	SchoolID          *uuid.UUID
	Search            string
	Page              int
	PerPage           int
}

// ListTeachersResult contains the result of a list operation.
type ListTeachersResult struct {
	Teachers []*domain.TeacherProfile
	Page     int
	PerPage  int
	Total    int64
}

// ListTeachersHandler handles fetching a list of teachers with pagination.
type ListTeachersHandler struct {
	teacherRepo domain.TeacherRepository
}

// NewListTeachersHandler creates a new ListTeachersHandler.
func NewListTeachersHandler(teacherRepo domain.TeacherRepository) *ListTeachersHandler {
	return &ListTeachersHandler{teacherRepo: teacherRepo}
}

// Handle retrieves a paginated list of teachers with authorization checks.
func (h *ListTeachersHandler) Handle(ctx context.Context, q ListTeachersQuery) (*ListTeachersResult, error) {
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

	var schoolID *uuid.UUID
	switch q.RequesterRole {
	case authdomain.RoleSuperadmin:
		schoolID = q.SchoolID
	case authdomain.RoleAdminSekolah:
		if q.RequesterSchoolID == nil {
			return nil, apperror.New(apperror.ErrForbidden, "user not assigned to a school")
		}
		schoolID = q.RequesterSchoolID
	default:
		return nil, apperror.New(apperror.ErrForbidden, "insufficient permissions")
	}

	filter := domain.TeacherFilter{
		SchoolID: schoolID,
		Search:   q.Search,
		Offset:   (q.Page - 1) * q.PerPage,
		Limit:    q.PerPage,
	}

	teachers, total, err := h.teacherRepo.ListTeachers(ctx, filter)
	if err != nil {
		return nil, apperror.New(apperror.ErrBadRequest, "failed to list teachers")
	}

	return &ListTeachersResult{
		Teachers: teachers,
		Page:     q.Page,
		PerPage:  q.PerPage,
		Total:    total,
	}, nil
}
