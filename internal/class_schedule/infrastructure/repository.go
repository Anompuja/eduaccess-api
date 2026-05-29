package infrastructure

import (
	"context"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/class_schedule/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ── GORM models ───────────────────────────────────────────────────────────────

type classScheduleModel struct {
	ID                    uuid.UUID  `gorm:"column:id;primaryKey"`
	SchoolID              uuid.UUID  `gorm:"column:school_id"`
	ClassroomID           uuid.UUID  `gorm:"column:school_classroom_id"`
	SubjectID             uuid.UUID  `gorm:"column:school_subject_id"`
	TeacherID             uuid.UUID  `gorm:"column:teacher_id"`
	StartPeriodID         *uuid.UUID `gorm:"column:start_period_id"`
	EndPeriodID           *uuid.UUID `gorm:"column:end_period_id"`
	Date                  time.Time  `gorm:"column:date"`
	StartTime             string     `gorm:"column:start_time"`
	EndTime               string     `gorm:"column:end_time"`
	TeacherAttendanceTime *time.Time `gorm:"column:teacher_attendance_time"`
	Status                string     `gorm:"column:status"`
	DeletedAt             *time.Time `gorm:"column:deleted_at"`
	CreatedAt             time.Time  `gorm:"column:created_at"`
	UpdatedAt             time.Time  `gorm:"column:updated_at"`
}

func (classScheduleModel) TableName() string { return "class_schedules" }

type classScheduleStudentModel struct {
	ID                    uuid.UUID  `gorm:"column:id;primaryKey"`
	SchoolID              uuid.UUID  `gorm:"column:school_id"`
	ClassScheduleID       uuid.UUID  `gorm:"column:class_schedule_id"`
	StudentID             uuid.UUID  `gorm:"column:student_id"`
	Type                  string     `gorm:"column:type"`
	PhotoPath             string     `gorm:"column:photo_path"`
	Note                  string     `gorm:"column:note"`
	StudentAttendanceTime *time.Time `gorm:"column:student_attendance_time"`
	Status                string     `gorm:"column:status"`
	DeletedAt             *time.Time `gorm:"column:deleted_at"`
	CreatedAt             time.Time  `gorm:"column:created_at"`
	UpdatedAt             time.Time  `gorm:"column:updated_at"`
}

func (classScheduleStudentModel) TableName() string { return "class_schedule_students" }

// ── Repository ────────────────────────────────────────────────────────────────

type GormClassScheduleRepository struct {
	db *gorm.DB
}

func NewGormClassScheduleRepository(db *gorm.DB) *GormClassScheduleRepository {
	return &GormClassScheduleRepository{db: db}
}

func scheduleFromModel(m classScheduleModel) *domain.ClassSchedule {
	return &domain.ClassSchedule{
		ID: m.ID, SchoolID: m.SchoolID, ClassroomID: m.ClassroomID,
		SubjectID: m.SubjectID, TeacherID: m.TeacherID,
		StartPeriodID: m.StartPeriodID, EndPeriodID: m.EndPeriodID,
		Date: m.Date, StartTime: m.StartTime, EndTime: m.EndTime,
		TeacherAttendanceTime: m.TeacherAttendanceTime,
		Status:                m.Status,
		DeletedAt:             m.DeletedAt, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func attendanceFromModel(m classScheduleStudentModel) *domain.ClassScheduleStudent {
	return &domain.ClassScheduleStudent{
		ID: m.ID, SchoolID: m.SchoolID, ClassScheduleID: m.ClassScheduleID,
		StudentID: m.StudentID, Type: m.Type, PhotoPath: m.PhotoPath, Note: m.Note,
		StudentAttendanceTime: m.StudentAttendanceTime,
		Status:                m.Status,
		DeletedAt:             m.DeletedAt, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

// ── Schedule CRUD ─────────────────────────────────────────────────────────────

func (r *GormClassScheduleRepository) CreateClassSchedule(ctx context.Context, cs *domain.ClassSchedule) error {
	m := classScheduleModel{
		ID: cs.ID, SchoolID: cs.SchoolID, ClassroomID: cs.ClassroomID,
		SubjectID: cs.SubjectID, TeacherID: cs.TeacherID,
		StartPeriodID: cs.StartPeriodID, EndPeriodID: cs.EndPeriodID,
		Date: cs.Date, StartTime: cs.StartTime, EndTime: cs.EndTime,
		Status: cs.Status, CreatedAt: cs.CreatedAt, UpdatedAt: cs.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *GormClassScheduleRepository) FindClassScheduleByID(ctx context.Context, id uuid.UUID) (*domain.ClassSchedule, error) {
	var m classScheduleModel
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.ErrNotFound, "class schedule not found")
		}
		return nil, err
	}
	return scheduleFromModel(m), nil
}

func (r *GormClassScheduleRepository) ListClassSchedules(ctx context.Context, filter domain.ClassScheduleFilter) ([]*domain.ClassSchedule, error) {
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if filter.SchoolID != nil {
		q = q.Where("school_id = ?", *filter.SchoolID)
	}
	if filter.ClassroomID != nil {
		q = q.Where("school_classroom_id = ?", *filter.ClassroomID)
	}
	if filter.TeacherID != nil {
		q = q.Where("teacher_id = ?", *filter.TeacherID)
	}
	if filter.SubjectID != nil {
		q = q.Where("school_subject_id = ?", *filter.SubjectID)
	}
	if filter.Date != nil {
		q = q.Where("date = ?", filter.Date.Format("2006-01-02"))
	}
	if filter.Status != nil {
		q = q.Where("status = ?", *filter.Status)
	}
	var rows []classScheduleModel
	if err := q.Order("date ASC, start_time ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]*domain.ClassSchedule, 0, len(rows))
	for _, m := range rows {
		list = append(list, scheduleFromModel(m))
	}
	return list, nil
}

func (r *GormClassScheduleRepository) UpdateClassSchedule(ctx context.Context, cs *domain.ClassSchedule) error {
	return r.db.WithContext(ctx).Table("class_schedules").Where("id = ?", cs.ID).
		Updates(map[string]interface{}{
			"school_classroom_id": cs.ClassroomID,
			"school_subject_id":   cs.SubjectID,
			"teacher_id":          cs.TeacherID,
			"start_period_id":     cs.StartPeriodID,
			"end_period_id":       cs.EndPeriodID,
			"date":                cs.Date,
			"start_time":          cs.StartTime,
			"end_time":            cs.EndTime,
			"updated_at":          cs.UpdatedAt,
		}).Error
}

func (r *GormClassScheduleRepository) SoftDeleteClassSchedule(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Table("class_schedules").Where("id = ?", id).
		Updates(map[string]interface{}{"deleted_at": now, "updated_at": now}).Error
}

// ── Status Transitions ────────────────────────────────────────────────────────

func (r *GormClassScheduleRepository) StartClassSchedule(ctx context.Context, id uuid.UUID, teacherTime time.Time) error {
	now := time.Now()
	return r.db.WithContext(ctx).Table("class_schedules").Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":                  "ongoing",
			"teacher_attendance_time": teacherTime,
			"updated_at":              now,
		}).Error
}

func (r *GormClassScheduleRepository) CompleteClassSchedule(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Table("class_schedules").Where("id = ?", id).
		Updates(map[string]interface{}{"status": "completed", "updated_at": now}).Error
}

func (r *GormClassScheduleRepository) CancelClassSchedule(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Table("class_schedules").Where("id = ?", id).
		Updates(map[string]interface{}{"status": "cancelled", "updated_at": now}).Error
}

// ── Student Attendance ────────────────────────────────────────────────────────

func (r *GormClassScheduleRepository) AutoPopulateStudents(ctx context.Context, scheduleID uuid.UUID, schoolID uuid.UUID, classroomID uuid.UUID) error {
	type studentRow struct {
		StudentID uuid.UUID `gorm:"column:student_id"`
	}
	var studies []studentRow
	if err := r.db.WithContext(ctx).Table("student_studies").Select("student_id").
		Where("school_classroom_id = ? AND status = 'active' AND deleted_at IS NULL", classroomID).
		Find(&studies).Error; err != nil {
		return err
	}
	if len(studies) == 0 {
		return nil
	}
	now := time.Now()
	rows := make([]classScheduleStudentModel, 0, len(studies))
	for _, s := range studies {
		rows = append(rows, classScheduleStudentModel{
			ID:              uuid.New(),
			SchoolID:        schoolID,
			ClassScheduleID: scheduleID,
			StudentID:       s.StudentID,
			Status:          "scheduled",
			CreatedAt:       now,
			UpdatedAt:       now,
		})
	}
	return r.db.WithContext(ctx).Create(&rows).Error
}

func (r *GormClassScheduleRepository) SyncStudents(ctx context.Context, scheduleID uuid.UUID, schoolID uuid.UUID, classroomID uuid.UUID) error {
	type studentRow struct {
		StudentID uuid.UUID `gorm:"column:student_id"`
	}
	var studies []studentRow
	if err := r.db.WithContext(ctx).Table("student_studies").Select("student_id").
		Where("school_classroom_id = ? AND status = 'active' AND deleted_at IS NULL", classroomID).
		Find(&studies).Error; err != nil {
		return err
	}
	var existing []classScheduleStudentModel
	if err := r.db.WithContext(ctx).
		Where("class_schedule_id = ? AND deleted_at IS NULL", scheduleID).
		Find(&existing).Error; err != nil {
		return err
	}
	existingMap := make(map[uuid.UUID]bool, len(existing))
	for _, e := range existing {
		existingMap[e.StudentID] = true
	}
	now := time.Now()
	var newRows []classScheduleStudentModel
	for _, s := range studies {
		if !existingMap[s.StudentID] {
			newRows = append(newRows, classScheduleStudentModel{
				ID:              uuid.New(),
				SchoolID:        schoolID,
				ClassScheduleID: scheduleID,
				StudentID:       s.StudentID,
				Status:          "scheduled",
				CreatedAt:       now,
				UpdatedAt:       now,
			})
		}
	}
	if len(newRows) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&newRows).Error
}

func (r *GormClassScheduleRepository) ListAttendances(ctx context.Context, scheduleID uuid.UUID) ([]*domain.ClassScheduleStudent, error) {
	var rows []classScheduleStudentModel
	if err := r.db.WithContext(ctx).
		Where("class_schedule_id = ? AND deleted_at IS NULL", scheduleID).
		Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]*domain.ClassScheduleStudent, 0, len(rows))
	for _, m := range rows {
		list = append(list, attendanceFromModel(m))
	}
	return list, nil
}

func (r *GormClassScheduleRepository) FindAttendance(ctx context.Context, scheduleID uuid.UUID, studentID uuid.UUID) (*domain.ClassScheduleStudent, error) {
	var m classScheduleStudentModel
	if err := r.db.WithContext(ctx).
		Where("class_schedule_id = ? AND student_id = ? AND deleted_at IS NULL", scheduleID, studentID).
		First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.ErrNotFound, "attendance record not found")
		}
		return nil, err
	}
	return attendanceFromModel(m), nil
}

func (r *GormClassScheduleRepository) UpdateAttendance(ctx context.Context, att *domain.ClassScheduleStudent) error {
	return r.db.WithContext(ctx).Table("class_schedule_students").
		Where("class_schedule_id = ? AND student_id = ? AND deleted_at IS NULL", att.ClassScheduleID, att.StudentID).
		Updates(map[string]interface{}{
			"status":                  att.Status,
			"note":                    att.Note,
			"photo_path":              att.PhotoPath,
			"student_attendance_time": att.StudentAttendanceTime,
			"updated_at":              att.UpdatedAt,
		}).Error
}
