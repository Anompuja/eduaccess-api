package http

import (
	"errors"
	"net/http"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	"github.com/eduaccess/eduaccess-api/internal/student/application"
	"github.com/eduaccess/eduaccess-api/internal/student/domain"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler wires student and parent-link use-cases to HTTP endpoints.
type Handler struct {
	createStudent     *application.CreateStudentHandler
	listStudents      *application.ListStudentsHandler
	getStudent        *application.GetStudentHandler
	updateStudent     *application.UpdateStudentHandler
	deactivateStudent *application.DeactivateStudentHandler
	linkParent        *application.LinkParentHandler
	unlinkParent      *application.UnlinkParentHandler
}

// NewHandler registers student routes and returns the handler.
func NewHandler(
	v1 *echo.Group,
	createStudent *application.CreateStudentHandler,
	listStudents *application.ListStudentsHandler,
	getStudent *application.GetStudentHandler,
	updateStudent *application.UpdateStudentHandler,
	deactivateStudent *application.DeactivateStudentHandler,
	linkParent *application.LinkParentHandler,
	unlinkParent *application.UnlinkParentHandler,
) *Handler {
	h := &Handler{
		createStudent:     createStudent,
		listStudents:      listStudents,
		getStudent:        getStudent,
		updateStudent:     updateStudent,
		deactivateStudent: deactivateStudent,
		linkParent:        linkParent,
		unlinkParent:      unlinkParent,
	}

	h.registerStudentRoutes(v1, authmw.RequireAuth)

	return h
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
