package infrastructure

import (
	"context"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/student/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type parentLinkModel struct {
	ID           uuid.UUID `gorm:"column:id;primaryKey"`
	SchoolID     uuid.UUID `gorm:"column:school_id"`
	StudentID    uuid.UUID `gorm:"column:student_id"`
	ParentID     uuid.UUID `gorm:"column:parent_id"`
	Relationship string    `gorm:"column:relationship"`
	IsPrimary    bool      `gorm:"column:is_primary"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (parentLinkModel) TableName() string { return "student_parent_links" }

// GormStudentRepository implements domain.StudentRepository.
type GormStudentRepository struct {
	db *gorm.DB
}

func NewGormStudentRepository(db *gorm.DB) *GormStudentRepository {
	return &GormStudentRepository{db: db}
}

func (r *GormStudentRepository) LinkParent(ctx context.Context, link *domain.ParentLink) error {
	m := parentLinkModel{
		ID:           link.ID,
		SchoolID:     link.SchoolID,
		StudentID:    link.StudentID,
		ParentID:     link.ParentID,
		Relationship: link.Relationship,
		IsPrimary:    link.IsPrimary,
		CreatedAt:    link.CreatedAt,
		UpdatedAt:    link.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *GormStudentRepository) UnlinkParent(ctx context.Context, studentID, parentID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Table("student_parent_links").
		Where("student_id = ? AND parent_id = ?", studentID, parentID).
		Delete(nil).Error
}

func (r *GormStudentRepository) ListParentLinks(ctx context.Context, studentID uuid.UUID) ([]*domain.ParentLink, error) {
	sql := `
SELECT spl.*, pp.id AS pp_id, pp.school_id AS pp_school_id,
       pp.user_id AS pp_user_id, pp.father_name, pp.mother_name,
       pp.phone_number AS pp_phone, pp.address AS pp_address,
       pp.father_religion, pp.mother_religion,
       u.name AS user_name, u.email AS user_email, u.username, u.avatar
FROM student_parent_links spl
JOIN parent_profiles pp ON pp.id = spl.parent_id AND pp.deleted_at IS NULL
JOIN users u ON u.id = pp.user_id
WHERE spl.student_id = ?`

	type linkRow struct {
		ID             uuid.UUID `gorm:"column:id"`
		SchoolID       uuid.UUID `gorm:"column:school_id"`
		StudentID      uuid.UUID `gorm:"column:student_id"`
		ParentID       uuid.UUID `gorm:"column:parent_id"`
		Relationship   string    `gorm:"column:relationship"`
		IsPrimary      bool      `gorm:"column:is_primary"`
		CreatedAt      time.Time `gorm:"column:created_at"`
		UpdatedAt      time.Time `gorm:"column:updated_at"`
		PpID           uuid.UUID `gorm:"column:pp_id"`
		PpSchoolID     uuid.UUID `gorm:"column:pp_school_id"`
		PpUserID       uuid.UUID `gorm:"column:pp_user_id"`
		FatherName     string    `gorm:"column:father_name"`
		MotherName     string    `gorm:"column:mother_name"`
		PpPhone        string    `gorm:"column:pp_phone"`
		PpAddress      string    `gorm:"column:pp_address"`
		FatherReligion string    `gorm:"column:father_religion"`
		MotherReligion string    `gorm:"column:mother_religion"`
		UserName       string    `gorm:"column:user_name"`
		UserEmail      string    `gorm:"column:user_email"`
		Username       string    `gorm:"column:username"`
		Avatar         string    `gorm:"column:avatar"`
	}

	var rows []linkRow
	if err := r.db.WithContext(ctx).Raw(sql, studentID).Scan(&rows).Error; err != nil {
		return nil, err
	}

	links := make([]*domain.ParentLink, 0, len(rows))
	for _, row := range rows {
		link := &domain.ParentLink{
			ID:           row.ID,
			SchoolID:     row.SchoolID,
			StudentID:    row.StudentID,
			ParentID:     row.ParentID,
			Relationship: row.Relationship,
			IsPrimary:    row.IsPrimary,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
			Parent: &domain.ParentProfile{
				ID:             row.PpID,
				UserID:         row.PpUserID,
				SchoolID:       row.PpSchoolID,
				FatherName:     row.FatherName,
				MotherName:     row.MotherName,
				FatherReligion: row.FatherReligion,
				MotherReligion: row.MotherReligion,
				PhoneNumber:    row.PpPhone,
				Address:        row.PpAddress,
				Name:           row.UserName,
				Email:          row.UserEmail,
				Username:       row.Username,
				Avatar:         row.Avatar,
			},
		}
		links = append(links, link)
	}

	return links, nil
}
