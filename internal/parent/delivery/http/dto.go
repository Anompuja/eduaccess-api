package http

import "time"

type ParentResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	SchoolID    string    `json:"school_id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	Avatar      string    `json:"avatar"`
	Religion    string    `json:"religion"`
	PhoneNumber string    `json:"phone_number"`
	Address     string    `json:"address"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateParentRequest struct {
	SchoolID    *string `json:"school_id"    validate:"omitempty,uuid"`
	Name        string  `json:"name"         validate:"required,min=2,max=191"`
	Email       string  `json:"email"        validate:"required,email,max=191"`
	Username    string  `json:"username"     validate:"omitempty,min=3,max=50"`
	Password    string  `json:"password"     validate:"omitempty,min=8"`
	Religion    string  `json:"religion"     validate:"omitempty,max=50"`
	PhoneNumber string  `json:"phone_number" validate:"omitempty,max=50"`
	Address     string  `json:"address"      validate:"omitempty"`
}

type UpdateParentRequest struct {
	Name        *string `json:"name"         validate:"omitempty,min=2,max=191"`
	Email       *string `json:"email"        validate:"omitempty,email,max=191"`
	Religion    *string `json:"religion"     validate:"omitempty,max=50"`
	PhoneNumber *string `json:"phone_number" validate:"omitempty,max=50"`
	Address     *string `json:"address"      validate:"omitempty"`
}
