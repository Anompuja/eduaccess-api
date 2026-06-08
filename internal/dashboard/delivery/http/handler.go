package http

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/eduaccess/eduaccess-api/internal/dashboard/application"
	"github.com/eduaccess/eduaccess-api/internal/dashboard/infrastructure"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler wires dashboard stats to HTTP endpoints.
type Handler struct {
	getStats *application.GetStatsHandler
	cache    *infrastructure.DashboardCache
}

// NewHandler registers dashboard routes and returns the handler.
func NewHandler(v1 *echo.Group, getStats *application.GetStatsHandler, cache *infrastructure.DashboardCache) *Handler {
	h := &Handler{getStats: getStats, cache: cache}

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

	effectiveSchoolID := schoolID
	if authmw.GetRole(c) != "superadmin" {
		effectiveSchoolID = authmw.GetSchoolID(c)
	}

	schoolIDStr := "all"
	if effectiveSchoolID != nil {
		schoolIDStr = effectiveSchoolID.String()
	}

	cacheKey := fmt.Sprintf("dashboard:stats:%s:%s", authmw.GetRole(c), schoolIDStr)

	if h.cache != nil {
		if cachedResp, found := h.cache.Get(cacheKey); found {
			cacheData := cachedResp.(map[string]interface{})
			etag := cacheData["etag"].(string)

			c.Response().Header().Set("Cache-Control", "private, max-age=30, must-revalidate")
			c.Response().Header().Set("ETag", `"`+etag+`"`)
			c.Response().Header().Set("Vary", "Authorization")

			if match := c.Request().Header.Get("If-None-Match"); match != "" {
				if match == `"`+etag+`"` {
					return c.NoContent(http.StatusNotModified)
				}
			}

			return c.JSON(http.StatusOK, cacheData["response"])
		}
	}

	stats, err := h.getStats.Handle(c.Request().Context(), application.GetStatsQuery{
		RequesterRole:     authmw.GetRole(c),
		RequesterSchoolID: authmw.GetSchoolID(c),
		SchoolID:          schoolID,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	resp := response.Response{
		Success: true,
		Message: "dashboard stats retrieved",
		Data:    toDashboardStatsResponse(stats),
	}

	respBytes, _ := json.Marshal(resp)
	hash := sha256.Sum256(respBytes)
	etag := hex.EncodeToString(hash[:])

	if h.cache != nil {
		h.cache.Set(cacheKey, map[string]interface{}{
			"response": resp,
			"etag":     etag,
		})
	}

	c.Response().Header().Set("Cache-Control", "private, max-age=30, must-revalidate")
	c.Response().Header().Set("ETag", `"`+etag+`"`)
	c.Response().Header().Set("Vary", "Authorization")

	return c.JSON(http.StatusOK, resp)
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
