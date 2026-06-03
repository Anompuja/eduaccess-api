package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/student_tracking/domain"
	"github.com/google/uuid"
)

// guardView allows school staff to read tracking data; students and parents
// are excluded from the school-wide views.
func guardView(role string) error {
	switch role {
	case "superadmin", "admin_sekolah", "kepala_sekolah", "guru", "staff":
		return nil
	}
	return apperror.New(apperror.ErrForbidden, "not allowed to view student tracking")
}

// ── List studies ──────────────────────────────────────────────────────────────

type ListStudiesQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ClassroomID       *uuid.UUID
	AcademicYearID    *uuid.UUID
	ClassID           *uuid.UUID
	StudentID         *uuid.UUID
	Status            *string
}

type ListStudiesHandler struct{ repo domain.Repository }

func NewListStudiesHandler(repo domain.Repository) *ListStudiesHandler {
	return &ListStudiesHandler{repo: repo}
}

func (h *ListStudiesHandler) Handle(ctx context.Context, q ListStudiesQuery) ([]domain.StudyView, error) {
	if err := guardView(q.RequesterRole); err != nil {
		return nil, err
	}
	// Default to active-only so promoted students don't bleed into the old class list.
	// Callers can override with an explicit Status value (e.g. "inactive", "all").
	status := q.Status
	if status == nil {
		active := "active"
		status = &active
	}
	return h.repo.ListStudies(ctx, domain.StudyFilter{
		SchoolID:       q.RequesterSchoolID,
		ClassroomID:    q.ClassroomID,
		AcademicYearID: q.AcademicYearID,
		ClassID:        q.ClassID,
		StudentID:      q.StudentID,
		Status:         status,
	})
}

// ── Student detail (enrollment history for one student) ─────────────────────────

type GetStudentDetailQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	StudentID         uuid.UUID
}

type GetStudentDetailHandler struct{ repo domain.Repository }

func NewGetStudentDetailHandler(repo domain.Repository) *GetStudentDetailHandler {
	return &GetStudentDetailHandler{repo: repo}
}

func (h *GetStudentDetailHandler) Handle(ctx context.Context, q GetStudentDetailQuery) ([]domain.StudyView, error) {
	if err := guardView(q.RequesterRole); err != nil {
		return nil, err
	}
	studies, err := h.repo.ListStudies(ctx, domain.StudyFilter{
		SchoolID:  q.RequesterSchoolID,
		StudentID: &q.StudentID,
	})
	if err != nil {
		return nil, err
	}
	if len(studies) == 0 {
		return nil, apperror.New(apperror.ErrNotFound, "no enrollment records found for this student")
	}
	return studies, nil
}
