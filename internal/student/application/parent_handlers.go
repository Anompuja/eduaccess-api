package application

import (
	"context"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/student/domain"
	"github.com/google/uuid"
)

// ── List Parents ──────────────────────────────────────────────────────────────

type ListParentsQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	Search            string
	Page              int
	PerPage           int
}

type ListParentsResult struct {
	Parents []*domain.ParentProfile
	Page    int
	PerPage int
	Total   int64
}

type ListParentsHandler struct {
	repo domain.StudentRepository
}

func NewListParentsHandler(repo domain.StudentRepository) *ListParentsHandler {
	return &ListParentsHandler{repo: repo}
}

func (h *ListParentsHandler) Handle(ctx context.Context, q ListParentsQuery) (*ListParentsResult, error) {
	page := q.Page
	if page < 1 {
		page = 1
	}
	perPage := q.PerPage
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var schoolID *uuid.UUID
	if q.RequesterRole != "superadmin" {
		schoolID = q.RequesterSchoolID
	}

	parents, total, err := h.repo.ListParents(ctx, domain.ParentFilter{
		SchoolID: schoolID,
		Search:   q.Search,
		Offset:   (page - 1) * perPage,
		Limit:    perPage,
	})
	if err != nil {
		return nil, err
	}
	return &ListParentsResult{Parents: parents, Page: page, PerPage: perPage, Total: total}, nil
}

// ── Get Parent ────────────────────────────────────────────────────────────────

type GetParentQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ParentID          uuid.UUID
}

type GetParentHandler struct {
	repo domain.StudentRepository
}

func NewGetParentHandler(repo domain.StudentRepository) *GetParentHandler {
	return &GetParentHandler{repo: repo}
}

func (h *GetParentHandler) Handle(ctx context.Context, q GetParentQuery) (*domain.ParentProfile, error) {
	parent, err := h.repo.FindParentByID(ctx, q.ParentID)
	if err != nil {
		return nil, err
	}
	if q.RequesterRole != "superadmin" {
		if q.RequesterSchoolID == nil || parent.SchoolID != *q.RequesterSchoolID {
			return nil, apperror.New(apperror.ErrForbidden, "access denied to this parent")
		}
	}
	return parent, nil
}

// ── Update Parent ─────────────────────────────────────────────────────────────

type UpdateParentCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ParentID          uuid.UUID
	FatherName        *string
	MotherName        *string
	FatherReligion    *string
	MotherReligion    *string
	PhoneNumber       *string
	Address           *string
}

type UpdateParentHandler struct {
	repo domain.StudentRepository
}

func NewUpdateParentHandler(repo domain.StudentRepository) *UpdateParentHandler {
	return &UpdateParentHandler{repo: repo}
}

func (h *UpdateParentHandler) Handle(ctx context.Context, cmd UpdateParentCommand) (*domain.ParentProfile, error) {
	if cmd.RequesterRole != "superadmin" && cmd.RequesterRole != "admin_sekolah" {
		return nil, apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can update parent profiles")
	}

	profile, err := h.repo.FindParentByID(ctx, cmd.ParentID)
	if err != nil {
		return nil, err
	}
	if cmd.RequesterRole != "superadmin" {
		if cmd.RequesterSchoolID == nil || profile.SchoolID != *cmd.RequesterSchoolID {
			return nil, apperror.New(apperror.ErrForbidden, "access denied to this parent")
		}
	}

	if cmd.FatherName != nil {
		profile.FatherName = *cmd.FatherName
	}
	if cmd.MotherName != nil {
		profile.MotherName = *cmd.MotherName
	}
	if cmd.FatherReligion != nil {
		profile.FatherReligion = *cmd.FatherReligion
	}
	if cmd.MotherReligion != nil {
		profile.MotherReligion = *cmd.MotherReligion
	}
	if cmd.PhoneNumber != nil {
		profile.PhoneNumber = *cmd.PhoneNumber
	}
	if cmd.Address != nil {
		profile.Address = *cmd.Address
	}

	profile.UpdatedAt = time.Now()
	if err := h.repo.UpdateParentProfile(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

// ── Deactivate Parent ─────────────────────────────────────────────────────────

type DeactivateParentCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ParentID          uuid.UUID
}

type DeactivateParentHandler struct {
	repo domain.StudentRepository
}

func NewDeactivateParentHandler(repo domain.StudentRepository) *DeactivateParentHandler {
	return &DeactivateParentHandler{repo: repo}
}

func (h *DeactivateParentHandler) Handle(ctx context.Context, cmd DeactivateParentCommand) error {
	if cmd.RequesterRole != "superadmin" && cmd.RequesterRole != "admin_sekolah" {
		return apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can deactivate parents")
	}

	profile, err := h.repo.FindParentByID(ctx, cmd.ParentID)
	if err != nil {
		return err
	}
	if cmd.RequesterRole != "superadmin" {
		if cmd.RequesterSchoolID == nil || profile.SchoolID != *cmd.RequesterSchoolID {
			return apperror.New(apperror.ErrForbidden, "access denied to this parent")
		}
	}

	return h.repo.SoftDeleteParent(ctx, cmd.ParentID)
}
