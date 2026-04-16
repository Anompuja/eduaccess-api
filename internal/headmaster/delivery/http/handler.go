package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/eduaccess/eduaccess-api/internal/headmaster/application"
	"github.com/eduaccess/eduaccess-api/internal/headmaster/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	"github.com/eduaccess/eduaccess-api/internal/shared/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler wires headmaster use-cases to HTTP endpoints.
type Handler struct {
	create     *application.CreateHeadmasterHandler
	list       *application.ListHeadmastersHandler
	get        *application.GetHeadmasterHandler
	update     *application.UpdateHeadmasterHandler
	deactivate *application.DeactivateHeadmasterHandler
}

// NewHandler registers headmaster routes under /api/v1 and returns the handler.
func NewHandler(
	v1 *echo.Group,
	create *application.CreateHeadmasterHandler,
	list *application.ListHeadmastersHandler,
	get *application.GetHeadmasterHandler,
	update *application.UpdateHeadmasterHandler,
	deactivate *application.DeactivateHeadmasterHandler,
) *Handler {
	h := &Handler{
		create:     create,
		list:       list,
		get:        get,
		update:     update,
		deactivate: deactivate,
	}

	hm := v1.Group("/headmasters", authmw.RequireAuth)
	hm.POST("", h.Create)
	hm.GET("", h.List)
	hm.GET("/:id", h.Get)
	hm.PUT("/:id", h.Update)
	hm.DELETE("/:id", h.Deactivate)

	return h
}

// Create godoc
//
//	@Summary      Create headmaster
//	@Description  Creates a user account (kepala_sekolah) and headmaster profile, then sets the school's current headmaster.
//	@Tags         headmasters
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      CreateHeadmasterRequest  true  "Headmaster data"
//	@Success      201   {object}  response.Response{data=HeadmasterResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      403   {object}  response.Response
//	@Failure      409   {object}  response.Response
//	@Router       /headmasters [post]
func (h *Handler) Create(c echo.Context) error {
	var req CreateHeadmasterRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	var schoolID *uuid.UUID
	if req.SchoolID != "" {
		parsed, err := uuid.Parse(req.SchoolID)
		if err != nil {
			return response.BadRequest(c, "invalid school_id")
		}
		schoolID = &parsed
	}

	profile, err := h.create.Handle(c.Request().Context(), application.CreateHeadmasterCommand{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		SchoolID:          schoolID,
		Name:              req.Name,
		Email:             req.Email,
		Username:          req.Username,
		Password:          req.Password,
		PhoneNumber:       req.PhoneNumber,
		Address:           req.Address,
		Gender:            req.Gender,
		Religion:          req.Religion,
		BirthPlace:        req.BirthPlace,
		BirthDate:         req.BirthDate,
		NIK:               req.NIK,
		KTPImagePath:      req.KTPImagePath,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	return c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Message: "headmaster created",
		Data:    toHeadmasterResponse(profile),
	})
}

// List godoc
//
//	@Summary      List headmasters
//	@Description  Returns a paginated list of headmaster profiles scoped to the requester's school.
//	@Tags         headmasters
//	@Produce      json
//	@Security     BearerAuth
//	@Param        search   query  string  false  "Search by name, email or username"
//	@Param        page     query  int     false  "Page number (default 1)"
//	@Param        per_page query  int     false  "Page size (default 20, max 100)"
//	@Success      200  {object}  response.PaginatedResponse{data=[]HeadmasterResponse}
//	@Failure      401  {object}  response.Response
//	@Failure      403  {object}  response.Response
//	@Router       /headmasters [get]
func (h *Handler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	result, err := h.list.Handle(c.Request().Context(), application.ListHeadmastersQuery{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		Search:            c.QueryParam("search"),
		Page:              page,
		PerPage:           perPage,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	dtos := make([]HeadmasterResponse, 0, len(result.Headmasters))
	for _, p := range result.Headmasters {
		dtos = append(dtos, toHeadmasterResponse(p))
	}

	return response.Paginated(c, "headmasters retrieved", dtos, result.Page, result.PerPage, result.Total)
}

// Get godoc
//
//	@Summary      Get headmaster by ID
//	@Description  Returns a single headmaster profile.
//	@Tags         headmasters
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "HeadmasterProfile UUID"
//	@Success      200  {object}  response.Response{data=HeadmasterResponse}
//	@Failure      403  {object}  response.Response
//	@Failure      404  {object}  response.Response
//	@Router       /headmasters/{id} [get]
func (h *Handler) Get(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}

	profile, err := h.get.Handle(c.Request().Context(), application.GetHeadmasterQuery{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		HeadmasterID:      id,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "headmaster retrieved", toHeadmasterResponse(profile))
}

// Update godoc
//
//	@Summary      Update headmaster profile
//	@Description  Updates mutable fields of a headmaster profile.
//	@Tags         headmasters
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id    path      string                  true  "HeadmasterProfile UUID"
//	@Param        body  body      UpdateHeadmasterRequest true  "Fields to update"
//	@Success      200   {object}  response.Response{data=HeadmasterResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      403   {object}  response.Response
//	@Failure      404   {object}  response.Response
//	@Router       /headmasters/{id} [put]
func (h *Handler) Update(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}

	var req UpdateHeadmasterRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	profile, err := h.update.Handle(c.Request().Context(), application.UpdateHeadmasterCommand{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		HeadmasterID:      id,
		PhoneNumber:       req.PhoneNumber,
		Address:           req.Address,
		Gender:            req.Gender,
		Religion:          req.Religion,
		BirthPlace:        req.BirthPlace,
		BirthDate:         req.BirthDate,
		NIK:               req.NIK,
		KTPImagePath:      req.KTPImagePath,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "headmaster updated", toHeadmasterResponse(profile))
}

// Deactivate godoc
//
//	@Summary      Deactivate headmaster
//	@Description  Soft-deletes a headmaster profile.
//	@Tags         headmasters
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "HeadmasterProfile UUID"
//	@Success      200  {object}  response.Response
//	@Failure      403  {object}  response.Response
//	@Failure      404  {object}  response.Response
//	@Router       /headmasters/{id} [delete]
func (h *Handler) Deactivate(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}

	if err := h.deactivate.Handle(c.Request().Context(), application.DeactivateHeadmasterCommand{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		HeadmasterID:      id,
	}); err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "headmaster deactivated", nil)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func toHeadmasterResponse(p *domain.HeadmasterProfile) HeadmasterResponse {
	return HeadmasterResponse{
		ID:           p.ID.String(),
		UserID:       p.UserID.String(),
		SchoolID:     p.SchoolID.String(),
		Name:         p.Name,
		Email:        p.Email,
		Username:     p.Username,
		Avatar:       p.Avatar,
		PhoneNumber:  p.PhoneNumber,
		Address:      p.Address,
		Gender:       p.Gender,
		Religion:     p.Religion,
		BirthPlace:   p.BirthPlace,
		BirthDate:    p.BirthDate,
		NIK:          p.NIK,
		KTPImagePath: p.KTPImagePath,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

func parseUUID(c echo.Context, param string) (uuid.UUID, error) {
	raw := c.Param(param)
	id, err := uuid.Parse(raw)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response.Response{
			Success: false,
			Message: "invalid UUID: " + param,
		})
		return uuid.UUID{}, echo.ErrBadRequest
	}
	return id, nil
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
	// sentinel errors (ErrNotFound used directly without wrapping)
	if errors.Is(err, apperror.ErrNotFound) {
		return response.NotFound(c, "not found")
	}
	return c.JSON(http.StatusInternalServerError, response.Response{
		Success: false,
		Message: "internal server error",
	})
}
