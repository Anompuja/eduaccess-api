package application

import (
	"context"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/school/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

// CreateSchoolCommand holds data needed to create a new school.
type CreateSchoolCommand struct {
	RequesterRole string
	Name          string
	Address       string
	Phone         string
	Email         string
	Description   string
	ImagePath     string
	TimeZone      string
}

// CreateSchoolHandler creates a new school tenant. Only superadmin may call this.
type CreateSchoolHandler struct {
	repo domain.SchoolRepository
}

func NewCreateSchoolHandler(repo domain.SchoolRepository) *CreateSchoolHandler {
	return &CreateSchoolHandler{repo: repo}
}

func (h *CreateSchoolHandler) Handle(ctx context.Context, cmd CreateSchoolCommand) (*domain.School, error) {
	if cmd.RequesterRole != "superadmin" {
		return nil, apperror.New(apperror.ErrForbidden, "only superadmin can create schools")
	}

	exists, err := h.repo.ExistsByName(ctx, cmd.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.New(apperror.ErrConflict, "a school with this name already exists")
	}

	tz := cmd.TimeZone
	if tz == "" {
		tz = "Asia/Jakarta"
	}

	school := &domain.School{
		ID:          uuid.New(),
		Name:        cmd.Name,
		Address:     cmd.Address,
		Phone:       cmd.Phone,
		Email:       cmd.Email,
		Description: cmd.Description,
		ImagePath:   cmd.ImagePath,
		TimeZone:    tz,
		Status:      domain.StatusNonactive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.repo.Create(ctx, school); err != nil {
		return nil, err
	}
	return school, nil
}
