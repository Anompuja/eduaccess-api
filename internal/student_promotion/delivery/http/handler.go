package http

import (
	"errors"
	"net/http"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/shared/httpcache"
	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	"github.com/eduaccess/eduaccess-api/internal/shared/validator"
	"github.com/eduaccess/eduaccess-api/internal/student_promotion/application"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	listPromotions *application.ListPromotionsHandler
	promote        *application.PromoteHandler
}

func NewHandler(
	v1 *echo.Group,
	listPromotions *application.ListPromotionsHandler,
	promote *application.PromoteHandler,
) *Handler {
	h := &Handler{listPromotions: listPromotions, promote: promote}

	// POST mutates (cache middleware auto-skips it); GET history is short-lived.
	g := v1.Group("/student-promotions", authmw.RequireAuth, httpcache.Middleware(httpcache.ShortLived))
	g.GET("", h.ListPromotions)
	g.POST("/promote", h.Promote)

	return h
}

func (h *Handler) ListPromotions(c echo.Context) error {
	q := application.ListPromotionsQuery{
		RequesterSchoolID: getSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		StudentID:         optionalUUID(c.QueryParam("student_id")),
		AcademicYearID:    optionalUUID(c.QueryParam("academic_year_id")),
	}
	rows, err := h.listPromotions.Handle(c.Request().Context(), q)
	if err != nil {
		return handleAppError(c, err)
	}
	dtos := make([]PromotionResponse, 0, len(rows))
	for _, v := range rows {
		dtos = append(dtos, toPromotionResponse(v))
	}
	return response.OK(c, "promotions retrieved", dtos)
}

func (h *Handler) Promote(c echo.Context) error {
	var req PromoteRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}
	toClassroomID, err := uuid.Parse(req.ToClassroomID)
	if err != nil {
		return response.BadRequest(c, "invalid to_classroom_id")
	}
	studentIDs := make([]uuid.UUID, 0, len(req.StudentIDs))
	for _, raw := range req.StudentIDs {
		id, perr := uuid.Parse(raw)
		if perr != nil {
			return response.BadRequest(c, "invalid student id: "+raw)
		}
		studentIDs = append(studentIDs, id)
	}

	result, err := h.promote.Handle(c.Request().Context(), application.PromoteCommand{
		RequesterSchoolID: getSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		StudentIDs:        studentIDs,
		ToClassroomID:     toClassroomID,
		Status:            req.Status,
		Notes:             req.Notes,
		PromotionDate:     parsePromotionDate(req.PromotionDate),
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "promotion processed", result)
}

// ── Shared utilities ──────────────────────────────────────────────────────────

func optionalUUID(raw string) *uuid.UUID {
	if raw == "" {
		return nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil
	}
	return &id
}

func getSchoolID(c echo.Context) *uuid.UUID {
	if id := authmw.GetSchoolID(c); id != nil {
		return id
	}
	if authmw.GetRole(c) == "superadmin" {
		if raw := c.QueryParam("school_id"); raw != "" {
			if id, err := uuid.Parse(raw); err == nil {
				return &id
			}
		}
	}
	return nil
}

func handleAppError(c echo.Context, err error) error {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		switch appErr.Err {
		case apperror.ErrNotFound:
			return response.NotFound(c, appErr.Message)
		case apperror.ErrForbidden:
			return response.Forbidden(c, appErr.Message)
		case apperror.ErrBadRequest, apperror.ErrWrongPassword:
			return response.BadRequest(c, appErr.Message)
		case apperror.ErrUnauthorized, apperror.ErrInvalidToken, apperror.ErrTokenRevoked:
			return response.Unauthorized(c, appErr.Message)
		case apperror.ErrConflict:
			return response.Conflict(c, appErr.Message)
		}
	}
	return c.JSON(http.StatusInternalServerError, response.Response{Success: false, Message: "internal server error"})
}
