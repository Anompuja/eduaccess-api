package domain

import (
	"time"

	"github.com/google/uuid"
)

// HeadmasterProfile is the extended profile for a user with role "kepala_sekolah".
type HeadmasterProfile struct {
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

	Name     string
	Email    string
	Username string
	Avatar   string
}

func (h *HeadmasterProfile) IsActive() bool { return h.DeletedAt == nil }
