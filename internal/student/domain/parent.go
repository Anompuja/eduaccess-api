package domain

import (
	"time"

	"github.com/google/uuid"
)

// ParentProfile is the extended profile for a user with role "orangtua".
type ParentProfile struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	SchoolID       uuid.UUID
	FatherName     string
	MotherName     string
	FatherReligion string
	MotherReligion string
	PhoneNumber    string
	Address        string
	DeletedAt      *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// Joined from users table
	Name     string
	Email    string
	Username string
	Avatar   string
}

func (p *ParentProfile) IsActive() bool { return p.DeletedAt == nil }

// ParentLink represents the many-to-many relationship between students and parents.
type ParentLink struct {
	ID           uuid.UUID
	SchoolID     uuid.UUID
	StudentID    uuid.UUID
	ParentID     uuid.UUID
	Relationship string // father | mother | guardian | other
	IsPrimary    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Parent *ParentProfile
}

// Relationship constants
const (
	RelFather   = "father"
	RelMother   = "mother"
	RelGuardian = "guardian"
	RelOther    = "other"
)
