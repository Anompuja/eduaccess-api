package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/school/domain"
	"github.com/google/uuid"
)

// ListSchoolsQuery holds pagination/filter parameters.
type ListSchoolsQuery struct {
	// RequesterSchoolID is non-nil for school-scoped roles — limits result to own school.
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	Search            string
	Status            string
	Page              int
	PerPage           int
}

// ListSchoolsResult is the paginated response.
type ListSchoolsResult struct {
	Schools []*domain.School
	Page    int
	PerPage int
	Total   int64
}

// ListSchoolsHandler returns a paginated list of schools.
type ListSchoolsHandler struct {
	repo domain.SchoolRepository
}

func NewListSchoolsHandler(repo domain.SchoolRepository) *ListSchoolsHandler {
	return &ListSchoolsHandler{repo: repo}
}

func (h *ListSchoolsHandler) Handle(ctx context.Context, q ListSchoolsQuery) (*ListSchoolsResult, error) {
	// Non-superadmin sees only their own school.
	if q.RequesterRole != "superadmin" {
		if q.RequesterSchoolID == nil {
			return &ListSchoolsResult{Schools: []*domain.School{}, Page: 1, PerPage: 1, Total: 0}, nil
		}
		school, err := h.repo.FindByID(ctx, *q.RequesterSchoolID)
		if err != nil {
			return nil, err
		}
		return &ListSchoolsResult{Schools: []*domain.School{school}, Page: 1, PerPage: 1, Total: 1}, nil
	}

	page := q.Page
	if page < 1 {
		page = 1
	}
	perPage := q.PerPage
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	schools, total, err := h.repo.List(ctx, domain.SchoolFilter{
		Search: q.Search,
		Status: q.Status,
		Offset: (page - 1) * perPage,
		Limit:  perPage,
	})
	if err != nil {
		return nil, err
	}
	return &ListSchoolsResult{Schools: schools, Page: page, PerPage: perPage, Total: total}, nil
}
