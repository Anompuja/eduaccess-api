package infrastructure

import (
	"context"
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

type GormPaymentRepository struct {
	db *gorm.DB
}

func NewGormPaymentRepository(db *gorm.DB) *GormPaymentRepository {
	return &GormPaymentRepository{db: db}
}

func (r *GormPaymentRepository) Create(ctx context.Context, payment *billingdomain.PaymentTransaction) error {
	return r.db.WithContext(ctx).Create(toPaymentModel(payment)).Error
}

func (r *GormPaymentRepository) FindByID(ctx context.Context, id uuid.UUID) (*billingdomain.PaymentTransaction, error) {
	var row paymentTransactionModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.ErrNotFound, "payment transaction not found")
		}
		return nil, err
	}
	return toPaymentDomain(row), nil
}

func (r *GormPaymentRepository) FindByProviderOrderID(ctx context.Context, orderID string) (*billingdomain.PaymentTransaction, error) {
	var row paymentTransactionModel
	if err := r.db.WithContext(ctx).Where("provider_order_id = ?", orderID).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.New(apperror.ErrNotFound, "payment transaction not found")
		}
		return nil, err
	}
	return toPaymentDomain(row), nil
}

func (r *GormPaymentRepository) FindLatestPendingBySchool(ctx context.Context, schoolID uuid.UUID) (*billingdomain.PaymentTransaction, error) {
	var row paymentTransactionModel
	if err := r.db.WithContext(ctx).
		Where("school_id = ? AND status = ?", schoolID, billingdomain.PaymentStatusPending).
		Order("created_at DESC").
		First(&row).Error; err != nil {
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

func toPaymentDomain(row paymentTransactionModel) *billingdomain.PaymentTransaction {
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

func normalizeRawNotification(raw string) string {
	if raw == "" {
		return "{}"
	}
	return raw
}
