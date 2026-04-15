package application

import (
	"context"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/student/domain"
	"github.com/google/uuid"
)

// ── Education Levels ──────────────────────────────────────────────────────────

type CreateLevelCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	Name              string
}

type CreateLevelHandler struct{ repo domain.AcademicRepository }

func NewCreateLevelHandler(repo domain.AcademicRepository) *CreateLevelHandler {
	return &CreateLevelHandler{repo: repo}
}

func (h *CreateLevelHandler) Handle(ctx context.Context, cmd CreateLevelCommand) (*domain.EducationLevel, error) {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return nil, err
	}
	schoolID := resolveSchoolID(cmd.RequesterRole, cmd.RequesterSchoolID)
	if schoolID == nil {
		return nil, apperror.New(apperror.ErrBadRequest, "school context required")
	}
	level := &domain.EducationLevel{
		ID:        uuid.New(),
		SchoolID:  *schoolID,
		Name:      cmd.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := h.repo.CreateLevel(ctx, level); err != nil {
		return nil, err
	}
	return level, nil
}

type ListLevelsQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
}

type ListLevelsHandler struct{ repo domain.AcademicRepository }

func NewListLevelsHandler(repo domain.AcademicRepository) *ListLevelsHandler {
	return &ListLevelsHandler{repo: repo}
}

func (h *ListLevelsHandler) Handle(ctx context.Context, q ListLevelsQuery) ([]*domain.EducationLevel, error) {
	schoolID := resolveSchoolID(q.RequesterRole, q.RequesterSchoolID)
	if schoolID == nil {
		return nil, apperror.New(apperror.ErrBadRequest, "school context required")
	}
	return h.repo.ListLevels(ctx, *schoolID)
}

type UpdateLevelCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	LevelID           uuid.UUID
	Name              string
}

type UpdateLevelHandler struct{ repo domain.AcademicRepository }

func NewUpdateLevelHandler(repo domain.AcademicRepository) *UpdateLevelHandler {
	return &UpdateLevelHandler{repo: repo}
}

func (h *UpdateLevelHandler) Handle(ctx context.Context, cmd UpdateLevelCommand) (*domain.EducationLevel, error) {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return nil, err
	}
	level, err := h.repo.FindLevelByID(ctx, cmd.LevelID)
	if err != nil {
		return nil, err
	}
	if err := guardSchoolIDMatch(cmd.RequesterRole, cmd.RequesterSchoolID, level.SchoolID); err != nil {
		return nil, err
	}
	level.Name = cmd.Name
	level.UpdatedAt = time.Now()
	if err := h.repo.UpdateLevel(ctx, level); err != nil {
		return nil, err
	}
	return level, nil
}

type DeleteLevelCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	LevelID           uuid.UUID
}

type DeleteLevelHandler struct{ repo domain.AcademicRepository }

func NewDeleteLevelHandler(repo domain.AcademicRepository) *DeleteLevelHandler {
	return &DeleteLevelHandler{repo: repo}
}

func (h *DeleteLevelHandler) Handle(ctx context.Context, cmd DeleteLevelCommand) error {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return err
	}
	level, err := h.repo.FindLevelByID(ctx, cmd.LevelID)
	if err != nil {
		return err
	}
	if err := guardSchoolIDMatch(cmd.RequesterRole, cmd.RequesterSchoolID, level.SchoolID); err != nil {
		return err
	}
	return h.repo.SoftDeleteLevel(ctx, cmd.LevelID)
}

// ── Classes ───────────────────────────────────────────────────────────────────

type CreateClassCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	LevelID           uuid.UUID
	Name              string
}

type CreateClassHandler struct{ repo domain.AcademicRepository }

func NewCreateClassHandler(repo domain.AcademicRepository) *CreateClassHandler {
	return &CreateClassHandler{repo: repo}
}

func (h *CreateClassHandler) Handle(ctx context.Context, cmd CreateClassCommand) (*domain.Class, error) {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return nil, err
	}
	schoolID := resolveSchoolID(cmd.RequesterRole, cmd.RequesterSchoolID)
	if schoolID == nil {
		return nil, apperror.New(apperror.ErrBadRequest, "school context required")
	}
	class := &domain.Class{
		ID:               uuid.New(),
		SchoolID:         *schoolID,
		EducationLevelID: cmd.LevelID,
		Name:             cmd.Name,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if err := h.repo.CreateClass(ctx, class); err != nil {
		return nil, err
	}
	return class, nil
}

type ListClassesQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	LevelID           *uuid.UUID
}

type ListClassesHandler struct{ repo domain.AcademicRepository }

func NewListClassesHandler(repo domain.AcademicRepository) *ListClassesHandler {
	return &ListClassesHandler{repo: repo}
}

func (h *ListClassesHandler) Handle(ctx context.Context, q ListClassesQuery) ([]*domain.Class, error) {
	schoolID := resolveSchoolID(q.RequesterRole, q.RequesterSchoolID)
	if schoolID == nil {
		return nil, apperror.New(apperror.ErrBadRequest, "school context required")
	}
	return h.repo.ListClasses(ctx, *schoolID, q.LevelID)
}

type UpdateClassCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ClassID           uuid.UUID
	Name              string
}

type UpdateClassHandler struct{ repo domain.AcademicRepository }

func NewUpdateClassHandler(repo domain.AcademicRepository) *UpdateClassHandler {
	return &UpdateClassHandler{repo: repo}
}

func (h *UpdateClassHandler) Handle(ctx context.Context, cmd UpdateClassCommand) (*domain.Class, error) {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return nil, err
	}
	class, err := h.repo.FindClassByID(ctx, cmd.ClassID)
	if err != nil {
		return nil, err
	}
	if err := guardSchoolIDMatch(cmd.RequesterRole, cmd.RequesterSchoolID, class.SchoolID); err != nil {
		return nil, err
	}
	class.Name = cmd.Name
	class.UpdatedAt = time.Now()
	if err := h.repo.UpdateClass(ctx, class); err != nil {
		return nil, err
	}
	return class, nil
}

type DeleteClassCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ClassID           uuid.UUID
}

type DeleteClassHandler struct{ repo domain.AcademicRepository }

func NewDeleteClassHandler(repo domain.AcademicRepository) *DeleteClassHandler {
	return &DeleteClassHandler{repo: repo}
}

func (h *DeleteClassHandler) Handle(ctx context.Context, cmd DeleteClassCommand) error {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return err
	}
	class, err := h.repo.FindClassByID(ctx, cmd.ClassID)
	if err != nil {
		return err
	}
	if err := guardSchoolIDMatch(cmd.RequesterRole, cmd.RequesterSchoolID, class.SchoolID); err != nil {
		return err
	}
	return h.repo.SoftDeleteClass(ctx, cmd.ClassID)
}

// ── Sub-classes ───────────────────────────────────────────────────────────────

type CreateSubClassCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ClassID           uuid.UUID
	Name              string
}

type CreateSubClassHandler struct{ repo domain.AcademicRepository }

func NewCreateSubClassHandler(repo domain.AcademicRepository) *CreateSubClassHandler {
	return &CreateSubClassHandler{repo: repo}
}

func (h *CreateSubClassHandler) Handle(ctx context.Context, cmd CreateSubClassCommand) (*domain.SubClass, error) {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return nil, err
	}
	schoolID := resolveSchoolID(cmd.RequesterRole, cmd.RequesterSchoolID)
	if schoolID == nil {
		return nil, apperror.New(apperror.ErrBadRequest, "school context required")
	}
	sub := &domain.SubClass{
		ID:        uuid.New(),
		SchoolID:  *schoolID,
		ClassID:   cmd.ClassID,
		Name:      cmd.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := h.repo.CreateSubClass(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

type ListSubClassesQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ClassID           *uuid.UUID
}

type ListSubClassesHandler struct{ repo domain.AcademicRepository }

func NewListSubClassesHandler(repo domain.AcademicRepository) *ListSubClassesHandler {
	return &ListSubClassesHandler{repo: repo}
}

func (h *ListSubClassesHandler) Handle(ctx context.Context, q ListSubClassesQuery) ([]*domain.SubClass, error) {
	schoolID := resolveSchoolID(q.RequesterRole, q.RequesterSchoolID)
	if schoolID == nil {
		return nil, apperror.New(apperror.ErrBadRequest, "school context required")
	}
	return h.repo.ListSubClasses(ctx, *schoolID, q.ClassID)
}

type UpdateSubClassCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	SubClassID        uuid.UUID
	Name              string
}

type UpdateSubClassHandler struct{ repo domain.AcademicRepository }

func NewUpdateSubClassHandler(repo domain.AcademicRepository) *UpdateSubClassHandler {
	return &UpdateSubClassHandler{repo: repo}
}

func (h *UpdateSubClassHandler) Handle(ctx context.Context, cmd UpdateSubClassCommand) (*domain.SubClass, error) {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return nil, err
	}
	sub, err := h.repo.FindSubClassByID(ctx, cmd.SubClassID)
	if err != nil {
		return nil, err
	}
	if err := guardSchoolIDMatch(cmd.RequesterRole, cmd.RequesterSchoolID, sub.SchoolID); err != nil {
		return nil, err
	}
	sub.Name = cmd.Name
	sub.UpdatedAt = time.Now()
	if err := h.repo.UpdateSubClass(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

type DeleteSubClassCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	SubClassID        uuid.UUID
}

type DeleteSubClassHandler struct{ repo domain.AcademicRepository }

func NewDeleteSubClassHandler(repo domain.AcademicRepository) *DeleteSubClassHandler {
	return &DeleteSubClassHandler{repo: repo}
}

func (h *DeleteSubClassHandler) Handle(ctx context.Context, cmd DeleteSubClassCommand) error {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return err
	}
	sub, err := h.repo.FindSubClassByID(ctx, cmd.SubClassID)
	if err != nil {
		return err
	}
	if err := guardSchoolIDMatch(cmd.RequesterRole, cmd.RequesterSchoolID, sub.SchoolID); err != nil {
		return err
	}
	return h.repo.SoftDeleteSubClass(ctx, cmd.SubClassID)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func guardWrite(role string) error {
	if role != "superadmin" && role != "admin_sekolah" {
		return apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can manage academic structure")
	}
	return nil
}

func resolveSchoolID(role string, id *uuid.UUID) *uuid.UUID {
	if role == "superadmin" {
		return id // may be nil; caller must validate
	}
	return id
}

func guardSchoolIDMatch(role string, requesterID *uuid.UUID, targetID uuid.UUID) error {
	if role == "superadmin" {
		return nil
	}
	if requesterID != nil && *requesterID != targetID {
		return apperror.New(apperror.ErrForbidden, "access denied to this resource")
	}
	return nil
}
