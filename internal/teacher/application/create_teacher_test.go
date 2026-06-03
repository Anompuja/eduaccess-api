package application

import (
	"context"
	"errors"
	"testing"
	"time"

	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/teacher/domain"
	"github.com/google/uuid"
)

type fakeTeacherRepo struct {
	created *domain.TeacherProfile
}

func (f *fakeTeacherRepo) CreateTeacherProfile(ctx context.Context, teacher *domain.TeacherProfile) error {
	f.created = teacher
	return nil
}

func (f *fakeTeacherRepo) FindTeacherByID(ctx context.Context, id uuid.UUID) (*domain.TeacherProfile, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeTeacherRepo) UpdateTeacherProfile(ctx context.Context, teacher *domain.TeacherProfile) error {
	return errors.New("not implemented")
}

func (f *fakeTeacherRepo) SoftDeleteTeacher(ctx context.Context, id uuid.UUID) error {
	return errors.New("not implemented")
}

func (f *fakeTeacherRepo) ListTeachers(ctx context.Context, filter domain.TeacherFilter) ([]*domain.TeacherProfile, int64, error) {
	return nil, 0, errors.New("not implemented")
}

type fakeUserCreator struct {
	assignedID uuid.UUID
	created    *authdomain.User
}

func (f *fakeUserCreator) Create(ctx context.Context, user *authdomain.User) error {
	user.ID = f.assignedID
	f.created = user
	return nil
}

func (f *fakeUserCreator) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return false, nil
}

func (f *fakeUserCreator) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return false, nil
}

func (f *fakeUserCreator) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func TestCreateTeacherUsesPersistedUserID(t *testing.T) {
	ctx := context.Background()
	schoolID := uuid.New()
	authUserID := uuid.New()

	repo := &fakeTeacherRepo{}
	users := &fakeUserCreator{assignedID: authUserID}
	handler := NewCreateTeacherHandler(repo, users)

	created, err := handler.Handle(ctx, CreateTeacherCommand{
		RequesterSchoolID: &schoolID,
		RequesterRole:     authdomain.RoleSuperadmin,
		Name:              "Teacher One",
		Email:             "teacher@example.com",
		Username:          "teacher1",
		Password:          "Password123!",
		SchoolID:          &schoolID,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if repo.created == nil {
		t.Fatal("expected teacher profile to be created")
	}
	if repo.created.UserID != authUserID {
		t.Fatalf("expected teacher user_id to use persisted auth ID %s, got %s", authUserID, repo.created.UserID)
	}
	if created == nil {
		t.Fatal("expected handler to return created teacher")
	}
	if created.UserID != authUserID {
		t.Fatalf("expected returned teacher to use persisted auth ID %s, got %s", authUserID, created.UserID)
	}
	if repo.created.CreatedAt.IsZero() || repo.created.UpdatedAt.IsZero() {
		t.Fatal("expected timestamps to be set")
	}
	if repo.created.CreatedAt.After(time.Now().Add(1 * time.Minute)) {
		t.Fatal("unexpected created_at timestamp")
	}
}