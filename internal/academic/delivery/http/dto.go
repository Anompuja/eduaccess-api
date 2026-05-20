package http

import "time"

type EducationLevelResponse struct {
	ID        string    `json:"id"`
	SchoolID  string    `json:"school_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ClassResponse struct {
	ID               string    `json:"id"`
	SchoolID         string    `json:"school_id"`
	EducationLevelID string    `json:"education_level_id"`
	Name             string    `json:"name"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type SubClassResponse struct {
	ID        string    `json:"id"`
	SchoolID  string    `json:"school_id"`
	ClassID   string    `json:"class_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AcademicNameRequest struct {
	Name string `json:"name" validate:"required,min=1,max=191"`
}

type CreateClassRequest struct {
	LevelID string `json:"level_id" validate:"required,uuid"`
	Name    string `json:"name"     validate:"required,min=1,max=191"`
}

type CreateSubClassRequest struct {
	ClassID string `json:"class_id" validate:"required,uuid"`
	Name    string `json:"name"     validate:"required,min=1,max=191"`
}

// ── Academic Year DTOs ────────────────────────────────────────────────────────

type CreateAcademicYearRequest struct {
	Name        string `json:"name"        validate:"required,min=1,max=191"`
	StartDate   string `json:"start_date"  validate:"required"`
	EndDate     string `json:"end_date"    validate:"required"`
	Description string `json:"description"`
}

type UpdateAcademicYearRequest struct {
	Name        string `json:"name"        validate:"required,min=1,max=191"`
	StartDate   string `json:"start_date"  validate:"required"`
	EndDate     string `json:"end_date"    validate:"required"`
	Description string `json:"description"`
}

type AcademicYearResponse struct {
	ID          string    `json:"id"`
	SchoolID    string    `json:"school_id"`
	Name        string    `json:"name"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	IsActive    bool      `json:"is_active"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ── Subject DTOs ──────────────────────────────────────────────────────────────

type CreateSubjectRequest struct {
	Name     string `json:"name"     validate:"required,min=1,max=191"`
	Category string `json:"category" validate:"required,oneof=core elective extracurricular specialized vocational"`
}

type UpdateSubjectRequest struct {
	Name     string `json:"name"     validate:"required,min=1,max=191"`
	Category string `json:"category" validate:"required,oneof=core elective extracurricular specialized vocational"`
}

type SubjectResponse struct {
	ID        string    `json:"id"`
	SchoolID  string    `json:"school_id"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ── Classroom DTOs ────────────────────────────────────────────────────────────

type CreateClassroomRequest struct {
	Name       string `json:"name"       validate:"required,min=1,max=191"`
	Capacity   int    `json:"capacity"   validate:"min=0"`
	Floor      int    `json:"floor"      validate:"min=0"`
	Building   string `json:"building"`
	RoomType   string `json:"room_type"`
	Status     string `json:"status"     validate:"omitempty,oneof=available occupied maintenance"`
	Facilities string `json:"facilities"`
}

type UpdateClassroomRequest struct {
	Name       string `json:"name"       validate:"required,min=1,max=191"`
	Capacity   int    `json:"capacity"   validate:"min=0"`
	Floor      int    `json:"floor"      validate:"min=0"`
	Building   string `json:"building"`
	RoomType   string `json:"room_type"`
	Status     string `json:"status"     validate:"omitempty,oneof=available occupied maintenance"`
	Facilities string `json:"facilities"`
}

type ClassroomResponse struct {
	ID         string    `json:"id"`
	SchoolID   string    `json:"school_id"`
	Name       string    `json:"name"`
	Capacity   int       `json:"capacity"`
	Floor      int       `json:"floor"`
	Building   string    `json:"building"`
	RoomType   string    `json:"room_type"`
	Status     string    `json:"status"`
	Facilities string    `json:"facilities"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ── Schedule DTOs ─────────────────────────────────────────────────────────────

type CreateScheduleRequest struct {
	ShiftType string `json:"shift_type" validate:"required,oneof=morning afternoon full_day"`
	StartTime string `json:"start_time" validate:"required"`
	EndTime   string `json:"end_time"   validate:"required"`
}

type UpdateScheduleRequest struct {
	ShiftType string `json:"shift_type" validate:"required,oneof=morning afternoon full_day"`
	StartTime string `json:"start_time" validate:"required"`
	EndTime   string `json:"end_time"   validate:"required"`
}

type ScheduleResponse struct {
	ID        string    `json:"id"`
	SchoolID  string    `json:"school_id"`
	ShiftType string    `json:"shift_type"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
