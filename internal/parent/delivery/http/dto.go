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
