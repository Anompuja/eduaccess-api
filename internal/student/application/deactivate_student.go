package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/student/domain"
	"github.com/google/uuid"
)

// DeactivateStudentCommand holds soft-delete parameters.
type DeactivateStudentCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	StudentID         uuid.UUID
}

// DeactivateStudentHandler soft-deletes a student profile.
type DeactivateStudentHandler struct {
	repo domain.StudentRepository
}

func NewDeactivateStudentHandler(repo domain.StudentRepository) *DeactivateStudentHandler {
	return &DeactivateStudentHandler{repo: repo}
}

func (h *DeactivateStudentHandler) Handle(ctx context.Context, cmd DeactivateStudentCommand) error {
	if cmd.RequesterRole != "superadmin" && cmd.RequesterRole != "admin_sekolah" {
		return apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can deactivate students")
	}

	profile, err := h.repo.FindStudentByID(ctx, cmd.StudentID)
	if err != nil {
		return err
	}

	if cmd.RequesterRole != "superadmin" {
		if cmd.RequesterSchoolID != nil && profile.SchoolID != *cmd.RequesterSchoolID {
			return apperror.New(apperror.ErrForbidden, "access denied to this student")
		}
	}

	return h.repo.SoftDeleteStudent(ctx, cmd.StudentID)
}
