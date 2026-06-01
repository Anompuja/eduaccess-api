package http

import "github.com/google/uuid"

// RegisterRequest is the body for POST /auth/register.
type RegisterRequest struct {
	Name     string     `json:"name"     validate:"required"`
	Username string     `json:"username"`
	Email    string     `json:"email"    validate:"required,email"`
	Password string     `json:"password" validate:"required,min=8"`
	Role     string     `json:"role"     validate:"required"`
	SchoolID *uuid.UUID `json:"school_id"`
}

// RegisterResponse is returned on successful registration.
type RegisterResponse struct {
	UserID string `json:"user_id"`
}

// LoginRequest is the body for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshRequest is the body for POST /auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LoginUserInfo is the user profile embedded in a login response.
type LoginUserInfo struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Role     string  `json:"role"`
	SchoolID *string `json:"school_id,omitempty"`
	Avatar   string  `json:"avatar"`
}

// LoginResponse is returned on successful login.
type LoginResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	TokenType    string        `json:"token_type"`
	ExpiresIn    int           `json:"expires_in"`
	User         LoginUserInfo `json:"user"`
}

// MeResponse is returned by GET /auth/me.
type MeResponse struct {
	UserID   string  `json:"user_id"`
	SchoolID *string `json:"school_id,omitempty"`
	Role     string  `json:"role"`
}
