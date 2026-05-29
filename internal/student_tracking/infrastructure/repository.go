package infrastructure

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/student_tracking/domain"
	"gorm.io/gorm"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

const studyViewSelect = `
	ss.id,
	ss.student_id,
	u.name           AS student_name,
	sp.nis           AS nis,
	ss.school_classroom_id,
	cr.name          AS classroom_name,
	ss.school_class_id,
	sc.name          AS class_name,
	ssc.name         AS sub_class_name,
	ss.school_academic_year_id,
	ay.name          AS academic_year_name,
	ss.status,
	ss.enrollment_date`

func (r *GormRepository) ListStudies(ctx context.Context, filter domain.StudyFilter) ([]domain.StudyView, error) {
	q := r.db.WithContext(ctx).
		Table("student_studies AS ss").
		Select(studyViewSelect).
		Joins("LEFT JOIN users u ON u.id = ss.student_id").
		Joins("LEFT JOIN student_profiles sp ON sp.user_id = ss.student_id AND sp.school_id = ss.school_id AND sp.deleted_at IS NULL").
		Joins("LEFT JOIN school_classrooms cr ON cr.id = ss.school_classroom_id").
		Joins("LEFT JOIN school_classes sc ON sc.id = ss.school_class_id").
		Joins("LEFT JOIN school_sub_classes ssc ON ssc.id = ss.school_sub_class_id").
		Joins("LEFT JOIN school_academic_years ay ON ay.id = ss.school_academic_year_id").
		Where("ss.deleted_at IS NULL")

	if filter.SchoolID != nil {
		q = q.Where("ss.school_id = ?", *filter.SchoolID)
	}
	if filter.ClassroomID != nil {
		q = q.Where("ss.school_classroom_id = ?", *filter.ClassroomID)
	}
	if filter.AcademicYearID != nil {
		q = q.Where("ss.school_academic_year_id = ?", *filter.AcademicYearID)
	}
	if filter.ClassID != nil {
		q = q.Where("ss.school_class_id = ?", *filter.ClassID)
	}
	if filter.StudentID != nil {
		q = q.Where("ss.student_id = ?", *filter.StudentID)
	}
	if filter.Status != nil {
		q = q.Where("ss.status = ?", *filter.Status)
	}

	var rows []domain.StudyView
	if err := q.Order("ss.enrollment_date DESC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
