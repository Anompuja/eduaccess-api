package application

import (
	"context"

	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/headmaster/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

// GetHeadmasterQuery selects a headmaster profile by its profile ID.
type GetHeadmasterQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	HeadmasterID      uuid.UUID
}

// GetHeadmasterHandler handles GetHeadmasterQuery.
type GetHeadmasterHandler struct {
	repo domain.HeadmasterRepository
}

func NewGetHeadmasterHandler(repo domain.HeadmasterRepository) *GetHeadmasterHandler {
	return &GetHeadmasterHandler{repo: repo}
}

func (h *GetHeadmasterHandler) Handle(ctx context.Context, q GetHeadmasterQuery) (*domain.HeadmasterProfile, error) {
	profile, err := h.repo.FindHeadmasterByID(ctx, q.HeadmasterID)
	if err != nil {
		return nil, err
	}

	// Non-superadmin can only see headmasters of their own school.
	if q.RequesterRole != authdomain.RoleSuperadmin {
		if q.RequesterSchoolID == nil || profile.SchoolID != *q.RequesterSchoolID {
			return nil, apperror.New(apperror.ErrForbidden, "access denied to this headmaster")
		}
	}

	return profile, nil
}
