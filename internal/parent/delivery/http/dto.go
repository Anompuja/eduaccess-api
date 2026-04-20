package http

import "time"

type ParentResponse struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	SchoolID       string    `json:"school_id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Username       string    `json:"username"`
	Avatar         string    `json:"avatar"`
	FatherName     string    `json:"father_name"`
	MotherName     string    `json:"mother_name"`
	FatherReligion string    `json:"father_religion"`
	MotherReligion string    `json:"mother_religion"`
	PhoneNumber    string    `json:"phone_number"`
	Address        string    `json:"address"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateParentRequest struct {
	SchoolID       *string `json:"school_id"       validate:"omitempty,uuid"`
	Name           string  `json:"name"            validate:"required,min=2,max=191"`
	Email          string  `json:"email"           validate:"required,email,max=191"`
	Username       string  `json:"username"        validate:"omitempty,min=3,max=50"`
	Password       string  `json:"password"        validate:"omitempty,min=8"`
	FatherName     string  `json:"father_name"     validate:"omitempty,max=191"`
	MotherName     string  `json:"mother_name"     validate:"omitempty,max=191"`
	FatherReligion string  `json:"father_religion" validate:"omitempty,max=100"`
	MotherReligion string  `json:"mother_religion" validate:"omitempty,max=100"`
	PhoneNumber    string  `json:"phone_number"    validate:"omitempty,max=50"`
	Address        string  `json:"address"         validate:"omitempty"`
}

type UpdateParentRequest struct {
	FatherName     *string `json:"father_name"     validate:"omitempty,max=191"`
	MotherName     *string `json:"mother_name"     validate:"omitempty,max=191"`
	FatherReligion *string `json:"father_religion" validate:"omitempty,max=100"`
	MotherReligion *string `json:"mother_religion" validate:"omitempty,max=100"`
	PhoneNumber    *string `json:"phone_number"    validate:"omitempty,max=50"`
	Address        *string `json:"address"         validate:"omitempty"`
}
