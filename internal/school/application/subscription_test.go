package application

import (
	"context"
	"testing"

	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/school/domain"
	"github.com/google/uuid"
)

type fakeSchoolRepo struct {
	createdSchool        *domain.School
	defaultSubscription  *domain.Subscription
	findByIDResult       *domain.School
	findPlanByIDResult   *domain.Plan
	replacedSubscription *domain.Subscription
	existsByNameResult   bool
}

func (f *fakeSchoolRepo) Create(context.Context, *domain.School) error { return nil }

func (f *fakeSchoolRepo) CreateWithDefaultSubscription(_ context.Context, school *domain.School) (*domain.Subscription, error) {
	f.createdSchool = school
	return f.defaultSubscription, nil
}

func (f *fakeSchoolRepo) FindByID(context.Context, uuid.UUID) (*domain.School, error) {
	return f.findByIDResult, nil
}

func (f *fakeSchoolRepo) List(context.Context, domain.SchoolFilter) ([]*domain.School, int64, error) {
	return nil, 0, nil
}

func (f *fakeSchoolRepo) Update(context.Context, *domain.School) error { return nil }

func (f *fakeSchoolRepo) SoftDelete(context.Context, uuid.UUID) error { return nil }

func (f *fakeSchoolRepo) ExistsByName(context.Context, string) (bool, error) {
	return f.existsByNameResult, nil
}

func (f *fakeSchoolRepo) ListRules(context.Context, uuid.UUID) ([]*domain.SchoolRule, error) {
	return nil, nil
}

func (f *fakeSchoolRepo) UpsertRule(context.Context, *domain.SchoolRule) error { return nil }

func (f *fakeSchoolRepo) DeleteRule(context.Context, uuid.UUID, string) error { return nil }

func (f *fakeSchoolRepo) ListPlans(context.Context) ([]*domain.Plan, error) { return nil, nil }

func (f *fakeSchoolRepo) FindPlanByID(context.Context, uuid.UUID) (*domain.Plan, error) {
	return f.findPlanByIDResult, nil
}

func (f *fakeSchoolRepo) FindActiveSubscription(context.Context, uuid.UUID) (*domain.Subscription, error) {
	return nil, nil
}

func (f *fakeSchoolRepo) ReplaceSubscription(_ context.Context, sub *domain.Subscription) error {
	f.replacedSubscription = sub
	return nil
}

func (f *fakeSchoolRepo) SetHeadmasterID(context.Context, uuid.UUID, uuid.UUID) error { return nil }

type fakeStudentCounter struct {
	total int64
}

func (f *fakeStudentCounter) CountActiveStudents(context.Context, uuid.UUID) (int64, error) {
	return f.total, nil
}

func TestCreateSchoolHandler_AssignsDefaultSubscription(t *testing.T) {
	repo := &fakeSchoolRepo{
		defaultSubscription: &domain.Subscription{
			ID:     uuid.New(),
			Status: "trial",
			Plan: &domain.Plan{
				ID:          uuid.New(),
				Name:        "Trial",
				MaxStudents: 100,
			},
		},
	}
	handler := NewCreateSchoolHandler(repo)

	school, err := handler.Handle(context.Background(), CreateSchoolCommand{
		RequesterRole: authdomain.RoleSuperadmin,
		Name:          "SMK Nusantara",
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if repo.createdSchool == nil {
		t.Fatal("expected school to be persisted")
	}
	if school.Subscription == nil {
		t.Fatal("expected default subscription to be attached")
	}
	if school.Subscription.Plan == nil || school.Subscription.Plan.Name != "Trial" {
		t.Fatalf("expected trial plan, got %#v", school.Subscription.Plan)
	}
	if school.TimeZone != "Asia/Jakarta" {
		t.Fatalf("expected default timezone Asia/Jakarta, got %s", school.TimeZone)
	}
}

func TestUpdateSubscriptionHandler_UsesSelectedPlanPricing(t *testing.T) {
	planID := uuid.New()
	schoolID := uuid.New()
	repo := &fakeSchoolRepo{
		findByIDResult: &domain.School{ID: schoolID},
		findPlanByIDResult: &domain.Plan{
			ID:           planID,
			Name:         "Pro",
			MaxStudents:  1500,
			MonthlyPrice: 1299000,
			YearlyPrice:  12990000,
		},
	}
	handler := NewUpdateSubscriptionHandler(repo)

	sub, err := handler.Handle(context.Background(), UpdateSubscriptionCommand{
		RequesterRole: authdomain.RoleSuperadmin,
		SchoolID:      schoolID,
		PlanID:        planID,
		Cycle:         "year",
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if repo.replacedSubscription == nil {
		t.Fatal("expected subscription replacement to be persisted")
	}
	if sub.Price != 12990000 {
		t.Fatalf("expected yearly price 12990000, got %d", sub.Price)
	}
	if sub.Status != "active" {
		t.Fatalf("expected active status, got %s", sub.Status)
	}
	if sub.Plan == nil || sub.Plan.Name != "Pro" {
		t.Fatalf("expected plan Pro, got %#v", sub.Plan)
	}
	if sub.EndsAt == nil {
		t.Fatal("expected yearly subscription to have an end date")
	}
}

func TestUpdateSubscriptionHandler_RejectsPlanBelowCurrentStudentCount(t *testing.T) {
	planID := uuid.New()
	schoolID := uuid.New()
	repo := &fakeSchoolRepo{
		findByIDResult: &domain.School{ID: schoolID},
		findPlanByIDResult: &domain.Plan{
			ID:           planID,
			Name:         "Basic",
			MaxStudents:  500,
			MonthlyPrice: 499000,
			YearlyPrice:  4990000,
		},
	}
	handler := NewUpdateSubscriptionHandler(repo, &fakeStudentCounter{total: 700})

	_, err := handler.Handle(context.Background(), UpdateSubscriptionCommand{
		RequesterRole: authdomain.RoleSuperadmin,
		SchoolID:      schoolID,
		PlanID:        planID,
		Cycle:         "month",
	})
	if err == nil {
		t.Fatal("expected quota validation error, got nil")
	}
}
