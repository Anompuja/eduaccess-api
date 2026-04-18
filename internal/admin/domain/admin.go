package domain

import (
	"time"

	"github.com/google/uuid"
)

// AdminProfile is the extended profile for a user with role "admin_sekolah".
type AdminProfile struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	SchoolID     uuid.UUID
	PhoneNumber  string
	Address      string
	Gender       string
	Religion     string
	BirthPlace   string
	BirthDate    *time.Time
	NIK          string
	KTPImagePath string
	DeletedAt    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time

	// Joined from users table
	Name     string
	Email    string
	Username string
	Avatar   string
}

func (a *AdminProfile) IsActive() bool { return a.DeletedAt == nil }
