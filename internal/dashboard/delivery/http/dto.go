package http

import (
	"time"

	"github.com/eduaccess/eduaccess-api/internal/dashboard/domain"
	"github.com/google/uuid"
)

// DashboardStatsResponse mirrors the dashboard aggregation payload.
type DashboardStatsResponse struct {
	School       DashboardSchoolResponse        `json:"school"`
	Counts       DashboardCountsResponse        `json:"counts"`
	Attendance   DashboardAttendanceResponse    `json:"attendance"`
	Subscription *DashboardSubscriptionResponse `json:"subscription,omitempty"`
}

type DashboardSchoolResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Status   string    `json:"status"`
	TimeZone string    `json:"time_zone"`
}

type DashboardCountsResponse = domain.Counts

type DashboardAttendanceResponse = domain.AttendanceSummary

type DashboardSubscriptionResponse struct {
	PlanName string     `json:"plan_name"`
	Status   string     `json:"status"`
	Cycle    string     `json:"cycle"`
	Price    int64      `json:"price"`
	EndsAt   *time.Time `json:"ends_at,omitempty"`
}

func toDashboardStatsResponse(stats *domain.Stats) DashboardStatsResponse {
	resp := DashboardStatsResponse{
		School: DashboardSchoolResponse{
			ID:       stats.School.ID,
			Name:     stats.School.Name,
			Status:   stats.School.Status,
			TimeZone: stats.School.TimeZone,
		},
		Counts:     stats.Counts,
		Attendance: stats.Attendance,
	}
	if stats.Subscription != nil {
		resp.Subscription = &DashboardSubscriptionResponse{
			PlanName: stats.Subscription.PlanName,
			Status:   stats.Subscription.Status,
			Cycle:    stats.Subscription.Cycle,
			Price:    stats.Subscription.Price,
			EndsAt:   stats.Subscription.EndsAt,
		}
	}
	return resp
}