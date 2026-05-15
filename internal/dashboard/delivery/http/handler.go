package http

import (
	"errors"
	"net/http"

	"github.com/eduaccess/eduaccess-api/internal/dashboard/application"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler wires dashboard stats to HTTP endpoints.
type Handler struct {
	getStats *application.GetStatsHandler
}

// NewHandler registers dashboard routes and returns the handler.
func NewHandler(v1 *echo.Group, getStats *application.GetStatsHandler) *Handler {
	h := &Handler{getStats: getStats}

	dashboard := v1.Group("/dashboard", authmw.RequireAuth)
	dashboard.GET("/stats", h.GetStats)

	return h
}

// GetStats godoc
//
//	@Summary      Get dashboard stats
//	@Description  Returns a school summary with counts for users, academics, attendance, and subscription status.
//	@Tags         dashboard
//	@Produce      json
//	@Security     BearerAuth
//	@Param        school_id  query     string  false  "School UUID (superadmin only)"
//	@Success      200        {object}  response.Response{data=DashboardStatsResponse}
//	@Failure      400        {object}  response.Response
//	@Failure      403        {object}  response.Response
//	@Failure      404        {object}  response.Response
//	@Router       /dashboard/stats [get]
func (h *Handler) GetStats(c echo.Context) error {
	var schoolID *uuid.UUID
	if raw := c.QueryParam("school_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			return response.BadRequest(c, "invalid school_id")
		}
		schoolID = &parsed
	}

	stats, err := h.getStats.Handle(c.Request().Context(), application.GetStatsQuery{
		RequesterRole:     authmw.GetRole(c),
		RequesterSchoolID: authmw.GetSchoolID(c),
		SchoolID:          schoolID,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "dashboard stats retrieved", toDashboardStatsResponse(stats))
}

func handleAppError(c echo.Context, err error) error {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		switch appErr.Err {
		case apperror.ErrNotFound:
			return response.NotFound(c, appErr.Message)
		case apperror.ErrUnauthorized, apperror.ErrInvalidToken:
			return response.Unauthorized(c, appErr.Message)
		case apperror.ErrForbidden:
			return response.Forbidden(c, appErr.Message)
		case apperror.ErrConflict:
			return response.Conflict(c, appErr.Message)
		case apperror.ErrBadRequest:
			return response.BadRequest(c, appErr.Message)
		}
	}
	return c.JSON(http.StatusInternalServerError, response.Response{
		Success: false,
		Message: "internal server error",
	})
}