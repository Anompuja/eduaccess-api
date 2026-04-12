package http

import "time"

// ── School ────────────────────────────────────────────────────────────────────

// SchoolResponse is the public representation of a school.
type SchoolResponse struct {
	ID           string              `json:"id"`
	HeadmasterID *string             `json:"headmaster_id,omitempty"`
	Name         string              `json:"name"`
	Address      string              `json:"address"`
	Phone        string              `json:"phone"`
	Email        string              `json:"email"`
	Description  string              `json:"description"`
	ImagePath    string              `json:"image_path"`
	TimeZone     string              `json:"time_zone"`
	Status       string              `json:"status"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
	Subscription *SubscriptionResponse `json:"subscription,omitempty"`
}

// CreateSchoolRequest is the body for POST /schools.
type CreateSchoolRequest struct {
	Name        string `json:"name"        validate:"required,min=2,max=191"`
	Address     string `json:"address"     validate:"omitempty,max=191"`
	Phone       string `json:"phone"       validate:"omitempty,max=50"`
	Email       string `json:"email"       validate:"omitempty,email,max=191"`
	Description string `json:"description" validate:"omitempty"`
	ImagePath   string `json:"image_path"  validate:"omitempty,max=191"`
	TimeZone    string `json:"time_zone"   validate:"omitempty,max=100"`
}

// UpdateSchoolRequest is the body for PUT /schools/:id.
type UpdateSchoolRequest struct {
	Name        *string `json:"name"        validate:"omitempty,min=2,max=191"`
	Address     *string `json:"address"     validate:"omitempty,max=191"`
	Phone       *string `json:"phone"       validate:"omitempty,max=50"`
	Email       *string `json:"email"       validate:"omitempty,email,max=191"`
	Description *string `json:"description" validate:"omitempty"`
	ImagePath   *string `json:"image_path"  validate:"omitempty,max=191"`
	TimeZone    *string `json:"time_zone"   validate:"omitempty,max=100"`
	Status      *string `json:"status"      validate:"omitempty,oneof=active nonactive"`
}

// ── Rules ─────────────────────────────────────────────────────────────────────

// SchoolRuleResponse is the public representation of a school rule.
type SchoolRuleResponse struct {
	ID        string    `json:"id"`
	SchoolID  string    `json:"school_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Note      string    `json:"note"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RuleInput is a single key-value pair in the upsert request.
type RuleInput struct {
	Key   string `json:"key"   validate:"required,min=1,max=191"`
	Value string `json:"value" validate:"required,max=191"`
	Note  string `json:"note"  validate:"omitempty"`
}

// UpsertRulesRequest is the body for PUT /schools/:id/rules.
type UpsertRulesRequest struct {
	Rules []RuleInput `json:"rules" validate:"required,min=1,dive"`
}

// ── Subscription ──────────────────────────────────────────────────────────────

// SubscriptionResponse is the public representation of a subscription.
type SubscriptionResponse struct {
	ID        string        `json:"id"`
	SchoolID  string        `json:"school_id"`
	Status    string        `json:"status"`
	Cycle     string        `json:"cycle"`
	Quantity  int           `json:"quantity"`
	Price     int64         `json:"price"`
	EndsAt    *time.Time    `json:"ends_at,omitempty"`
	Plan      *PlanResponse `json:"plan,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// PlanResponse is the public representation of a billing plan.
type PlanResponse struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Features     []string `json:"features"`
	MonthlyPrice int64    `json:"monthly_price"`
	YearlyPrice  int64    `json:"yearly_price"`
}
