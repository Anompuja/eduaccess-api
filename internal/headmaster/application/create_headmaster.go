package application

import (
	"context"
	"strings"
	"time"

	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/headmaster/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateHeadmasterCommand holds data needed to register a new headmaster.
type CreateHeadmasterCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	// SchoolID may be set by superadmin to target a specific school.
	// For admin_sekolah this is ignored; their JWT school is used.
	SchoolID *uuid.UUID
	// User account fields
	Name     string
	Email    string
	Username string
	Password string // if empty, a default is used
	// Profile fields
	PhoneNumber  string
	Address      string
	Gender       string
	Religion     string
	BirthPlace   string
	BirthDate    *time.Time
	NIK          string
	KTPImagePath string
}

// CreateHeadmasterHandler creates a user (role=kepala_sekolah) + headmaster_profile
// and sets the school's headmaster_id atomically via three sequential writes.
type CreateHeadmasterHandler struct {
	users   UserCreator
	repo    domain.HeadmasterRepository
	schools SchoolHeadmasterSetter
}

func NewCreateHeadmasterHandler(
	users UserCreator,
	repo domain.HeadmasterRepository,
	schools SchoolHeadmasterSetter,
) *CreateHeadmasterHandler {
	return &CreateHeadmasterHandler{users: users, repo: repo, schools: schools}
}

func (h *CreateHeadmasterHandler) Handle(ctx context.Context, cmd CreateHeadmasterCommand) (*domain.HeadmasterProfile, error) {
	if cmd.RequesterRole != authdomain.RoleSuperadmin && cmd.RequesterRole != authdomain.RoleAdminSekolah {
		return nil, apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can create a headmaster")
	}

	// admin_sekolah is always scoped to their own school.
	// superadmin must explicitly provide a school_id in the request body.
	schoolID := cmd.RequesterSchoolID
	if schoolID == nil && cmd.RequesterRole == authdomain.RoleSuperadmin {
		schoolID = cmd.SchoolID
	}
	if schoolID == nil {
		return nil, apperror.New(apperror.ErrBadRequest, "school_id is required (superadmin must provide it in the request body)")
	}

	emailExists, err := h.users.ExistsByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, err
	}
	if emailExists {
		return nil, apperror.New(apperror.ErrConflict, "email already in use")
	}

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

	pwd := cmd.Password
	if pwd == "" {
		pwd = "KepSek@12345" // default — must be changed on first login
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userID := uuid.New()
	user := &authdomain.User{
		ID:        userID,
		SchoolID:  schoolID,
		Role:      authdomain.RoleKepalaSekolah,
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

	profile := &domain.HeadmasterProfile{
		ID:           uuid.New(),
		UserID:       userID,
		SchoolID:     *schoolID,
		PhoneNumber:  cmd.PhoneNumber,
		Address:      cmd.Address,
		Gender:       cmd.Gender,
		Religion:     cmd.Religion,
		BirthPlace:   cmd.BirthPlace,
		BirthDate:    cmd.BirthDate,
		NIK:          cmd.NIK,
		KTPImagePath: cmd.KTPImagePath,
		Name:         cmd.Name,
		Email:        cmd.Email,
		Username:     username,
		Avatar:       "default.png",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := h.repo.CreateHeadmasterProfile(ctx, profile); err != nil {
		return nil, err
	}

	// Mark this user as the school's current headmaster.
	if err := h.schools.SetHeadmasterID(ctx, *schoolID, userID); err != nil {
		return nil, err
	}

	return profile, nil
}
