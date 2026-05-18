package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository defines the dashboard statistics use case.
// A nil schoolID means "aggregate across all schools" — only valid for superadmin.
type Repository interface {
	GetStats(ctx context.Context, schoolID *uuid.UUID) (*Stats, error)
}

// Stats is the aggregated dashboard payload for one school.
type Stats struct {
	School       SchoolSummary         `json:"school"`
	Counts       Counts                `json:"counts"`
	Attendance   AttendanceSummary     `json:"attendance"`
	Subscription *SubscriptionSummary `json:"subscription,omitempty"`
}

// SchoolSummary contains the school identity shown at the top of the dashboard.
type SchoolSummary struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Status   string    `json:"status"`
	TimeZone string    `json:"time_zone"`
}

// Counts groups the core totals the frontend can display as cards.
type Counts struct {
	SchoolUsers       int64 `json:"school_users"`
	Admins            int64 `json:"admins"`
	Headmasters       int64 `json:"headmasters"`
	Teachers          int64 `json:"teachers"`
	Staff             int64 `json:"staff"`
	Parents           int64 `json:"parents"`
	Students          int64 `json:"students"`
	ActiveStudents    int64 `json:"active_students"`
	Enrollments       int64 `json:"enrollments"`
	ActiveEnrollments int64 `json:"active_enrollments"`
	AcademicYears     int64 `json:"academic_years"`
	EducationLevels   int64 `json:"education_levels"`
	Classes           int64 `json:"classes"`
	SubClasses        int64 `json:"sub_classes"`
	Classrooms        int64 `json:"classrooms"`
	Subjects          int64 `json:"subjects"`
	Schedules         int64 `json:"schedules"`
}

// AttendanceSummary summarizes today's attendance outcome.
type AttendanceSummary struct {
	Present int64   `json:"present"`
	Late    int64   `json:"late"`
	Absent  int64   `json:"absent"`
	Excused int64   `json:"excused"`
	Total   int64   `json:"total"`
	Rate    float64 `json:"rate"`
}

// SubscriptionSummary shows the active school subscription snapshot.
type SubscriptionSummary struct {
	PlanName string     `json:"plan_name"`
	Status   string     `json:"status"`
	Cycle    string     `json:"cycle"`
	Price    int64      `json:"price"`
	EndsAt   *time.Time `json:"ends_at,omitempty"`
}