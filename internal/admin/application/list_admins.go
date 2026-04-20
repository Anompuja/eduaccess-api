package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/admin/domain"
	"github.com/google/uuid"
)

// ListAdminsQuery holds pagination and filter parameters.
type ListAdminsQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	Search            string // name, email, username
	Page              int
	PerPage           int
}

// ListAdminsResult is the paginated response.
type ListAdminsResult struct {
	Admins  []*domain.AdminProfile
	Page    int
	PerPage int
	Total   int64
}

// ListAdminsHandler returns a paginated list of admins, tenant-scoped.
type ListAdminsHandler struct {
	repo domain.AdminRepository
}

func NewListAdminsHandler(repo domain.AdminRepository) *ListAdminsHandler {
	return &ListAdminsHandler{repo: repo}
}

func (h *ListAdminsHandler) Handle(ctx context.Context, q ListAdminsQuery) (*ListAdminsResult, error) {
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

	admins, total, err := h.repo.ListAdmins(ctx, domain.AdminFilter{
		SchoolID: schoolID,
		Search:   q.Search,
		Offset:   (page - 1) * perPage,
		Limit:    perPage,
	})
	if err != nil {
		return nil, err
	}
	return &ListAdminsResult{Admins: admins, Page: page, PerPage: perPage, Total: total}, nil
}
