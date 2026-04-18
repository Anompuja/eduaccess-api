package application

import (
	"context"
	"strings"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/admin/domain"
	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateAdminCommand holds data needed to register a new admin sekolah.
type CreateAdminCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string

	// Required for superadmin to decide tenant placement.
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

// CreateAdminHandler creates a user (role=admin_sekolah) + admin_profile.
type CreateAdminHandler struct {
	users UserCreator
	repo  domain.AdminRepository
}

func NewCreateAdminHandler(users UserCreator, repo domain.AdminRepository) *CreateAdminHandler {
	return &CreateAdminHandler{users: users, repo: repo}
}

func (h *CreateAdminHandler) Handle(ctx context.Context, cmd CreateAdminCommand) (*domain.AdminProfile, error) {
	if cmd.RequesterRole != authdomain.RoleSuperadmin && cmd.RequesterRole != authdomain.RoleAdminSekolah {
		return nil, apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can create admin")
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
		pwd = "Admin@12345" // default - must be changed after first login
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	userID := uuid.New()
	user := &authdomain.User{
		ID:        userID,
		SchoolID:  schoolID,
		Role:      authdomain.RoleAdminSekolah,
		Name:      cmd.Name,
		Username:  username,
		Email:     cmd.Email,
		Password:  string(hash),
		Avatar:    "default.png",
		Verified:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := h.users.Create(ctx, user); err != nil {
		return nil, err
	}

	profile := &domain.AdminProfile{
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
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := h.repo.CreateAdminProfile(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}
