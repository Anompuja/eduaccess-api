package http

import "time"

// ── Schedule DTOs ─────────────────────────────────────────────────────────────

type CreateClassScheduleRequest struct {
	ClassroomID   string `json:"classroom_id"    validate:"required,uuid"`
	SubjectID     string `json:"subject_id"      validate:"required,uuid"`
	TeacherID     string `json:"teacher_id"      validate:"required,uuid"`
	StartPeriodID string `json:"start_period_id" validate:"omitempty,uuid"`
	EndPeriodID   string `json:"end_period_id"   validate:"omitempty,uuid"`
	Date          string `json:"date"            validate:"required"` // YYYY-MM-DD
	StartTime     string `json:"start_time"      validate:"required"` // HH:MM
	EndTime       string `json:"end_time"        validate:"required"` // HH:MM
}

type UpdateClassScheduleRequest struct {
	ClassroomID   string `json:"classroom_id"    validate:"required,uuid"`
	SubjectID     string `json:"subject_id"      validate:"required,uuid"`
	TeacherID     string `json:"teacher_id"      validate:"required,uuid"`
	StartPeriodID string `json:"start_period_id" validate:"omitempty,uuid"`
	EndPeriodID   string `json:"end_period_id"   validate:"omitempty,uuid"`
	Date          string `json:"date"            validate:"required"`
	StartTime     string `json:"start_time"      validate:"required"`
	EndTime       string `json:"end_time"        validate:"required"`
}

type ClassScheduleResponse struct {
	ID                    string     `json:"id"`
	SchoolID              string     `json:"school_id"`
	ClassroomID           string     `json:"classroom_id"`
	SubjectID             string     `json:"subject_id"`
	TeacherID             string     `json:"teacher_id"`
	StartPeriodID         *string    `json:"start_period_id"`
	EndPeriodID           *string    `json:"end_period_id"`
	Date                  string     `json:"date"`
	StartTime             string     `json:"start_time"`
	EndTime               string     `json:"end_time"`
	TeacherAttendanceTime *time.Time `json:"teacher_attendance_time"`
	Status                string     `json:"status"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// ── Attendance DTOs ───────────────────────────────────────────────────────────

type UpdateAttendanceRequest struct {
	Status    string `json:"status"     validate:"required,oneof=present sick permission absent scheduled"`
	Note      string `json:"note"`
	PhotoPath string `json:"photo_path"`
}

type AttendanceResponse struct {
	ID                    string     `json:"id"`
	ClassScheduleID       string     `json:"class_schedule_id"`
	StudentID             string     `json:"student_id"`
	Status                string     `json:"status"`
	Type                  string     `json:"type"`
	Note                  string     `json:"note"`
	PhotoPath             string     `json:"photo_path"`
	StudentAttendanceTime *time.Time `json:"student_attendance_time"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}
