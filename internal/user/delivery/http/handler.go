package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	"github.com/eduaccess/eduaccess-api/internal/shared/validator"
	"github.com/eduaccess/eduaccess-api/internal/user/application"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler wires user management use-cases to HTTP endpoints.
type Handler struct {
	listUsers      *application.ListUsersHandler
	getUser        *application.GetUserHandler
	updateUser     *application.UpdateUserHandler
	deactivateUser *application.DeactivateUserHandler
	changePassword *application.ChangePasswordHandler
}

// NewHandler registers user management routes and returns the handler.
func NewHandler(
	v1 *echo.Group,
	listUsers *application.ListUsersHandler,
	getUser *application.GetUserHandler,
	updateUser *application.UpdateUserHandler,
	deactivateUser *application.DeactivateUserHandler,
	changePassword *application.ChangePasswordHandler,
) *Handler {
	h := &Handler{
		listUsers:      listUsers,
		getUser:        getUser,
		updateUser:     updateUser,
		deactivateUser: deactivateUser,
		changePassword: changePassword,
	}

	// All user routes require auth
	users := v1.Group("/users", authmw.RequireAuth)
	users.GET("", h.ListUsers)
	users.GET("/:id", h.GetUser)
	users.PUT("/:id", h.UpdateUser)
	users.DELETE("/:id", h.DeactivateUser)
	users.PUT("/:id/password", h.ChangePassword)

	// Profile: authenticated user's own record
	profile := v1.Group("/profile", authmw.RequireAuth)
	profile.GET("", h.GetProfile)
	profile.PUT("", h.UpdateProfile)

	return h
}

// ListUsers godoc
//
//	@Summary      List users
//	@Description  Returns a paginated list of users. Admin sekolah sees only their school. Superadmin sees all.
//	@Tags         users
//	@Produce      json
//	@Security     BearerAuth
//	@Param        role    query     string  false  "Filter by role"
//	@Param        search  query     string  false  "Search by name, email or username"
//	@Param        page    query     int     false  "Page number (default 1)"
//	@Param        per_page query    int     false  "Page size (default 20, max 100)"
//	@Success      200   {object}  response.PaginatedResponse{data=[]UserResponse}
//	@Failure      401   {object}  response.Response
//	@Failure      403   {object}  response.Response
//	@Router       /users [get]
func (h *Handler) ListUsers(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	result, err := h.listUsers.Handle(c.Request().Context(), application.ListUsersQuery{
		SchoolID: authmw.GetSchoolID(c),
		Role:     c.QueryParam("role"),
		Search:   c.QueryParam("search"),
		Page:     page,
		PerPage:  perPage,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	dtos := make([]UserResponse, 0, len(result.Users))
	for _, u := range result.Users {
		dtos = append(dtos, toUserResponse(u))
	}

	return response.Paginated(c, "users retrieved", dtos, result.Page, result.PerPage, result.Total)
}

// GetUser godoc
//
//	@Summary      Get user by ID
//	@Description  Returns a single user. Admin sekolah can only fetch users within their school.
//	@Tags         users
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "User UUID"
//	@Success      200  {object}  response.Response{data=UserResponse}
//	@Failure      403  {object}  response.Response
//	@Failure      404  {object}  response.Response
//	@Router       /users/{id} [get]
func (h *Handler) GetUser(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}

	user, err := h.getUser.Handle(c.Request().Context(), application.GetUserQuery{
		RequesterID:       authmw.GetUserID(c),
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		UserID:            id,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "user retrieved", toUserResponse(user))
}

// UpdateUser godoc
//
//	@Summary      Update user
//	@Description  Updates name and/or avatar. Admin sekolah can only update users in their school.
//	@Tags         users
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id    path      string            true  "User UUID"
//	@Param        body  body      UpdateUserRequest true  "Fields to update"
//	@Success      200   {object}  response.Response{data=UserResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      403   {object}  response.Response
//	@Failure      404   {object}  response.Response
//	@Router       /users/{id} [put]
func (h *Handler) UpdateUser(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}

	var req UpdateUserRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	user, err := h.updateUser.Handle(c.Request().Context(), application.UpdateUserCommand{
		RequesterID:       authmw.GetUserID(c),
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		UserID:            id,
		Name:              req.Name,
		Avatar:            req.Avatar,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "user updated", toUserResponse(user))
}

// DeactivateUser godoc
//
//	@Summary      Deactivate user
//	@Description  Soft-deletes a user. Admin sekolah cannot deactivate other admins.
//	@Tags         users
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "User UUID"
//	@Success      200  {object}  response.Response
//	@Failure      400  {object}  response.Response
//	@Failure      403  {object}  response.Response
//	@Failure      404  {object}  response.Response
//	@Router       /users/{id} [delete]
func (h *Handler) DeactivateUser(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}

	if err := h.deactivateUser.Handle(c.Request().Context(), application.DeactivateUserCommand{
		RequesterID:       authmw.GetUserID(c),
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		UserID:            id,
	}); err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "user deactivated", nil)
}

// ChangePassword godoc
//
//	@Summary      Change password
//	@Description  Changes a user's password. Users must supply their current password; superadmin can skip it.
//	@Tags         users
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id    path      string                true  "User UUID"
//	@Param        body  body      ChangePasswordRequest true  "Passwords"
//	@Success      200   {object}  response.Response
//	@Failure      400   {object}  response.Response
//	@Failure      403   {object}  response.Response
//	@Router       /users/{id}/password [put]
func (h *Handler) ChangePassword(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}

	var req ChangePasswordRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	if err := h.changePassword.Handle(c.Request().Context(), application.ChangePasswordCommand{
		RequesterID:   authmw.GetUserID(c),
		RequesterRole: authmw.GetRole(c),
		UserID:        id,
		OldPassword:   req.OldPassword,
		NewPassword:   req.NewPassword,
	}); err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "password changed", nil)
}

// GetProfile godoc
//
//	@Summary      Get own profile
//	@Description  Returns the profile of the currently authenticated user.
//	@Tags         profile
//	@Produce      json
//	@Security     BearerAuth
//	@Success      200  {object}  response.Response{data=UserResponse}
//	@Failure      401  {object}  response.Response
//	@Router       /profile [get]
func (h *Handler) GetProfile(c echo.Context) error {
	userID := authmw.GetUserID(c)

	user, err := h.getUser.Handle(c.Request().Context(), application.GetUserQuery{
		RequesterID:       userID,
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		UserID:            userID,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "profile retrieved", toUserResponse(user))
}

// UpdateProfile godoc
//
//	@Summary      Update own profile
//	@Description  Updates the authenticated user's name and/or avatar.
//	@Tags         profile
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      UpdateUserRequest true  "Fields to update"
//	@Success      200   {object}  response.Response{data=UserResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      422   {object}  response.Response
//	@Router       /profile [put]
func (h *Handler) UpdateProfile(c echo.Context) error {
	var req UpdateUserRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	user, err := h.updateUser.Handle(c.Request().Context(), application.UpdateUserCommand{
		RequesterID:       authmw.GetUserID(c),
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		UserID:            authmw.GetUserID(c),
		Name:              req.Name,
		Avatar:            req.Avatar,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "profile updated", toUserResponse(user))
}

// ── helpers ───────────────────────────────────────────────────────────────────

func toUserResponse(u *domain.User) UserResponse {
	dto := UserResponse{
		ID:        u.ID.String(),
		Role:      u.Role,
		Name:      u.Name,
		Username:  u.Username,
		Email:     u.Email,
		Avatar:    u.Avatar,
		Verified:  u.Verified,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
	if u.SchoolID != nil {
		s := u.SchoolID.String()
		dto.SchoolID = &s
	}
	return dto
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
