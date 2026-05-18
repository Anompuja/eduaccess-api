package application

import (
	"context"

	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	dashboarddomain "github.com/eduaccess/eduaccess-api/internal/dashboard/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

// GetStatsQuery resolves which school dashboard should be loaded.
type GetStatsQuery struct {
	RequesterRole     string
	RequesterSchoolID *uuid.UUID
	SchoolID          *uuid.UUID
}

// GetStatsHandler fetches aggregated dashboard statistics.
type GetStatsHandler struct {
	repo dashboarddomain.Repository
}

// NewGetStatsHandler creates a dashboard stats handler.
func NewGetStatsHandler(repo dashboarddomain.Repository) *GetStatsHandler {
	return &GetStatsHandler{repo: repo}
}

// Handle resolves the school scope and returns the dashboard snapshot.
// Superadmin: schoolID is optional; nil means aggregate across all schools.
// Scoped roles: schoolID always derived from JWT; query param must match if provided.
func (h *GetStatsHandler) Handle(ctx context.Context, q GetStatsQuery) (*dashboarddomain.Stats, error) {
	var schoolID *uuid.UUID

	switch q.RequesterRole {
	case authdomain.RoleSuperadmin:
		schoolID = q.SchoolID // nil = aggregate all schools
	default:
		if q.RequesterSchoolID == nil {
			return nil, apperror.New(apperror.ErrForbidden, "school_id is missing from the token")
		}
		if q.SchoolID != nil && *q.SchoolID != *q.RequesterSchoolID {
			return nil, apperror.New(apperror.ErrForbidden, "cannot access another school dashboard")
		}
		schoolID = q.RequesterSchoolID
	}

	return h.repo.GetStats(ctx, schoolID)
}