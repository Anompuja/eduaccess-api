package http

import "time"

type CreateCheckoutRequest struct {
	PlanID string `json:"plan_id" validate:"required,uuid4"`
	Cycle  string `json:"cycle" validate:"required,oneof=month year"`
}

type PaymentResponse struct {
	ID                      string     `json:"id"`
	SchoolID                string     `json:"school_id"`
	PlanID                  string     `json:"plan_id"`
	CreatedByUserID         string     `json:"created_by_user_id"`
	ActivatedSubscriptionID *string    `json:"activated_subscription_id,omitempty"`
	Status                  string     `json:"status"`
	Cycle                   string     `json:"cycle"`
	Amount                  int64      `json:"amount"`
	Currency                string     `json:"currency"`
	Provider                string     `json:"provider"`
	ProviderOrderID         string     `json:"provider_order_id"`
	ProviderTransactionID   string     `json:"provider_transaction_id,omitempty"`
	ProviderSnapToken       string     `json:"provider_snap_token,omitempty"`
	ProviderRedirectURL     string     `json:"provider_redirect_url,omitempty"`
	PaymentType             string     `json:"payment_type,omitempty"`
	TransactionStatus       string     `json:"transaction_status,omitempty"`
	FraudStatus             string     `json:"fraud_status,omitempty"`
	PaidAt                  *time.Time `json:"paid_at,omitempty"`
	ExpiresAt               *time.Time `json:"expires_at,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

type MidtransNotificationRequest struct {
	OrderID           string `json:"order_id"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
	TransactionID     string `json:"transaction_id"`
	TransactionStatus string `json:"transaction_status"`
	PaymentType       string `json:"payment_type"`
	FraudStatus       string `json:"fraud_status"`
}
