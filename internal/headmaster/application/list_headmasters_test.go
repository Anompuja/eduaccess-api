package application

import (
	"context"
	"testing"

	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/headmaster/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

type fakeHeadmasterRepo struct {
	lastFilter domain.HeadmasterFilter
}

func (f *fakeHeadmasterRepo) CreateHeadmasterProfile(context.Context, *domain.HeadmasterProfile) error {
	return nil
}

func (f *fakeHeadmasterRepo) FindHeadmasterByID(context.Context, uuid.UUID) (*domain.HeadmasterProfile, error) {
	return nil, nil
}

func (f *fakeHeadmasterRepo) FindHeadmasterByUserID(context.Context, uuid.UUID) (*domain.HeadmasterProfile, error) {
	return nil, nil
}

func (f *fakeHeadmasterRepo) ListHeadmasters(_ context.Context, filter domain.HeadmasterFilter) ([]*domain.HeadmasterProfile, int64, error) {
	f.lastFilter = filter
	return []*domain.HeadmasterProfile{}, 0, nil
}

func (f *fakeHeadmasterRepo) UpdateHeadmasterProfile(context.Context, *domain.HeadmasterProfile) error {
	return nil
}

func (f *fakeHeadmasterRepo) SoftDeleteHeadmaster(context.Context, uuid.UUID) error {
	return nil
}

func TestListHeadmastersHandler_SuperadminUsesRequestedSchoolFilter(t *testing.T) {
	repo := &fakeHeadmasterRepo{}
	handler := NewListHeadmastersHandler(repo)
	schoolID := uuid.New()

	_, err := handler.Handle(context.Background(), ListHeadmastersQuery{
		RequesterRole: authdomain.RoleSuperadmin,
		SchoolID:      &schoolID,
		Page:          1,
		PerPage:       20,
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if repo.lastFilter.SchoolID == nil || *repo.lastFilter.SchoolID != schoolID {
		t.Fatalf("expected school filter %s, got %#v", schoolID, repo.lastFilter.SchoolID)
	}
}

func TestListHeadmastersHandler_AdminSekolahUsesRequesterSchool(t *testing.T) {
	repo := &fakeHeadmasterRepo{}
	handler := NewListHeadmastersHandler(repo)
	requesterSchoolID := uuid.New()
	otherSchoolID := uuid.New()

	_, err := handler.Handle(context.Background(), ListHeadmastersQuery{
		RequesterRole:     authdomain.RoleAdminSekolah,
		RequesterSchoolID: &requesterSchoolID,
		SchoolID:          &otherSchoolID,
		Page:              1,
		PerPage:           20,
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if repo.lastFilter.SchoolID == nil || *repo.lastFilter.SchoolID != requesterSchoolID {
		t.Fatalf("expected requester school filter %s, got %#v", requesterSchoolID, repo.lastFilter.SchoolID)
	}
}

func TestListHeadmastersHandler_RejectsUnauthorizedRole(t *testing.T) {
	repo := &fakeHeadmasterRepo{}
	handler := NewListHeadmastersHandler(repo)

	_, err := handler.Handle(context.Background(), ListHeadmastersQuery{
		RequesterRole: authdomain.RoleGuru,
		Page:          1,
		PerPage:       20,
	})
	if err == nil {
		t.Fatal("expected forbidden error, got nil")
	}

	if !apperror.Is(err, apperror.ErrForbidden) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}
