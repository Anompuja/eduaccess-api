package infrastructure

import (
	"context"
	"fmt"
	"strings"
	"time"

	billingdomain "github.com/eduaccess/eduaccess-api/internal/billing/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type paymentTransactionModel struct {
	ID                      uuid.UUID  `gorm:"column:id;primaryKey"`
	SchoolID                uuid.UUID  `gorm:"column:school_id"`
	PlanID                  uuid.UUID  `gorm:"column:plan_id"`
	CreatedByUserID         uuid.UUID  `gorm:"column:created_by_user_id"`
	ActivatedSubscriptionID *uuid.UUID `gorm:"column:activated_subscription_id"`
	Status                  string     `gorm:"column:status"`
	Cycle                   string     `gorm:"column:cycle"`
	Amount                  int64      `gorm:"column:amount"`
	Currency                string     `gorm:"column:currency"`
	Provider                string     `gorm:"column:provider"`
	ProviderOrderID         string     `gorm:"column:provider_order_id"`
	ProviderTransactionID   string     `gorm:"column:provider_transaction_id"`
	ProviderSnapToken       string     `gorm:"column:provider_snap_token"`
	ProviderRedirectURL     string     `gorm:"column:provider_redirect_url"`
	PaymentType             string     `gorm:"column:payment_type"`
	TransactionStatus       string     `gorm:"column:transaction_status"`
	FraudStatus             string     `gorm:"column:fraud_status"`
	RawNotification         string     `gorm:"column:raw_notification"`
	PaidAt                  *time.Time `gorm:"column:paid_at"`
	ExpiresAt               *time.Time `gorm:"column:expires_at"`
	CreatedAt               time.Time  `gorm:"column:created_at"`
	UpdatedAt               time.Time  `gorm:"column:updated_at"`
}

func (paymentTransactionModel) TableName() string { return "payment_transactions" }

type paymentListRow struct {
	Payment    paymentTransactionModel `gorm:"embedded"`
	SchoolName string                  `gorm:"column:school_name"`
	PlanName   string                  `gorm:"column:plan_name"`
}

type GormPaymentRepository struct {
	db *gorm.DB
}

func NewGormPaymentRepository(db *gorm.DB) *GormPaymentRepository {
	return &GormPaymentRepository{db: db}
}

func (r *GormPaymentRepository) Create(ctx context.Context, payment *billingdomain.PaymentTransaction) error {
	return r.db.WithContext(ctx).Create(toPaymentModel(payment)).Error
}

func (r *GormPaymentRepository) List(ctx context.Context, filter billingdomain.PaymentFilter) ([]*billingdomain.PaymentTransaction, int64, error) {
	base := `
FROM payment_transactions pt
JOIN schools s ON s.id = pt.school_id
JOIN plans p ON p.id = pt.plan_id`

	args := []any{}
	conditions := []string{}

	if filter.SchoolID != nil {
		conditions = append(conditions, "pt.school_id = ?")
		args = append(args, *filter.SchoolID)
	}
	if filter.Status != "" {
		conditions = append(conditions, "pt.status = ?")
		args = append(args, filter.Status)
	}
	if filter.Search != "" {
		conditions = append(conditions, "(s.name ILIKE ? OR p.name ILIKE ? OR pt.provider_order_id ILIKE ?)")
		like := "%" + filter.Search + "%"
		args = append(args, like, like, like)
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(*) %s%s", base, where)
	if err := r.db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	dataSQL := fmt.Sprintf(`
SELECT
	pt.*,
	s.name AS school_name,
	p.name AS plan_name
%s%s
ORDER BY pt.created_at DESC
LIMIT ? OFFSET ?`, base, where)
	queryArgs := append(args, filter.Limit, filter.Offset)

	var rows []paymentListRow
	if err := r.db.WithContext(ctx).Raw(dataSQL, queryArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	payments := make([]*billingdomain.PaymentTransaction, 0, len(rows))
	for _, row := range rows {
		payments = append(payments, toPaymentDomain(row))
	}

	return payments, total, nil
}

func (r *GormPaymentRepository) FindByID(ctx context.Context, id uuid.UUID) (*billingdomain.PaymentTransaction, error) {
	var row paymentListRow
	if err := r.baseQuery(ctx).Where("pt.id = ?", id).Take(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.ErrNotFound, "payment transaction not found")
		}
		return nil, err
	}
	return toPaymentDomain(row), nil
}

func (r *GormPaymentRepository) FindByProviderOrderID(ctx context.Context, orderID string) (*billingdomain.PaymentTransaction, error) {
	var row paymentListRow
	if err := r.baseQuery(ctx).Where("pt.provider_order_id = ?", orderID).Take(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.ErrNotFound, "payment transaction not found")
		}
		return nil, err
	}
	return toPaymentDomain(row), nil
}

func (r *GormPaymentRepository) FindLatestPendingBySchool(ctx context.Context, schoolID uuid.UUID) (*billingdomain.PaymentTransaction, error) {
	var row paymentListRow
	if err := r.baseQuery(ctx).
		Where("pt.school_id = ? AND pt.status = ?", schoolID, billingdomain.PaymentStatusPending).
		Order("pt.created_at DESC").
		Take(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.ErrNotFound, "pending payment transaction not found")
		}
		return nil, err
	}
	return toPaymentDomain(row), nil
}

func (r *GormPaymentRepository) Update(ctx context.Context, payment *billingdomain.PaymentTransaction) error {
	return r.db.WithContext(ctx).
		Table("payment_transactions").
		Where("id = ?", payment.ID).
		Updates(map[string]any{
			"activated_subscription_id": payment.ActivatedSubscriptionID,
			"status":                    payment.Status,
			"cycle":                     payment.Cycle,
			"amount":                    payment.Amount,
			"currency":                  payment.Currency,
			"provider_transaction_id":   payment.ProviderTransactionID,
			"provider_snap_token":       payment.ProviderSnapToken,
			"provider_redirect_url":     payment.ProviderRedirectURL,
			"payment_type":              payment.PaymentType,
			"transaction_status":        payment.TransactionStatus,
			"fraud_status":              payment.FraudStatus,
			"raw_notification":          normalizeRawNotification(payment.RawNotification),
			"paid_at":                   payment.PaidAt,
			"expires_at":                payment.ExpiresAt,
			"updated_at":                payment.UpdatedAt,
		}).Error
}

func toPaymentModel(payment *billingdomain.PaymentTransaction) paymentTransactionModel {
	return paymentTransactionModel{
		ID:                      payment.ID,
		SchoolID:                payment.SchoolID,
		PlanID:                  payment.PlanID,
		CreatedByUserID:         payment.CreatedByUserID,
		ActivatedSubscriptionID: payment.ActivatedSubscriptionID,
		Status:                  payment.Status,
		Cycle:                   payment.Cycle,
		Amount:                  payment.Amount,
		Currency:                payment.Currency,
		Provider:                payment.Provider,
		ProviderOrderID:         payment.ProviderOrderID,
		ProviderTransactionID:   payment.ProviderTransactionID,
		ProviderSnapToken:       payment.ProviderSnapToken,
		ProviderRedirectURL:     payment.ProviderRedirectURL,
		PaymentType:             payment.PaymentType,
		TransactionStatus:       payment.TransactionStatus,
		FraudStatus:             payment.FraudStatus,
		RawNotification:         normalizeRawNotification(payment.RawNotification),
		PaidAt:                  payment.PaidAt,
		ExpiresAt:               payment.ExpiresAt,
		CreatedAt:               payment.CreatedAt,
		UpdatedAt:               payment.UpdatedAt,
	}
}

func toPaymentDomainModel(row paymentTransactionModel) *billingdomain.PaymentTransaction {
	return &billingdomain.PaymentTransaction{
		ID:                      row.ID,
		SchoolID:                row.SchoolID,
		PlanID:                  row.PlanID,
		CreatedByUserID:         row.CreatedByUserID,
		ActivatedSubscriptionID: row.ActivatedSubscriptionID,
		Status:                  row.Status,
		Cycle:                   row.Cycle,
		Amount:                  row.Amount,
		Currency:                row.Currency,
		Provider:                row.Provider,
		ProviderOrderID:         row.ProviderOrderID,
		ProviderTransactionID:   row.ProviderTransactionID,
		ProviderSnapToken:       row.ProviderSnapToken,
		ProviderRedirectURL:     row.ProviderRedirectURL,
		PaymentType:             row.PaymentType,
		TransactionStatus:       row.TransactionStatus,
		FraudStatus:             row.FraudStatus,
		RawNotification:         row.RawNotification,
		PaidAt:                  row.PaidAt,
		ExpiresAt:               row.ExpiresAt,
		CreatedAt:               row.CreatedAt,
		UpdatedAt:               row.UpdatedAt,
	}
}

func toPaymentDomain(row paymentListRow) *billingdomain.PaymentTransaction {
	payment := toPaymentDomainModel(row.Payment)
	payment.SchoolName = row.SchoolName
	payment.PlanName = row.PlanName
	return payment
}

func normalizeRawNotification(raw string) string {
	if raw == "" {
		return "{}"
	}
	return raw
}

func (r *GormPaymentRepository) baseQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("payment_transactions AS pt").
		Select("pt.*, s.name AS school_name, p.name AS plan_name").
		Joins("JOIN schools s ON s.id = pt.school_id").
		Joins("JOIN plans p ON p.id = pt.plan_id")
}
