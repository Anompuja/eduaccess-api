package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/headmaster/application"
	"github.com/eduaccess/eduaccess-api/internal/headmaster/domain"
	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type capturingHeadmasterRepo struct {
	lastFilter domain.HeadmasterFilter
}

func (f *capturingHeadmasterRepo) CreateHeadmasterProfile(context.Context, *domain.HeadmasterProfile) error {
	return nil
}

func (f *capturingHeadmasterRepo) FindHeadmasterByID(context.Context, uuid.UUID) (*domain.HeadmasterProfile, error) {
	return nil, nil
}

func (f *capturingHeadmasterRepo) FindHeadmasterByUserID(context.Context, uuid.UUID) (*domain.HeadmasterProfile, error) {
	return nil, nil
}

func (f *capturingHeadmasterRepo) ListHeadmasters(_ context.Context, filter domain.HeadmasterFilter) ([]*domain.HeadmasterProfile, int64, error) {
	f.lastFilter = filter
	return []*domain.HeadmasterProfile{}, 0, nil
}

func (f *capturingHeadmasterRepo) UpdateHeadmasterProfile(context.Context, *domain.HeadmasterProfile) error {
	return nil
}

func (f *capturingHeadmasterRepo) SoftDeleteHeadmaster(context.Context, uuid.UUID) error {
	return nil
}

func TestHandlerList_ParsesSchoolIDForSuperadmin(t *testing.T) {
	e := echo.New()
	repo := &capturingHeadmasterRepo{}
	handler := &Handler{
		list: application.NewListHeadmastersHandler(repo),
	}

	schoolID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/headmasters?school_id="+schoolID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.ContextKeyRole, authdomain.RoleSuperadmin)

	if err := handler.List(c); err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if repo.lastFilter.SchoolID == nil || *repo.lastFilter.SchoolID != schoolID {
		t.Fatalf("expected school filter %s, got %#v", schoolID, repo.lastFilter.SchoolID)
	}
}
