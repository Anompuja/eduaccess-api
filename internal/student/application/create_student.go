package application

import (
	"context"
	"strings"
	"time"

	academicdomain "github.com/eduaccess/eduaccess-api/internal/academic/domain"
	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/student/domain"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateStudentCommand holds data needed to register a new student.
type CreateStudentCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	// User account fields
	Name     string
	Email    string
	Username string
	Password string // if empty, a default is used
	// Profile fields
	NIS               string
	NISN              string
	PhoneNumber       string
	Address           string
	Gender            string
	Religion          string
	BirthPlace        string
	BirthDate         *time.Time
	TahunMasuk        string
	JalurMasukSekolah string
	EducationLevelID  *uuid.UUID
	ClassID           *uuid.UUID
	SubClassID        *uuid.UUID
}

// CreateStudentHandler creates a user (role=siswa) + student_profile atomically.
type CreateStudentHandler struct {
	users    UserCreator
	repo     domain.StudentRepository
	academic academicdomain.AcademicRepository
}

func NewCreateStudentHandler(users UserCreator, repo domain.StudentRepository, academic academicdomain.AcademicRepository) *CreateStudentHandler {
	return &CreateStudentHandler{users: users, repo: repo, academic: academic}
}

func (h *CreateStudentHandler) Handle(ctx context.Context, cmd CreateStudentCommand) (*domain.StudentProfile, error) {
	if cmd.RequesterRole != "superadmin" && cmd.RequesterRole != "admin_sekolah" {
		return nil, apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can create students")
	}

	// Resolve school ID from requester context; if missing, fallback to referenced academic entities.
	schoolID := cmd.RequesterSchoolID
	if schoolID == nil && cmd.SubClassID != nil {
		subClass, err := h.academic.FindSubClassByID(ctx, *cmd.SubClassID)
		if err != nil {
			return nil, err
		}
		schoolID = &subClass.SchoolID
	}
	if schoolID == nil && cmd.ClassID != nil {
		class, err := h.academic.FindClassByID(ctx, *cmd.ClassID)
		if err != nil {
			return nil, err
		}
		schoolID = &class.SchoolID
	}
	if schoolID == nil && cmd.EducationLevelID != nil {
		level, err := h.academic.FindLevelByID(ctx, *cmd.EducationLevelID)
		if err != nil {
			return nil, err
		}
		schoolID = &level.SchoolID
	}
	if schoolID == nil {
		return nil, apperror.New(apperror.ErrBadRequest, "unable to resolve school context; provide class_id, sub_class_id, or education_level_id")
	}

	// Check email uniqueness
	emailExists, err := h.users.ExistsByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, err
	}
	if emailExists {
		return nil, apperror.New(apperror.ErrConflict, "email already in use")
	}

	// Derive username from email if blank
	username := cmd.Username
	if username == "" {
		username = strings.Split(cmd.Email, "@")[0]
	}
	usernameExists, err := h.users.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if usernameExists {
		return nil, apperror.New(apperror.ErrConflict, "username already in use")
	}

	// Hash password
	pwd := cmd.Password
	if pwd == "" {
		pwd = "Siswa@12345" // default — must be changed on first login
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userID := uuid.New()
	user := &authdomain.User{
		ID:        userID,
		SchoolID:  schoolID,
		Role:      authdomain.RoleSiswa,
		Name:      cmd.Name,
		Username:  username,
		Email:     cmd.Email,
		Password:  string(hash),
		Avatar:    "default.png",
		Verified:  false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := h.users.Create(ctx, user); err != nil {
		return nil, err
	}

	profile := &domain.StudentProfile{
		ID:                uuid.New(),
		UserID:            userID,
		SchoolID:          *schoolID,
		NIS:               cmd.NIS,
		NISN:              cmd.NISN,
		PhoneNumber:       cmd.PhoneNumber,
		Address:           cmd.Address,
		Gender:            cmd.Gender,
		Religion:          cmd.Religion,
		BirthPlace:        cmd.BirthPlace,
		BirthDate:         cmd.BirthDate,
		TahunMasuk:        cmd.TahunMasuk,
		JalurMasukSekolah: cmd.JalurMasukSekolah,
		EducationLevelID:  cmd.EducationLevelID,
		ClassID:           cmd.ClassID,
		SubClassID:        cmd.SubClassID,
		Name:              cmd.Name,
		Email:             cmd.Email,
		Username:          username,
		Avatar:            "default.png",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	if err := h.repo.CreateStudentProfile(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}
