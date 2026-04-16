package http

import "time"

// HeadmasterResponse is the public representation of a headmaster profile.
type HeadmasterResponse struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	SchoolID     string     `json:"school_id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	Username     string     `json:"username"`
	Avatar       string     `json:"avatar"`
	PhoneNumber  string     `json:"phone_number,omitempty"`
	Address      string     `json:"address,omitempty"`
	Gender       string     `json:"gender,omitempty"`
	Religion     string     `json:"religion,omitempty"`
	BirthPlace   string     `json:"birth_place,omitempty"`
	BirthDate    *time.Time `json:"birth_date,omitempty"`
	NIK          string     `json:"nik,omitempty"`
	KTPImagePath string     `json:"ktp_image_path,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CreateHeadmasterRequest is the body for POST /headmasters.
type CreateHeadmasterRequest struct {
	// SchoolID is required when called by superadmin (who has no school in their JWT).
	// admin_sekolah may omit it — their JWT school is used automatically.
	SchoolID     string     `json:"school_id"     validate:"omitempty,uuid"`
	Name         string     `json:"name"          validate:"required,min=2,max=100"`
	Email        string     `json:"email"         validate:"required,email"`
	Username     string     `json:"username"      validate:"omitempty,min=3,max=50"`
	Password     string     `json:"password"      validate:"omitempty,min=8"`
	PhoneNumber  string     `json:"phone_number"  validate:"omitempty,max=50"`
	Address      string     `json:"address"       validate:"omitempty,max=500"`
	Gender       string     `json:"gender"        validate:"omitempty,oneof=male female"`
	Religion     string     `json:"religion"      validate:"omitempty,max=100"`
	BirthPlace   string     `json:"birth_place"   validate:"omitempty,max=191"`
	BirthDate    *time.Time `json:"birth_date"`
	NIK          string     `json:"nik"           validate:"omitempty,max=50"`
	KTPImagePath string     `json:"ktp_image_path" validate:"omitempty,max=191"`
}

// UpdateHeadmasterRequest is the body for PUT /headmasters/:id.
type UpdateHeadmasterRequest struct {
	PhoneNumber  *string    `json:"phone_number"  validate:"omitempty,max=50"`
	Address      *string    `json:"address"       validate:"omitempty,max=500"`
	Gender       *string    `json:"gender"        validate:"omitempty,oneof=male female"`
	Religion     *string    `json:"religion"      validate:"omitempty,max=100"`
	BirthPlace   *string    `json:"birth_place"   validate:"omitempty,max=191"`
	BirthDate    *time.Time `json:"birth_date"`
	NIK          *string    `json:"nik"           validate:"omitempty,max=50"`
	KTPImagePath *string    `json:"ktp_image_path" validate:"omitempty,max=191"`
}
