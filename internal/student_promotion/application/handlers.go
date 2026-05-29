package application

import (
	"context"
	"errors"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/student_promotion/domain"
	"github.com/google/uuid"
)

// ── Guards ────────────────────────────────────────────────────────────────────

func guardView(role string) error {
	switch role {
	case "superadmin", "admin_sekolah", "kepala_sekolah", "guru", "staff":
		return nil
	}
	return apperror.New(apperror.ErrForbidden, "not allowed to view promotions")
}

// guardManage restricts promotion (a structural academic action) to school
// leadership and the platform superadmin.
func guardManage(role string) error {
	switch role {
	case "superadmin", "admin_sekolah", "kepala_sekolah":
		return nil
	}
	return apperror.New(apperror.ErrForbidden, "not allowed to promote students")
}

func guardSchoolMatch(role string, requesterSchoolID *uuid.UUID, targetSchoolID uuid.UUID) error {
	if role == "superadmin" {
		return nil
	}
	if requesterSchoolID == nil {
		return apperror.New(apperror.ErrForbidden, "school context required")
	}
	if *requesterSchoolID != targetSchoolID {
		return apperror.New(apperror.ErrForbidden, "access denied to this resource")
	}
	return nil
}

// ── List promotions (history) ─────────────────────────────────────────────────

type ListPromotionsQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	StudentID         *uuid.UUID
	AcademicYearID    *uuid.UUID
}

type ListPromotionsHandler struct{ repo domain.Repository }

func NewListPromotionsHandler(repo domain.Repository) *ListPromotionsHandler {
	return &ListPromotionsHandler{repo: repo}
}

func (h *ListPromotionsHandler) Handle(ctx context.Context, q ListPromotionsQuery) ([]domain.PromotionView, error) {
	if err := guardView(q.RequesterRole); err != nil {
		return nil, err
	}
	return h.repo.ListPromotions(ctx, domain.PromotionFilter{
		SchoolID:       q.RequesterSchoolID,
		StudentID:      q.StudentID,
		AcademicYearID: q.AcademicYearID,
	})
}

// ── Promote (kenaikan kelas) ───────────────────────────────────────────────────

type PromoteCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	StudentIDs        []uuid.UUID
	ToClassroomID     uuid.UUID
	Status            string
	Notes             string
	PromotionDate     time.Time
}

type PromoteResult struct {
	Success int                `json:"success"`
	Failed  int                `json:"failed"`
	Errors  []PromoteItemError `json:"errors,omitempty"`
}

type PromoteItemError struct {
	StudentID string `json:"student_id"`
	Reason    string `json:"reason"`
}

type PromoteHandler struct{ repo domain.Repository }

func NewPromoteHandler(repo domain.Repository) *PromoteHandler {
	return &PromoteHandler{repo: repo}
}

func (h *PromoteHandler) Handle(ctx context.Context, cmd PromoteCommand) (*PromoteResult, error) {
	if err := guardManage(cmd.RequesterRole); err != nil {
		return nil, err
	}
	if cmd.RequesterSchoolID == nil {
		return nil, apperror.New(apperror.ErrBadRequest, "school context required")
	}
	if len(cmd.StudentIDs) == 0 {
		return nil, apperror.New(apperror.ErrBadRequest, "student_ids is required")
	}

	target, err := h.repo.FindClassroomTarget(ctx, cmd.ToClassroomID)
	if err != nil {
		return nil, err
	}
	if err := guardSchoolMatch(cmd.RequesterRole, cmd.RequesterSchoolID, target.SchoolID); err != nil {
		return nil, err
	}

	promotionDate := cmd.PromotionDate
	if promotionDate.IsZero() {
		promotionDate = time.Now()
	}

	result := &PromoteResult{}
	for _, studentID := range cmd.StudentIDs {
		err := h.repo.PromoteStudent(ctx, domain.PromotionInput{
			SchoolID:      target.SchoolID,
			StudentID:     studentID,
			Target:        *target,
			PromotionDate: promotionDate,
			Status:        cmd.Status,
			Notes:         cmd.Notes,
		})
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, PromoteItemError{
				StudentID: studentID.String(),
				Reason:    messageOf(err),
			})
			continue
		}
		result.Success++
	}
	return result, nil
}

func messageOf(err error) string {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		return appErr.Message
	}
	return "promotion failed"
}
