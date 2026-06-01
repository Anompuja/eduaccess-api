package application

import (
	"context"
	"errors"
	"testing"
	"time"

	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/student/domain"
	"github.com/google/uuid"
)

type fakeStudentRepo struct {
	created   *domain.StudentProfile
	createErr error
}

func (f *fakeStudentRepo) CreateStudentProfile(ctx context.Context, profile *domain.StudentProfile) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = profile
	return nil
}

func (f *fakeStudentRepo) FindStudentByID(ctx context.Context, id uuid.UUID) (*domain.StudentProfile, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeStudentRepo) FindStudentByUserID(ctx context.Context, userID uuid.UUID) (*domain.StudentProfile, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeStudentRepo) ListStudents(ctx context.Context, filter domain.StudentFilter) ([]*domain.StudentProfile, int64, error) {
	return nil, 0, errors.New("not implemented")
}

func (f *fakeStudentRepo) UpdateStudentProfile(ctx context.Context, profile *domain.StudentProfile) error {
	return errors.New("not implemented")
}

func (f *fakeStudentRepo) SoftDeleteStudent(ctx context.Context, id uuid.UUID) error {
	return errors.New("not implemented")
}

func (f *fakeStudentRepo) CreateParentProfile(ctx context.Context, profile *domain.ParentProfile) error {
	return errors.New("not implemented")
}

func (f *fakeStudentRepo) FindParentByID(ctx context.Context, id uuid.UUID) (*domain.ParentProfile, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeStudentRepo) FindParentByUserID(ctx context.Context, userID uuid.UUID) (*domain.ParentProfile, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeStudentRepo) ListParents(ctx context.Context, filter domain.ParentFilter) ([]*domain.ParentProfile, int64, error) {
	return nil, 0, errors.New("not implemented")
}

func (f *fakeStudentRepo) UpdateParentProfile(ctx context.Context, profile *domain.ParentProfile) error {
	return errors.New("not implemented")
}

func (f *fakeStudentRepo) SoftDeleteParent(ctx context.Context, id uuid.UUID) error {
	return errors.New("not implemented")
}

func (f *fakeStudentRepo) LinkParent(ctx context.Context, link *domain.ParentLink) error {
	return errors.New("not implemented")
}

func (f *fakeStudentRepo) UnlinkParent(ctx context.Context, studentID, parentID uuid.UUID) error {
	return errors.New("not implemented")
}

func (f *fakeStudentRepo) ListParentLinks(ctx context.Context, studentID uuid.UUID) ([]*domain.ParentLink, error) {
	return nil, errors.New("not implemented")
}

type fakeStudentUserCreator struct {
	assignedID  uuid.UUID
	created     *authdomain.User
	softDeleted uuid.UUID
}

func (f *fakeStudentUserCreator) Create(ctx context.Context, user *authdomain.User) error {
	user.ID = f.assignedID
	f.created = user
	return nil
}

func (f *fakeStudentUserCreator) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return false, nil
}

func (f *fakeStudentUserCreator) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return false, nil
}

func (f *fakeStudentUserCreator) SoftDelete(ctx context.Context, id uuid.UUID) error {
	f.softDeleted = id
	return nil
}

func TestCreateStudentUsesPersistedUserID(t *testing.T) {
	ctx := context.Background()
	schoolID := uuid.New()
	authUserID := uuid.New()

	repo := &fakeStudentRepo{}
	users := &fakeStudentUserCreator{assignedID: authUserID}
	handler := NewCreateStudentHandler(users, repo, nil)

	created, err := handler.Handle(ctx, CreateStudentCommand{
		RequesterSchoolID: &schoolID,
		RequesterRole:     authdomain.RoleSuperadmin,
		Name:              "Student One",
		Email:             "student@example.com",
		Username:          "student1",
		Password:          "Password123!",
		NIS:               "12345",
		NISN:              "67890",
		JalurMasukSekolah: domain.JalurReguler,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if repo.created == nil {
		t.Fatal("expected student profile to be created")
	}
	if repo.created.UserID != authUserID {
		t.Fatalf("expected student user_id to use persisted auth ID %s, got %s", authUserID, repo.created.UserID)
	}
	if created == nil {
		t.Fatal("expected handler to return created student")
	}
	if created.UserID != authUserID {
		t.Fatalf("expected returned student to use persisted auth ID %s, got %s", authUserID, created.UserID)
	}
	if repo.created.CreatedAt.IsZero() || repo.created.UpdatedAt.IsZero() {
		t.Fatal("expected timestamps to be set")
	}
	if repo.created.CreatedAt.After(time.Now().Add(1 * time.Minute)) {
		t.Fatal("unexpected created_at timestamp")
	}
}

func TestCreateStudentRollsBackUserWhenProfileCreateFails(t *testing.T) {
	ctx := context.Background()
	schoolID := uuid.New()
	authUserID := uuid.New()

	repo := &fakeStudentRepo{createErr: errors.New("insert student profile failed")}
	users := &fakeStudentUserCreator{assignedID: authUserID}
	handler := NewCreateStudentHandler(users, repo, nil)

	_, err := handler.Handle(ctx, CreateStudentCommand{
		RequesterSchoolID: &schoolID,
		RequesterRole:     authdomain.RoleSuperadmin,
		Name:              "Student Two",
		Email:             "student2@example.com",
		Username:          "student2",
		Password:          "Password123!",
		NIS:               "12346",
		NISN:              "67891",
		JalurMasukSekolah: domain.JalurReguler,
	})
	if err == nil {
		t.Fatal("expected error when student profile creation fails")
	}
	if users.softDeleted != authUserID {
		t.Fatalf("expected soft delete for auth user %s, got %s", authUserID, users.softDeleted)
	}
}