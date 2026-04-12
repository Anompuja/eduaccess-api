package infrastructure

import (
	"context"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/student/domain"
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

func (r *GormAcademicRepository) ListLevels(ctx context.Context, schoolID uuid.UUID) ([]*domain.EducationLevel, error) {
	var rows []educationLevelModel
	if err := r.db.WithContext(ctx).Where("school_id = ? AND deleted_at IS NULL", schoolID).Order("name ASC").Find(&rows).Error; err != nil {
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

func (r *GormAcademicRepository) ListClasses(ctx context.Context, schoolID uuid.UUID, levelID *uuid.UUID) ([]*domain.Class, error) {
	q := r.db.WithContext(ctx).Where("school_id = ? AND deleted_at IS NULL", schoolID)
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

func (r *GormAcademicRepository) ListSubClasses(ctx context.Context, schoolID uuid.UUID, classID *uuid.UUID) ([]*domain.SubClass, error) {
	q := r.db.WithContext(ctx).Where("school_id = ? AND deleted_at IS NULL", schoolID)
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
