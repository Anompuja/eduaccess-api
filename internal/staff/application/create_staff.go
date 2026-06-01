package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/staff/domain"
	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateStaffCommand represents the command to create a new staff.
type CreateStaffCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	Name              string
	Email             string
	Username          string
	Password          string
	SchoolID          *uuid.UUID
	PhoneNumber       *string
	Address           *string
	Gender            *string
	Religion          *string
	BirthPlace        *string
	BirthDate         *string
	NIK               *string
	KTPImagePath      *string
}

// CreateStaffHandler handles the creation of a new staff profile.
type CreateStaffHandler struct {
	staffRepo   domain.StaffRepository
	userCreator UserCreator
}

// NewCreateStaffHandler creates a new CreateStaffHandler.
func NewCreateStaffHandler(
	staffRepo domain.StaffRepository,
	userCreator UserCreator,
) *CreateStaffHandler {
	return &CreateStaffHandler{
		staffRepo:   staffRepo,
		userCreator: userCreator,
	}
}

// Handle creates a new staff profile and user account.
func (h *CreateStaffHandler) Handle(ctx context.Context, cmd CreateStaffCommand) (*domain.StaffProfile, error) {
	if cmd.RequesterRole != authdomain.RoleSuperadmin && cmd.RequesterRole != authdomain.RoleAdminSekolah {
		return nil, apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can create staff")
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
		pwd = "Staff@12345" // default - must be changed after first login
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
	user := &authdomain.User{
		ID:        uuid.New(),
		SchoolID:  schoolID,
		Role:      authdomain.RoleStaff,
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

	staff := &domain.StaffProfile{
		ID:           uuid.New(),
		UserID:       user.ID,
		SchoolID:     *schoolID,
		Name:         cmd.Name,
		Email:        cmd.Email,
		Username:     username,
		Avatar:       "default.png",
		PhoneNumber:  cmd.PhoneNumber,
		Address:      cmd.Address,
		Gender:       cmd.Gender,
		Religion:     cmd.Religion,
		BirthPlace:   cmd.BirthPlace,
		BirthDate:    birthDate,
		NIK:          cmd.NIK,
		KTPImagePath: cmd.KTPImagePath,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := h.staffRepo.CreateStaffProfile(ctx, staff); err != nil {
		if rollbackErr := h.userCreator.SoftDelete(ctx, user.ID); rollbackErr != nil {
			return nil, fmt.Errorf("create staff failed: %w (rollback failed: %v)", err, rollbackErr)
		}
		return nil, err
	}

	return staff, nil
}
