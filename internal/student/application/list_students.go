package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/student/domain"
	"github.com/google/uuid"
)

// ListStudentsQuery holds pagination and filter parameters.
type ListStudentsQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	EducationLevelID  *uuid.UUID
	ClassID           *uuid.UUID
	SubClassID        *uuid.UUID
	Search            string
	Page              int
	PerPage           int
}

// ListStudentsResult is the paginated response.
type ListStudentsResult struct {
	Students []*domain.StudentProfile
	Page     int
	PerPage  int
	Total    int64
}

// ListStudentsHandler returns a paginated list of students, tenant-scoped.
type ListStudentsHandler struct {
	repo domain.StudentRepository
}

func NewListStudentsHandler(repo domain.StudentRepository) *ListStudentsHandler {
	return &ListStudentsHandler{repo: repo}
}

func (h *ListStudentsHandler) Handle(ctx context.Context, q ListStudentsQuery) (*ListStudentsResult, error) {
	page := q.Page
	if page < 1 {
		page = 1
	}
	perPage := q.PerPage
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var schoolID *uuid.UUID
	if q.RequesterRole != "superadmin" {
		schoolID = q.RequesterSchoolID
	}

	students, total, err := h.repo.ListStudents(ctx, domain.StudentFilter{
		SchoolID:         schoolID,
		EducationLevelID: q.EducationLevelID,
		ClassID:          q.ClassID,
		SubClassID:       q.SubClassID,
		Search:           q.Search,
		Offset:           (page - 1) * perPage,
		Limit:            perPage,
	})
	if err != nil {
		return nil, err
	}
	return &ListStudentsResult{Students: students, Page: page, PerPage: perPage, Total: total}, nil
}
