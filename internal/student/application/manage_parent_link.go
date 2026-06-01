package application

import (
	"context"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/student/domain"
	"github.com/google/uuid"
)

// LinkParentCommand links an existing parent profile to a student.
type LinkParentCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	StudentID         uuid.UUID
	ParentID          uuid.UUID
	Relationship      string // father | mother | guardian | other
	IsPrimary         bool
}

// LinkParentHandler links a parent to a student.
type LinkParentHandler struct {
	repo domain.StudentRepository
}

func NewLinkParentHandler(repo domain.StudentRepository) *LinkParentHandler {
	return &LinkParentHandler{repo: repo}
}

func (h *LinkParentHandler) Handle(ctx context.Context, cmd LinkParentCommand) error {
	if cmd.RequesterRole != "superadmin" && cmd.RequesterRole != "admin_sekolah" {
		return apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can manage parent links")
	}

	student, err := h.repo.FindStudentByID(ctx, cmd.StudentID)
	if err != nil {
		return err
	}
	if cmd.RequesterRole != "superadmin" {
		if cmd.RequesterSchoolID != nil && student.SchoolID != *cmd.RequesterSchoolID {
			return apperror.New(apperror.ErrForbidden, "access denied to this student")
		}
	}

	parent, err := h.repo.FindParentByID(ctx, cmd.ParentID)
	if err != nil {
		return err
	}
	// Parent must belong to the same school
	if parent.SchoolID != student.SchoolID {
		return apperror.New(apperror.ErrBadRequest, "parent does not belong to the same school")
	}

	link := &domain.ParentLink{
		ID:           uuid.New(),
		SchoolID:     student.SchoolID,
		StudentID:    cmd.StudentID,
		ParentID:     cmd.ParentID,
		Relationship: cmd.Relationship,
		IsPrimary:    cmd.IsPrimary,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	return h.repo.LinkParent(ctx, link)
}

// UnlinkParentCommand removes a parent-student link.
type UnlinkParentCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	StudentID         uuid.UUID
	ParentID          uuid.UUID
}

// UnlinkParentHandler removes a parent-student link.
type UnlinkParentHandler struct {
	repo domain.StudentRepository
}

func NewUnlinkParentHandler(repo domain.StudentRepository) *UnlinkParentHandler {
	return &UnlinkParentHandler{repo: repo}
}

func (h *UnlinkParentHandler) Handle(ctx context.Context, cmd UnlinkParentCommand) error {
	if cmd.RequesterRole != "superadmin" && cmd.RequesterRole != "admin_sekolah" {
		return apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can manage parent links")
	}

	student, err := h.repo.FindStudentByID(ctx, cmd.StudentID)
	if err != nil {
		return err
	}
	if cmd.RequesterRole != "superadmin" {
		if cmd.RequesterSchoolID != nil && student.SchoolID != *cmd.RequesterSchoolID {
			return apperror.New(apperror.ErrForbidden, "access denied to this student")
		}
	}

	return h.repo.UnlinkParent(ctx, cmd.StudentID, cmd.ParentID)
}
