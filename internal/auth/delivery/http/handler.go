package http

import (
	"errors"
	"net/http"

	"github.com/eduaccess/eduaccess-api/internal/auth/application"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	"github.com/eduaccess/eduaccess-api/internal/shared/validator"
	"github.com/labstack/echo/v4"
)

// Handler wires the auth use-cases to HTTP endpoints.
type Handler struct {
	register *application.RegisterHandler
	login    *application.LoginHandler
	refresh  *application.RefreshHandler
	logout   *application.LogoutHandler
}

// NewHandler creates a Handler and registers routes on the given group.
func NewHandler(
	v1 *echo.Group,
	register *application.RegisterHandler,
	login *application.LoginHandler,
	refresh *application.RefreshHandler,
	logout *application.LogoutHandler,
) *Handler {
	h := &Handler{
		register: register,
		login:    login,
		refresh:  refresh,
		logout:   logout,
	}

	auth := v1.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
	auth.POST("/logout", h.Logout)

	return h
}

// Register godoc
//
//	@Summary      Register a new user
//	@Description  Creates a new user account. Superadmin accounts cannot be created via this endpoint.
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        body  body      RegisterRequest                           true  "Registration payload"
//	@Success      201   {object}  response.Response{data=RegisterResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      409   {object}  response.Response
//	@Failure      422   {object}  response.Response
//	@Failure      500   {object}  response.Response
//	@Router       /auth/register [post]
func (h *Handler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	result, err := h.register.Handle(c.Request().Context(), application.RegisterCommand{
		Name:     req.Name,
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	return response.Created(c, "user registered successfully", RegisterResponse{
		UserID: result.UserID.String(),
	})
}

// Login godoc
//
//	@Summary      Login
//	@Description  Authenticates a user and returns an access + refresh token pair.
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        body  body      LoginRequest                              true  "Login credentials"
//	@Success      200   {object}  response.Response{data=TokenResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      401   {object}  response.Response
//	@Failure      403   {object}  response.Response
//	@Failure      422   {object}  response.Response
//	@Router       /auth/login [post]
func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	pair, err := h.login.Handle(c.Request().Context(), application.LoginCommand{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "login successful", TokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	})
}

// Refresh godoc
//
//	@Summary      Refresh tokens
//	@Description  Issues a new access + refresh token pair, rotating the refresh token.
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        body  body      RefreshRequest                            true  "Refresh token"
//	@Success      200   {object}  response.Response{data=TokenResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      401   {object}  response.Response
//	@Failure      422   {object}  response.Response
//	@Router       /auth/refresh [post]
func (h *Handler) Refresh(c echo.Context) error {
	var req RefreshRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	pair, err := h.refresh.Handle(c.Request().Context(), application.RefreshCommand{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "token refreshed", TokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	})
}

// Logout godoc
//
//	@Summary      Logout
//	@Description  Revokes the provided refresh token.
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        body  body      LogoutRequest   true  "Refresh token to revoke"
//	@Success      200   {object}  response.Response
//	@Failure      400   {object}  response.Response
//	@Failure      422   {object}  response.Response
//	@Router       /auth/logout [post]
func (h *Handler) Logout(c echo.Context) error {
	var req LogoutRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	if err := h.logout.Handle(c.Request().Context(), application.LogoutCommand{
		RefreshToken: req.RefreshToken,
	}); err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "logged out successfully", nil)
}

// handleAppError maps domain/application errors to HTTP responses.
func handleAppError(c echo.Context, err error) error {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		switch appErr.Err {
		case apperror.ErrNotFound:
			return response.NotFound(c, appErr.Message)
		case apperror.ErrUnauthorized, apperror.ErrWrongPassword, apperror.ErrInvalidToken, apperror.ErrTokenRevoked:
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
