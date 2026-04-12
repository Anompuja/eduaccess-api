package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

// GetUserQuery is the input for fetching a single user.
type GetUserQuery struct {
	RequesterSchoolID *uuid.UUID // nil = superadmin
	RequesterRole     string
	UserID            uuid.UUID
}

// GetUserHandler handles the GetUserQuery.
type GetUserHandler struct {
	users UserReadRepository
}

func NewGetUserHandler(users UserReadRepository) *GetUserHandler {
	return &GetUserHandler{users: users}
}

func (h *GetUserHandler) Handle(ctx context.Context, q GetUserQuery) (*domain.User, error) {
	user, err := h.users.FindByID(ctx, q.UserID)
	if err != nil {
		return nil, apperror.New(apperror.ErrNotFound, "user not found")
	}

	// Tenant guard: non-superadmin can only see users in their own school
	if q.RequesterRole != domain.RoleSuperadmin {
		if user.SchoolID == nil || q.RequesterSchoolID == nil || *user.SchoolID != *q.RequesterSchoolID {
			return nil, apperror.New(apperror.ErrForbidden, "access denied")
		}
	}

	return user, nil
}
