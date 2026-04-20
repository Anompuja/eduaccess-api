package infrastructure

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/parent/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

type parentWithUser struct {
	ID             uuid.UUID  `gorm:"column:id"`
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
	UserName       string     `gorm:"column:user_name"`
	UserEmail      string     `gorm:"column:user_email"`
	Username       string     `gorm:"column:username"`
	Avatar         string     `gorm:"column:avatar"`
}

// GormParentRepository implements domain.ParentRepository.
type GormParentRepository struct {
	db *gorm.DB
}

func NewGormParentRepository(db *gorm.DB) *GormParentRepository {
	return &GormParentRepository{db: db}
}

func (r *GormParentRepository) CreateParentProfile(ctx context.Context, p *domain.ParentProfile) error {
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

func (r *GormParentRepository) FindParentByID(ctx context.Context, id uuid.UUID) (*domain.ParentProfile, error) {
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

func (r *GormParentRepository) ListParents(ctx context.Context, f domain.ParentFilter) ([]*domain.ParentProfile, int64, error) {
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
