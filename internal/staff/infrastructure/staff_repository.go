package infrastructure

import (
	"context"
	"errors"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/staff/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// staffProfileModel represents the ORM model for staff profiles.
type staffProfileModel struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID       uuid.UUID  `gorm:"type:uuid;index"`
	SchoolID     uuid.UUID  `gorm:"type:uuid;index"`
	PhoneNumber  *string    `gorm:"type:varchar(50)"`
	Address      *string    `gorm:"type:text"`
	Gender       *string    `gorm:"type:varchar(50)"`
	Religion     *string    `gorm:"type:varchar(100)"`
	BirthPlace   *string    `gorm:"type:varchar(191)"`
	BirthDate    *time.Time `gorm:"type:date"`
	NIK          *string    `gorm:"type:varchar(50)"`
	KTPImagePath *string    `gorm:"type:varchar(191)"`
	DeletedAt    *time.Time `gorm:"type:timestamptz;index"`
	CreatedAt    time.Time  `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt    time.Time  `gorm:"type:timestamptz;autoUpdateTime"`
}

// TableName specifies the table name for the ORM.
func (s *staffProfileModel) TableName() string {
	return "staff_profiles"
}

// StaffRepository implements the domain repository interface for staff.
type StaffRepository struct {
	db *gorm.DB
}

// NewStaffRepository creates a new StaffRepository.
func NewStaffRepository(db *gorm.DB) *StaffRepository {
	return &StaffRepository{db: db}
}

// CreateStaffProfile creates a new staff profile in the database.
func (r *StaffRepository) CreateStaffProfile(ctx context.Context, staff *domain.StaffProfile) error {
	model := &staffProfileModel{
		ID:           staff.ID,
		UserID:       staff.UserID,
		SchoolID:     staff.SchoolID,
		PhoneNumber:  staff.PhoneNumber,
		Address:      staff.Address,
		Gender:       staff.Gender,
		Religion:     staff.Religion,
		BirthPlace:   staff.BirthPlace,
		BirthDate:    staff.BirthDate,
		NIK:          staff.NIK,
		KTPImagePath: staff.KTPImagePath,
	}

	return r.db.WithContext(ctx).Create(model).Error
}

// FindStaffByID retrieves a staff by ID.
func (r *StaffRepository) FindStaffByID(ctx context.Context, id uuid.UUID) (*domain.StaffProfile, error) {
	var model staffProfileModel
	result := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}

	// Fetch associated user data
	var user struct {
		Name     string
		Email    string
		Username string
		Avatar   string
	}
	if err := r.db.WithContext(ctx).Table("users").Where("id = ?", model.UserID).
		Select("name, email, username, avatar").First(&user).Error; err != nil {
		return nil, err
	}

	return r.modelToDomain(&model, user.Name, user.Email, user.Username, user.Avatar), nil
}

// UpdateStaffProfile updates an existing staff profile.
func (r *StaffRepository) UpdateStaffProfile(ctx context.Context, staff *domain.StaffProfile) error {
	model := &staffProfileModel{
		ID:           staff.ID,
		UserID:       staff.UserID,
		SchoolID:     staff.SchoolID,
		PhoneNumber:  staff.PhoneNumber,
		Address:      staff.Address,
		Gender:       staff.Gender,
		Religion:     staff.Religion,
		BirthPlace:   staff.BirthPlace,
		BirthDate:    staff.BirthDate,
		NIK:          staff.NIK,
		KTPImagePath: staff.KTPImagePath,
	}

	return r.db.WithContext(ctx).Model(&staffProfileModel{}).Where("id = ?", staff.ID).
		Updates(model).Error
}

// SoftDeleteStaff soft-deletes a staff profile by setting deleted_at.
func (r *StaffRepository) SoftDeleteStaff(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&staffProfileModel{}).Where("id = ?", id).
		Update("deleted_at", time.Now()).Error
}

// ListStaff retrieves a paginated list of staff with optional filtering.
func (r *StaffRepository) ListStaff(ctx context.Context, filter domain.StaffFilter) ([]*domain.StaffProfile, int64, error) {
	var models []staffProfileModel
	var total int64

	query := r.db.WithContext(ctx).
		Where("school_id = ? AND deleted_at IS NULL", filter.SchoolID)

	// Apply search filter
	if filter.Search != "" {
		query = query.Where("(u.name ILIKE ? OR u.email ILIKE ? OR u.username ILIKE ?)",
			"%"+filter.Search+"%", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	// Get total count
	if err := query.Model(&staffProfileModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch paginated results with user data via JOIN
	result := query.
		Joins("LEFT JOIN users u ON staff_profiles.user_id = u.id").
		Select("staff_profiles.*, u.name, u.email, u.username, u.avatar").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Order("staff_profiles.created_at DESC").
		Find(&models)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	// Convert models to domain entities
	staff := make([]*domain.StaffProfile, len(models))
	for i, model := range models {
		// Fetch user data for this staff
		var user struct {
			Name     string
			Email    string
			Username string
			Avatar   string
		}
		r.db.WithContext(ctx).Table("users").Where("id = ?", model.UserID).
			Select("name, email, username, avatar").First(&user)

		staff[i] = r.modelToDomain(&model, user.Name, user.Email, user.Username, user.Avatar)
	}

	return staff, total, nil
}

// modelToDomain converts a model to a domain entity.
func (r *StaffRepository) modelToDomain(m *staffProfileModel, name, email, username, avatar string) *domain.StaffProfile {
	return &domain.StaffProfile{
		ID:           m.ID,
		UserID:       m.UserID,
		SchoolID:     m.SchoolID,
		Name:         name,
		Email:        email,
		Username:     username,
		Avatar:       avatar,
		PhoneNumber:  m.PhoneNumber,
		Address:      m.Address,
		Gender:       m.Gender,
		Religion:     m.Religion,
		BirthPlace:   m.BirthPlace,
		BirthDate:    m.BirthDate,
		NIK:          m.NIK,
		KTPImagePath: m.KTPImagePath,
		DeletedAt:    m.DeletedAt,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}
