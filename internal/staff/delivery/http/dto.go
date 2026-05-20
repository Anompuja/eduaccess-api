package http

import "time"

// CreateStaffRequest represents the request body for creating a staff.
type CreateStaffRequest struct {
	Name         string  `json:"name" validate:"required,min=1,max=191"`
	Email        string  `json:"email" validate:"required,email"`
	Username     string  `json:"username" validate:"required,min=3,max=191"`
	Password     string  `json:"password" validate:"required,min=8"`
	SchoolID     *string `json:"school_id" validate:"omitempty,uuid"`
	PhoneNumber  *string `json:"phone_number"`
	Address      *string `json:"address"`
	Gender       *string `json:"gender" validate:"omitempty,oneof=male female other"`
	Religion     *string `json:"religion"`
	BirthPlace   *string `json:"birth_place"`
	BirthDate    *string `json:"birth_date"`
	NIK          *string `json:"nik"`
	KTPImagePath *string `json:"ktp_image_path"`
}

// UpdateStaffRequest represents the request body for updating a staff.
type UpdateStaffRequest struct {
	Name         *string `json:"name" validate:"omitempty,min=1,max=191"`
	Email        *string `json:"email" validate:"omitempty,email"`
	Username     *string `json:"username" validate:"omitempty,min=3,max=191"`
	PhoneNumber  *string `json:"phone_number"`
	Address      *string `json:"address"`
	Gender       *string `json:"gender" validate:"omitempty,oneof=male female other"`
	Religion     *string `json:"religion"`
	BirthPlace   *string `json:"birth_place"`
	BirthDate    *string `json:"birth_date"`
	NIK          *string `json:"nik"`
	KTPImagePath *string `json:"ktp_image_path"`
}

// StaffResponse represents the response body for a staff.
type StaffResponse struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	SchoolID     string     `json:"school_id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	Username     string     `json:"username"`
	Avatar       string     `json:"avatar"`
	PhoneNumber  *string    `json:"phone_number"`
	Address      *string    `json:"address"`
	Gender       *string    `json:"gender"`
	Religion     *string    `json:"religion"`
	BirthPlace   *string    `json:"birth_place"`
	BirthDate    *time.Time `json:"birth_date"`
	NIK          *string    `json:"nik"`
	KTPImagePath *string    `json:"ktp_image_path"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
