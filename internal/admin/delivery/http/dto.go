package http

import "time"

type AdminResponse struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	SchoolID     string     `json:"school_id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	Username     string     `json:"username"`
	Avatar       string     `json:"avatar"`
	PhoneNumber  string     `json:"phone_number"`
	Address      string     `json:"address"`
	Gender       string     `json:"gender"`
	Religion     string     `json:"religion"`
	BirthPlace   string     `json:"birth_place"`
	BirthDate    *time.Time `json:"birth_date,omitempty"`
	NIK          string     `json:"nik"`
	KTPImagePath string     `json:"ktp_image_path"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type CreateAdminRequest struct {
	SchoolID     *string `json:"school_id"      validate:"omitempty,uuid"`
	Name         string  `json:"name"           validate:"required,min=2,max=191"`
	Email        string  `json:"email"          validate:"required,email,max=191"`
	Username     string  `json:"username"       validate:"omitempty,min=3,max=50"`
	Password     string  `json:"password"       validate:"omitempty,min=8"`
	PhoneNumber  string  `json:"phone_number"   validate:"omitempty,max=50"`
	Address      string  `json:"address"        validate:"omitempty"`
	Gender       string  `json:"gender"         validate:"omitempty,oneof=L P"`
	Religion     string  `json:"religion"       validate:"omitempty,max=100"`
	BirthPlace   string  `json:"birth_place"    validate:"omitempty,max=191"`
	BirthDate    *string `json:"birth_date"     validate:"omitempty"`
	NIK          string  `json:"nik"            validate:"omitempty,max=50"`
	KTPImagePath string  `json:"ktp_image_path" validate:"omitempty,max=191"`
}

type UpdateAdminRequest struct {
	Name         *string `json:"name"           validate:"omitempty,min=2,max=191"`
	Email        *string `json:"email"          validate:"omitempty,email,max=191"`
	Username     *string `json:"username"       validate:"omitempty,min=3,max=50"`
	PhoneNumber  *string `json:"phone_number"   validate:"omitempty,max=50"`
	Address      *string `json:"address"        validate:"omitempty"`
	Gender       *string `json:"gender"         validate:"omitempty,oneof=L P"`
	Religion     *string `json:"religion"       validate:"omitempty,max=100"`
	BirthPlace   *string `json:"birth_place"    validate:"omitempty,max=191"`
	BirthDate    *string `json:"birth_date"     validate:"omitempty"`
	NIK          *string `json:"nik"            validate:"omitempty,max=50"`
	KTPImagePath *string `json:"ktp_image_path" validate:"omitempty,max=191"`
}
