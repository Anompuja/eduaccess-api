package infrastructure

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/student/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ── GORM models ───────────────────────────────────────────────────────────────

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

type parentProfileModel struct {
	ID             uuid.UUID  `gorm:"column:id;primaryKey"`
	UserID         uuid.UUID  `gorm:"column:user_id"`
	SchoolID       uuid.UUID  `gorm:"column:school_id"`
	FatherName     string     `gorm:"column:father_name"`
	MotherName     string     `gorm:"column:mother_name"`
	FatherReligion string     `gorm:"column:father_religion"`
	MotherReligion string     `gorm:"column:mother_religion"`
	PhoneNumber    string     `gorm:"column:phone_number"`
	Address        string     `gorm:"column:address"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

func (parentProfileModel) TableName() string { return "parent_profiles" }

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

// studentWithUser is the scan target for the JOIN query.
type studentWithUser struct {
	studentProfileModel
	UserName  string `gorm:"column:user_name"`
	UserEmail string `gorm:"column:user_email"`
	Username  string `gorm:"column:username"`
	Avatar    string `gorm:"column:avatar"`
}

// parentWithUser is the scan target for the parent JOIN query.
type parentWithUser struct {
	parentProfileModel
	UserName  string `gorm:"column:user_name"`
	UserEmail string `gorm:"column:user_email"`
	Username  string `gorm:"column:username"`
	Avatar    string `gorm:"column:avatar"`
}

// ── Repository ────────────────────────────────────────────────────────────────

// GormStudentRepository implements domain.StudentRepository.
type GormStudentRepository struct {
	db *gorm.DB
}

func NewGormStudentRepository(db *gorm.DB) *GormStudentRepository {
	return &GormStudentRepository{db: db}
}

// ── Student profiles ──────────────────────────────────────────────────────────

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

// ── Parent profiles ───────────────────────────────────────────────────────────

func (r *GormStudentRepository) CreateParentProfile(ctx context.Context, p *domain.ParentProfile) error {
	m := parentProfileModel{
		ID:             p.ID,
		UserID:         p.UserID,
		SchoolID:       p.SchoolID,
		FatherName:     p.FatherName,
		MotherName:     p.MotherName,
		FatherReligion: p.FatherReligion,
		MotherReligion: p.MotherReligion,
		PhoneNumber:    p.PhoneNumber,
		Address:        p.Address,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *GormStudentRepository) FindParentByID(ctx context.Context, id uuid.UUID) (*domain.ParentProfile, error) {
	var row parentWithUser
	sql := `
SELECT pp.*, u.name AS user_name, u.email AS user_email, u.username, u.avatar
FROM parent_profiles pp
JOIN users u ON u.id = pp.user_id
WHERE pp.id = ? AND pp.deleted_at IS NULL
LIMIT 1`
	if err := r.db.WithContext(ctx).Raw(sql, id).Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.ID == uuid.Nil {
		return nil, apperror.New(apperror.ErrNotFound, "parent not found")
	}
	return toParentDomain(row), nil
}

func (r *GormStudentRepository) FindParentByUserID(ctx context.Context, userID uuid.UUID) (*domain.ParentProfile, error) {
	var row parentWithUser
	sql := `
SELECT pp.*, u.name AS user_name, u.email AS user_email, u.username, u.avatar
FROM parent_profiles pp
JOIN users u ON u.id = pp.user_id
WHERE pp.user_id = ? AND pp.deleted_at IS NULL
LIMIT 1`
	if err := r.db.WithContext(ctx).Raw(sql, userID).Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.ID == uuid.Nil {
		return nil, apperror.New(apperror.ErrNotFound, "parent profile not found")
	}
	return toParentDomain(row), nil
}

func (r *GormStudentRepository) ListParents(ctx context.Context, f domain.ParentFilter) ([]*domain.ParentProfile, int64, error) {
	base := `
FROM parent_profiles pp
JOIN users u ON u.id = pp.user_id
WHERE pp.deleted_at IS NULL`

	args := []interface{}{}
	conditions := []string{}

	if f.SchoolID != nil {
		conditions = append(conditions, "pp.school_id = ?")
		args = append(args, *f.SchoolID)
	}
	if f.Search != "" {
		conditions = append(conditions, "(u.name ILIKE ? OR u.email ILIKE ? OR u.username ILIKE ?)")
		like := "%" + f.Search + "%"
		args = append(args, like, like, like)
	}

	where := ""
	if len(conditions) > 0 {
		where = " AND " + strings.Join(conditions, " AND ")
	}

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(DISTINCT pp.id) %s%s", base, where)
	if err := r.db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	dataSQL := fmt.Sprintf(`
SELECT pp.*, u.name AS user_name, u.email AS user_email, u.username, u.avatar
%s%s
ORDER BY u.name ASC
LIMIT ? OFFSET ?`, base, where)

	queryArgs := append(args, f.Limit, f.Offset)
	var rows []parentWithUser
	if err := r.db.WithContext(ctx).Raw(dataSQL, queryArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	parents := make([]*domain.ParentProfile, 0, len(rows))
	for _, row := range rows {
		parents = append(parents, toParentDomain(row))
	}
	return parents, total, nil
}

func (r *GormStudentRepository) UpdateParentProfile(ctx context.Context, p *domain.ParentProfile) error {
	return r.db.WithContext(ctx).
		Table("parent_profiles").
		Where("id = ?", p.ID).
		Updates(map[string]interface{}{
			"father_name":     p.FatherName,
			"mother_name":     p.MotherName,
			"father_religion": p.FatherReligion,
			"mother_religion": p.MotherReligion,
			"phone_number":    p.PhoneNumber,
			"address":         p.Address,
			"updated_at":      p.UpdatedAt,
		}).Error
}

func (r *GormStudentRepository) SoftDeleteParent(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Table("parent_profiles").
		Where("id = ?", id).
		Updates(map[string]interface{}{"deleted_at": now, "updated_at": now}).Error
}

// ── Parent links ──────────────────────────────────────────────────────────────

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
		parentLinkModel
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

// ── helpers ───────────────────────────────────────────────────────────────────

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

func toParentDomain(row parentWithUser) *domain.ParentProfile {
	return &domain.ParentProfile{
		ID:             row.ID,
		UserID:         row.UserID,
		SchoolID:       row.SchoolID,
		FatherName:     row.FatherName,
		MotherName:     row.MotherName,
		FatherReligion: row.FatherReligion,
		MotherReligion: row.MotherReligion,
		PhoneNumber:    row.PhoneNumber,
		Address:        row.Address,
		DeletedAt:      row.DeletedAt,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
		Name:           row.UserName,
		Email:          row.UserEmail,
		Username:       row.Username,
		Avatar:         row.Avatar,
	}
}
