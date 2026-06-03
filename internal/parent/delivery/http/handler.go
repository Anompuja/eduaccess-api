package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/eduaccess/eduaccess-api/internal/parent/application"
	"github.com/eduaccess/eduaccess-api/internal/parent/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/shared/httpcache"
	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	"github.com/eduaccess/eduaccess-api/internal/shared/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler wires parent use-cases to HTTP endpoints.
type Handler struct {
	createParent *application.CreateParentHandler
	listParents  *application.ListParentsHandler
	getParent    *application.GetParentHandler
	updateParent *application.UpdateParentHandler
	deactivate   *application.DeactivateParentHandler
}

// NewHandler registers parent routes and returns the handler.
func NewHandler(
	v1 *echo.Group,
	createParent *application.CreateParentHandler,
	listParents *application.ListParentsHandler,
	getParent *application.GetParentHandler,
	updateParent *application.UpdateParentHandler,
	deactivate *application.DeactivateParentHandler,
) *Handler {
	h := &Handler{
		createParent: createParent,
		listParents:  listParents,
		getParent:    getParent,
		updateParent: updateParent,
		deactivate:   deactivate,
	}

	parents := v1.Group("/parents", authmw.RequireAuth, httpcache.Middleware(httpcache.ShortLived))
	parents.POST("", h.CreateParent)
	parents.GET("", h.ListParents)
	parents.GET("/:id", h.GetParent)
	parents.PUT("/:id", h.UpdateParent)
	parents.DELETE("/:id", h.DeactivateParent)

	return h
}

// CreateParent godoc
//
//	@Summary      Create parent
//	@Description  Creates a parent profile. Superadmin may provide school_id in the request body; admin_sekolah is scoped to their own school.
//	@Tags         parents
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      CreateParentRequest  true  "Parent data"
//	@Success      201   {object}  response.Response{data=ParentResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      403   {object}  response.Response
//	@Failure      409   {object}  response.Response
//	@Router       /parents [post]
func (h *Handler) CreateParent(c echo.Context) error {
	var req CreateParentRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	var schoolID *uuid.UUID
	if req.SchoolID != nil && *req.SchoolID != "" {
		id, err := uuid.Parse(*req.SchoolID)
		if err != nil {
			return response.BadRequest(c, "invalid school_id")
		}
		schoolID = &id
	}

	parent, err := h.createParent.Handle(c.Request().Context(), application.CreateParentCommand{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		SchoolID:          schoolID,
		Name:              req.Name,
		Email:             req.Email,
		Username:          req.Username,
		Password:          req.Password,
		Religion:          req.Religion,
		PhoneNumber:       req.PhoneNumber,
		Address:           req.Address,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	return c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Message: "parent created",
		Data:    toParentResponse(parent),
	})
}

// ListParents godoc
//
//	@Summary      List parents
//	@Description  Returns a paginated list of parents. Superadmin may filter by school_id; admin_sekolah is scoped to their own school.
//	@Tags         parents
//	@Produce      json
//	@Security     BearerAuth
//	@Param        school_id query  string  false  "School UUID (superadmin only)"
//	@Param        search    query  string  false  "Search by name, email or username"
//	@Param        page      query  int     false  "Page number (default 1)"
//	@Param        per_page  query  int     false  "Page size (default 20, max 100)"
//	@Success      200       {object}  response.PaginatedResponse{data=[]ParentResponse}
//	@Failure      400       {object}  response.Response
//	@Failure      403       {object}  response.Response
//	@Router       /parents [get]
func (h *Handler) ListParents(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	var schoolFilter *uuid.UUID
	if raw := c.QueryParam("school_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			return response.BadRequest(c, "invalid school_id")
		}
		schoolFilter = &parsed
	}

	result, err := h.listParents.Handle(c.Request().Context(), application.ListParentsQuery{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		SchoolIDFilter:    schoolFilter,
		Search:            c.QueryParam("search"),
		Page:              page,
		PerPage:           perPage,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	dtos := make([]ParentResponse, 0, len(result.Parents))
	for _, p := range result.Parents {
		dtos = append(dtos, toParentResponse(p))
	}
	return response.Paginated(c, "parents retrieved", dtos, result.Page, result.PerPage, result.Total)
}

// GetParent godoc
//
//	@Summary      Get parent by ID
//	@Description  Returns a single parent profile.
//	@Tags         parents
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "Parent profile UUID"
//	@Success      200  {object}  response.Response{data=ParentResponse}
//	@Failure      403  {object}  response.Response
//	@Failure      404  {object}  response.Response
//	@Router       /parents/{id} [get]
func (h *Handler) GetParent(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}

	parent, err := h.getParent.Handle(c.Request().Context(), application.GetParentQuery{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		ParentID:          id,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "parent retrieved", toParentResponse(parent))
}

// UpdateParent godoc
//
//	@Summary      Update parent
//	@Description  Updates mutable fields of a parent profile.
//	@Tags         parents
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id    path      string               true  "Parent profile UUID"
//	@Param        body  body      UpdateParentRequest  true  "Parent fields to update"
//	@Success      200   {object}  response.Response{data=ParentResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      403   {object}  response.Response
//	@Failure      404   {object}  response.Response
//	@Router       /parents/{id} [put]
func (h *Handler) UpdateParent(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}

	var req UpdateParentRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	parent, err := h.updateParent.Handle(c.Request().Context(), application.UpdateParentCommand{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		ParentID:          id,
		Name:              req.Name,
		Email:             req.Email,
		Religion:          req.Religion,
		PhoneNumber:       req.PhoneNumber,
		Address:           req.Address,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "parent updated", toParentResponse(parent))
}

// DeactivateParent godoc
//
//	@Summary      Deactivate parent
//	@Description  Soft-deletes a parent profile.
//	@Tags         parents
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "Parent profile UUID"
//	@Success      200  {object}  response.Response
//	@Failure      403  {object}  response.Response
//	@Failure      404  {object}  response.Response
//	@Router       /parents/{id} [delete]
func (h *Handler) DeactivateParent(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}

	if err := h.deactivate.Handle(c.Request().Context(), application.DeactivateParentCommand{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		ParentID:          id,
	}); err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "parent deactivated", nil)
}

func toParentResponse(p *domain.ParentProfile) ParentResponse {
	return ParentResponse{
		ID:          p.ID.String(),
		UserID:      p.UserID.String(),
		SchoolID:    p.SchoolID.String(),
		Name:        p.Name,
		Email:       p.Email,
		Username:    p.Username,
		Avatar:      p.Avatar,
		Religion:    p.Religion,
		PhoneNumber: p.PhoneNumber,
		Address:     p.Address,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
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

	return c.JSON(http.StatusInternalServerError, response.Response{
		Success: false,
		Message: "internal server error",
	})
}
