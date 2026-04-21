package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	"github.com/eduaccess/eduaccess-api/internal/shared/validator"
	"github.com/eduaccess/eduaccess-api/internal/student/application"
	"github.com/eduaccess/eduaccess-api/internal/student/domain"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler wires student and parent use-cases to HTTP endpoints.
type Handler struct {
	// Students
	createStudent     *application.CreateStudentHandler
	listStudents      *application.ListStudentsHandler
	getStudent        *application.GetStudentHandler
	updateStudent     *application.UpdateStudentHandler
	deactivateStudent *application.DeactivateStudentHandler
	linkParent        *application.LinkParentHandler
	unlinkParent      *application.UnlinkParentHandler
	// Parents
	createParent     *application.CreateParentHandler
	listParents      *application.ListParentsHandler
	getParent        *application.GetParentHandler
	updateParent     *application.UpdateParentHandler
	deactivateParent *application.DeactivateParentHandler
}

// NewHandler registers all student routes and returns the handler.
func NewHandler(
	v1 *echo.Group,
	createStudent *application.CreateStudentHandler,
	listStudents *application.ListStudentsHandler,
	getStudent *application.GetStudentHandler,
	updateStudent *application.UpdateStudentHandler,
	deactivateStudent *application.DeactivateStudentHandler,
	linkParent *application.LinkParentHandler,
	unlinkParent *application.UnlinkParentHandler,
	createParent *application.CreateParentHandler,
	listParents *application.ListParentsHandler,
	getParent *application.GetParentHandler,
	updateParent *application.UpdateParentHandler,
	deactivateParent *application.DeactivateParentHandler,
) *Handler {
	h := &Handler{
		createStudent:     createStudent,
		listStudents:      listStudents,
		getStudent:        getStudent,
		updateStudent:     updateStudent,
		deactivateStudent: deactivateStudent,
		linkParent:        linkParent,
		unlinkParent:      unlinkParent,
		createParent:      createParent,
		listParents:       listParents,
		getParent:         getParent,
		updateParent:      updateParent,
		deactivateParent:  deactivateParent,
	}

	auth := authmw.RequireAuth

	// Students
	h.registerStudentRoutes(v1, auth)

	return h
}

// CreateParent godoc
//
//	@Summary      Create parent
//	@Description  Creates a user account (role=orangtua) and parent profile atomically.
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
	parent, err := h.createParent.Handle(c.Request().Context(), application.CreateParentCommand{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		Name:              req.Name,
		Email:             req.Email,
		Username:          req.Username,
		Password:          req.Password,
		FatherName:        req.FatherName,
		MotherName:        req.MotherName,
		FatherReligion:    req.FatherReligion,
		MotherReligion:    req.MotherReligion,
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
//	@Description  Returns a paginated list of parents. Tenant-scoped.
//	@Tags         parents
//	@Produce      json
//	@Security     BearerAuth
//	@Param        search   query  string  false  "Search by name or email"
//	@Param        page     query  int     false  "Page number (default 1)"
//	@Param        per_page query  int     false  "Page size (default 20)"
//	@Success      200  {object}  response.PaginatedResponse{data=[]ParentResponse}
//	@Router       /parents [get]
func (h *Handler) ListParents(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	result, err := h.listParents.Handle(c.Request().Context(), application.ListParentsQuery{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
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
//	@Tags         parents
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id    path      string             true  "Parent profile UUID"
//	@Param        body  body      UpdateParentRequest true  "Fields to update"
//	@Success      200   {object}  response.Response{data=ParentResponse}
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
		FatherName:        req.FatherName,
		MotherName:        req.MotherName,
		FatherReligion:    req.FatherReligion,
		MotherReligion:    req.MotherReligion,
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
//	@Tags         parents
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "Parent profile UUID"
//	@Success      200  {object}  response.Response
//	@Router       /parents/{id} [delete]
func (h *Handler) DeactivateParent(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	if err := h.deactivateParent.Handle(c.Request().Context(), application.DeactivateParentCommand{
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
		ID:             p.ID.String(),
		UserID:         p.UserID.String(),
		SchoolID:       p.SchoolID.String(),
		Name:           p.Name,
		Email:          p.Email,
		Username:       p.Username,
		Avatar:         p.Avatar,
		FatherName:     p.FatherName,
		MotherName:     p.MotherName,
		FatherReligion: p.FatherReligion,
		MotherReligion: p.MotherReligion,
		PhoneNumber:    p.PhoneNumber,
		Address:        p.Address,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
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
