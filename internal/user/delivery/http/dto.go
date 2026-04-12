package http

import "time"

// UserResponse is the public representation of a user.
type UserResponse struct {
	ID        string     `json:"id"`
	SchoolID  *string    `json:"school_id,omitempty"`
	Role      string     `json:"role"`
	Name      string     `json:"name"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Avatar    string     `json:"avatar"`
	Verified  bool       `json:"verified"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// UpdateUserRequest is the body for PUT /users/:id and PUT /profile.
type UpdateUserRequest struct {
	Name   *string `json:"name"   validate:"omitempty,min=2,max=100"`
	Avatar *string `json:"avatar" validate:"omitempty,max=255"`
}

// ChangePasswordRequest is the body for PUT /users/:id/password.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"omitempty,min=8"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}
