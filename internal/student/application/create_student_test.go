package application

import (
	"context"
	"testing"
	"time"

	academicdomain "github.com/eduaccess/eduaccess-api/internal/academic/domain"
	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	schooldomain "github.com/eduaccess/eduaccess-api/internal/school/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	studentdomain "github.com/eduaccess/eduaccess-api/internal/student/domain"
	"github.com/google/uuid"
)

type fakeStudentUserCreator struct {
	created       *authdomain.User
	softDeletedID *uuid.UUID
}

func (f *fakeStudentUserCreator) Create(_ context.Context, user *authdomain.User) error {
	f.created = user
	return nil
}

func (f *fakeStudentUserCreator) ExistsByEmail(context.Context, string) (bool, error) {
	return false, nil
}

func (f *fakeStudentUserCreator) ExistsByUsername(context.Context, string) (bool, error) {
	return false, nil
}

func (f *fakeStudentUserCreator) SoftDelete(_ context.Context, id uuid.UUID) error {
	f.softDeletedID = &id
	return nil
}

type fakeStudentRepo struct {
	count   int64
	created *studentdomain.StudentProfile
	err     error
}

func (f *fakeStudentRepo) CreateStudentProfile(_ context.Context, profile *studentdomain.StudentProfile) error {
	if f.err != nil {
		return f.err
	}
	f.created = profile
	return nil
}

func (f *fakeStudentRepo) FindStudentByID(context.Context, uuid.UUID) (*studentdomain.StudentProfile, error) {
	return nil, nil
}

func (f *fakeStudentRepo) FindStudentByUserID(context.Context, uuid.UUID) (*studentdomain.StudentProfile, error) {
	return nil, nil
}

func (f *fakeStudentRepo) ListStudents(context.Context, studentdomain.StudentFilter) ([]*studentdomain.StudentProfile, int64, error) {
	return nil, 0, nil
}

func (f *fakeStudentRepo) CountActiveStudents(context.Context, uuid.UUID) (int64, error) {
	return f.count, nil
}

func (f *fakeStudentRepo) UpdateStudentProfile(context.Context, *studentdomain.StudentProfile) error {
	return nil
}

func (f *fakeStudentRepo) SoftDeleteStudent(context.Context, uuid.UUID) error { return nil }

func (f *fakeStudentRepo) AutoEnrollStudent(context.Context, uuid.UUID, uuid.UUID, *uuid.UUID, *uuid.UUID) error {
	return nil
}

func (f *fakeStudentRepo) CreateParentProfile(context.Context, *studentdomain.ParentProfile) error {
	return nil
}

func (f *fakeStudentRepo) FindParentByID(context.Context, uuid.UUID) (*studentdomain.ParentProfile, error) {
	return nil, nil
}

func (f *fakeStudentRepo) FindParentByUserID(context.Context, uuid.UUID) (*studentdomain.ParentProfile, error) {
	return nil, nil
}

func (f *fakeStudentRepo) ListParents(context.Context, studentdomain.ParentFilter) ([]*studentdomain.ParentProfile, int64, error) {
	return nil, 0, nil
}

func (f *fakeStudentRepo) UpdateParentProfile(context.Context, *studentdomain.ParentProfile) error {
	return nil
}

func (f *fakeStudentRepo) SoftDeleteParent(context.Context, uuid.UUID) error { return nil }

func (f *fakeStudentRepo) LinkParent(context.Context, *studentdomain.ParentLink) error { return nil }

func (f *fakeStudentRepo) UnlinkParent(context.Context, uuid.UUID, uuid.UUID) error { return nil }

func (f *fakeStudentRepo) ListParentLinks(context.Context, uuid.UUID) ([]*studentdomain.ParentLink, error) {
	return nil, nil
}

type fakeAcademicRepo struct{}

func (f *fakeAcademicRepo) CreateLevel(context.Context, *academicdomain.EducationLevel) error {
	return nil
}
func (f *fakeAcademicRepo) FindLevelByID(context.Context, uuid.UUID) (*academicdomain.EducationLevel, error) {
	return nil, nil
}
func (f *fakeAcademicRepo) ListLevels(context.Context, *uuid.UUID) ([]*academicdomain.EducationLevel, error) {
	return nil, nil
}
func (f *fakeAcademicRepo) UpdateLevel(context.Context, *academicdomain.EducationLevel) error {
	return nil
}
func (f *fakeAcademicRepo) SoftDeleteLevel(context.Context, uuid.UUID) error         { return nil }
func (f *fakeAcademicRepo) CreateClass(context.Context, *academicdomain.Class) error { return nil }
func (f *fakeAcademicRepo) FindClassByID(context.Context, uuid.UUID) (*academicdomain.Class, error) {
	return nil, nil
}
func (f *fakeAcademicRepo) ListClasses(context.Context, *uuid.UUID, *uuid.UUID) ([]*academicdomain.Class, error) {
	return nil, nil
}
func (f *fakeAcademicRepo) UpdateClass(context.Context, *academicdomain.Class) error { return nil }
func (f *fakeAcademicRepo) SoftDeleteClass(context.Context, uuid.UUID) error         { return nil }
func (f *fakeAcademicRepo) CreateSubClass(context.Context, *academicdomain.SubClass) error {
	return nil
}
func (f *fakeAcademicRepo) FindSubClassByID(context.Context, uuid.UUID) (*academicdomain.SubClass, error) {
	return nil, nil
}
func (f *fakeAcademicRepo) ListSubClasses(context.Context, *uuid.UUID, *uuid.UUID) ([]*academicdomain.SubClass, error) {
	return nil, nil
}
func (f *fakeAcademicRepo) UpdateSubClass(context.Context, *academicdomain.SubClass) error {
	return nil
}
func (f *fakeAcademicRepo) SoftDeleteSubClass(context.Context, uuid.UUID) error { return nil }
func (f *fakeAcademicRepo) CreateAcademicYear(context.Context, *academicdomain.AcademicYear) error {
	return nil
}
func (f *fakeAcademicRepo) FindAcademicYearByID(context.Context, uuid.UUID) (*academicdomain.AcademicYear, error) {
	return nil, nil
}
func (f *fakeAcademicRepo) ListAcademicYears(context.Context, *uuid.UUID) ([]*academicdomain.AcademicYear, error) {
	return nil, nil
}
func (f *fakeAcademicRepo) UpdateAcademicYear(context.Context, *academicdomain.AcademicYear) error {
	return nil
}
func (f *fakeAcademicRepo) SoftDeleteAcademicYear(context.Context, uuid.UUID) error { return nil }
func (f *fakeAcademicRepo) ActivateAcademicYear(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}
func (f *fakeAcademicRepo) CreateSubject(context.Context, *academicdomain.Subject) error { return nil }
func (f *fakeAcademicRepo) FindSubjectByID(context.Context, uuid.UUID) (*academicdomain.Subject, error) {
	return nil, nil
}
func (f *fakeAcademicRepo) ListSubjects(context.Context, *uuid.UUID) ([]*academicdomain.Subject, error) {
	return nil, nil
}
func (f *fakeAcademicRepo) UpdateSubject(context.Context, *academicdomain.Subject) error { return nil }
func (f *fakeAcademicRepo) SoftDeleteSubject(context.Context, uuid.UUID) error           { return nil }
func (f *fakeAcademicRepo) CreateClassroom(context.Context, *academicdomain.Classroom) error {
	return nil
}
func (f *fakeAcademicRepo) FindClassroomByID(context.Context, uuid.UUID) (*academicdomain.Classroom, error) {
	return nil, nil
}
func (f *fakeAcademicRepo) ListClassrooms(context.Context, *uuid.UUID) ([]*academicdomain.Classroom, error) {
	return nil, nil
}
func (f *fakeAcademicRepo) UpdateClassroom(context.Context, *academicdomain.Classroom) error {
	return nil
}
func (f *fakeAcademicRepo) SoftDeleteClassroom(context.Context, uuid.UUID) error { return nil }
func (f *fakeAcademicRepo) CreateSchedule(context.Context, *academicdomain.Schedule) error {
	return nil
}
func (f *fakeAcademicRepo) FindScheduleByID(context.Context, uuid.UUID) (*academicdomain.Schedule, error) {
	return nil, nil
}
func (f *fakeAcademicRepo) ListSchedules(context.Context, *uuid.UUID, *string) ([]*academicdomain.Schedule, error) {
	return nil, nil
}
func (f *fakeAcademicRepo) UpdateSchedule(context.Context, *academicdomain.Schedule) error {
	return nil
}
func (f *fakeAcademicRepo) SoftDeleteSchedule(context.Context, uuid.UUID) error { return nil }

type fakeSchoolSubscriptionReader struct {
	sub *schooldomain.Subscription
	err error
}

func (f *fakeSchoolSubscriptionReader) FindActiveSubscription(context.Context, uuid.UUID) (*schooldomain.Subscription, error) {
	return f.sub, f.err
}

func TestCreateStudentHandler_RejectsWhenQuotaIsReached(t *testing.T) {
	schoolID := uuid.New()
	repo := &fakeStudentRepo{count: 500}
	users := &fakeStudentUserCreator{}
	handler := NewCreateStudentHandler(
		users,
		repo,
		&fakeAcademicRepo{},
		&fakeSchoolSubscriptionReader{
			sub: &schooldomain.Subscription{
				Plan: &schooldomain.Plan{Name: "Basic", MaxStudents: 500},
			},
		},
	)

	_, err := handler.Handle(context.Background(), CreateStudentCommand{
		RequesterSchoolID: &schoolID,
		RequesterRole:     authdomain.RoleAdminSekolah,
		Name:              "Student A",
		Email:             "studenta@example.com",
		NIS:               "1001",
	})
	if err == nil {
		t.Fatal("expected quota error, got nil")
	}
	if !apperror.Is(err, apperror.ErrConflict) {
		t.Fatalf("expected conflict error, got %v", err)
	}
	if users.created != nil {
		t.Fatal("expected user creation to be skipped when quota is full")
	}
}

func TestCreateStudentHandler_CreatesStudentWhenQuotaAvailable(t *testing.T) {
	schoolID := uuid.New()
	repo := &fakeStudentRepo{count: 499}
	users := &fakeStudentUserCreator{}
	handler := NewCreateStudentHandler(
		users,
		repo,
		&fakeAcademicRepo{},
		&fakeSchoolSubscriptionReader{
			sub: &schooldomain.Subscription{
				Plan: &schooldomain.Plan{Name: "Basic", MaxStudents: 500},
			},
		},
	)
	birthDate := time.Now().AddDate(-10, 0, 0)

	profile, err := handler.Handle(context.Background(), CreateStudentCommand{
		RequesterSchoolID: &schoolID,
		RequesterRole:     authdomain.RoleAdminSekolah,
		Name:              "Student B",
		Email:             "studentb@example.com",
		NIS:               "1002",
		BirthDate:         &birthDate,
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if profile == nil {
		t.Fatal("expected student profile to be returned")
	}
	if repo.created == nil {
		t.Fatal("expected student profile to be persisted")
	}
	if users.created == nil {
		t.Fatal("expected user account to be created")
	}
	if repo.created.SchoolID != schoolID {
		t.Fatalf("expected school ID %s, got %s", schoolID, repo.created.SchoolID)
	}
}

func TestCreateStudentHandler_RollsBackUserWhenProfileCreationFails(t *testing.T) {
	schoolID := uuid.New()
	repo := &fakeStudentRepo{err: apperror.New(apperror.ErrInternal, "insert failed")}
	users := &fakeStudentUserCreator{}
	handler := NewCreateStudentHandler(
		users,
		repo,
		&fakeAcademicRepo{},
		&fakeSchoolSubscriptionReader{
			sub: &schooldomain.Subscription{
				Plan: &schooldomain.Plan{Name: "Basic", MaxStudents: 500},
			},
		},
	)

	_, err := handler.Handle(context.Background(), CreateStudentCommand{
		RequesterSchoolID: &schoolID,
		RequesterRole:     authdomain.RoleAdminSekolah,
		Name:              "Student C",
		Email:             "studentc@example.com",
		NIS:               "1003",
	})
	if err == nil {
		t.Fatal("expected profile creation error, got nil")
	}
	if users.created == nil {
		t.Fatal("expected auth user to be created before profile failure")
	}
	if users.softDeletedID == nil {
		t.Fatal("expected auth user rollback on profile failure")
	}
	if *users.softDeletedID != users.created.ID {
		t.Fatalf("expected rollback for user %s, got %s", users.created.ID, *users.softDeletedID)
	}
}
