package application

import (
	"context"

	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/teacher/domain"
	"github.com/google/uuid"
)

// GetTeacherQuery represents a query to get a teacher by ID.
type GetTeacherQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	TeacherID         uuid.UUID
}

// GetTeacherHandler handles fetching a single teacher by ID.
type GetTeacherHandler struct {
	teacherRepo domain.TeacherRepository
}

// NewGetTeacherHandler creates a new GetTeacherHandler.
func NewGetTeacherHandler(teacherRepo domain.TeacherRepository) *GetTeacherHandler {
	return &GetTeacherHandler{teacherRepo: teacherRepo}
}

// Handle retrieves a single teacher by ID with authorization checks.
func (h *GetTeacherHandler) Handle(ctx context.Context, q GetTeacherQuery) (*domain.TeacherProfile, error) {
	teacher, err := h.teacherRepo.FindTeacherByID(ctx, q.TeacherID)
	if err != nil {
		return nil, apperror.New(apperror.ErrNotFound, "teacher not found")
	}

	// Authorization check
	if q.RequesterRole == authdomain.RoleAdminSekolah {
		if q.RequesterSchoolID == nil {
			return nil, apperror.New(apperror.ErrForbidden, "user not assigned to a school")
		}
		if teacher.SchoolID != *q.RequesterSchoolID {
			return nil, apperror.New(apperror.ErrForbidden, "not authorized to access this teacher")
		}
	}

	return teacher, nil
}
