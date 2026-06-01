package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/teacher/domain"
	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateTeacherCommand represents the command to create a new teacher.
type CreateTeacherCommand struct {
	RequesterSchoolID   *uuid.UUID
	RequesterRole       string
	Name                string
	Email               string
	Username            string
	Password            string
	SchoolID            *uuid.UUID
	NIP                 *string
	NUPTK               *string
	PhoneNumber         *string
	Address             *string
	Gender              *string
	Religion            *string
	BirthPlace          *string
	BirthDate           *string
	NIK                 *string
	KTPImagePath        *string
	Kewarganegaraan     *string
	GolonganDarah       *string
	BeratBadan          *string
	TinggiBadan         *string
	PenyakitYangSeringKambuh *string
	KelainanJasmani     *string
	PenyakitKronisYangPernahDiderita *string
	RTRW                *string
	KodePos             *string
	PendidikanTerakhir  *string
	Jurusan             *string
	TahunLulus          *string
	TahunMasuk          *string
}

// CreateTeacherHandler handles the creation of a new teacher profile.
type CreateTeacherHandler struct {
	teacherRepo domain.TeacherRepository
	userCreator UserCreator
}

// NewCreateTeacherHandler creates a new CreateTeacherHandler.
func NewCreateTeacherHandler(
	teacherRepo domain.TeacherRepository,
	userCreator UserCreator,
) *CreateTeacherHandler {
	return &CreateTeacherHandler{
		teacherRepo: teacherRepo,
		userCreator: userCreator,
	}
}

// Handle creates a new teacher profile and user account.
func (h *CreateTeacherHandler) Handle(ctx context.Context, cmd CreateTeacherCommand) (*domain.TeacherProfile, error) {
	if cmd.RequesterRole != authdomain.RoleSuperadmin && cmd.RequesterRole != authdomain.RoleAdminSekolah {
		return nil, apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can create teacher")
	}

	// Resolve school ID
	schoolID := cmd.RequesterSchoolID
	if cmd.RequesterRole == authdomain.RoleSuperadmin {
		if cmd.SchoolID == nil {
			return nil, apperror.New(apperror.ErrBadRequest, "school_id required for superadmin")
		}
		schoolID = cmd.SchoolID
	}
	if schoolID == nil {
		return nil, apperror.New(apperror.ErrForbidden, "school context required")
	}

	// Check email uniqueness
	emailExists, err := h.userCreator.ExistsByEmail(ctx, cmd.Email)
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
	usernameExists, err := h.userCreator.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if usernameExists {
		return nil, apperror.New(apperror.ErrConflict, "username already in use")
	}

	// Hash password
	pwd := cmd.Password
	if pwd == "" {
		pwd = "Teacher@12345" // default - must be changed after first login
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Parse birth date if provided
	var birthDate *time.Time
	if cmd.BirthDate != nil && *cmd.BirthDate != "" {
		t, err := time.Parse("2006-01-02", *cmd.BirthDate)
		if err != nil {
			return nil, apperror.New(apperror.ErrBadRequest, "birth_date must be YYYY-MM-DD")
		}
		birthDate = &t
	}

	now := time.Now()
	userID := uuid.New()
	user := &authdomain.User{
		ID:        userID,
		SchoolID:  schoolID,
		Role:      authdomain.RoleGuru,
		Name:      cmd.Name,
		Username:  username,
		Email:     cmd.Email,
		Password:  string(hash),
		Avatar:    "default.png",
		Verified:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := h.userCreator.Create(ctx, user); err != nil {
		return nil, err
	}

	teacher := &domain.TeacherProfile{
		ID:                                  uuid.New(),
		UserID:                              user.ID,
		SchoolID:                            *schoolID,
		Name:                                cmd.Name,
		Email:                               cmd.Email,
		Username:                            username,
		Avatar:                              "default.png",
		NIP:                                 cmd.NIP,
		NUPTK:                               cmd.NUPTK,
		PhoneNumber:                         cmd.PhoneNumber,
		Address:                             cmd.Address,
		Gender:                              cmd.Gender,
		Religion:                            cmd.Religion,
		BirthPlace:                          cmd.BirthPlace,
		BirthDate:                           birthDate,
		NIK:                                 cmd.NIK,
		KTPImagePath:                        cmd.KTPImagePath,
		Kewarganegaraan:                     cmd.Kewarganegaraan,
		GolonganDarah:                       cmd.GolonganDarah,
		BeratBadan:                          cmd.BeratBadan,
		TinggiBadan:                         cmd.TinggiBadan,
		PenyakitYangSeringKambuh:            cmd.PenyakitYangSeringKambuh,
		KelainanJasmani:                     cmd.KelainanJasmani,
		PenyakitKronisYangPernahDiderita:    cmd.PenyakitKronisYangPernahDiderita,
		RTRW:                                cmd.RTRW,
		KodePos:                             cmd.KodePos,
		PendidikanTerakhir:                  cmd.PendidikanTerakhir,
		Jurusan:                             cmd.Jurusan,
		TahunLulus:                          cmd.TahunLulus,
		TahunMasuk:                          cmd.TahunMasuk,
		CreatedAt:                           now,
		UpdatedAt:                           now,
	}
	if err := h.teacherRepo.CreateTeacherProfile(ctx, teacher); err != nil {
		if rollbackErr := h.userCreator.SoftDelete(ctx, user.ID); rollbackErr != nil {
			return nil, fmt.Errorf("create teacher failed: %w (rollback failed: %v)", err, rollbackErr)
		}
		return nil, err
	}

	return teacher, nil
}
