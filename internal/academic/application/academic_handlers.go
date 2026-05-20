package application

import (
	"context"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/academic/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
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
	return h.repo.ListLevels(ctx, schoolID)
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
	return h.repo.ListClasses(ctx, schoolID, q.LevelID)
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
	return h.repo.ListSubClasses(ctx, schoolID, q.ClassID)
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

// ── Academic Years ────────────────────────────────────────────────────────────

type CreateAcademicYearCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	Name              string
	StartDate         time.Time
	EndDate           time.Time
	Description       string
}

type CreateAcademicYearHandler struct{ repo domain.AcademicRepository }

func NewCreateAcademicYearHandler(repo domain.AcademicRepository) *CreateAcademicYearHandler {
	return &CreateAcademicYearHandler{repo: repo}
}

func (h *CreateAcademicYearHandler) Handle(ctx context.Context, cmd CreateAcademicYearCommand) (*domain.AcademicYear, error) {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return nil, err
	}
	schoolID := resolveSchoolID(cmd.RequesterRole, cmd.RequesterSchoolID)
	if schoolID == nil {
		return nil, apperror.New(apperror.ErrBadRequest, "school context required")
	}
	ay := &domain.AcademicYear{
		ID: uuid.New(), SchoolID: *schoolID, Name: cmd.Name, StartDate: cmd.StartDate,
		EndDate: cmd.EndDate, Description: cmd.Description, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := h.repo.CreateAcademicYear(ctx, ay); err != nil {
		return nil, err
	}
	return ay, nil
}

type ListAcademicYearsQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
}

type ListAcademicYearsHandler struct{ repo domain.AcademicRepository }

func NewListAcademicYearsHandler(repo domain.AcademicRepository) *ListAcademicYearsHandler {
	return &ListAcademicYearsHandler{repo: repo}
}

func (h *ListAcademicYearsHandler) Handle(ctx context.Context, q ListAcademicYearsQuery) ([]*domain.AcademicYear, error) {
	schoolID := resolveSchoolID(q.RequesterRole, q.RequesterSchoolID)
	return h.repo.ListAcademicYears(ctx, schoolID)
}

type UpdateAcademicYearCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	AcademicYearID    uuid.UUID
	Name              string
	StartDate         time.Time
	EndDate           time.Time
	Description       string
}

type UpdateAcademicYearHandler struct{ repo domain.AcademicRepository }

func NewUpdateAcademicYearHandler(repo domain.AcademicRepository) *UpdateAcademicYearHandler {
	return &UpdateAcademicYearHandler{repo: repo}
}

func (h *UpdateAcademicYearHandler) Handle(ctx context.Context, cmd UpdateAcademicYearCommand) (*domain.AcademicYear, error) {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return nil, err
	}
	ay, err := h.repo.FindAcademicYearByID(ctx, cmd.AcademicYearID)
	if err != nil {
		return nil, err
	}
	if err := guardSchoolIDMatch(cmd.RequesterRole, cmd.RequesterSchoolID, ay.SchoolID); err != nil {
		return nil, err
	}
	ay.Name = cmd.Name
	ay.StartDate = cmd.StartDate
	ay.EndDate = cmd.EndDate
	ay.Description = cmd.Description
	ay.UpdatedAt = time.Now()
	if err := h.repo.UpdateAcademicYear(ctx, ay); err != nil {
		return nil, err
	}
	return ay, nil
}

type DeleteAcademicYearCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	AcademicYearID    uuid.UUID
}

type DeleteAcademicYearHandler struct{ repo domain.AcademicRepository }

func NewDeleteAcademicYearHandler(repo domain.AcademicRepository) *DeleteAcademicYearHandler {
	return &DeleteAcademicYearHandler{repo: repo}
}

func (h *DeleteAcademicYearHandler) Handle(ctx context.Context, cmd DeleteAcademicYearCommand) error {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return err
	}
	ay, err := h.repo.FindAcademicYearByID(ctx, cmd.AcademicYearID)
	if err != nil {
		return err
	}
	if err := guardSchoolIDMatch(cmd.RequesterRole, cmd.RequesterSchoolID, ay.SchoolID); err != nil {
		return err
	}
	return h.repo.SoftDeleteAcademicYear(ctx, cmd.AcademicYearID)
}

type ActivateAcademicYearCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	AcademicYearID    uuid.UUID
}

type ActivateAcademicYearHandler struct{ repo domain.AcademicRepository }

func NewActivateAcademicYearHandler(repo domain.AcademicRepository) *ActivateAcademicYearHandler {
	return &ActivateAcademicYearHandler{repo: repo}
}

func (h *ActivateAcademicYearHandler) Handle(ctx context.Context, cmd ActivateAcademicYearCommand) error {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return err
	}
	schoolID := resolveSchoolID(cmd.RequesterRole, cmd.RequesterSchoolID)
	if schoolID == nil {
		return apperror.New(apperror.ErrBadRequest, "school context required")
	}
	ay, err := h.repo.FindAcademicYearByID(ctx, cmd.AcademicYearID)
	if err != nil {
		return err
	}
	if err := guardSchoolIDMatch(cmd.RequesterRole, cmd.RequesterSchoolID, ay.SchoolID); err != nil {
		return err
	}
	return h.repo.ActivateAcademicYear(ctx, cmd.AcademicYearID, *schoolID)
}

// ── Subjects ──────────────────────────────────────────────────────────────────

type CreateSubjectCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	Name              string
	Category          string
}

type CreateSubjectHandler struct{ repo domain.AcademicRepository }

func NewCreateSubjectHandler(repo domain.AcademicRepository) *CreateSubjectHandler {
	return &CreateSubjectHandler{repo: repo}
}

func (h *CreateSubjectHandler) Handle(ctx context.Context, cmd CreateSubjectCommand) (*domain.Subject, error) {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return nil, err
	}
	schoolID := resolveSchoolID(cmd.RequesterRole, cmd.RequesterSchoolID)
	if schoolID == nil {
		return nil, apperror.New(apperror.ErrBadRequest, "school context required")
	}
	s := &domain.Subject{ID: uuid.New(), SchoolID: *schoolID, Name: cmd.Name, Category: cmd.Category, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := h.repo.CreateSubject(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

type ListSubjectsQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
}

type ListSubjectsHandler struct{ repo domain.AcademicRepository }

func NewListSubjectsHandler(repo domain.AcademicRepository) *ListSubjectsHandler {
	return &ListSubjectsHandler{repo: repo}
}

func (h *ListSubjectsHandler) Handle(ctx context.Context, q ListSubjectsQuery) ([]*domain.Subject, error) {
	schoolID := resolveSchoolID(q.RequesterRole, q.RequesterSchoolID)
	return h.repo.ListSubjects(ctx, schoolID)
}

type UpdateSubjectCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	SubjectID         uuid.UUID
	Name              string
	Category          string
}

type UpdateSubjectHandler struct{ repo domain.AcademicRepository }

func NewUpdateSubjectHandler(repo domain.AcademicRepository) *UpdateSubjectHandler {
	return &UpdateSubjectHandler{repo: repo}
}

func (h *UpdateSubjectHandler) Handle(ctx context.Context, cmd UpdateSubjectCommand) (*domain.Subject, error) {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return nil, err
	}
	s, err := h.repo.FindSubjectByID(ctx, cmd.SubjectID)
	if err != nil {
		return nil, err
	}
	if err := guardSchoolIDMatch(cmd.RequesterRole, cmd.RequesterSchoolID, s.SchoolID); err != nil {
		return nil, err
	}
	s.Name = cmd.Name
	s.Category = cmd.Category
	s.UpdatedAt = time.Now()
	if err := h.repo.UpdateSubject(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

type DeleteSubjectCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	SubjectID         uuid.UUID
}

type DeleteSubjectHandler struct{ repo domain.AcademicRepository }

func NewDeleteSubjectHandler(repo domain.AcademicRepository) *DeleteSubjectHandler {
	return &DeleteSubjectHandler{repo: repo}
}

func (h *DeleteSubjectHandler) Handle(ctx context.Context, cmd DeleteSubjectCommand) error {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return err
	}
	s, err := h.repo.FindSubjectByID(ctx, cmd.SubjectID)
	if err != nil {
		return err
	}
	if err := guardSchoolIDMatch(cmd.RequesterRole, cmd.RequesterSchoolID, s.SchoolID); err != nil {
		return err
	}
	return h.repo.SoftDeleteSubject(ctx, cmd.SubjectID)
}

// ── Classrooms ────────────────────────────────────────────────────────────────

type CreateClassroomCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	Name              string
	Capacity          int
	Floor             int
	Building          string
	RoomType          string
	Status            string
	Facilities        string
}

type CreateClassroomHandler struct{ repo domain.AcademicRepository }

func NewCreateClassroomHandler(repo domain.AcademicRepository) *CreateClassroomHandler {
	return &CreateClassroomHandler{repo: repo}
}

func (h *CreateClassroomHandler) Handle(ctx context.Context, cmd CreateClassroomCommand) (*domain.Classroom, error) {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return nil, err
	}
	schoolID := resolveSchoolID(cmd.RequesterRole, cmd.RequesterSchoolID)
	if schoolID == nil {
		return nil, apperror.New(apperror.ErrBadRequest, "school context required")
	}
	status := cmd.Status
	if status == "" {
		status = "available"
	}
	c := &domain.Classroom{ID: uuid.New(), SchoolID: *schoolID, Name: cmd.Name, Capacity: cmd.Capacity, Floor: cmd.Floor, Building: cmd.Building, RoomType: cmd.RoomType, Status: status, Facilities: cmd.Facilities, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := h.repo.CreateClassroom(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

type ListClassroomsQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
}

type ListClassroomsHandler struct{ repo domain.AcademicRepository }

func NewListClassroomsHandler(repo domain.AcademicRepository) *ListClassroomsHandler {
	return &ListClassroomsHandler{repo: repo}
}

func (h *ListClassroomsHandler) Handle(ctx context.Context, q ListClassroomsQuery) ([]*domain.Classroom, error) {
	schoolID := resolveSchoolID(q.RequesterRole, q.RequesterSchoolID)
	return h.repo.ListClassrooms(ctx, schoolID)
}

type UpdateClassroomCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ClassroomID       uuid.UUID
	Name              string
	Capacity          int
	Floor             int
	Building          string
	RoomType          string
	Status            string
	Facilities        string
}

type UpdateClassroomHandler struct{ repo domain.AcademicRepository }

func NewUpdateClassroomHandler(repo domain.AcademicRepository) *UpdateClassroomHandler {
	return &UpdateClassroomHandler{repo: repo}
}

func (h *UpdateClassroomHandler) Handle(ctx context.Context, cmd UpdateClassroomCommand) (*domain.Classroom, error) {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return nil, err
	}
	c, err := h.repo.FindClassroomByID(ctx, cmd.ClassroomID)
	if err != nil {
		return nil, err
	}
	if err := guardSchoolIDMatch(cmd.RequesterRole, cmd.RequesterSchoolID, c.SchoolID); err != nil {
		return nil, err
	}
	c.Name = cmd.Name
	c.Capacity = cmd.Capacity
	c.Floor = cmd.Floor
	c.Building = cmd.Building
	c.RoomType = cmd.RoomType
	c.Status = cmd.Status
	c.Facilities = cmd.Facilities
	c.UpdatedAt = time.Now()
	if err := h.repo.UpdateClassroom(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

type DeleteClassroomCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ClassroomID       uuid.UUID
}

type DeleteClassroomHandler struct{ repo domain.AcademicRepository }

func NewDeleteClassroomHandler(repo domain.AcademicRepository) *DeleteClassroomHandler {
	return &DeleteClassroomHandler{repo: repo}
}

func (h *DeleteClassroomHandler) Handle(ctx context.Context, cmd DeleteClassroomCommand) error {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return err
	}
	c, err := h.repo.FindClassroomByID(ctx, cmd.ClassroomID)
	if err != nil {
		return err
	}
	if err := guardSchoolIDMatch(cmd.RequesterRole, cmd.RequesterSchoolID, c.SchoolID); err != nil {
		return err
	}
	return h.repo.SoftDeleteClassroom(ctx, cmd.ClassroomID)
}

// ── Schedules ─────────────────────────────────────────────────────────────────

type CreateScheduleCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ShiftType         string
	StartTime         string
	EndTime           string
}

type CreateScheduleHandler struct{ repo domain.AcademicRepository }

func NewCreateScheduleHandler(repo domain.AcademicRepository) *CreateScheduleHandler {
	return &CreateScheduleHandler{repo: repo}
}

func (h *CreateScheduleHandler) Handle(ctx context.Context, cmd CreateScheduleCommand) (*domain.Schedule, error) {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return nil, err
	}
	schoolID := resolveSchoolID(cmd.RequesterRole, cmd.RequesterSchoolID)
	if schoolID == nil {
		return nil, apperror.New(apperror.ErrBadRequest, "school context required")
	}
	s := &domain.Schedule{ID: uuid.New(), SchoolID: *schoolID, ShiftType: cmd.ShiftType, StartTime: cmd.StartTime, EndTime: cmd.EndTime, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := h.repo.CreateSchedule(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

type ListSchedulesQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
}

type ListSchedulesHandler struct{ repo domain.AcademicRepository }

func NewListSchedulesHandler(repo domain.AcademicRepository) *ListSchedulesHandler {
	return &ListSchedulesHandler{repo: repo}
}

func (h *ListSchedulesHandler) Handle(ctx context.Context, q ListSchedulesQuery) ([]*domain.Schedule, error) {
	schoolID := resolveSchoolID(q.RequesterRole, q.RequesterSchoolID)
	return h.repo.ListSchedules(ctx, schoolID)
}

type UpdateScheduleCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ScheduleID        uuid.UUID
	ShiftType         string
	StartTime         string
	EndTime           string
}

type UpdateScheduleHandler struct{ repo domain.AcademicRepository }

func NewUpdateScheduleHandler(repo domain.AcademicRepository) *UpdateScheduleHandler {
	return &UpdateScheduleHandler{repo: repo}
}

func (h *UpdateScheduleHandler) Handle(ctx context.Context, cmd UpdateScheduleCommand) (*domain.Schedule, error) {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return nil, err
	}
	s, err := h.repo.FindScheduleByID(ctx, cmd.ScheduleID)
	if err != nil {
		return nil, err
	}
	if err := guardSchoolIDMatch(cmd.RequesterRole, cmd.RequesterSchoolID, s.SchoolID); err != nil {
		return nil, err
	}
	s.ShiftType = cmd.ShiftType
	s.StartTime = cmd.StartTime
	s.EndTime = cmd.EndTime
	s.UpdatedAt = time.Now()
	if err := h.repo.UpdateSchedule(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

type DeleteScheduleCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ScheduleID        uuid.UUID
}

type DeleteScheduleHandler struct{ repo domain.AcademicRepository }

func NewDeleteScheduleHandler(repo domain.AcademicRepository) *DeleteScheduleHandler {
	return &DeleteScheduleHandler{repo: repo}
}

func (h *DeleteScheduleHandler) Handle(ctx context.Context, cmd DeleteScheduleCommand) error {
	if err := guardWrite(cmd.RequesterRole); err != nil {
		return err
	}
	s, err := h.repo.FindScheduleByID(ctx, cmd.ScheduleID)
	if err != nil {
		return err
	}
	if err := guardSchoolIDMatch(cmd.RequesterRole, cmd.RequesterSchoolID, s.SchoolID); err != nil {
		return err
	}
	return h.repo.SoftDeleteSchedule(ctx, cmd.ScheduleID)
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
