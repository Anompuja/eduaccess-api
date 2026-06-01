package infrastructure

import (
	"context"
	"errors"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	supabasePkg "github.com/eduaccess/eduaccess-api/pkg/supabase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SupabaseUserRepository implements domain.UserRepository.
// Reads (FindByEmail, FindByID, Exists*) delegate to GormUserRepository against
// public.users. Writes (Create) call the Supabase Admin API first so that the
// identity record in auth.users is created before the profile row in public.users.
type SupabaseUserRepository struct {
	gorm     *GormUserRepository
	db       *gorm.DB
	supabase *supabasePkg.Client
}

func NewSupabaseUserRepository(db *gorm.DB, supabase *supabasePkg.Client) *SupabaseUserRepository {
	return &SupabaseUserRepository{
		gorm:     NewGormUserRepository(db),
		db:       db,
		supabase: supabase,
	}
}

func (r *SupabaseUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.gorm.FindByEmail(ctx, email)
}

func (r *SupabaseUserRepository) FindByEmailIncludingDeleted(ctx context.Context, email string) (*domain.User, error) {
	return r.gorm.FindByEmailIncludingDeleted(ctx, email)
}

func (r *SupabaseUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return r.gorm.FindByID(ctx, id)
}

func (r *SupabaseUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return r.gorm.ExistsByEmail(ctx, email)
}

func (r *SupabaseUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return r.gorm.ExistsByUsername(ctx, username)
}

func (r *SupabaseUserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&userModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", time.Now()).Error; err != nil {
		return err
	}

	r.supabase.DeleteUser(ctx, id)
	return nil
}

// Create registers the user in Supabase Auth then inserts the profile row,
// assigns the role, and links to a school — all in one transaction.
// If the GORM transaction fails, we attempt to roll back the Supabase Auth user.
func (r *SupabaseUserRepository) Create(ctx context.Context, user *domain.User) error {
	// 1. Create the identity record in Supabase Auth. user.Password holds the
	//    plaintext credential at this point; Supabase hashes it internally.
	authUser, err := r.supabase.CreateUser(ctx, user.Email, user.Password)
	if err != nil {
		return err // already an AppError
	}
	user.ID = authUser.ID

	// 2. Persist the profile + role + school link inside a transaction.
	txErr := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m := &userModel{
			ID:        user.ID,
			Name:      user.Name,
			Username:  user.Username,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Verified:  true, // admin-created accounts are immediately verified
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
		if err := tx.Create(m).Error; err != nil {
			return err
		}

		var role roleModel
		if err := tx.Where("name = ?", user.Role).First(&role).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperror.New(apperror.ErrBadRequest, "unknown role: "+user.Role)
			}
			return err
		}
		user.RoleID = &role.ID

		if err := tx.Create(&modelHasRole{UserID: user.ID, RoleID: role.ID}).Error; err != nil {
			return err
		}

		if user.SchoolID != nil {
			su := &schoolUserModel{
				ID:        uuid.New(),
				UserID:    user.ID,
				SchoolID:  *user.SchoolID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := tx.Create(su).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if txErr != nil {
		// Best-effort rollback: remove the orphaned Supabase Auth user.
		r.supabase.DeleteUser(ctx, user.ID)
		return txErr
	}

	return nil
}
