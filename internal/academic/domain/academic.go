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

// AcademicYear represents a school academic year (e.g. 2024/2025).
type AcademicYear struct {
	ID          uuid.UUID
	SchoolID    uuid.UUID
	Name        string
	StartDate   time.Time
	EndDate     time.Time
	IsActive    bool
	Description string
	DeletedAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Subject represents a school subject (Mata Pelajaran).
type Subject struct {
	ID        uuid.UUID
	SchoolID  uuid.UUID
	Name      string
	Category  string // core, elective, extracurricular, specialized, vocational
	DeletedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Classroom represents a physical classroom (Ruang Kelas).
type Classroom struct {
	ID         uuid.UUID
	SchoolID   uuid.UUID
	Name       string
	Capacity   int
	Floor      int
	Building   string
	RoomType   string
	Status     string // available, occupied, maintenance
	Facilities string
	DeletedAt  *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Schedule represents a school shift template (Jadwal / shift pagi/siang/full_day).
type Schedule struct {
	ID        uuid.UUID
	SchoolID  uuid.UUID
	ShiftType string // morning, afternoon, full_day
	StartTime string // e.g. "07:00"
	EndTime   string // e.g. "13:00"
	DeletedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
