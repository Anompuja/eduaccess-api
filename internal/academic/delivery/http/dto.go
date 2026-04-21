package http

import "time"

type EducationLevelResponse struct {
	ID        string    `json:"id"`
	SchoolID  string    `json:"school_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ClassResponse struct {
	ID               string    `json:"id"`
	SchoolID         string    `json:"school_id"`
	EducationLevelID string    `json:"education_level_id"`
	Name             string    `json:"name"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type SubClassResponse struct {
	ID        string    `json:"id"`
	SchoolID  string    `json:"school_id"`
	ClassID   string    `json:"class_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AcademicNameRequest struct {
	Name string `json:"name" validate:"required,min=1,max=191"`
}

type CreateClassRequest struct {
	LevelID string `json:"level_id" validate:"required,uuid"`
	Name    string `json:"name"     validate:"required,min=1,max=191"`
}

type CreateSubClassRequest struct {
	ClassID string `json:"class_id" validate:"required,uuid"`
	Name    string `json:"name"     validate:"required,min=1,max=191"`
}
