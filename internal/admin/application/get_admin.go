package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/admin/domain"
	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

// GetAdminQuery holds parameters for a single admin fetch.
type GetAdminQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	AdminID           uuid.UUID
}

// GetAdminHandler fetches a single admin profile.
type GetAdminHandler struct {
	repo domain.AdminRepository
}

func NewGetAdminHandler(repo domain.AdminRepository) *GetAdminHandler {
	return &GetAdminHandler{repo: repo}
}

func (h *GetAdminHandler) Handle(ctx context.Context, q GetAdminQuery) (*domain.AdminProfile, error) {
	if q.RequesterRole != authdomain.RoleSuperadmin && q.RequesterRole != authdomain.RoleAdminSekolah {
		return nil, apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can get admin")
	}

	admin, err := h.repo.FindAdminByID(ctx, q.AdminID)
	if err != nil {
		return nil, err
	}

	if q.RequesterRole != authdomain.RoleSuperadmin {
		if q.RequesterSchoolID == nil || admin.SchoolID != *q.RequesterSchoolID {
			return nil, apperror.New(apperror.ErrForbidden, "access denied to this admin")
		}
	}

	return admin, nil
}
