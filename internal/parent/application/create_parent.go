package application

import (
	"context"
	"strings"
	"time"

	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/parent/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateParentCommand holds data needed to register a new parent.
type CreateParentCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	// Required for superadmin to decide tenant placement.
	SchoolID *uuid.UUID
	// User account fields
	Name     string
	Email    string
	Username string
	Password string
	// Profile fields
	FatherName     string
	MotherName     string
	FatherReligion string
	MotherReligion string
	PhoneNumber    string
	Address        string
}

// CreateParentHandler creates a user (role=orangtua) + parent_profile atomically.
type CreateParentHandler struct {
	users UserCreator
	repo  domain.ParentRepository
}

func NewCreateParentHandler(users UserCreator, repo domain.ParentRepository) *CreateParentHandler {
	return &CreateParentHandler{users: users, repo: repo}
}

func (h *CreateParentHandler) Handle(ctx context.Context, cmd CreateParentCommand) (*domain.ParentProfile, error) {
	if cmd.RequesterRole != authdomain.RoleSuperadmin && cmd.RequesterRole != authdomain.RoleAdminSekolah {
		return nil, apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can create parents")
	}

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
		pwd = "Ortu@12345"
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
		Role:      authdomain.RoleOrangTua,
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

	profile := &domain.ParentProfile{
		ID:             uuid.New(),
		UserID:         userID,
		SchoolID:       *schoolID,
		FatherName:     cmd.FatherName,
		MotherName:     cmd.MotherName,
		FatherReligion: cmd.FatherReligion,
		MotherReligion: cmd.MotherReligion,
		PhoneNumber:    cmd.PhoneNumber,
		Address:        cmd.Address,
		Name:           cmd.Name,
		Email:          cmd.Email,
		Username:       username,
		Avatar:         "default.png",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := h.repo.CreateParentProfile(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}
