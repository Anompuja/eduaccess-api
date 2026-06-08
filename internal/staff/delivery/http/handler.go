package http

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	"github.com/eduaccess/eduaccess-api/internal/shared/validator"
	"github.com/eduaccess/eduaccess-api/internal/staff/application"
	"github.com/eduaccess/eduaccess-api/internal/staff/domain"
	"github.com/eduaccess/eduaccess-api/internal/staff/infrastructure"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler wires staff use-cases to HTTP endpoints.
type Handler struct {
	createStaff     *application.CreateStaffHandler
	getStaff        *application.GetStaffHandler
	listStaff       *application.ListStaffHandler
	updateStaff     *application.UpdateStaffHandler
	deactivateStaff *application.DeactivateStaffHandler
	staffCache      *infrastructure.StaffCache
}

// NewHandler registers staff routes and returns the handler.
func NewHandler(
	v1 *echo.Group,
	createStaff *application.CreateStaffHandler,
	getStaff *application.GetStaffHandler,
	listStaff *application.ListStaffHandler,
	updateStaff *application.UpdateStaffHandler,
	deactivateStaff *application.DeactivateStaffHandler,
	staffCache *infrastructure.StaffCache,
) *Handler {
	h := &Handler{
		createStaff:     createStaff,
		getStaff:        getStaff,
		listStaff:       listStaff,
		updateStaff:     updateStaff,
		deactivateStaff: deactivateStaff,
		staffCache:      staffCache,
	}

	staff := v1.Group("/staff", authmw.RequireAuth)
	staff.POST("", h.CreateStaff)
	staff.GET("", h.ListStaff)
	staff.GET("/:id", h.GetStaff)
	staff.PUT("/:id", h.UpdateStaff)
	staff.DELETE("/:id", h.DeactivateStaff)

	return h
}

// GetStaff godoc
//
//	@Summary      Get staff by ID
//	@Tags         staff
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "Staff profile UUID"
//	@Success      200  {object}  response.Response{data=StaffResponse}
//	@Failure      403  {object}  response.Response
//	@Failure      404  {object}  response.Response
//	@Router       /staff/{id} [get]
func (h *Handler) GetStaff(c echo.Context) error {
	staffID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}

	staff, err := h.getStaff.Handle(c.Request().Context(), application.GetStaffQuery{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		StaffID:           staffID,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "staff retrieved", toStaffResponse(staff))
}

// ListStaff godoc
//
//	@Summary      List staff
//	@Description  Returns a paginated list of staff. Superadmin may filter by school_id; admin_sekolah is scoped to their own school.
//	@Tags         staff
//	@Produce      json
//	@Security     BearerAuth
//	@Param        school_id query  string  false  "School UUID (superadmin only)"
//	@Param        search   query  string  false  "Search by name, email or username"
//	@Param        page     query  int     false  "Page number (default 1)"
//	@Param        per_page query  int     false  "Page size (default 20)"
//	@Success      200  {object}  response.PaginatedResponse{data=[]StaffResponse}
//	@Router       /staff [get]
func (h *Handler) ListStaff(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	var schoolID *uuid.UUID
	if rawSchoolID := c.QueryParam("school_id"); rawSchoolID != "" {
		parsedSchoolID, err := uuid.Parse(rawSchoolID)
		if err != nil {
			return response.BadRequest(c, "invalid school_id")
		}
		schoolID = &parsedSchoolID
	}

	q := application.ListStaffQuery{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		SchoolID:          schoolID,
		Search:            c.QueryParam("search"),
		Page:              page,
		PerPage:           perPage,
	}

	schoolIDStr := "all"
	if schoolID != nil {
		schoolIDStr = schoolID.String()
	}
	cacheKey := fmt.Sprintf("staff:list:%s:%s:%d:%d:%s", q.RequesterRole, schoolIDStr, page, perPage, q.Search)

	if cachedResp, found := h.staffCache.Get(cacheKey); found {
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

	result, err := h.listStaff.Handle(c.Request().Context(), q)
	if err != nil {
		return handleAppError(c, err)
	}

	dtos := make([]StaffResponse, 0, len(result.Staff))
	for _, s := range result.Staff {
		dtos = append(dtos, toStaffResponse(s))
	}

	totalPages := int(result.Total) / result.PerPage
	if int(result.Total)%result.PerPage != 0 {
		totalPages++
	}

	resp := response.PaginatedResponse{
		Success: true,
		Message: "staff retrieved",
		Data:    dtos,
		Pagination: response.Pagination{
			Page:       result.Page,
			PerPage:    result.PerPage,
			Total:      result.Total,
			TotalPages: totalPages,
		},
	}

	respBytes, _ := json.Marshal(resp)
	hash := sha256.Sum256(respBytes)
	etag := hex.EncodeToString(hash[:])

	h.staffCache.Set(cacheKey, map[string]interface{}{
		"response": resp,
		"etag":     etag,
	})

	c.Response().Header().Set("Cache-Control", "private, max-age=30, must-revalidate")
	c.Response().Header().Set("ETag", `"`+etag+`"`)
	c.Response().Header().Set("Vary", "Authorization")

	return c.JSON(http.StatusOK, resp)
}

// CreateStaff godoc
//
//	@Summary      Create staff
//	@Description  Creates a user account (role=staff) and staff profile atomically.
//	@Tags         staff
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      CreateStaffRequest  true  "Staff data"
//	@Success      201   {object}  response.Response{data=StaffResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      403   {object}  response.Response
//	@Failure      409   {object}  response.Response
//	@Router       /staff [post]
func (h *Handler) CreateStaff(c echo.Context) error {
	var req CreateStaffRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	cmd := application.CreateStaffCommand{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		Name:              req.Name,
		Email:             req.Email,
		Username:          req.Username,
		Password:          req.Password,
		PhoneNumber:       req.PhoneNumber,
		Address:           req.Address,
		Gender:            req.Gender,
		Religion:          req.Religion,
		BirthPlace:        req.BirthPlace,
		NIK:               req.NIK,
		KTPImagePath:      req.KTPImagePath,
	}

	if err := parseDateField(req.BirthDate, &cmd.BirthDate); err != nil {
		return response.BadRequest(c, "birth_date must be YYYY-MM-DD")
	}
	if err := parseUUIDField(req.SchoolID, &cmd.SchoolID); err != nil {
		return response.BadRequest(c, "invalid school_id")
	}

	staff, err := h.createStaff.Handle(c.Request().Context(), cmd)
	if err != nil {
		return handleAppError(c, err)
	}

	h.staffCache.InvalidatePrefix("staff:list:")

	return c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Message: "staff created",
		Data:    toStaffResponse(staff),
	})
}

// UpdateStaff godoc
//
//	@Summary      Update staff
//	@Description  Updates mutable fields in users and staff_profiles.
//	@Tags         staff
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id    path      string              true  "Staff profile UUID"
//	@Param        body  body      UpdateStaffRequest  true  "Staff fields to update"
//	@Success      200   {object}  response.Response{data=StaffResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      403   {object}  response.Response
//	@Failure      404   {object}  response.Response
//	@Failure      409   {object}  response.Response
//	@Router       /staff/{id} [put]
func (h *Handler) UpdateStaff(c echo.Context) error {
	staffID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}

	var req UpdateStaffRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	cmd := application.UpdateStaffCommand{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		StaffID:           staffID,
		Name:              req.Name,
		Email:             req.Email,
		Username:          req.Username,
		PhoneNumber:       req.PhoneNumber,
		Address:           req.Address,
		Gender:            req.Gender,
		Religion:          req.Religion,
		BirthPlace:        req.BirthPlace,
		NIK:               req.NIK,
		KTPImagePath:      req.KTPImagePath,
	}

	if err := parseDateField(req.BirthDate, &cmd.BirthDate); err != nil {
		return response.BadRequest(c, "birth_date must be YYYY-MM-DD")
	}

	staff, err := h.updateStaff.Handle(c.Request().Context(), cmd)
	if err != nil {
		return handleAppError(c, err)
	}

	h.staffCache.InvalidatePrefix("staff:list:")

	return response.OK(c, "staff updated", toStaffResponse(staff))
}

// DeactivateStaff godoc
//
//	@Summary      Deactivate staff
//	@Tags         staff
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "Staff profile UUID"
//	@Success      200  {object}  response.Response
//	@Failure      403  {object}  response.Response
//	@Failure      404  {object}  response.Response
//	@Router       /staff/{id} [delete]
func (h *Handler) DeactivateStaff(c echo.Context) error {
	staffID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}

	if err := h.deactivateStaff.Handle(c.Request().Context(), application.DeactivateStaffCommand{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		StaffID:           staffID,
	}); err != nil {
		return handleAppError(c, err)
	}

	h.staffCache.InvalidatePrefix("staff:list:")

	return response.OK(c, "staff deactivated", nil)
}

func toStaffResponse(s *domain.StaffProfile) StaffResponse {
	return StaffResponse{
		ID:           s.ID.String(),
		UserID:       s.UserID.String(),
		SchoolID:     s.SchoolID.String(),
		Name:         s.Name,
		Email:        s.Email,
		Username:     s.Username,
		Avatar:       s.Avatar,
		PhoneNumber:  s.PhoneNumber,
		Address:      s.Address,
		Gender:       s.Gender,
		Religion:     s.Religion,
		BirthPlace:   s.BirthPlace,
		BirthDate:    s.BirthDate,
		NIK:          s.NIK,
		KTPImagePath: s.KTPImagePath,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}
}

// parseDateField parses an optional *string "YYYY-MM-DD" into *string.
func parseDateField(src *string, dst **string) error {
	if src == nil || *src == "" {
		return nil
	}
	_, err := time.Parse("2006-01-02", *src)
	if err != nil {
		return err
	}
	*dst = src
	return nil
}

func parseUUIDField(src *string, dst **uuid.UUID) error {
	if src == nil || *src == "" {
		return nil
	}
	id, err := uuid.Parse(*src)
	if err != nil {
		return err
	}
	*dst = &id
	return nil
}

func handleAppError(c echo.Context, err error) error {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		switch appErr.Err {
		case apperror.ErrNotFound:
			return response.NotFound(c, appErr.Message)
		case apperror.ErrUnauthorized, apperror.ErrInvalidToken, apperror.ErrTokenRevoked:
			return response.Unauthorized(c, appErr.Message)
		case apperror.ErrForbidden:
			return response.Forbidden(c, appErr.Message)
		case apperror.ErrConflict:
			return response.Conflict(c, appErr.Message)
		case apperror.ErrBadRequest, apperror.ErrWrongPassword:
			return response.BadRequest(c, appErr.Message)
		}
	}

	return c.JSON(http.StatusInternalServerError, response.Response{
		Success: false,
		Message: "internal server error",
	})
}
