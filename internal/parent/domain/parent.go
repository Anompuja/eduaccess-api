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
