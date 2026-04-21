package domain

import (
	"time"

	"github.com/google/uuid"
)

// EducationLevel (e.g. SD, SMP, SMA) scoped to a school.
type EducationLevel struct {
	ID        uuid.UUID
	SchoolID  uuid.UUID
	Name      string
	DeletedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Class belongs to an EducationLevel (e.g. Kelas 7, Kelas 8).
type Class struct {
	ID               uuid.UUID
	SchoolID         uuid.UUID
	EducationLevelID uuid.UUID
	Name             string
	DeletedAt        *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// SubClass belongs to a Class (e.g. 7A, 7B).
type SubClass struct {
	ID        uuid.UUID
	SchoolID  uuid.UUID
	ClassID   uuid.UUID
	Name      string
	DeletedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
