package application

import (
	"context"
	"strings"
	"time"

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
	users   UserCreator
	repo    domain.StudentRepository
}

func NewCreateStudentHandler(users UserCreator, repo domain.StudentRepository) *CreateStudentHandler {
	return &CreateStudentHandler{users: users, repo: repo}
}

func (h *CreateStudentHandler) Handle(ctx context.Context, cmd CreateStudentCommand) (*domain.StudentProfile, error) {
	if cmd.RequesterRole != "superadmin" && cmd.RequesterRole != "admin_sekolah" {
		return nil, apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can create students")
	}
	if cmd.RequesterRole != "superadmin" && cmd.RequesterSchoolID == nil {
		return nil, apperror.New(apperror.ErrForbidden, "school context required")
	}

	// Resolve school ID
	schoolID := cmd.RequesterSchoolID
	if cmd.RequesterRole == "superadmin" && schoolID == nil {
		return nil, apperror.New(apperror.ErrBadRequest, "school_id required for superadmin")
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
