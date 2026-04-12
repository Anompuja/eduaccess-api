package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/school/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ── GORM models ───────────────────────────────────────────────────────────────

type schoolModel struct {
	ID           uuid.UUID  `gorm:"column:id;primaryKey"`
	HeadmasterID *uuid.UUID `gorm:"column:headmaster_id"`
	Name         string     `gorm:"column:name"`
	Address      string     `gorm:"column:address"`
	Phone        string     `gorm:"column:phone"`
	Email        string     `gorm:"column:email"`
	Description  string     `gorm:"column:description"`
	ImagePath    string     `gorm:"column:image_path"`
	TimeZone     string     `gorm:"column:time_zone"`
	Status       string     `gorm:"column:status"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

func (schoolModel) TableName() string { return "schools" }

type schoolRuleModel struct {
	ID        uuid.UUID  `gorm:"column:id;primaryKey"`
	SchoolID  uuid.UUID  `gorm:"column:school_id"`
	Key       string     `gorm:"column:key"`
	Value     string     `gorm:"column:value"`
	Note      string     `gorm:"column:note"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
}

func (schoolRuleModel) TableName() string { return "school_rules" }

type subscriptionModel struct {
	ID        uuid.UUID  `gorm:"column:id;primaryKey"`
	SchoolID  uuid.UUID  `gorm:"column:school_id"`
	PlanID    uuid.UUID  `gorm:"column:plan_id"`
	Status    string     `gorm:"column:status"`
	Cycle     string     `gorm:"column:cycle"`
	Quantity  int        `gorm:"column:quantity"`
	Price     int64      `gorm:"column:price"`
	EndsAt    *time.Time `gorm:"column:ends_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
}

func (subscriptionModel) TableName() string { return "subscriptions" }

type planModel struct {
	ID           uuid.UUID `gorm:"column:id;primaryKey"`
	Name         string    `gorm:"column:name"`
	Description  string    `gorm:"column:description"`
	Features     []byte    `gorm:"column:features"` // JSONB stored as []byte
	MonthlyPrice int64     `gorm:"column:monthly_price"`
	YearlyPrice  int64     `gorm:"column:yearly_price"`
	OnetimePrice *int64    `gorm:"column:onetime_price"`
	Active       bool      `gorm:"column:active"`
	IsDefault    bool      `gorm:"column:is_default"`
}

func (planModel) TableName() string { return "plans" }

// ── Repository ────────────────────────────────────────────────────────────────

// GormSchoolRepository implements domain.SchoolRepository.
type GormSchoolRepository struct {
	db *gorm.DB
}

func NewGormSchoolRepository(db *gorm.DB) *GormSchoolRepository {
	return &GormSchoolRepository{db: db}
}

func (r *GormSchoolRepository) Create(ctx context.Context, school *domain.School) error {
	m := schoolModel{
		ID:           school.ID,
		HeadmasterID: school.HeadmasterID,
		Name:         school.Name,
		Address:      school.Address,
		Phone:        school.Phone,
		Email:        school.Email,
		Description:  school.Description,
		ImagePath:    school.ImagePath,
		TimeZone:     school.TimeZone,
		Status:       school.Status,
		CreatedAt:    school.CreatedAt,
		UpdatedAt:    school.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *GormSchoolRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.School, error) {
	var m schoolModel
	if err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.ErrNotFound, "school not found")
		}
		return nil, err
	}
	return toSchoolDomain(m), nil
}

func (r *GormSchoolRepository) List(ctx context.Context, f domain.SchoolFilter) ([]*domain.School, int64, error) {
	base := `FROM schools WHERE deleted_at IS NULL`
	args := []interface{}{}
	conditions := []string{}

	if f.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, f.Status)
	}
	if f.Search != "" {
		conditions = append(conditions, "(name ILIKE ? OR email ILIKE ?)")
		like := "%" + f.Search + "%"
		args = append(args, like, like)
	}

	where := ""
	if len(conditions) > 0 {
		where = " AND " + strings.Join(conditions, " AND ")
	}

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(*) %s%s", base, where)
	if err := r.db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	dataSQL := fmt.Sprintf("SELECT * %s%s ORDER BY created_at DESC LIMIT ? OFFSET ?", base, where)
	queryArgs := append(args, f.Limit, f.Offset)
	var rows []schoolModel
	if err := r.db.WithContext(ctx).Raw(dataSQL, queryArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	schools := make([]*domain.School, 0, len(rows))
	for _, row := range rows {
		schools = append(schools, toSchoolDomain(row))
	}
	return schools, total, nil
}

func (r *GormSchoolRepository) Update(ctx context.Context, school *domain.School) error {
	return r.db.WithContext(ctx).
		Table("schools").
		Where("id = ?", school.ID).
		Updates(map[string]interface{}{
			"name":        school.Name,
			"address":     school.Address,
			"phone":       school.Phone,
			"email":       school.Email,
			"description": school.Description,
			"image_path":  school.ImagePath,
			"time_zone":   school.TimeZone,
			"status":      school.Status,
			"updated_at":  school.UpdatedAt,
		}).Error
}

func (r *GormSchoolRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Table("schools").
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted_at": now,
			"updated_at": now,
		}).Error
}

func (r *GormSchoolRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("schools").
		Where("name = ? AND deleted_at IS NULL", name).
		Count(&count).Error
	return count > 0, err
}

// ── Rules ─────────────────────────────────────────────────────────────────────

func (r *GormSchoolRepository) ListRules(ctx context.Context, schoolID uuid.UUID) ([]*domain.SchoolRule, error) {
	var rows []schoolRuleModel
	if err := r.db.WithContext(ctx).
		Where("school_id = ? AND deleted_at IS NULL", schoolID).
		Order("key ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	rules := make([]*domain.SchoolRule, 0, len(rows))
	for _, row := range rows {
		rules = append(rules, toRuleDomain(row))
	}
	return rules, nil
}

func (r *GormSchoolRepository) UpsertRule(ctx context.Context, rule *domain.SchoolRule) error {
	// ON CONFLICT (school_id, key) DO UPDATE
	sql := `
INSERT INTO school_rules (id, school_id, key, value, note, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (school_id, key) DO UPDATE
  SET value = EXCLUDED.value,
      note = EXCLUDED.note,
      updated_at = EXCLUDED.updated_at`

	return r.db.WithContext(ctx).Exec(sql,
		rule.ID, rule.SchoolID, rule.Key, rule.Value, rule.Note,
		rule.CreatedAt, rule.UpdatedAt,
	).Error
}

func (r *GormSchoolRepository) DeleteRule(ctx context.Context, schoolID uuid.UUID, key string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Table("school_rules").
		Where("school_id = ? AND key = ?", schoolID, key).
		Updates(map[string]interface{}{
			"deleted_at": now,
			"updated_at": now,
		}).Error
}

// ── Subscription ──────────────────────────────────────────────────────────────

func (r *GormSchoolRepository) FindActiveSubscription(ctx context.Context, schoolID uuid.UUID) (*domain.Subscription, error) {
	var sub subscriptionModel
	err := r.db.WithContext(ctx).
		Where("school_id = ? AND status IN ('active','trial')", schoolID).
		Order("created_at DESC").
		First(&sub).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.ErrNotFound, "no active subscription found")
		}
		return nil, err
	}

	result := &domain.Subscription{
		ID:        sub.ID,
		SchoolID:  sub.SchoolID,
		PlanID:    sub.PlanID,
		Status:    sub.Status,
		Cycle:     sub.Cycle,
		Quantity:  sub.Quantity,
		Price:     sub.Price,
		EndsAt:    sub.EndsAt,
		CreatedAt: sub.CreatedAt,
		UpdatedAt: sub.UpdatedAt,
	}

	// Load plan
	var plan planModel
	if err := r.db.WithContext(ctx).Where("id = ?", sub.PlanID).First(&plan).Error; err == nil {
		var features []string
		_ = json.Unmarshal(plan.Features, &features)
		result.Plan = &domain.Plan{
			ID:           plan.ID,
			Name:         plan.Name,
			Description:  plan.Description,
			Features:     features,
			MonthlyPrice: plan.MonthlyPrice,
			YearlyPrice:  plan.YearlyPrice,
			OnetimePrice: plan.OnetimePrice,
			Active:       plan.Active,
			IsDefault:    plan.IsDefault,
		}
	}

	return result, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func toSchoolDomain(m schoolModel) *domain.School {
	return &domain.School{
		ID:           m.ID,
		HeadmasterID: m.HeadmasterID,
		Name:         m.Name,
		Address:      m.Address,
		Phone:        m.Phone,
		Email:        m.Email,
		Description:  m.Description,
		ImagePath:    m.ImagePath,
		TimeZone:     m.TimeZone,
		Status:       m.Status,
		DeletedAt:    m.DeletedAt,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func toRuleDomain(m schoolRuleModel) *domain.SchoolRule {
	return &domain.SchoolRule{
		ID:        m.ID,
		SchoolID:  m.SchoolID,
		Key:       m.Key,
		Value:     m.Value,
		Note:      m.Note,
		DeletedAt: m.DeletedAt,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
