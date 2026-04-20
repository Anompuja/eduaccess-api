package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/eduaccess/eduaccess-api/internal/parent/application"
	"github.com/eduaccess/eduaccess-api/internal/parent/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
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
}

// NewHandler registers parent routes and returns the handler.
func NewHandler(
	v1 *echo.Group,
	createParent *application.CreateParentHandler,
	listParents *application.ListParentsHandler,
	getParent *application.GetParentHandler,
) *Handler {
	h := &Handler{
		createParent: createParent,
		listParents:  listParents,
		getParent:    getParent,
	}

	parents := v1.Group("/parents", authmw.RequireAuth)
	parents.POST("", h.CreateParent)
	parents.GET("", h.ListParents)
	parents.GET("/:id", h.GetParent)

	return h
}

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
