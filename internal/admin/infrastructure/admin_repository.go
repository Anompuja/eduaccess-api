package infrastructure

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/admin/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type adminProfileModel struct {
	ID           uuid.UUID  `gorm:"column:id;primaryKey"`
	UserID       uuid.UUID  `gorm:"column:user_id"`
	SchoolID     uuid.UUID  `gorm:"column:school_id"`
	PhoneNumber  string     `gorm:"column:phone_number"`
	Address      string     `gorm:"column:address"`
	Gender       string     `gorm:"column:gender"`
	Religion     string     `gorm:"column:religion"`
	BirthPlace   string     `gorm:"column:birth_place"`
	BirthDate    *time.Time `gorm:"column:birth_date"`
	NIK          string     `gorm:"column:nik"`
	KTPImagePath string     `gorm:"column:ktp_image_path"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

func (adminProfileModel) TableName() string { return "admin_profiles" }

type adminWithUser struct {
	ID           uuid.UUID  `gorm:"column:id"`
	UserID       uuid.UUID  `gorm:"column:user_id"`
	SchoolID     uuid.UUID  `gorm:"column:school_id"`
	PhoneNumber  string     `gorm:"column:phone_number"`
	Address      string     `gorm:"column:address"`
	Gender       string     `gorm:"column:gender"`
	Religion     string     `gorm:"column:religion"`
	BirthPlace   string     `gorm:"column:birth_place"`
	BirthDate    *time.Time `gorm:"column:birth_date"`
	NIK          string     `gorm:"column:nik"`
	KTPImagePath string     `gorm:"column:ktp_image_path"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
	Name         string     `gorm:"column:user_name"`
	Email        string     `gorm:"column:user_email"`
	Username     string     `gorm:"column:username"`
	Avatar       string     `gorm:"column:avatar"`
}

// GormAdminRepository implements domain.AdminRepository.
type GormAdminRepository struct {
	db *gorm.DB
}

func NewGormAdminRepository(db *gorm.DB) *GormAdminRepository {
	return &GormAdminRepository{db: db}
}

func (r *GormAdminRepository) CreateAdminProfile(ctx context.Context, p *domain.AdminProfile) error {
	m := adminProfileModel{
		ID:           p.ID,
		UserID:       p.UserID,
		SchoolID:     p.SchoolID,
		PhoneNumber:  p.PhoneNumber,
		Address:      p.Address,
		Gender:       p.Gender,
		Religion:     p.Religion,
		BirthPlace:   p.BirthPlace,
		BirthDate:    p.BirthDate,
		NIK:          p.NIK,
		KTPImagePath: p.KTPImagePath,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *GormAdminRepository) FindAdminByID(ctx context.Context, id uuid.UUID) (*domain.AdminProfile, error) {
	var row adminWithUser
	sql := `
SELECT ap.*, u.name AS user_name, u.email AS user_email, u.username, u.avatar
FROM admin_profiles ap
JOIN users u ON u.id = ap.user_id
WHERE ap.id = ? AND ap.deleted_at IS NULL
LIMIT 1`
	if err := r.db.WithContext(ctx).Raw(sql, id).Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.ID == uuid.Nil {
		return nil, apperror.New(apperror.ErrNotFound, "admin not found")
	}

	return &domain.AdminProfile{
		ID:           row.ID,
		UserID:       row.UserID,
		SchoolID:     row.SchoolID,
		PhoneNumber:  row.PhoneNumber,
		Address:      row.Address,
		Gender:       row.Gender,
		Religion:     row.Religion,
		BirthPlace:   row.BirthPlace,
		BirthDate:    row.BirthDate,
		NIK:          row.NIK,
		KTPImagePath: row.KTPImagePath,
		DeletedAt:    row.DeletedAt,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		Name:         row.Name,
		Email:        row.Email,
		Username:     row.Username,
		Avatar:       row.Avatar,
	}, nil
}

func (r *GormAdminRepository) UpdateAdminProfile(ctx context.Context, p *domain.AdminProfile) error {
	return r.db.WithContext(ctx).
		Table("admin_profiles").
		Where("id = ?", p.ID).
		Updates(map[string]interface{}{
			"phone_number":   p.PhoneNumber,
			"address":        p.Address,
			"gender":         p.Gender,
			"religion":       p.Religion,
			"birth_place":    p.BirthPlace,
			"birth_date":     p.BirthDate,
			"nik":            p.NIK,
			"ktp_image_path": p.KTPImagePath,
			"updated_at":     p.UpdatedAt,
		}).Error
}

func (r *GormAdminRepository) SoftDeleteAdmin(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Table("admin_profiles").
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted_at": now,
			"updated_at": now,
		}).Error
}

func (r *GormAdminRepository) ListAdmins(ctx context.Context, f domain.AdminFilter) ([]*domain.AdminProfile, int64, error) {
	base := `
FROM admin_profiles ap
JOIN users u ON u.id = ap.user_id
WHERE ap.deleted_at IS NULL`

	args := []interface{}{}
	conditions := []string{}

	if f.SchoolID != nil {
		conditions = append(conditions, "ap.school_id = ?")
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
	countSQL := fmt.Sprintf("SELECT COUNT(DISTINCT ap.id) %s%s", base, where)
	if err := r.db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	dataSQL := fmt.Sprintf(`
SELECT ap.*, u.name AS user_name, u.email AS user_email, u.username, u.avatar
%s%s
ORDER BY u.name ASC
LIMIT ? OFFSET ?`, base, where)

	queryArgs := append(args, f.Limit, f.Offset)
	var rows []adminWithUser
	if err := r.db.WithContext(ctx).Raw(dataSQL, queryArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	admins := make([]*domain.AdminProfile, 0, len(rows))
	for _, row := range rows {
		admins = append(admins, &domain.AdminProfile{
			ID:           row.ID,
			UserID:       row.UserID,
			SchoolID:     row.SchoolID,
			PhoneNumber:  row.PhoneNumber,
			Address:      row.Address,
			Gender:       row.Gender,
			Religion:     row.Religion,
			BirthPlace:   row.BirthPlace,
			BirthDate:    row.BirthDate,
			NIK:          row.NIK,
			KTPImagePath: row.KTPImagePath,
			DeletedAt:    row.DeletedAt,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
			Name:         row.Name,
			Email:        row.Email,
			Username:     row.Username,
			Avatar:       row.Avatar,
		})
	}
	return admins, total, nil
}
