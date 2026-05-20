package domain

import (
	"time"

	"github.com/google/uuid"
)

// StaffProfile represents a staff member's extended profile in the school system.
type StaffProfile struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	SchoolID    uuid.UUID
	Name        string
	Email       string
	Username    string
	Avatar      string
	PhoneNumber *string
	Address     *string
	Gender      *string
	Religion    *string
	BirthPlace  *string
	BirthDate   *time.Time
	NIK         *string
	KTPImagePath *string
	DeletedAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IsActive returns true if the staff profile has not been soft-deleted.
func (s *StaffProfile) IsActive() bool {
	return s.DeletedAt == nil
}
