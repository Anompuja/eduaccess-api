package http

import "github.com/eduaccess/eduaccess-api/internal/student_tracking/domain"

type StudyResponse struct {
	ID               string  `json:"id"`
	StudentID        string  `json:"student_id"`
	StudentName      string  `json:"student_name"`
	NIS              string  `json:"nis"`
	ClassroomID      string  `json:"classroom_id"`
	ClassroomName    string  `json:"classroom_name"`
	ClassID          *string `json:"class_id"`
	ClassName        string  `json:"class_name"`
	SubClassName     string  `json:"sub_class_name"`
	AcademicYearID   string  `json:"academic_year_id"`
	AcademicYearName string  `json:"academic_year_name"`
	Status           string  `json:"status"`
	EnrollmentDate   string  `json:"enrollment_date"`
}

func toStudyResponse(v domain.StudyView) StudyResponse {
	var classID *string
	if v.ClassID != nil {
		s := v.ClassID.String()
		classID = &s
	}
	return StudyResponse{
		ID:               v.ID.String(),
		StudentID:        v.StudentID.String(),
		StudentName:      v.StudentName,
		NIS:              v.NIS,
		ClassroomID:      v.ClassroomID.String(),
		ClassroomName:    v.ClassroomName,
		ClassID:          classID,
		ClassName:        v.ClassName,
		SubClassName:     v.SubClassName,
		AcademicYearID:   v.AcademicYearID.String(),
		AcademicYearName: v.AcademicYearName,
		Status:           v.Status,
		EnrollmentDate:   v.EnrollmentDate.Format("2006-01-02"),
	}
}

func toStudyResponses(views []domain.StudyView) []StudyResponse {
	out := make([]StudyResponse, 0, len(views))
	for _, v := range views {
		out = append(out, toStudyResponse(v))
	}
	return out
}
