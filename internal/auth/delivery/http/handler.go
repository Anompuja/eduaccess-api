package http

import (
	"errors"
	"net/http"
	"net/mail"
	"strings"

	authApp "github.com/eduaccess/eduaccess-api/internal/auth/application"
	authDomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/shared/httpcache"
	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	supabasePkg "github.com/eduaccess/eduaccess-api/pkg/supabase"
	"github.com/labstack/echo/v4"
)

// Handler exposes auth endpoints.
type Handler struct {
	register *authApp.RegisterHandler
	supabase *supabasePkg.Client
	userRepo authDomain.UserRepository
}

// NewHandler registers auth routes on the given group.
func NewHandler(v1 *echo.Group, register *authApp.RegisterHandler, supabase *supabasePkg.Client, userRepo authDomain.UserRepository) *Handler {
	h := &Handler{register: register, supabase: supabase, userRepo: userRepo}

	auth := v1.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
	auth.GET("/me", h.Me, authmw.RequireAuth, httpcache.Middleware(httpcache.Profile))

	return h
}

// Register godoc
//
//	@Summary      Register a new user
//	@Description  Creates a Supabase Auth account and a public profile with role and school assignment.
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        body  body      RegisterRequest  true  "Registration payload"
//	@Success      201   {object}  response.Response{data=RegisterResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      409   {object}  response.Response
//	@Router       /auth/register [post]
func (h *Handler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	result, err := h.register.Handle(c.Request().Context(), authApp.RegisterCommand{
		SchoolID: req.SchoolID,
		Role:     req.Role,
		Name:     req.Name,
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	return response.Created(c, "user registered", RegisterResponse{UserID: result.UserID.String()})
}

// Login godoc
//
//	@Summary      Login
//	@Description  Authenticates via Supabase Auth and returns a JWT. Use the access_token as Bearer token for subsequent requests.
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        body  body      LoginRequest  true  "Login credentials"
//	@Success      200   {object}  response.Response{data=LoginResponse}
//	@Failure      401   {object}  response.Response
//	@Router       /auth/login [post]
func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if validationErr := validateLoginRequest(req); validationErr != nil {
		return c.JSON(validationErr.status, response.Response{
			Success: false,
			Message: validationErr.message,
			Errors:  validationErr.errors,
		})
	}

	email := strings.TrimSpace(req.Email)
	user, err := h.userRepo.FindByEmailIncludingDeleted(c.Request().Context(), email)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return response.Unauthorized(c, "login failed", map[string]string{
				"Email": "tidak ditemukan",
			})
		}
		return handleAppError(c, err)
	}

	if !user.IsActive() {
		return response.Forbidden(c, "akun tidak aktif", map[string]string{
			"Account": "tidak aktif",
		})
	}

	token, err := h.supabase.SignIn(c.Request().Context(), email, req.Password)
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) && errors.Is(appErr.Err, apperror.ErrUnauthorized) {
			return response.Unauthorized(c, "login failed", map[string]string{
				"Password": "salah",
			})
		}
		return handleAppError(c, err)
	}

	userInfo := LoginUserInfo{
		ID:     user.ID.String(),
		Name:   user.Name,
		Email:  user.Email,
		Role:   user.Role,
		Avatar: user.Avatar,
	}
	if user.SchoolID != nil {
		s := user.SchoolID.String()
		userInfo.SchoolID = &s
	}

	return response.OK(c, "login successful", LoginResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresIn:    token.ExpiresIn,
		User:         userInfo,
	})
}

// Refresh godoc
//
//	@Summary      Refresh access token
//	@Description  Exchanges a refresh token for a new access token.
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        body  body      RefreshRequest  true  "Refresh token payload"
//	@Success      200   {object}  response.Response{data=LoginResponse}
//	@Failure      401   {object}  response.Response
//	@Router       /auth/refresh [post]
func (h *Handler) Refresh(c echo.Context) error {
	var req RefreshRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	token, err := h.supabase.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "token refreshed", LoginResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresIn:    token.ExpiresIn,
	})
}

// Me godoc
//
//	@Summary      Current user identity
//	@Description  Returns the user ID, school, and role extracted from the Supabase JWT.
//	@Tags         auth
//	@Produce      json
//	@Security     BearerAuth
//	@Success      200  {object}  response.Response{data=MeResponse}
//	@Failure      401  {object}  response.Response
//	@Router       /auth/me [get]
func (h *Handler) Me(c echo.Context) error {
	userID := authmw.GetUserID(c)
	role := authmw.GetRole(c)
	schoolID := authmw.GetSchoolID(c)

	resp := MeResponse{
		UserID: userID.String(),
		Role:   role,
	}
	if schoolID != nil {
		s := schoolID.String()
		resp.SchoolID = &s
	}

	return response.OK(c, "authenticated", resp)
}

func handleAppError(c echo.Context, err error) error {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		switch appErr.Err {
		case apperror.ErrNotFound:
			return response.NotFound(c, appErr.Message)
		case apperror.ErrUnauthorized, apperror.ErrInvalidToken:
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

type loginValidationError struct {
	status  int
	message string
	errors  map[string]string
}

func validateLoginRequest(req LoginRequest) *loginValidationError {
	fieldErrors := map[string]string{}
	status := http.StatusBadRequest

	email := strings.TrimSpace(req.Email)
	if email == "" {
		fieldErrors["email"] = "Email wajib diisi"
	} else if _, err := mail.ParseAddress(email); err != nil {
		fieldErrors["email"] = "Format email tidak valid"
		status = http.StatusUnprocessableEntity
	}

	if strings.TrimSpace(req.Password) == "" {
		fieldErrors["password"] = "Password wajib diisi"
	}

	if len(fieldErrors) == 0 {
		return nil
	}

	if fieldErrors["email"] == "Format email tidak valid" {
		status = http.StatusUnprocessableEntity
	}

	return &loginValidationError{
		status:  status,
		message: "validasi login gagal",
		errors:  fieldErrors,
	}
}
