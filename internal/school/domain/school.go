package domain

import (
	"time"

	"github.com/google/uuid"
)

// School represents a tenant school.
type School struct {
	ID           uuid.UUID
	HeadmasterID *uuid.UUID
	Name         string
	Address      string
	Phone        string
	Email        string
	Description  string
	ImagePath    string
	TimeZone     string
	Status       string // "active" | "nonactive"
	DeletedAt    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time

	// Eager-loaded relations
	Subscription *Subscription
}

func (s *School) IsActive() bool { return s.DeletedAt == nil && s.Status == "active" }

// SchoolRule is a key-value setting scoped to a school.
type SchoolRule struct {
	ID        uuid.UUID
	SchoolID  uuid.UUID
	Key       string
	Value     string
	Note      string
	DeletedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Subscription represents an active billing subscription for a school.
type Subscription struct {
	ID        uuid.UUID
	SchoolID  uuid.UUID
	PlanID    uuid.UUID
	Status    string // active | inactive | trial | expired | cancelled
	Cycle     string // month | year | onetime
	Quantity  int
	Price     int64
	EndsAt    *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time

	Plan *Plan
}

// Plan is a billing plan definition.
type Plan struct {
	ID           uuid.UUID
	Name         string
	Description  string
	Features     []string
	MonthlyPrice int64
	YearlyPrice  int64
	OnetimePrice *int64
	Active       bool
	IsDefault    bool
}

// Status constants
const (
	StatusActive    = "active"
	StatusNonactive = "nonactive"
)
