package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/admin/application"
	"github.com/eduaccess/eduaccess-api/internal/admin/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	"github.com/eduaccess/eduaccess-api/internal/shared/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler wires admin use-cases to HTTP endpoints.
type Handler struct {
	createAdmin     *application.CreateAdminHandler
	getAdmin        *application.GetAdminHandler
	listAdmins      *application.ListAdminsHandler
	updateAdmin     *application.UpdateAdminHandler
	deactivateAdmin *application.DeactivateAdminHandler
}

// NewHandler registers admin routes and returns the handler.
func NewHandler(
	v1 *echo.Group,
	createAdmin *application.CreateAdminHandler,
	getAdmin *application.GetAdminHandler,
	listAdmins *application.ListAdminsHandler,
	updateAdmin *application.UpdateAdminHandler,
	deactivateAdmin *application.DeactivateAdminHandler,
) *Handler {
	h := &Handler{
		createAdmin:     createAdmin,
		getAdmin:        getAdmin,
		listAdmins:      listAdmins,
		updateAdmin:     updateAdmin,
		deactivateAdmin: deactivateAdmin,
	}

	admins := v1.Group("/admins", authmw.RequireAuth)
	admins.POST("", h.CreateAdmin)
	admins.GET("", h.ListAdmins)
	admins.GET("/:id", h.GetAdmin)
	admins.PUT("/:id", h.UpdateAdmin)
	admins.DELETE("/:id", h.DeactivateAdmin)

	return h
}

// GetAdmin godoc
//
//	@Summary      Get admin by ID
//	@Tags         admins
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "Admin profile UUID"
//	@Success      200  {object}  response.Response{data=AdminResponse}
//	@Failure      403  {object}  response.Response
//	@Failure      404  {object}  response.Response
//	@Router       /admins/{id} [get]
func (h *Handler) GetAdmin(c echo.Context) error {
	adminID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}

	admin, err := h.getAdmin.Handle(c.Request().Context(), application.GetAdminQuery{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		AdminID:           adminID,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "admin retrieved", toAdminResponse(admin))
}

// ListAdmins godoc
//
//	@Summary      List admins
//	@Description  Returns a paginated list of admins. Tenant-scoped.
//	@Tags         admins
//	@Produce      json
//	@Security     BearerAuth
//	@Param        search   query  string  false  "Search by name, email or username"
//	@Param        page     query  int     false  "Page number (default 1)"
//	@Param        per_page query  int     false  "Page size (default 20)"
//	@Success      200  {object}  response.PaginatedResponse{data=[]AdminResponse}
//	@Router       /admins [get]
func (h *Handler) ListAdmins(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	q := application.ListAdminsQuery{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		Search:            c.QueryParam("search"),
		Page:              page,
		PerPage:           perPage,
	}

	result, err := h.listAdmins.Handle(c.Request().Context(), q)
	if err != nil {
		return handleAppError(c, err)
	}
	dtos := make([]AdminResponse, 0, len(result.Admins))
	for _, a := range result.Admins {
		dtos = append(dtos, toAdminResponse(a))
	}
	return response.Paginated(c, "admins retrieved", dtos, result.Page, result.PerPage, result.Total)
}

// CreateAdmin godoc
//
//	@Summary      Create admin sekolah
//	@Description  Creates a user account (role=admin_sekolah) and admin profile atomically.
//	@Tags         admins
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      CreateAdminRequest  true  "Admin data"
//	@Success      201   {object}  response.Response{data=AdminResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      403   {object}  response.Response
//	@Failure      409   {object}  response.Response
//	@Router       /admins [post]
func (h *Handler) CreateAdmin(c echo.Context) error {
	var req CreateAdminRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	cmd := application.CreateAdminCommand{
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

	admin, err := h.createAdmin.Handle(c.Request().Context(), cmd)
	if err != nil {
		return handleAppError(c, err)
	}

	return c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Message: "admin created",
		Data:    toAdminResponse(admin),
	})
}

// UpdateAdmin godoc
//
//	@Summary      Update admin sekolah
//	@Description  Updates mutable fields in users and admin_profiles.
//	@Tags         admins
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id    path      string              true  "Admin profile UUID"
//	@Param        body  body      UpdateAdminRequest  true  "Admin fields to update"
//	@Success      200   {object}  response.Response{data=AdminResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      403   {object}  response.Response
//	@Failure      404   {object}  response.Response
//	@Failure      409   {object}  response.Response
//	@Router       /admins/{id} [put]
func (h *Handler) UpdateAdmin(c echo.Context) error {
	adminID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}

	var req UpdateAdminRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	cmd := application.UpdateAdminCommand{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		AdminID:           adminID,
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

	admin, err := h.updateAdmin.Handle(c.Request().Context(), cmd)
	if err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "admin updated", toAdminResponse(admin))
}

// DeactivateAdmin godoc
//
//	@Summary      Deactivate admin
//	@Tags         admins
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "Admin profile UUID"
//	@Success      200  {object}  response.Response
//	@Failure      403  {object}  response.Response
//	@Failure      404  {object}  response.Response
//	@Router       /admins/{id} [delete]
func (h *Handler) DeactivateAdmin(c echo.Context) error {
	adminID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}

	if err := h.deactivateAdmin.Handle(c.Request().Context(), application.DeactivateAdminCommand{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		AdminID:           adminID,
	}); err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "admin deactivated", nil)
}

func toAdminResponse(a *domain.AdminProfile) AdminResponse {
	return AdminResponse{
		ID:           a.ID.String(),
		UserID:       a.UserID.String(),
		SchoolID:     a.SchoolID.String(),
		Name:         a.Name,
		Email:        a.Email,
		Username:     a.Username,
		Avatar:       a.Avatar,
		PhoneNumber:  a.PhoneNumber,
		Address:      a.Address,
		Gender:       a.Gender,
		Religion:     a.Religion,
		BirthPlace:   a.BirthPlace,
		BirthDate:    a.BirthDate,
		NIK:          a.NIK,
		KTPImagePath: a.KTPImagePath,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
}

// parseDateField parses an optional *string "YYYY-MM-DD" into *time.Time.
func parseDateField(src *string, dst **time.Time) error {
	if src == nil || *src == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *src)
	if err != nil {
		return err
	}
	*dst = &t
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
