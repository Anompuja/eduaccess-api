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

// classScheduleRow is a flat struct for JOIN scans — does NOT embed classScheduleModel
// because GORM's Scan() skips fields of embedded structs that define TableName().
type classScheduleRow struct {
	ID                    uuid.UUID  `gorm:"column:id"`
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
	ClassroomName         string     `gorm:"column:classroom_name"`
	SubjectName           string     `gorm:"column:subject_name"`
	TeacherName           string     `gorm:"column:teacher_name"`
	StartPeriodNumber     *int       `gorm:"column:start_period_number"`
	StartPeriodLabel      *string    `gorm:"column:start_period_label"`
	EndPeriodNumber       *int       `gorm:"column:end_period_number"`
}

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

// classScheduleStudentRow is a flat struct for JOIN scans — same reason as classScheduleRow.
type classScheduleStudentRow struct {
	ID                    uuid.UUID  `gorm:"column:id"`
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
	StudentName           string     `gorm:"column:student_name"`
}

// ── Repository ────────────────────────────────────────────────────────────────

type GormClassScheduleRepository struct {
	db *gorm.DB
}

func NewGormClassScheduleRepository(db *gorm.DB) *GormClassScheduleRepository {
	return &GormClassScheduleRepository{db: db}
}

func scheduleFromRow(r classScheduleRow) *domain.ClassSchedule {
	return &domain.ClassSchedule{
		ID: r.ID, SchoolID: r.SchoolID, ClassroomID: r.ClassroomID,
		SubjectID: r.SubjectID, TeacherID: r.TeacherID,
		StartPeriodID: r.StartPeriodID, EndPeriodID: r.EndPeriodID,
		Date: r.Date, StartTime: r.StartTime, EndTime: r.EndTime,
		TeacherAttendanceTime: r.TeacherAttendanceTime,
		Status:                r.Status,
		DeletedAt:             r.DeletedAt, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		ClassroomName:     r.ClassroomName,
		SubjectName:       r.SubjectName,
		TeacherName:       r.TeacherName,
		StartPeriodNumber: r.StartPeriodNumber,
		StartPeriodLabel:  r.StartPeriodLabel,
		EndPeriodNumber:   r.EndPeriodNumber,
	}
}

func attendanceFromRow(r classScheduleStudentRow) *domain.ClassScheduleStudent {
	return &domain.ClassScheduleStudent{
		ID: r.ID, SchoolID: r.SchoolID, ClassScheduleID: r.ClassScheduleID,
		StudentID: r.StudentID, Type: r.Type, PhotoPath: r.PhotoPath, Note: r.Note,
		StudentAttendanceTime: r.StudentAttendanceTime,
		Status:                r.Status,
		DeletedAt:             r.DeletedAt, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		StudentName: r.StudentName,
	}
}

const scheduleSelectCols = `
	class_schedules.id, class_schedules.school_id,
	class_schedules.school_classroom_id, class_schedules.school_subject_id,
	class_schedules.teacher_id, class_schedules.start_period_id, class_schedules.end_period_id,
	class_schedules.date, class_schedules.start_time, class_schedules.end_time,
	class_schedules.teacher_attendance_time, class_schedules.status,
	class_schedules.deleted_at, class_schedules.created_at, class_schedules.updated_at,
	sc.name AS classroom_name,
	ss.name AS subject_name,
	u.name  AS teacher_name,
	sp.period_number AS start_period_number,
	sp.label         AS start_period_label,
	ep.period_number AS end_period_number`

const scheduleJoins = `
	LEFT JOIN school_classrooms sc ON sc.id = class_schedules.school_classroom_id
	LEFT JOIN school_subjects   ss ON ss.id = class_schedules.school_subject_id
	LEFT JOIN users              u  ON u.id  = class_schedules.teacher_id
	LEFT JOIN school_schedules  sp ON sp.id = class_schedules.start_period_id
	LEFT JOIN school_schedules  ep ON ep.id = class_schedules.end_period_id`

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
	var row classScheduleRow
	err := r.db.WithContext(ctx).
		Table("class_schedules").
		Select(scheduleSelectCols).
		Joins(scheduleJoins).
		Where("class_schedules.id = ? AND class_schedules.deleted_at IS NULL", id).
		First(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.ErrNotFound, "class schedule not found")
		}
		return nil, err
	}
	return scheduleFromRow(row), nil
}

func (r *GormClassScheduleRepository) ListClassSchedules(ctx context.Context, filter domain.ClassScheduleFilter) ([]*domain.ClassSchedule, error) {
	q := r.db.WithContext(ctx).
		Table("class_schedules").
		Select(scheduleSelectCols).
		Joins(scheduleJoins).
		Where("class_schedules.deleted_at IS NULL")
	if filter.SchoolID != nil {
		q = q.Where("class_schedules.school_id = ?", *filter.SchoolID)
	}
	if filter.ClassroomID != nil {
		q = q.Where("class_schedules.school_classroom_id = ?", *filter.ClassroomID)
	}
	if filter.TeacherID != nil {
		q = q.Where("class_schedules.teacher_id = ?", *filter.TeacherID)
	}
	if filter.SubjectID != nil {
		q = q.Where("class_schedules.school_subject_id = ?", *filter.SubjectID)
	}
	if filter.Date != nil {
		q = q.Where("class_schedules.date = ?", filter.Date.Format("2006-01-02"))
	}
	if filter.Status != nil {
		q = q.Where("class_schedules.status = ?", *filter.Status)
	}
	var rows []classScheduleRow
	if err := q.Order("class_schedules.date ASC, class_schedules.start_time ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]*domain.ClassSchedule, 0, len(rows))
	for _, row := range rows {
		list = append(list, scheduleFromRow(row))
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
			Type:            "check_in",
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
				Type:            "check_in",
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
	var rows []classScheduleStudentRow
	err := r.db.WithContext(ctx).
		Table("class_schedule_students").
		Select("class_schedule_students.*, u.name AS student_name").
		Joins("LEFT JOIN users u ON u.id = class_schedule_students.student_id").
		Where("class_schedule_students.class_schedule_id = ? AND class_schedule_students.deleted_at IS NULL", scheduleID).
		Order("class_schedule_students.created_at ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	list := make([]*domain.ClassScheduleStudent, 0, len(rows))
	for _, row := range rows {
		list = append(list, attendanceFromRow(row))
	}
	return list, nil
}

func (r *GormClassScheduleRepository) FindAttendance(ctx context.Context, scheduleID uuid.UUID, studentID uuid.UUID) (*domain.ClassScheduleStudent, error) {
	var row classScheduleStudentRow
	err := r.db.WithContext(ctx).
		Table("class_schedule_students").
		Select("class_schedule_students.*, u.name AS student_name").
		Joins("LEFT JOIN users u ON u.id = class_schedule_students.student_id").
		Where("class_schedule_students.class_schedule_id = ? AND class_schedule_students.student_id = ? AND class_schedule_students.deleted_at IS NULL", scheduleID, studentID).
		First(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.ErrNotFound, "attendance record not found")
		}
		return nil, err
	}
	return attendanceFromRow(row), nil
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
