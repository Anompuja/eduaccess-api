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
	"github.com/labstack/echo/v4"
)

// Handler wires parent use-cases to HTTP endpoints.
type Handler struct {
	listParents *application.ListParentsHandler
}

// NewHandler registers parent routes and returns the handler.
func NewHandler(
	v1 *echo.Group,
	listParents *application.ListParentsHandler,
) *Handler {
	h := &Handler{
		listParents: listParents,
	}

	parents := v1.Group("/parents", authmw.RequireAuth)
	parents.GET("", h.ListParents)

	return h
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
