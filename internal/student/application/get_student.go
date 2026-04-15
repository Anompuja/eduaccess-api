package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/student/domain"
	"github.com/google/uuid"
)

// GetStudentQuery holds parameters for a single-student fetch.
type GetStudentQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	RequesterUserID   uuid.UUID
	StudentID         uuid.UUID
}

// GetStudentHandler fetches a single student with their parent links.
type GetStudentHandler struct {
	repo domain.StudentRepository
}

func NewGetStudentHandler(repo domain.StudentRepository) *GetStudentHandler {
	return &GetStudentHandler{repo: repo}
}

func (h *GetStudentHandler) Handle(ctx context.Context, q GetStudentQuery) (*domain.StudentProfile, error) {
	student, err := h.repo.FindStudentByID(ctx, q.StudentID)
	if err != nil {
		return nil, err
	}

	// Tenant guard: non-superadmin can only see students in their school
	if q.RequesterRole != "superadmin" {
		if q.RequesterSchoolID != nil && student.SchoolID != *q.RequesterSchoolID {
			return nil, apperror.New(apperror.ErrForbidden, "access denied to this student")
		}
	}

	// Load parent links
	links, err := h.repo.ListParentLinks(ctx, q.StudentID)
	if err != nil {
		return nil, err
	}
	student.Parents = links

	return student, nil
}
