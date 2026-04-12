package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/google/uuid"
)

// ListUsersQuery filters for the list-users use-case.
type ListUsersQuery struct {
	SchoolID *uuid.UUID // nil = superadmin sees all
	Role     string     // optional filter
	Search   string     // optional name/email/username search
	Page     int
	PerPage  int
}

// ListUsersResult is the paginated result.
type ListUsersResult struct {
	Users   []*domain.User
	Total   int64
	Page    int
	PerPage int
}

// ListUsersHandler handles the ListUsersQuery.
type ListUsersHandler struct {
	users UserReadRepository
}

func NewListUsersHandler(users UserReadRepository) *ListUsersHandler {
	return &ListUsersHandler{users: users}
}

func (h *ListUsersHandler) Handle(ctx context.Context, q ListUsersQuery) (*ListUsersResult, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PerPage < 1 || q.PerPage > 100 {
		q.PerPage = 20
	}

	users, total, err := h.users.List(ctx, ListFilter{
		SchoolID: q.SchoolID,
		Role:     q.Role,
		Search:   q.Search,
		Offset:   (q.Page - 1) * q.PerPage,
		Limit:    q.PerPage,
	})
	if err != nil {
		return nil, err
	}

	return &ListUsersResult{
		Users:   users,
		Total:   total,
		Page:    q.Page,
		PerPage: q.PerPage,
	}, nil
}
