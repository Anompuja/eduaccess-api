package infrastructure

import (
	"context"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/academic/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ── GORM models ───────────────────────────────────────────────────────────────

type educationLevelModel struct {
	ID        uuid.UUID  `gorm:"column:id;primaryKey"`
	SchoolID  uuid.UUID  `gorm:"column:school_id"`
	Name      string     `gorm:"column:name"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
}

func (educationLevelModel) TableName() string { return "school_education_levels" }

type classModel struct {
	ID               uuid.UUID  `gorm:"column:id;primaryKey"`
	SchoolID         uuid.UUID  `gorm:"column:school_id"`
	EducationLevelID uuid.UUID  `gorm:"column:school_education_level_id"`
	Name             string     `gorm:"column:name"`
	DeletedAt        *time.Time `gorm:"column:deleted_at"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

func (classModel) TableName() string { return "school_classes" }

type subClassModel struct {
	ID        uuid.UUID  `gorm:"column:id;primaryKey"`
	SchoolID  uuid.UUID  `gorm:"column:school_id"`
	ClassID   uuid.UUID  `gorm:"column:school_class_id"`
	Name      string     `gorm:"column:name"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
}

func (subClassModel) TableName() string { return "school_sub_classes" }

// ── Repository ────────────────────────────────────────────────────────────────

// GormAcademicRepository implements domain.AcademicRepository.
type GormAcademicRepository struct {
	db *gorm.DB
}

func NewGormAcademicRepository(db *gorm.DB) *GormAcademicRepository {
	return &GormAcademicRepository{db: db}
}

// ── Education Levels ──────────────────────────────────────────────────────────

func (r *GormAcademicRepository) CreateLevel(ctx context.Context, l *domain.EducationLevel) error {
	m := educationLevelModel{ID: l.ID, SchoolID: l.SchoolID, Name: l.Name, CreatedAt: l.CreatedAt, UpdatedAt: l.UpdatedAt}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *GormAcademicRepository) FindLevelByID(ctx context.Context, id uuid.UUID) (*domain.EducationLevel, error) {
	var m educationLevelModel
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.ErrNotFound, "education level not found")
		}
		return nil, err
	}
	return &domain.EducationLevel{ID: m.ID, SchoolID: m.SchoolID, Name: m.Name, DeletedAt: m.DeletedAt, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}, nil
}

func (r *GormAcademicRepository) ListLevels(ctx context.Context, schoolID *uuid.UUID) ([]*domain.EducationLevel, error) {
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if schoolID != nil {
		q = q.Where("school_id = ?", *schoolID)
	}
	var rows []educationLevelModel
	if err := q.Order("name ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	levels := make([]*domain.EducationLevel, 0, len(rows))
	for _, m := range rows {
		levels = append(levels, &domain.EducationLevel{ID: m.ID, SchoolID: m.SchoolID, Name: m.Name, DeletedAt: m.DeletedAt, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt})
	}
	return levels, nil
}

func (r *GormAcademicRepository) UpdateLevel(ctx context.Context, l *domain.EducationLevel) error {
	return r.db.WithContext(ctx).Table("school_education_levels").Where("id = ?", l.ID).
		Updates(map[string]interface{}{"name": l.Name, "updated_at": l.UpdatedAt}).Error
}

func (r *GormAcademicRepository) SoftDeleteLevel(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Table("school_education_levels").Where("id = ?", id).
		Updates(map[string]interface{}{"deleted_at": now, "updated_at": now}).Error
}

// ── Classes ───────────────────────────────────────────────────────────────────

func (r *GormAcademicRepository) CreateClass(ctx context.Context, c *domain.Class) error {
	m := classModel{ID: c.ID, SchoolID: c.SchoolID, EducationLevelID: c.EducationLevelID, Name: c.Name, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *GormAcademicRepository) FindClassByID(ctx context.Context, id uuid.UUID) (*domain.Class, error) {
	var m classModel
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.ErrNotFound, "class not found")
		}
		return nil, err
	}
	return &domain.Class{ID: m.ID, SchoolID: m.SchoolID, EducationLevelID: m.EducationLevelID, Name: m.Name, DeletedAt: m.DeletedAt, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}, nil
}

func (r *GormAcademicRepository) ListClasses(ctx context.Context, schoolID *uuid.UUID, levelID *uuid.UUID) ([]*domain.Class, error) {
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if schoolID != nil {
		q = q.Where("school_id = ?", *schoolID)
	}
	if levelID != nil {
		q = q.Where("school_education_level_id = ?", *levelID)
	}
	var rows []classModel
	if err := q.Order("name ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	classes := make([]*domain.Class, 0, len(rows))
	for _, m := range rows {
		classes = append(classes, &domain.Class{ID: m.ID, SchoolID: m.SchoolID, EducationLevelID: m.EducationLevelID, Name: m.Name, DeletedAt: m.DeletedAt, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt})
	}
	return classes, nil
}

func (r *GormAcademicRepository) UpdateClass(ctx context.Context, c *domain.Class) error {
	return r.db.WithContext(ctx).Table("school_classes").Where("id = ?", c.ID).
		Updates(map[string]interface{}{"name": c.Name, "updated_at": c.UpdatedAt}).Error
}

func (r *GormAcademicRepository) SoftDeleteClass(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Table("school_classes").Where("id = ?", id).
		Updates(map[string]interface{}{"deleted_at": now, "updated_at": now}).Error
}

// ── Sub-classes ───────────────────────────────────────────────────────────────

func (r *GormAcademicRepository) CreateSubClass(ctx context.Context, s *domain.SubClass) error {
	m := subClassModel{ID: s.ID, SchoolID: s.SchoolID, ClassID: s.ClassID, Name: s.Name, CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *GormAcademicRepository) FindSubClassByID(ctx context.Context, id uuid.UUID) (*domain.SubClass, error) {
	var m subClassModel
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.ErrNotFound, "sub-class not found")
		}
		return nil, err
	}
	return &domain.SubClass{ID: m.ID, SchoolID: m.SchoolID, ClassID: m.ClassID, Name: m.Name, DeletedAt: m.DeletedAt, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}, nil
}

func (r *GormAcademicRepository) ListSubClasses(ctx context.Context, schoolID *uuid.UUID, classID *uuid.UUID) ([]*domain.SubClass, error) {
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if schoolID != nil {
		q = q.Where("school_id = ?", *schoolID)
	}
	if classID != nil {
		q = q.Where("school_class_id = ?", *classID)
	}
	var rows []subClassModel
	if err := q.Order("name ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	subs := make([]*domain.SubClass, 0, len(rows))
	for _, m := range rows {
		subs = append(subs, &domain.SubClass{ID: m.ID, SchoolID: m.SchoolID, ClassID: m.ClassID, Name: m.Name, DeletedAt: m.DeletedAt, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt})
	}
	return subs, nil
}

func (r *GormAcademicRepository) UpdateSubClass(ctx context.Context, s *domain.SubClass) error {
	return r.db.WithContext(ctx).Table("school_sub_classes").Where("id = ?", s.ID).
		Updates(map[string]interface{}{"name": s.Name, "updated_at": s.UpdatedAt}).Error
}

func (r *GormAcademicRepository) SoftDeleteSubClass(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Table("school_sub_classes").Where("id = ?", id).
		Updates(map[string]interface{}{"deleted_at": now, "updated_at": now}).Error
}

// ── Academic Years ─────────────────────────────────────────────────────────────

type academicYearModel struct {
	ID          uuid.UUID  `gorm:"column:id;primaryKey"`
	SchoolID    uuid.UUID  `gorm:"column:school_id"`
	Name        string     `gorm:"column:name"`
	StartDate   time.Time  `gorm:"column:start_date"`
	EndDate     time.Time  `gorm:"column:end_date"`
	IsActive    bool       `gorm:"column:is_active"`
	Description string     `gorm:"column:description"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
}

func (academicYearModel) TableName() string { return "school_academic_years" }

func (r *GormAcademicRepository) CreateAcademicYear(ctx context.Context, ay *domain.AcademicYear) error {
	m := academicYearModel{ID: ay.ID, SchoolID: ay.SchoolID, Name: ay.Name, StartDate: ay.StartDate, EndDate: ay.EndDate, IsActive: ay.IsActive, Description: ay.Description, CreatedAt: ay.CreatedAt, UpdatedAt: ay.UpdatedAt}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *GormAcademicRepository) FindAcademicYearByID(ctx context.Context, id uuid.UUID) (*domain.AcademicYear, error) {
	var m academicYearModel
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.ErrNotFound, "academic year not found")
		}
		return nil, err
	}
	return &domain.AcademicYear{ID: m.ID, SchoolID: m.SchoolID, Name: m.Name, StartDate: m.StartDate, EndDate: m.EndDate, IsActive: m.IsActive, Description: m.Description, DeletedAt: m.DeletedAt, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}, nil
}

func (r *GormAcademicRepository) ListAcademicYears(ctx context.Context, schoolID *uuid.UUID) ([]*domain.AcademicYear, error) {
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if schoolID != nil {
		q = q.Where("school_id = ?", *schoolID)
	}
	var rows []academicYearModel
	if err := q.Order("start_date DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]*domain.AcademicYear, 0, len(rows))
	for _, m := range rows {
		list = append(list, &domain.AcademicYear{ID: m.ID, SchoolID: m.SchoolID, Name: m.Name, StartDate: m.StartDate, EndDate: m.EndDate, IsActive: m.IsActive, Description: m.Description, DeletedAt: m.DeletedAt, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt})
	}
	return list, nil
}

func (r *GormAcademicRepository) UpdateAcademicYear(ctx context.Context, ay *domain.AcademicYear) error {
	return r.db.WithContext(ctx).Table("school_academic_years").Where("id = ?", ay.ID).
		Updates(map[string]interface{}{"name": ay.Name, "start_date": ay.StartDate, "end_date": ay.EndDate, "description": ay.Description, "updated_at": ay.UpdatedAt}).Error
}

func (r *GormAcademicRepository) SoftDeleteAcademicYear(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Table("school_academic_years").Where("id = ?", id).
		Updates(map[string]interface{}{"deleted_at": now, "updated_at": now}).Error
}

func (r *GormAcademicRepository) ActivateAcademicYear(ctx context.Context, id uuid.UUID, schoolID uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Table("school_academic_years").Where("school_id = ? AND deleted_at IS NULL", schoolID).
		Updates(map[string]interface{}{"is_active": false, "updated_at": now}).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Table("school_academic_years").Where("id = ?", id).
		Updates(map[string]interface{}{"is_active": true, "updated_at": now}).Error
}

// ── Subjects ──────────────────────────────────────────────────────────────────

type subjectModel struct {
	ID               uuid.UUID  `gorm:"column:id;primaryKey"`
	SchoolID         uuid.UUID  `gorm:"column:school_id"`
	EducationLevelID *uuid.UUID `gorm:"column:school_education_level_id"`
	Name             string     `gorm:"column:name"`
	Code             *string    `gorm:"column:code"`
	Category         string     `gorm:"column:category"`
	DeletedAt        *time.Time `gorm:"column:deleted_at"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

func (subjectModel) TableName() string { return "school_subjects" }

func subjectFromModel(m subjectModel) *domain.Subject {
	return &domain.Subject{
		ID: m.ID, SchoolID: m.SchoolID, EducationLevelID: m.EducationLevelID,
		Name: m.Name, Code: m.Code, Category: m.Category,
		DeletedAt: m.DeletedAt, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func (r *GormAcademicRepository) CreateSubject(ctx context.Context, s *domain.Subject) error {
	m := subjectModel{
		ID: s.ID, SchoolID: s.SchoolID, EducationLevelID: s.EducationLevelID,
		Name: s.Name, Code: s.Code, Category: s.Category,
		CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *GormAcademicRepository) FindSubjectByID(ctx context.Context, id uuid.UUID) (*domain.Subject, error) {
	var m subjectModel
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.ErrNotFound, "subject not found")
		}
		return nil, err
	}
	return subjectFromModel(m), nil
}

func (r *GormAcademicRepository) ListSubjects(ctx context.Context, schoolID *uuid.UUID) ([]*domain.Subject, error) {
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if schoolID != nil {
		q = q.Where("school_id = ?", *schoolID)
	}
	var rows []subjectModel
	if err := q.Order("name ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]*domain.Subject, 0, len(rows))
	for _, m := range rows {
		list = append(list, subjectFromModel(m))
	}
	return list, nil
}

func (r *GormAcademicRepository) UpdateSubject(ctx context.Context, s *domain.Subject) error {
	return r.db.WithContext(ctx).Table("school_subjects").Where("id = ?", s.ID).
		Updates(map[string]interface{}{
			"school_education_level_id": s.EducationLevelID,
			"name":                      s.Name,
			"code":                      s.Code,
			"category":                  s.Category,
			"updated_at":                s.UpdatedAt,
		}).Error
}

func (r *GormAcademicRepository) SoftDeleteSubject(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Table("school_subjects").Where("id = ?", id).
		Updates(map[string]interface{}{"deleted_at": now, "updated_at": now}).Error
}

// ── Classrooms ────────────────────────────────────────────────────────────────

type classroomModel struct {
	ID                uuid.UUID  `gorm:"column:id;primaryKey"`
	SchoolID          uuid.UUID  `gorm:"column:school_id"`
	ClassID           *uuid.UUID `gorm:"column:school_class_id"`
	SubClassID        *uuid.UUID `gorm:"column:school_sub_class_id"`
	AcademicYearID    *uuid.UUID `gorm:"column:school_academic_year_id"`
	HomeroomTeacherID *uuid.UUID `gorm:"column:homeroom_teacher_id"`
	Name              string     `gorm:"column:name"`
	CodeRoom          string     `gorm:"column:code_room"`
	Capacity          int        `gorm:"column:capacity"`
	Floor             string     `gorm:"column:floor"`
	Building          string     `gorm:"column:building"`
	RoomType          string     `gorm:"column:room_type"`
	Status            string     `gorm:"column:status"`
	Facilities        string     `gorm:"column:facilities"`
	DeletedAt         *time.Time `gorm:"column:deleted_at"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at"`
}

func (classroomModel) TableName() string { return "school_classrooms" }

func classroomFromModel(m classroomModel) *domain.Classroom {
	return &domain.Classroom{
		ID: m.ID, SchoolID: m.SchoolID,
		ClassID: m.ClassID, SubClassID: m.SubClassID,
		AcademicYearID: m.AcademicYearID, HomeroomTeacherID: m.HomeroomTeacherID,
		Name: m.Name, CodeRoom: m.CodeRoom, Capacity: m.Capacity,
		Floor: m.Floor, Building: m.Building, RoomType: m.RoomType,
		Status: m.Status, Facilities: m.Facilities,
		DeletedAt: m.DeletedAt, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func (r *GormAcademicRepository) CreateClassroom(ctx context.Context, c *domain.Classroom) error {
	m := classroomModel{
		ID: c.ID, SchoolID: c.SchoolID,
		ClassID: c.ClassID, SubClassID: c.SubClassID,
		AcademicYearID: c.AcademicYearID, HomeroomTeacherID: c.HomeroomTeacherID,
		Name: c.Name, CodeRoom: c.CodeRoom, Capacity: c.Capacity,
		Floor: c.Floor, Building: c.Building, RoomType: c.RoomType,
		Status: c.Status, Facilities: c.Facilities,
		CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *GormAcademicRepository) FindClassroomByID(ctx context.Context, id uuid.UUID) (*domain.Classroom, error) {
	var m classroomModel
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.ErrNotFound, "classroom not found")
		}
		return nil, err
	}
	return classroomFromModel(m), nil
}

func (r *GormAcademicRepository) ListClassrooms(ctx context.Context, schoolID *uuid.UUID) ([]*domain.Classroom, error) {
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if schoolID != nil {
		q = q.Where("school_id = ?", *schoolID)
	}
	var rows []classroomModel
	if err := q.Order("name ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]*domain.Classroom, 0, len(rows))
	for _, m := range rows {
		list = append(list, classroomFromModel(m))
	}
	return list, nil
}

func (r *GormAcademicRepository) UpdateClassroom(ctx context.Context, c *domain.Classroom) error {
	return r.db.WithContext(ctx).Table("school_classrooms").Where("id = ?", c.ID).
		Updates(map[string]interface{}{
			"school_class_id":        c.ClassID,
			"school_sub_class_id":    c.SubClassID,
			"school_academic_year_id": c.AcademicYearID,
			"homeroom_teacher_id":    c.HomeroomTeacherID,
			"name":                   c.Name,
			"code_room":              c.CodeRoom,
			"capacity":               c.Capacity,
			"floor":                  c.Floor,
			"building":               c.Building,
			"room_type":              c.RoomType,
			"status":                 c.Status,
			"facilities":             c.Facilities,
			"updated_at":             c.UpdatedAt,
		}).Error
}

func (r *GormAcademicRepository) SoftDeleteClassroom(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Table("school_classrooms").Where("id = ?", id).
		Updates(map[string]interface{}{"deleted_at": now, "updated_at": now}).Error
}

// ── Schedules ─────────────────────────────────────────────────────────────────

type scheduleModel struct {
	ID           uuid.UUID  `gorm:"column:id;primaryKey"`
	SchoolID     uuid.UUID  `gorm:"column:school_id"`
	DayOfWeek    string     `gorm:"column:day_of_week"`
	PeriodNumber int        `gorm:"column:period_number"`
	Label        string     `gorm:"column:label"`
	StartTime    string     `gorm:"column:start_time"`
	EndTime      string     `gorm:"column:end_time"`
	IsBreak      bool       `gorm:"column:is_break"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

func (scheduleModel) TableName() string { return "school_schedules" }

func (r *GormAcademicRepository) CreateSchedule(ctx context.Context, s *domain.Schedule) error {
	m := scheduleModel{
		ID: s.ID, SchoolID: s.SchoolID,
		DayOfWeek: s.DayOfWeek, PeriodNumber: s.PeriodNumber, Label: s.Label,
		StartTime: s.StartTime, EndTime: s.EndTime, IsBreak: s.IsBreak,
		CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *GormAcademicRepository) FindScheduleByID(ctx context.Context, id uuid.UUID) (*domain.Schedule, error) {
	var m scheduleModel
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.ErrNotFound, "schedule not found")
		}
		return nil, err
	}
	return scheduleFromModel(m), nil
}

func scheduleFromModel(m scheduleModel) *domain.Schedule {
	return &domain.Schedule{
		ID: m.ID, SchoolID: m.SchoolID,
		DayOfWeek: m.DayOfWeek, PeriodNumber: m.PeriodNumber, Label: m.Label,
		StartTime: m.StartTime, EndTime: m.EndTime, IsBreak: m.IsBreak,
		DeletedAt: m.DeletedAt, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func (r *GormAcademicRepository) ListSchedules(ctx context.Context, schoolID *uuid.UUID, dayOfWeek *string) ([]*domain.Schedule, error) {
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if schoolID != nil {
		q = q.Where("school_id = ?", *schoolID)
	}
	if dayOfWeek != nil {
		q = q.Where("day_of_week = ?", *dayOfWeek)
	}
	var rows []scheduleModel
	if err := q.Order("CASE day_of_week WHEN 'monday' THEN 1 WHEN 'tuesday' THEN 2 WHEN 'wednesday' THEN 3 WHEN 'thursday' THEN 4 WHEN 'friday' THEN 5 WHEN 'saturday' THEN 6 WHEN 'sunday' THEN 7 END, period_number ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]*domain.Schedule, 0, len(rows))
	for _, m := range rows {
		list = append(list, scheduleFromModel(m))
	}
	return list, nil
}

func (r *GormAcademicRepository) UpdateSchedule(ctx context.Context, s *domain.Schedule) error {
	return r.db.WithContext(ctx).Table("school_schedules").Where("id = ?", s.ID).
		Updates(map[string]interface{}{
			"day_of_week":   s.DayOfWeek,
			"period_number": s.PeriodNumber,
			"label":         s.Label,
			"start_time":    s.StartTime,
			"end_time":      s.EndTime,
			"is_break":      s.IsBreak,
			"updated_at":    s.UpdatedAt,
		}).Error
}

func (r *GormAcademicRepository) SoftDeleteSchedule(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Table("school_schedules").Where("id = ?", id).
		Updates(map[string]interface{}{"deleted_at": now, "updated_at": now}).Error
}
