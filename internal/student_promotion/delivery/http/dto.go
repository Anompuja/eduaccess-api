package http

import (
	"time"

	"github.com/eduaccess/eduaccess-api/internal/student_promotion/domain"
)

type PromoteRequest struct {
	StudentIDs    []string `json:"student_ids"     validate:"required,min=1,dive,uuid"`
	ToClassroomID string   `json:"to_classroom_id" validate:"required,uuid"`
	Status        string   `json:"status"          validate:"required,oneof=promoted retained transferred"`
	Notes         string   `json:"notes"`
	PromotionDate string   `json:"promotion_date"  validate:"omitempty,datetime=2006-01-02"`
}

type PromotionResponse struct {
	ID                string `json:"id"`
	StudentID         string `json:"student_id"`
	StudentName       string `json:"student_name"`
	NIS               string `json:"nis"`
	FromClassroomID   string `json:"from_classroom_id"`
	FromClassroomName string `json:"from_classroom_name"`
	ToClassroomID     string `json:"to_classroom_id"`
	ToClassroomName   string `json:"to_classroom_name"`
	AcademicYearID    string `json:"academic_year_id"`
	AcademicYearName  string `json:"academic_year_name"`
	PromotionDate     string `json:"promotion_date"`
	Status            string `json:"status"`
	Notes             string `json:"notes"`
}

func toPromotionResponse(v domain.PromotionView) PromotionResponse {
	return PromotionResponse{
		ID:                v.ID.String(),
		StudentID:         v.StudentID.String(),
		StudentName:       v.StudentName,
		NIS:               v.NIS,
		FromClassroomID:   v.FromClassroomID.String(),
		FromClassroomName: v.FromClassroomName,
		ToClassroomID:     v.ToClassroomID.String(),
		ToClassroomName:   v.ToClassroomName,
		AcademicYearID:    v.AcademicYearID.String(),
		AcademicYearName:  v.AcademicYearName,
		PromotionDate:     v.PromotionDate.Format("2006-01-02"),
		Status:            v.Status,
		Notes:             v.Notes,
	}
}

func parsePromotionDate(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}
	}
	return t
}
