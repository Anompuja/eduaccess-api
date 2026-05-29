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
	EducationLevelID string  `json:"education_level_id" validate:"omitempty,uuid"`
	Name             string  `json:"name"               validate:"required,min=1,max=191"`
	Code             *string `json:"code"               validate:"omitempty,max=50"`
	Category         string  `json:"category"           validate:"required,oneof=core elective extracurricular specialized vocational"`
}

type UpdateSubjectRequest struct {
	EducationLevelID string  `json:"education_level_id" validate:"omitempty,uuid"`
	Name             string  `json:"name"               validate:"required,min=1,max=191"`
	Code             *string `json:"code"               validate:"omitempty,max=50"`
	Category         string  `json:"category"           validate:"required,oneof=core elective extracurricular specialized vocational"`
}

type SubjectResponse struct {
	ID               string    `json:"id"`
	SchoolID         string    `json:"school_id"`
	EducationLevelID *string   `json:"education_level_id"`
	Name             string    `json:"name"`
	Code             *string   `json:"code"`
	Category         string    `json:"category"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ── Classroom DTOs ────────────────────────────────────────────────────────────

type CreateClassroomRequest struct {
	ClassID           string `json:"class_id"            validate:"omitempty,uuid"`
	SubClassID        string `json:"sub_class_id"        validate:"omitempty,uuid"`
	AcademicYearID    string `json:"academic_year_id"    validate:"omitempty,uuid"`
	HomeroomTeacherID string `json:"homeroom_teacher_id" validate:"omitempty,uuid"`
	Name              string `json:"name"                validate:"required,min=1,max=191"`
	CodeRoom          string `json:"code_room"           validate:"omitempty,max=50"`
	Capacity          int    `json:"capacity"            validate:"min=0"`
	Floor             string `json:"floor"               validate:"omitempty,max=50"`
	Building          string `json:"building"`
	RoomType          string `json:"room_type"`
	Status            string `json:"status"              validate:"omitempty,oneof=unknown available occupied maintenance"`
	Facilities        string `json:"facilities"`
}

type UpdateClassroomRequest struct {
	ClassID           string `json:"class_id"            validate:"omitempty,uuid"`
	SubClassID        string `json:"sub_class_id"        validate:"omitempty,uuid"`
	AcademicYearID    string `json:"academic_year_id"    validate:"omitempty,uuid"`
	HomeroomTeacherID string `json:"homeroom_teacher_id" validate:"omitempty,uuid"`
	Name              string `json:"name"                validate:"required,min=1,max=191"`
	CodeRoom          string `json:"code_room"           validate:"omitempty,max=50"`
	Capacity          int    `json:"capacity"            validate:"min=0"`
	Floor             string `json:"floor"               validate:"omitempty,max=50"`
	Building          string `json:"building"`
	RoomType          string `json:"room_type"`
	Status            string `json:"status"              validate:"omitempty,oneof=unknown available occupied maintenance"`
	Facilities        string `json:"facilities"`
}

type ClassroomResponse struct {
	ID                string    `json:"id"`
	SchoolID          string    `json:"school_id"`
	ClassID           *string   `json:"class_id"`
	SubClassID        *string   `json:"sub_class_id"`
	AcademicYearID    *string   `json:"academic_year_id"`
	HomeroomTeacherID *string   `json:"homeroom_teacher_id"`
	Name              string    `json:"name"`
	CodeRoom          string    `json:"code_room"`
	Capacity          int       `json:"capacity"`
	Floor             string    `json:"floor"`
	Building          string    `json:"building"`
	RoomType          string    `json:"room_type"`
	Status            string    `json:"status"`
	Facilities        string    `json:"facilities"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ── Schedule DTOs ─────────────────────────────────────────────────────────────

type CreateScheduleRequest struct {
	DayOfWeek    string `json:"day_of_week"   validate:"required,oneof=monday tuesday wednesday thursday friday saturday sunday"`
	PeriodNumber int    `json:"period_number" validate:"required,min=1"`
	Label        string `json:"label"         validate:"required,min=1,max=50"`
	StartTime    string `json:"start_time"    validate:"required,datetime=15:04"`
	EndTime      string `json:"end_time"      validate:"required,datetime=15:04"`
	IsBreak      bool   `json:"is_break"`
}

type UpdateScheduleRequest struct {
	DayOfWeek    string `json:"day_of_week"   validate:"required,oneof=monday tuesday wednesday thursday friday saturday sunday"`
	PeriodNumber int    `json:"period_number" validate:"required,min=1"`
	Label        string `json:"label"         validate:"required,min=1,max=50"`
	StartTime    string `json:"start_time"    validate:"required,datetime=15:04"`
	EndTime      string `json:"end_time"      validate:"required,datetime=15:04"`
	IsBreak      bool   `json:"is_break"`
}

type ScheduleResponse struct {
	ID           string    `json:"id"`
	SchoolID     string    `json:"school_id"`
	DayOfWeek    string    `json:"day_of_week"`
	PeriodNumber int       `json:"period_number"`
	Label        string    `json:"label"`
	StartTime    string    `json:"start_time"`
	EndTime      string    `json:"end_time"`
	IsBreak      bool      `json:"is_break"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
