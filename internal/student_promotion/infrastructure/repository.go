package infrastructure

import (
	"context"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/student_promotion/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

const promotionViewSelect = `
	p.id,
	p.student_id,
	u.name   AS student_name,
	sp.nis   AS nis,
	p.from_classroom_id,
	fcr.name AS from_classroom_name,
	p.to_classroom_id,
	tcr.name AS to_classroom_name,
	p.school_academic_year_id,
	ay.name  AS academic_year_name,
	p.promotion_date,
	p.status,
	p.notes`

func (r *GormRepository) ListPromotions(ctx context.Context, filter domain.PromotionFilter) ([]domain.PromotionView, error) {
	q := r.db.WithContext(ctx).
		Table("student_promotions AS p").
		Select(promotionViewSelect).
		Joins("LEFT JOIN users u ON u.id = p.student_id").
		Joins("LEFT JOIN student_profiles sp ON sp.user_id = p.student_id AND sp.school_id = p.school_id AND sp.deleted_at IS NULL").
		Joins("LEFT JOIN school_classrooms fcr ON fcr.id = p.from_classroom_id").
		Joins("LEFT JOIN school_classrooms tcr ON tcr.id = p.to_classroom_id").
		Joins("LEFT JOIN school_academic_years ay ON ay.id = p.school_academic_year_id").
		Where("p.deleted_at IS NULL")
	if filter.SchoolID != nil {
		q = q.Where("p.school_id = ?", *filter.SchoolID)
	}
	if filter.StudentID != nil {
		q = q.Where("p.student_id = ?", *filter.StudentID)
	}
	if filter.AcademicYearID != nil {
		q = q.Where("p.school_academic_year_id = ?", *filter.AcademicYearID)
	}
	var rows []domain.PromotionView
	if err := q.Order("p.promotion_date DESC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *GormRepository) FindClassroomTarget(ctx context.Context, classroomID uuid.UUID) (*domain.ClassroomTarget, error) {
	var row struct {
		ID                uuid.UUID  `gorm:"column:id"`
		SchoolID          uuid.UUID  `gorm:"column:school_id"`
		ClassID           *uuid.UUID `gorm:"column:school_class_id"`
		SubClassID        *uuid.UUID `gorm:"column:school_sub_class_id"`
		AcademicYearID    *uuid.UUID `gorm:"column:school_academic_year_id"`
		HomeroomTeacherID *uuid.UUID `gorm:"column:homeroom_teacher_id"`
		EducationLevelID  *uuid.UUID `gorm:"column:school_education_level_id"`
	}
	err := r.db.WithContext(ctx).
		Table("school_classrooms AS cr").
		Select("cr.id, cr.school_id, cr.school_class_id, cr.school_sub_class_id, cr.school_academic_year_id, cr.homeroom_teacher_id, sc.school_education_level_id").
		Joins("LEFT JOIN school_classes sc ON sc.id = cr.school_class_id").
		Where("cr.id = ? AND cr.deleted_at IS NULL", classroomID).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == uuid.Nil {
		return nil, apperror.New(apperror.ErrNotFound, "target classroom not found")
	}
	return &domain.ClassroomTarget{
		ClassroomID:       row.ID,
		SchoolID:          row.SchoolID,
		ClassID:           row.ClassID,
		SubClassID:        row.SubClassID,
		AcademicYearID:    row.AcademicYearID,
		HomeroomTeacherID: row.HomeroomTeacherID,
		EducationLevelID:  row.EducationLevelID,
	}, nil
}

func (r *GormRepository) findActiveStudyClassroom(ctx context.Context, tx *gorm.DB, schoolID, studentID uuid.UUID) (uuid.UUID, error) {
	var row struct {
		ClassroomID uuid.UUID `gorm:"column:school_classroom_id"`
	}
	err := tx.WithContext(ctx).
		Table("student_studies").
		Select("school_classroom_id").
		Where("school_id = ? AND student_id = ? AND status = 'active' AND deleted_at IS NULL", schoolID, studentID).
		Order("enrollment_date DESC").
		Limit(1).
		Scan(&row).Error
	if err != nil {
		return uuid.Nil, err
	}
	if row.ClassroomID == uuid.Nil {
		return uuid.Nil, apperror.New(apperror.ErrBadRequest, "student has no active enrollment to promote from")
	}
	return row.ClassroomID, nil
}

// PromoteStudent records the audit row, closes the current enrollment, opens a
// new active enrollment in the target classroom, and updates the student's
// current class pointer — all in one transaction.
func (r *GormRepository) PromoteStudent(ctx context.Context, in domain.PromotionInput) error {
	if in.Target.AcademicYearID == nil {
		return apperror.New(apperror.ErrBadRequest, "target classroom has no academic year assigned")
	}
	now := time.Now()

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		fromClassroom, err := r.findActiveStudyClassroom(ctx, tx, in.SchoolID, in.StudentID)
		if err != nil {
			return err
		}

		// 1. Audit row
		promotion := map[string]any{
			"id":                      uuid.New(),
			"school_id":               in.SchoolID,
			"student_id":              in.StudentID,
			"from_classroom_id":       fromClassroom,
			"to_classroom_id":         in.Target.ClassroomID,
			"school_academic_year_id": *in.Target.AcademicYearID,
			"promotion_date":          in.PromotionDate,
			"status":                  in.Status,
			"notes":                   in.Notes,
			"created_at":              now,
			"updated_at":              now,
		}
		if err := tx.Table("student_promotions").Create(promotion).Error; err != nil {
			return err
		}

		// 2. Close the current active enrollment(s)
		if err := tx.Table("student_studies").
			Where("school_id = ? AND student_id = ? AND status = 'active' AND deleted_at IS NULL", in.SchoolID, in.StudentID).
			Updates(map[string]any{"status": "inactive", "updated_at": now}).Error; err != nil {
			return err
		}

		// 3. Open the new active enrollment
		study := map[string]any{
			"id":                      uuid.New(),
			"school_id":               in.SchoolID,
			"student_id":              in.StudentID,
			"school_classroom_id":     in.Target.ClassroomID,
			"school_academic_year_id": *in.Target.AcademicYearID,
			"school_class_id":         in.Target.ClassID,
			"school_sub_class_id":     in.Target.SubClassID,
			"homeroom_teacher_id":     in.Target.HomeroomTeacherID,
			"status":                  "active",
			"enrollment_date":         in.PromotionDate,
			"created_at":              now,
			"updated_at":              now,
		}
		if err := tx.Table("student_studies").Create(study).Error; err != nil {
			return err
		}

		// 4. Update the student's denormalised current class pointer
		if err := tx.Table("student_profiles").
			Where("user_id = ? AND school_id = ? AND deleted_at IS NULL", in.StudentID, in.SchoolID).
			Updates(map[string]any{
				"school_class_id":           in.Target.ClassID,
				"school_sub_class_id":       in.Target.SubClassID,
				"school_education_level_id": in.Target.EducationLevelID,
				"updated_at":                now,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}
