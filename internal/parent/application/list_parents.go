package application

import (
	"context"

	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/parent/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

type ListParentsQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	Search            string
	Page              int
	PerPage           int
}

type ListParentsResult struct {
	Parents []*domain.ParentProfile
	Page    int
	PerPage int
	Total   int64
}

type ListParentsHandler struct {
	repo domain.ParentRepository
}

func NewListParentsHandler(repo domain.ParentRepository) *ListParentsHandler {
	return &ListParentsHandler{repo: repo}
}

func (h *ListParentsHandler) Handle(ctx context.Context, q ListParentsQuery) (*ListParentsResult, error) {
	page := q.Page
	if page < 1 {
		page = 1
	}
	perPage := q.PerPage
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var schoolID *uuid.UUID
	if q.RequesterRole != authdomain.RoleSuperadmin {
		if q.RequesterSchoolID == nil {
			return nil, apperror.New(apperror.ErrForbidden, "school context required")
		}
		schoolID = q.RequesterSchoolID
	}

	parents, total, err := h.repo.ListParents(ctx, domain.ParentFilter{
		SchoolID: schoolID,
		Search:   q.Search,
		Offset:   (page - 1) * perPage,
		Limit:    perPage,
	})
	if err != nil {
		return nil, err
	}
	return &ListParentsResult{Parents: parents, Page: page, PerPage: perPage, Total: total}, nil
}
