package infrastructure

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/headmaster/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ── GORM models ───────────────────────────────────────────────────────────────

type headmasterProfileModel struct {
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

func (headmasterProfileModel) TableName() string { return "headmaster_profiles" }

// headmasterWithUser is the scan target for JOIN queries.
type headmasterWithUser struct {
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
	UserName     string     `gorm:"column:user_name"`
	UserEmail    string     `gorm:"column:user_email"`
	Username     string     `gorm:"column:username"`
	Avatar       string     `gorm:"column:avatar"`
}

func (r *headmasterWithUser) toDomain() *domain.HeadmasterProfile {
	return &domain.HeadmasterProfile{
		ID:           r.ID,
		UserID:       r.UserID,
		SchoolID:     r.SchoolID,
		PhoneNumber:  r.PhoneNumber,
		Address:      r.Address,
		Gender:       r.Gender,
		Religion:     r.Religion,
		BirthPlace:   r.BirthPlace,
		BirthDate:    r.BirthDate,
		NIK:          r.NIK,
		KTPImagePath: r.KTPImagePath,
		DeletedAt:    r.DeletedAt,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
		Name:         r.UserName,
		Email:        r.UserEmail,
		Username:     r.Username,
		Avatar:       r.Avatar,
	}
}

const headmasterJoinSQL = `
SELECT
    hp.*,
    u.name  AS user_name,
    u.email AS user_email,
    u.username,
    u.avatar
FROM headmaster_profiles hp
JOIN users u ON u.id = hp.user_id
WHERE hp.deleted_at IS NULL`

// ── Repository ────────────────────────────────────────────────────────────────

// GormHeadmasterRepository implements domain.HeadmasterRepository.
type GormHeadmasterRepository struct {
	db *gorm.DB
}

func NewGormHeadmasterRepository(db *gorm.DB) *GormHeadmasterRepository {
	return &GormHeadmasterRepository{db: db}
}

func (r *GormHeadmasterRepository) CreateHeadmasterProfile(ctx context.Context, profile *domain.HeadmasterProfile) error {
	m := &headmasterProfileModel{
		ID:           profile.ID,
		UserID:       profile.UserID,
		SchoolID:     profile.SchoolID,
		PhoneNumber:  profile.PhoneNumber,
		Address:      profile.Address,
		Gender:       profile.Gender,
		Religion:     profile.Religion,
		BirthPlace:   profile.BirthPlace,
		BirthDate:    profile.BirthDate,
		NIK:          profile.NIK,
		KTPImagePath: profile.KTPImagePath,
		CreatedAt:    profile.CreatedAt,
		UpdatedAt:    profile.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *GormHeadmasterRepository) FindHeadmasterByID(ctx context.Context, id uuid.UUID) (*domain.HeadmasterProfile, error) {
	var row headmasterWithUser
	sql := fmt.Sprintf("%s AND hp.id = ? LIMIT 1", headmasterJoinSQL)
	if err := r.db.WithContext(ctx).Raw(sql, id).Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.ID == (uuid.UUID{}) {
		return nil, apperror.ErrNotFound
	}
	return row.toDomain(), nil
}

func (r *GormHeadmasterRepository) FindHeadmasterByUserID(ctx context.Context, userID uuid.UUID) (*domain.HeadmasterProfile, error) {
	var row headmasterWithUser
	sql := fmt.Sprintf("%s AND hp.user_id = ? LIMIT 1", headmasterJoinSQL)
	if err := r.db.WithContext(ctx).Raw(sql, userID).Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.ID == (uuid.UUID{}) {
		return nil, apperror.ErrNotFound
	}
	return row.toDomain(), nil
}

func (r *GormHeadmasterRepository) ListHeadmasters(ctx context.Context, f domain.HeadmasterFilter) ([]*domain.HeadmasterProfile, int64, error) {
	conditions := []string{}
	args := []interface{}{}

	if f.SchoolID != nil {
		conditions = append(conditions, "hp.school_id = ?")
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

	base := `
FROM headmaster_profiles hp
JOIN users u ON u.id = hp.user_id
WHERE hp.deleted_at IS NULL`

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(*) %s%s", base, where)
	if err := r.db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	dataSQL := fmt.Sprintf(`
SELECT
    hp.*,
    u.name  AS user_name,
    u.email AS user_email,
    u.username,
    u.avatar
%s%s
ORDER BY hp.created_at DESC
LIMIT ? OFFSET ?`, base, where)

	queryArgs := append(args, f.Limit, f.Offset)
	var rows []headmasterWithUser
	if err := r.db.WithContext(ctx).Raw(dataSQL, queryArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	profiles := make([]*domain.HeadmasterProfile, 0, len(rows))
	for i := range rows {
		profiles = append(profiles, rows[i].toDomain())
	}
	return profiles, total, nil
}

func (r *GormHeadmasterRepository) UpdateHeadmasterProfile(ctx context.Context, profile *domain.HeadmasterProfile) error {
	return r.db.WithContext(ctx).
		Table("headmaster_profiles").
		Where("id = ?", profile.ID).
		Updates(map[string]interface{}{
			"phone_number":   profile.PhoneNumber,
			"address":        profile.Address,
			"gender":         profile.Gender,
			"religion":       profile.Religion,
			"birth_place":    profile.BirthPlace,
			"birth_date":     profile.BirthDate,
			"nik":            profile.NIK,
			"ktp_image_path": profile.KTPImagePath,
			"updated_at":     profile.UpdatedAt,
		}).Error
}

func (r *GormHeadmasterRepository) SoftDeleteHeadmaster(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Table("headmaster_profiles").
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted_at": now,
			"updated_at": now,
		}).Error
}
