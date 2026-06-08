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
	MaxStudents  int       `gorm:"column:max_students"`
	MonthlyPrice int64     `gorm:"column:monthly_price"`
	YearlyPrice  int64     `gorm:"column:yearly_price"`
	OnetimePrice *int64    `gorm:"column:onetime_price"`
	Active       bool      `gorm:"column:active"`
	IsDefault    bool      `gorm:"column:is_default"`
}

func (planModel) TableName() string { return "plans" }

type subscriptionPlanRow struct {
	SubscriptionID        *uuid.UUID `gorm:"column:subscription_id"`
	SchoolID              uuid.UUID  `gorm:"column:school_id"`
	SubscriptionPlanID    *uuid.UUID `gorm:"column:subscription_plan_id"`
	SubscriptionStatus    *string    `gorm:"column:subscription_status"`
	SubscriptionCycle     *string    `gorm:"column:subscription_cycle"`
	SubscriptionQuantity  *int       `gorm:"column:subscription_quantity"`
	SubscriptionPrice     *int64     `gorm:"column:subscription_price"`
	SubscriptionEndsAt    *time.Time `gorm:"column:subscription_ends_at"`
	SubscriptionCreatedAt *time.Time `gorm:"column:subscription_created_at"`
	SubscriptionUpdatedAt *time.Time `gorm:"column:subscription_updated_at"`
	PlanID                *uuid.UUID `gorm:"column:plan_id"`
	PlanName              *string    `gorm:"column:plan_name"`
	PlanDescription       *string    `gorm:"column:plan_description"`
	PlanFeatures          []byte     `gorm:"column:plan_features"`
	PlanMaxStudents       *int       `gorm:"column:plan_max_students"`
	PlanMonthlyPrice      *int64     `gorm:"column:plan_monthly_price"`
	PlanYearlyPrice       *int64     `gorm:"column:plan_yearly_price"`
	PlanOnetimePrice      *int64     `gorm:"column:plan_onetime_price"`
	PlanActive            *bool      `gorm:"column:plan_active"`
	PlanIsDefault         *bool      `gorm:"column:plan_is_default"`
}

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

func (r *GormSchoolRepository) CreateWithDefaultSubscription(ctx context.Context, school *domain.School) (*domain.Subscription, error) {
	var created *domain.Subscription

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := &GormSchoolRepository{db: tx}
		if err := repo.Create(ctx, school); err != nil {
			return err
		}

		sub, err := repo.createDefaultSubscription(ctx, school.ID)
		if err != nil {
			return err
		}
		created = sub
		return nil
	})
	if err != nil {
		return nil, err
	}

	return created, nil
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
	school := toSchoolDomain(m)
	if err := r.attachActiveSubscriptions(ctx, []*domain.School{school}); err != nil {
		return nil, err
	}
	return school, nil
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
	if err := r.attachActiveSubscriptions(ctx, schools); err != nil {
		return nil, 0, err
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

func (r *GormSchoolRepository) ListPlans(ctx context.Context) ([]*domain.Plan, error) {
	var rows []planModel
	if err := r.db.WithContext(ctx).
		Where("active = TRUE AND deleted_at IS NULL").
		Order("is_default DESC").
		Order("monthly_price ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	plans := make([]*domain.Plan, 0, len(rows))
	for _, row := range rows {
		plans = append(plans, toPlanDomain(row))
	}
	return plans, nil
}

func (r *GormSchoolRepository) FindPlanByID(ctx context.Context, id uuid.UUID) (*domain.Plan, error) {
	var row planModel
	if err := r.db.WithContext(ctx).
		Where("id = ? AND active = TRUE AND deleted_at IS NULL", id).
		First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.ErrNotFound, "plan not found")
		}
		return nil, err
	}
	return toPlanDomain(row), nil
}

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
		result.Plan = toPlanDomain(plan)
	}

	return result, nil
}

func (r *GormSchoolRepository) ReplaceSubscription(ctx context.Context, sub *domain.Subscription) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if err := tx.Table("subscriptions").
			Where("school_id = ? AND status IN ('active','trial')", sub.SchoolID).
			Updates(map[string]interface{}{
				"status":     "inactive",
				"ends_at":    now,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}

		m := subscriptionModel{
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
		return tx.Create(&m).Error
	})
}

// SetHeadmasterID updates schools.headmaster_id for the given school.
// It satisfies both domain.SchoolRepository and the headmaster application's
// SchoolHeadmasterSetter port.
func (r *GormSchoolRepository) SetHeadmasterID(ctx context.Context, schoolID, headmasterUserID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Table("schools").
		Where("id = ?", schoolID).
		Updates(map[string]interface{}{
			"headmaster_id": headmasterUserID,
			"updated_at":    now,
		}).Error
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

func (r *GormSchoolRepository) createDefaultSubscription(ctx context.Context, schoolID uuid.UUID) (*domain.Subscription, error) {
	plan, err := r.findDefaultPlan(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	endsAt := now.Add(14 * 24 * time.Hour)
	sub := &domain.Subscription{
		ID:        uuid.New(),
		SchoolID:  schoolID,
		PlanID:    plan.ID,
		Status:    "trial",
		Cycle:     "month",
		Quantity:  1,
		Price:     0,
		EndsAt:    &endsAt,
		CreatedAt: now,
		UpdatedAt: now,
		Plan:      plan,
	}

	m := subscriptionModel{
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
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return nil, err
	}

	return sub, nil
}

func (r *GormSchoolRepository) findDefaultPlan(ctx context.Context) (*domain.Plan, error) {
	var row planModel
	if err := r.db.WithContext(ctx).
		Where("is_default = TRUE AND active = TRUE AND deleted_at IS NULL").
		Order("created_at ASC").
		First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.ErrNotFound, "default plan not found")
		}
		return nil, err
	}
	return toPlanDomain(row), nil
}

func toPlanDomain(m planModel) *domain.Plan {
	var features []string
	_ = json.Unmarshal(m.Features, &features)

	return &domain.Plan{
		ID:           m.ID,
		Name:         m.Name,
		Description:  m.Description,
		Features:     features,
		MaxStudents:  m.MaxStudents,
		MonthlyPrice: m.MonthlyPrice,
		YearlyPrice:  m.YearlyPrice,
		OnetimePrice: m.OnetimePrice,
		Active:       m.Active,
		IsDefault:    m.IsDefault,
	}
}

func (r *GormSchoolRepository) attachActiveSubscriptions(ctx context.Context, schools []*domain.School) error {
	if len(schools) == 0 {
		return nil
	}

	ids := make([]uuid.UUID, 0, len(schools))
	for _, school := range schools {
		ids = append(ids, school.ID)
	}

	const sql = `
SELECT DISTINCT ON (s.school_id)
	s.id AS subscription_id,
	s.school_id,
	s.plan_id AS subscription_plan_id,
	s.status AS subscription_status,
	s.cycle AS subscription_cycle,
	s.quantity AS subscription_quantity,
	s.price AS subscription_price,
	s.ends_at AS subscription_ends_at,
	s.created_at AS subscription_created_at,
	s.updated_at AS subscription_updated_at,
	p.id AS plan_id,
	p.name AS plan_name,
	p.description AS plan_description,
	p.features AS plan_features,
	p.max_students AS plan_max_students,
	p.monthly_price AS plan_monthly_price,
	p.yearly_price AS plan_yearly_price,
	p.onetime_price AS plan_onetime_price,
	p.active AS plan_active,
	p.is_default AS plan_is_default
FROM subscriptions s
JOIN plans p ON p.id = s.plan_id
WHERE s.school_id IN ?
  AND s.status IN ('active', 'trial')
ORDER BY s.school_id, s.created_at DESC`

	var rows []subscriptionPlanRow
	if err := r.db.WithContext(ctx).Raw(sql, ids).Scan(&rows).Error; err != nil {
		return err
	}

	bySchoolID := make(map[uuid.UUID]*domain.Subscription, len(rows))
	for _, row := range rows {
		sub := toSubscriptionDomain(row)
		if sub != nil {
			bySchoolID[row.SchoolID] = sub
		}
	}

	for _, school := range schools {
		school.Subscription = bySchoolID[school.ID]
	}

	return nil
}

func toSubscriptionDomain(row subscriptionPlanRow) *domain.Subscription {
	if row.SubscriptionID == nil || row.SubscriptionPlanID == nil || row.SubscriptionStatus == nil || row.SubscriptionCycle == nil ||
		row.SubscriptionQuantity == nil || row.SubscriptionPrice == nil || row.SubscriptionCreatedAt == nil || row.SubscriptionUpdatedAt == nil {
		return nil
	}

	sub := &domain.Subscription{
		ID:        *row.SubscriptionID,
		SchoolID:  row.SchoolID,
		PlanID:    *row.SubscriptionPlanID,
		Status:    *row.SubscriptionStatus,
		Cycle:     *row.SubscriptionCycle,
		Quantity:  *row.SubscriptionQuantity,
		Price:     *row.SubscriptionPrice,
		EndsAt:    row.SubscriptionEndsAt,
		CreatedAt: *row.SubscriptionCreatedAt,
		UpdatedAt: *row.SubscriptionUpdatedAt,
	}

	if row.PlanID != nil && row.PlanName != nil && row.PlanDescription != nil && row.PlanMaxStudents != nil &&
		row.PlanMonthlyPrice != nil && row.PlanYearlyPrice != nil && row.PlanActive != nil && row.PlanIsDefault != nil {
		sub.Plan = toPlanDomain(planModel{
			ID:           *row.PlanID,
			Name:         *row.PlanName,
			Description:  *row.PlanDescription,
			Features:     row.PlanFeatures,
			MaxStudents:  *row.PlanMaxStudents,
			MonthlyPrice: *row.PlanMonthlyPrice,
			YearlyPrice:  *row.PlanYearlyPrice,
			OnetimePrice: row.PlanOnetimePrice,
			Active:       *row.PlanActive,
			IsDefault:    *row.PlanIsDefault,
		})
	}

	return sub
}
