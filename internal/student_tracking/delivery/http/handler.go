package http

import (
	"errors"
	"net/http"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/shared/httpcache"
	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	"github.com/eduaccess/eduaccess-api/internal/student_tracking/application"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	listStudies      *application.ListStudiesHandler
	getStudentDetail *application.GetStudentDetailHandler
}

func NewHandler(
	v1 *echo.Group,
	listStudies *application.ListStudiesHandler,
	getStudentDetail *application.GetStudentDetailHandler,
) *Handler {
	h := &Handler{listStudies: listStudies, getStudentDetail: getStudentDetail}

	// Read-only, school-scoped data — safe to cache with ETag revalidation.
	g := v1.Group("/student-studies", authmw.RequireAuth, httpcache.Middleware(httpcache.AlwaysRevalidate))
	g.GET("", h.ListStudies)
	g.GET("/:student_id", h.GetStudentDetail)

	return h
}

func (h *Handler) ListStudies(c echo.Context) error {
	q := application.ListStudiesQuery{
		RequesterSchoolID: getSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		ClassroomID:       optionalUUID(c.QueryParam("classroom_id")),
		AcademicYearID:    optionalUUID(c.QueryParam("academic_year_id")),
		ClassID:           optionalUUID(c.QueryParam("class_id")),
		StudentID:         optionalUUID(c.QueryParam("student_id")),
	}
	if raw := c.QueryParam("status"); raw != "" {
		q.Status = &raw
	}
	rows, err := h.listStudies.Handle(c.Request().Context(), q)
	if err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "student studies retrieved", toStudyResponses(rows))
}

func (h *Handler) GetStudentDetail(c echo.Context) error {
	studentID, err := parseUUID(c, "student_id")
	if err != nil {
		return handleAppError(c, err)
	}
	rows, err := h.getStudentDetail.Handle(c.Request().Context(), application.GetStudentDetailQuery{
		RequesterSchoolID: getSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		StudentID:         studentID,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "student detail retrieved", toStudyResponses(rows))
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

func parseUUID(c echo.Context, param string) (uuid.UUID, error) {
	raw := c.Param(param)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.UUID{}, apperror.New(apperror.ErrBadRequest, "invalid UUID: "+param)
	}
	return id, nil
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
