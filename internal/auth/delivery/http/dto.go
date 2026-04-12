package http

// RegisterRequest is the body for POST /auth/register.
type RegisterRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Username string `json:"username" validate:"omitempty,min=3,max=50,alphanum"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role"     validate:"required,oneof=admin_sekolah kepala_sekolah guru staff orangtua siswa"`
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

// LogoutRequest is the body for POST /auth/logout.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// TokenResponse is returned on successful login or refresh.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// RegisterResponse is returned on successful registration.
type RegisterResponse struct {
	UserID string `json:"user_id"`
}
