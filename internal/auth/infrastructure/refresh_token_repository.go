package infrastructure

import (
	"context"
	"errors"

	"github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormRefreshTokenRepository implements domain.RefreshTokenRepository using GORM.
type GormRefreshTokenRepository struct {
	db *gorm.DB
}

func NewGormRefreshTokenRepository(db *gorm.DB) *GormRefreshTokenRepository {
	return &GormRefreshTokenRepository{db: db}
}

func (r *GormRefreshTokenRepository) Create(ctx context.Context, rt *domain.RefreshToken) error {
	m := refreshTokenModelFromDomain(rt)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *GormRefreshTokenRepository) FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	var m refreshTokenModel
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *GormRefreshTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&refreshTokenModel{}).Error
}

func (r *GormRefreshTokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&refreshTokenModel{}).Error
}
