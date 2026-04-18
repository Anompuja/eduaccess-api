package infrastructure

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/student/domain"
	"github.com/google/uuid"
)

type studentProfileModel struct {
	ID                uuid.UUID  `gorm:"column:id;primaryKey"`
	UserID            uuid.UUID  `gorm:"column:user_id"`
	SchoolID          uuid.UUID  `gorm:"column:school_id"`
	NIS               string     `gorm:"column:nis"`
	NISN              string     `gorm:"column:nisn"`
	PhoneNumber       string     `gorm:"column:phone_number"`
	Address           string     `gorm:"column:address"`
	Gender            string     `gorm:"column:gender"`
	Religion          string     `gorm:"column:religion"`
	BirthPlace        string     `gorm:"column:birth_place"`
	BirthDate         *time.Time `gorm:"column:birth_date"`
	TahunMasuk        string     `gorm:"column:tahun_masuk"`
	JalurMasukSekolah string     `gorm:"column:jalur_masuk_sekolah"`
	EducationLevelID  *uuid.UUID `gorm:"column:school_education_level_id"`
	ClassID           *uuid.UUID `gorm:"column:school_class_id"`
	SubClassID        *uuid.UUID `gorm:"column:school_sub_class_id"`
	DeletedAt         *time.Time `gorm:"column:deleted_at"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at"`
}

func (studentProfileModel) TableName() string { return "student_profiles" }

// studentWithUser is the scan target for student JOIN queries.
type studentWithUser struct {
	ID                uuid.UUID  `gorm:"column:id"`
	UserID            uuid.UUID  `gorm:"column:user_id"`
	SchoolID          uuid.UUID  `gorm:"column:school_id"`
	NIS               string     `gorm:"column:nis"`
	NISN              string     `gorm:"column:nisn"`
	PhoneNumber       string     `gorm:"column:phone_number"`
	Address           string     `gorm:"column:address"`
	Gender            string     `gorm:"column:gender"`
	Religion          string     `gorm:"column:religion"`
	BirthPlace        string     `gorm:"column:birth_place"`
	BirthDate         *time.Time `gorm:"column:birth_date"`
	TahunMasuk        string     `gorm:"column:tahun_masuk"`
	JalurMasukSekolah string     `gorm:"column:jalur_masuk_sekolah"`
	EducationLevelID  *uuid.UUID `gorm:"column:school_education_level_id"`
	ClassID           *uuid.UUID `gorm:"column:school_class_id"`
	SubClassID        *uuid.UUID `gorm:"column:school_sub_class_id"`
	DeletedAt         *time.Time `gorm:"column:deleted_at"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at"`
	UserName          string     `gorm:"column:user_name"`
	UserEmail         string     `gorm:"column:user_email"`
	Username          string     `gorm:"column:username"`
	Avatar            string     `gorm:"column:avatar"`
}

func (r *GormStudentRepository) CreateStudentProfile(ctx context.Context, p *domain.StudentProfile) error {
	m := studentProfileModel{
		ID:                p.ID,
		UserID:            p.UserID,
		SchoolID:          p.SchoolID,
		NIS:               p.NIS,
		NISN:              p.NISN,
		PhoneNumber:       p.PhoneNumber,
		Address:           p.Address,
		Gender:            p.Gender,
		Religion:          p.Religion,
		BirthPlace:        p.BirthPlace,
		BirthDate:         p.BirthDate,
		TahunMasuk:        p.TahunMasuk,
		JalurMasukSekolah: p.JalurMasukSekolah,
		EducationLevelID:  p.EducationLevelID,
		ClassID:           p.ClassID,
		SubClassID:        p.SubClassID,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *GormStudentRepository) FindStudentByID(ctx context.Context, id uuid.UUID) (*domain.StudentProfile, error) {
	var row studentWithUser
	sql := `
SELECT sp.*, u.name AS user_name, u.email AS user_email, u.username, u.avatar
FROM student_profiles sp
JOIN users u ON u.id = sp.user_id
WHERE sp.id = ? AND sp.deleted_at IS NULL
LIMIT 1`
	if err := r.db.WithContext(ctx).Raw(sql, id).Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.ID == uuid.Nil {
		return nil, apperror.New(apperror.ErrNotFound, "student not found")
	}
	return toStudentDomain(row), nil
}

func (r *GormStudentRepository) FindStudentByUserID(ctx context.Context, userID uuid.UUID) (*domain.StudentProfile, error) {
	var row studentWithUser
	sql := `
SELECT sp.*, u.name AS user_name, u.email AS user_email, u.username, u.avatar
FROM student_profiles sp
JOIN users u ON u.id = sp.user_id
WHERE sp.user_id = ? AND sp.deleted_at IS NULL
LIMIT 1`
	if err := r.db.WithContext(ctx).Raw(sql, userID).Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.ID == uuid.Nil {
		return nil, apperror.New(apperror.ErrNotFound, "student profile not found")
	}
	return toStudentDomain(row), nil
}

func (r *GormStudentRepository) ListStudents(ctx context.Context, f domain.StudentFilter) ([]*domain.StudentProfile, int64, error) {
	base := `
FROM student_profiles sp
JOIN users u ON u.id = sp.user_id
WHERE sp.deleted_at IS NULL`

	args := []interface{}{}
	conditions := []string{}

	if f.SchoolID != nil {
		conditions = append(conditions, "sp.school_id = ?")
		args = append(args, *f.SchoolID)
	}
	if f.EducationLevelID != nil {
		conditions = append(conditions, "sp.school_education_level_id = ?")
		args = append(args, *f.EducationLevelID)
	}
	if f.ClassID != nil {
		conditions = append(conditions, "sp.school_class_id = ?")
		args = append(args, *f.ClassID)
	}
	if f.SubClassID != nil {
		conditions = append(conditions, "sp.school_sub_class_id = ?")
		args = append(args, *f.SubClassID)
	}
	if f.Search != "" {
		conditions = append(conditions, "(u.name ILIKE ? OR u.email ILIKE ? OR u.username ILIKE ? OR sp.nis ILIKE ? OR sp.nisn ILIKE ?)")
		like := "%" + f.Search + "%"
		args = append(args, like, like, like, like, like)
	}

	where := ""
	if len(conditions) > 0 {
		where = " AND " + strings.Join(conditions, " AND ")
	}

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(DISTINCT sp.id) %s%s", base, where)
	if err := r.db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	dataSQL := fmt.Sprintf(`
SELECT sp.*, u.name AS user_name, u.email AS user_email, u.username, u.avatar
%s%s
ORDER BY u.name ASC
LIMIT ? OFFSET ?`, base, where)

	queryArgs := append(args, f.Limit, f.Offset)
	var rows []studentWithUser
	if err := r.db.WithContext(ctx).Raw(dataSQL, queryArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	students := make([]*domain.StudentProfile, 0, len(rows))
	for _, row := range rows {
		students = append(students, toStudentDomain(row))
	}
	return students, total, nil
}

func (r *GormStudentRepository) UpdateStudentProfile(ctx context.Context, p *domain.StudentProfile) error {
	return r.db.WithContext(ctx).
		Table("student_profiles").
		Where("id = ?", p.ID).
		Updates(map[string]interface{}{
			"nis":                       p.NIS,
			"nisn":                      p.NISN,
			"phone_number":              p.PhoneNumber,
			"address":                   p.Address,
			"gender":                    p.Gender,
			"religion":                  p.Religion,
			"birth_place":               p.BirthPlace,
			"birth_date":                p.BirthDate,
			"tahun_masuk":               p.TahunMasuk,
			"jalur_masuk_sekolah":       p.JalurMasukSekolah,
			"school_education_level_id": p.EducationLevelID,
			"school_class_id":           p.ClassID,
			"school_sub_class_id":       p.SubClassID,
			"updated_at":                p.UpdatedAt,
		}).Error
}

func (r *GormStudentRepository) SoftDeleteStudent(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Table("student_profiles").
		Where("id = ?", id).
		Updates(map[string]interface{}{"deleted_at": now, "updated_at": now}).Error
}

func toStudentDomain(row studentWithUser) *domain.StudentProfile {
	return &domain.StudentProfile{
		ID:                row.ID,
		UserID:            row.UserID,
		SchoolID:          row.SchoolID,
		NIS:               row.NIS,
		NISN:              row.NISN,
		PhoneNumber:       row.PhoneNumber,
		Address:           row.Address,
		Gender:            row.Gender,
		Religion:          row.Religion,
		BirthPlace:        row.BirthPlace,
		BirthDate:         row.BirthDate,
		TahunMasuk:        row.TahunMasuk,
		JalurMasukSekolah: row.JalurMasukSekolah,
		EducationLevelID:  row.EducationLevelID,
		ClassID:           row.ClassID,
		SubClassID:        row.SubClassID,
		DeletedAt:         row.DeletedAt,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
		Name:              row.UserName,
		Email:             row.UserEmail,
		Username:          row.Username,
		Avatar:            row.Avatar,
	}
}
