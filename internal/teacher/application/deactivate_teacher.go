package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/teacher/domain"
	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/google/uuid"
)

// DeactivateTeacherCommand represents the command to deactivate a teacher.
type DeactivateTeacherCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	TeacherID         uuid.UUID
}

// DeactivateTeacherHandler handles deactivating a teacher (soft delete).
type DeactivateTeacherHandler struct {
	teacherRepo domain.TeacherRepository
}

// NewDeactivateTeacherHandler creates a new DeactivateTeacherHandler.
func NewDeactivateTeacherHandler(teacherRepo domain.TeacherRepository) *DeactivateTeacherHandler {
	return &DeactivateTeacherHandler{teacherRepo: teacherRepo}
}

// Handle deactivates a teacher with authorization checks.
func (h *DeactivateTeacherHandler) Handle(ctx context.Context, cmd DeactivateTeacherCommand) error {
	// Fetch the teacher first to check authorization
	teacher, err := h.teacherRepo.FindTeacherByID(ctx, cmd.TeacherID)
	if err != nil {
		return apperror.New(apperror.ErrNotFound, "teacher not found")
	}

	// Authorization check
	if cmd.RequesterRole != authdomain.RoleSuperadmin {
		if cmd.RequesterSchoolID == nil || teacher.SchoolID != *cmd.RequesterSchoolID {
			return apperror.New(apperror.ErrForbidden, "access denied to this teacher")
		}
	}

	return h.teacherRepo.SoftDeleteTeacher(ctx, cmd.TeacherID)
}
