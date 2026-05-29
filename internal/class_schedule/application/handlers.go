package application

import (
	"context"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/class_schedule/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

// ── Guards ────────────────────────────────────────────────────────────────────

func guardWrite(role string) error {
	if role != "superadmin" && role != "admin_sekolah" {
		return apperror.New(apperror.ErrForbidden, "only admin_sekolah can manage class schedules")
	}
	return nil
}

func guardSessionControl(role string) error {
	if role == "student" || role == "parent" {
		return apperror.New(apperror.ErrForbidden, "students and parents cannot control sessions")
	}
	return nil
}

func guardSchoolMatch(role string, requesterSchoolID *uuid.UUID, targetSchoolID uuid.UUID) error {
	if role == "superadmin" {
		return nil
	}
	if requesterSchoolID == nil {
		return apperror.New(apperror.ErrForbidden, "school context required")
	}
	if *requesterSchoolID != targetSchoolID {
		return apperror.New(apperror.ErrForbidden, "access denied to this resource")
	}
	return nil
}

func resolveSchoolID(_ string, id *uuid.UUID) *uuid.UUID {
	return id
}

// ── Create ────────────────────────────────────────────────────────────────────

type CreateClassScheduleCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ClassroomID       uuid.UUID
	SubjectID         uuid.UUID
	TeacherID         uuid.UUID
	StartPeriodID     *uuid.UUID
	EndPeriodID       *uuid.UUID
	Date              time.Time
	StartTime         string
	EndTime           string
}

type CreateClassScheduleHandler struct{ repo domain.ClassScheduleRepository }

func NewCreateClassScheduleHandler(repo domain.ClassScheduleRepository) *CreateClassScheduleHandler {
	return &CreateClassScheduleHandler{repo: repo}
}

func (h *CreateClassScheduleHandler) Handle(ctx context.Context, cmd CreateClassScheduleCommand) (*domain.ClassSchedule, error) {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return nil, err
	}
	schoolID := resolveSchoolID(cmd.RequesterRole, cmd.RequesterSchoolID)
	if schoolID == nil {
		return nil, apperror.New(apperror.ErrBadRequest, "school context required")
	}
	cs := &domain.ClassSchedule{
		ID:            uuid.New(),
		SchoolID:      *schoolID,
		ClassroomID:   cmd.ClassroomID,
		SubjectID:     cmd.SubjectID,
		TeacherID:     cmd.TeacherID,
		StartPeriodID: cmd.StartPeriodID,
		EndPeriodID:   cmd.EndPeriodID,
		Date:          cmd.Date,
		StartTime:     cmd.StartTime,
		EndTime:       cmd.EndTime,
		Status:        "scheduled",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := h.repo.CreateClassSchedule(ctx, cs); err != nil {
		return nil, err
	}
	if err := h.repo.AutoPopulateStudents(ctx, cs.ID, cs.SchoolID, cs.ClassroomID); err != nil {
		return nil, err
	}
	return cs, nil
}

// ── List ──────────────────────────────────────────────────────────────────────

type ListClassSchedulesQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ClassroomID       *uuid.UUID
	TeacherID         *uuid.UUID
	SubjectID         *uuid.UUID
	Date              *time.Time
	Status            *string
}

type ListClassSchedulesHandler struct{ repo domain.ClassScheduleRepository }

func NewListClassSchedulesHandler(repo domain.ClassScheduleRepository) *ListClassSchedulesHandler {
	return &ListClassSchedulesHandler{repo: repo}
}

func (h *ListClassSchedulesHandler) Handle(ctx context.Context, q ListClassSchedulesQuery) ([]*domain.ClassSchedule, error) {
	schoolID := resolveSchoolID(q.RequesterRole, q.RequesterSchoolID)
	return h.repo.ListClassSchedules(ctx, domain.ClassScheduleFilter{
		SchoolID:    schoolID,
		ClassroomID: q.ClassroomID,
		TeacherID:   q.TeacherID,
		SubjectID:   q.SubjectID,
		Date:        q.Date,
		Status:      q.Status,
	})
}

// ── Get ───────────────────────────────────────────────────────────────────────

type GetClassScheduleQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ScheduleID        uuid.UUID
}

type GetClassScheduleHandler struct{ repo domain.ClassScheduleRepository }

func NewGetClassScheduleHandler(repo domain.ClassScheduleRepository) *GetClassScheduleHandler {
	return &GetClassScheduleHandler{repo: repo}
}

func (h *GetClassScheduleHandler) Handle(ctx context.Context, q GetClassScheduleQuery) (*domain.ClassSchedule, error) {
	cs, err := h.repo.FindClassScheduleByID(ctx, q.ScheduleID)
	if err != nil {
		return nil, err
	}
	if err := guardSchoolMatch(q.RequesterRole, q.RequesterSchoolID, cs.SchoolID); err != nil {
		return nil, err
	}
	return cs, nil
}

// ── Update ────────────────────────────────────────────────────────────────────

type UpdateClassScheduleCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ScheduleID        uuid.UUID
	ClassroomID       uuid.UUID
	SubjectID         uuid.UUID
	TeacherID         uuid.UUID
	StartPeriodID     *uuid.UUID
	EndPeriodID       *uuid.UUID
	Date              time.Time
	StartTime         string
	EndTime           string
}

type UpdateClassScheduleHandler struct{ repo domain.ClassScheduleRepository }

func NewUpdateClassScheduleHandler(repo domain.ClassScheduleRepository) *UpdateClassScheduleHandler {
	return &UpdateClassScheduleHandler{repo: repo}
}

func (h *UpdateClassScheduleHandler) Handle(ctx context.Context, cmd UpdateClassScheduleCommand) (*domain.ClassSchedule, error) {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return nil, err
	}
	cs, err := h.repo.FindClassScheduleByID(ctx, cmd.ScheduleID)
	if err != nil {
		return nil, err
	}
	if err := guardSchoolMatch(cmd.RequesterRole, cmd.RequesterSchoolID, cs.SchoolID); err != nil {
		return nil, err
	}
	if cs.Status != "scheduled" {
		return nil, apperror.New(apperror.ErrBadRequest, "can only update a scheduled (not yet started) class")
	}
	cs.ClassroomID = cmd.ClassroomID
	cs.SubjectID = cmd.SubjectID
	cs.TeacherID = cmd.TeacherID
	cs.StartPeriodID = cmd.StartPeriodID
	cs.EndPeriodID = cmd.EndPeriodID
	cs.Date = cmd.Date
	cs.StartTime = cmd.StartTime
	cs.EndTime = cmd.EndTime
	cs.UpdatedAt = time.Now()
	if err := h.repo.UpdateClassSchedule(ctx, cs); err != nil {
		return nil, err
	}
	return cs, nil
}

// ── Delete ────────────────────────────────────────────────────────────────────

type DeleteClassScheduleCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ScheduleID        uuid.UUID
}

type DeleteClassScheduleHandler struct{ repo domain.ClassScheduleRepository }

func NewDeleteClassScheduleHandler(repo domain.ClassScheduleRepository) *DeleteClassScheduleHandler {
	return &DeleteClassScheduleHandler{repo: repo}
}

func (h *DeleteClassScheduleHandler) Handle(ctx context.Context, cmd DeleteClassScheduleCommand) error {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return err
	}
	cs, err := h.repo.FindClassScheduleByID(ctx, cmd.ScheduleID)
	if err != nil {
		return err
	}
	if err := guardSchoolMatch(cmd.RequesterRole, cmd.RequesterSchoolID, cs.SchoolID); err != nil {
		return err
	}
	if cs.Status != "scheduled" {
		return apperror.New(apperror.ErrBadRequest, "can only delete a scheduled (not yet started) class")
	}
	return h.repo.SoftDeleteClassSchedule(ctx, cmd.ScheduleID)
}

// ── Start ─────────────────────────────────────────────────────────────────────

type StartClassScheduleCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ScheduleID        uuid.UUID
}

type StartClassScheduleHandler struct{ repo domain.ClassScheduleRepository }

func NewStartClassScheduleHandler(repo domain.ClassScheduleRepository) *StartClassScheduleHandler {
	return &StartClassScheduleHandler{repo: repo}
}

func (h *StartClassScheduleHandler) Handle(ctx context.Context, cmd StartClassScheduleCommand) error {
	if err := guardSessionControl(cmd.RequesterRole); err != nil {
		return err
	}
	cs, err := h.repo.FindClassScheduleByID(ctx, cmd.ScheduleID)
	if err != nil {
		return err
	}
	if err := guardSchoolMatch(cmd.RequesterRole, cmd.RequesterSchoolID, cs.SchoolID); err != nil {
		return err
	}
	if cs.Status != "scheduled" {
		return apperror.New(apperror.ErrBadRequest, "class is not in scheduled status")
	}
	return h.repo.StartClassSchedule(ctx, cmd.ScheduleID, time.Now())
}

// ── Complete ──────────────────────────────────────────────────────────────────

type CompleteClassScheduleCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ScheduleID        uuid.UUID
}

type CompleteClassScheduleHandler struct{ repo domain.ClassScheduleRepository }

func NewCompleteClassScheduleHandler(repo domain.ClassScheduleRepository) *CompleteClassScheduleHandler {
	return &CompleteClassScheduleHandler{repo: repo}
}

func (h *CompleteClassScheduleHandler) Handle(ctx context.Context, cmd CompleteClassScheduleCommand) error {
	if err := guardSessionControl(cmd.RequesterRole); err != nil {
		return err
	}
	cs, err := h.repo.FindClassScheduleByID(ctx, cmd.ScheduleID)
	if err != nil {
		return err
	}
	if err := guardSchoolMatch(cmd.RequesterRole, cmd.RequesterSchoolID, cs.SchoolID); err != nil {
		return err
	}
	if cs.Status != "ongoing" {
		return apperror.New(apperror.ErrBadRequest, "class is not ongoing")
	}
	return h.repo.CompleteClassSchedule(ctx, cmd.ScheduleID)
}

// ── Cancel ────────────────────────────────────────────────────────────────────

type CancelClassScheduleCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ScheduleID        uuid.UUID
}

type CancelClassScheduleHandler struct{ repo domain.ClassScheduleRepository }

func NewCancelClassScheduleHandler(repo domain.ClassScheduleRepository) *CancelClassScheduleHandler {
	return &CancelClassScheduleHandler{repo: repo}
}

func (h *CancelClassScheduleHandler) Handle(ctx context.Context, cmd CancelClassScheduleCommand) error {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return err
	}
	cs, err := h.repo.FindClassScheduleByID(ctx, cmd.ScheduleID)
	if err != nil {
		return err
	}
	if err := guardSchoolMatch(cmd.RequesterRole, cmd.RequesterSchoolID, cs.SchoolID); err != nil {
		return err
	}
	if cs.Status != "scheduled" {
		return apperror.New(apperror.ErrBadRequest, "can only cancel a scheduled class")
	}
	return h.repo.CancelClassSchedule(ctx, cmd.ScheduleID)
}

// ── Sync Students ─────────────────────────────────────────────────────────────

type SyncStudentsCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ScheduleID        uuid.UUID
}

type SyncStudentsHandler struct{ repo domain.ClassScheduleRepository }

func NewSyncStudentsHandler(repo domain.ClassScheduleRepository) *SyncStudentsHandler {
	return &SyncStudentsHandler{repo: repo}
}

func (h *SyncStudentsHandler) Handle(ctx context.Context, cmd SyncStudentsCommand) error {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return err
	}
	cs, err := h.repo.FindClassScheduleByID(ctx, cmd.ScheduleID)
	if err != nil {
		return err
	}
	if err := guardSchoolMatch(cmd.RequesterRole, cmd.RequesterSchoolID, cs.SchoolID); err != nil {
		return err
	}
	if cs.Status == "completed" || cs.Status == "cancelled" {
		return apperror.New(apperror.ErrBadRequest, "cannot sync students on a completed or cancelled class")
	}
	return h.repo.SyncStudents(ctx, cs.ID, cs.SchoolID, cs.ClassroomID)
}

// ── List Attendances ──────────────────────────────────────────────────────────

type ListAttendancesQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ScheduleID        uuid.UUID
}

type ListAttendancesHandler struct{ repo domain.ClassScheduleRepository }

func NewListAttendancesHandler(repo domain.ClassScheduleRepository) *ListAttendancesHandler {
	return &ListAttendancesHandler{repo: repo}
}

func (h *ListAttendancesHandler) Handle(ctx context.Context, q ListAttendancesQuery) ([]*domain.ClassScheduleStudent, error) {
	if err := guardSessionControl(q.RequesterRole); err != nil {
		return nil, err
	}
	cs, err := h.repo.FindClassScheduleByID(ctx, q.ScheduleID)
	if err != nil {
		return nil, err
	}
	if err := guardSchoolMatch(q.RequesterRole, q.RequesterSchoolID, cs.SchoolID); err != nil {
		return nil, err
	}
	return h.repo.ListAttendances(ctx, q.ScheduleID)
}

// ── Update Attendance ─────────────────────────────────────────────────────────

type UpdateAttendanceCommand struct {
	RequesterSchoolID     *uuid.UUID
	RequesterRole         string
	ScheduleID            uuid.UUID
	StudentID             uuid.UUID
	Status                string
	Note                  string
	PhotoPath             string
	StudentAttendanceTime *time.Time
}

type UpdateAttendanceHandler struct{ repo domain.ClassScheduleRepository }

func NewUpdateAttendanceHandler(repo domain.ClassScheduleRepository) *UpdateAttendanceHandler {
	return &UpdateAttendanceHandler{repo: repo}
}

func (h *UpdateAttendanceHandler) Handle(ctx context.Context, cmd UpdateAttendanceCommand) (*domain.ClassScheduleStudent, error) {
	if err := guardSessionControl(cmd.RequesterRole); err != nil {
		return nil, err
	}
	cs, err := h.repo.FindClassScheduleByID(ctx, cmd.ScheduleID)
	if err != nil {
		return nil, err
	}
	if err := guardSchoolMatch(cmd.RequesterRole, cmd.RequesterSchoolID, cs.SchoolID); err != nil {
		return nil, err
	}
	if cs.Status == "cancelled" {
		return nil, apperror.New(apperror.ErrBadRequest, "cannot update attendance on a cancelled class")
	}
	att, err := h.repo.FindAttendance(ctx, cmd.ScheduleID, cmd.StudentID)
	if err != nil {
		return nil, err
	}
	att.Status = cmd.Status
	att.Note = cmd.Note
	att.PhotoPath = cmd.PhotoPath
	att.StudentAttendanceTime = cmd.StudentAttendanceTime
	att.UpdatedAt = time.Now()
	if err := h.repo.UpdateAttendance(ctx, att); err != nil {
		return nil, err
	}
	return att, nil
}
