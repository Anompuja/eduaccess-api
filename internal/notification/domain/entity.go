package domain

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID
	SchoolID  *uuid.UUID
	UserID    uuid.UUID
	Type      string
	Title     string
	Body      string
	Data      map[string]any
	ReadAt    *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
