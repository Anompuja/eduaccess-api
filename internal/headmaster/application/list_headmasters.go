package application

import (
	"context"

	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/headmaster/domain"
	"github.com/google/uuid"
)

// ListHeadmastersQuery filters the headmaster list.
type ListHeadmastersQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	Search            string
	Page              int
	PerPage           int
}

// ListHeadmastersResult is the paginated result.
type ListHeadmastersResult struct {
	Headmasters []*domain.HeadmasterProfile
	Total       int64
	Page        int
	PerPage     int
}

// ListHeadmastersHandler handles ListHeadmastersQuery.
type ListHeadmastersHandler struct {
	repo domain.HeadmasterRepository
}

func NewListHeadmastersHandler(repo domain.HeadmasterRepository) *ListHeadmastersHandler {
	return &ListHeadmastersHandler{repo: repo}
}

func (h *ListHeadmastersHandler) Handle(ctx context.Context, q ListHeadmastersQuery) (*ListHeadmastersResult, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PerPage < 1 || q.PerPage > 100 {
		q.PerPage = 20
	}

	// Non-superadmin is always scoped to their own school.
	schoolID := q.RequesterSchoolID
	if q.RequesterRole == authdomain.RoleSuperadmin {
		schoolID = nil // superadmin may see all
	}

	profiles, total, err := h.repo.ListHeadmasters(ctx, domain.HeadmasterFilter{
		SchoolID: schoolID,
		Search:   q.Search,
		Offset:   (q.Page - 1) * q.PerPage,
		Limit:    q.PerPage,
	})
	if err != nil {
		return nil, err
	}

	return &ListHeadmastersResult{
		Headmasters: profiles,
		Total:       total,
		Page:        q.Page,
		PerPage:     q.PerPage,
	}, nil
}
