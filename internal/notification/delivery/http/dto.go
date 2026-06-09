package http

import (
	"time"

	"github.com/eduaccess/eduaccess-api/internal/notification/domain"
)

type NotificationResponse struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	Data      map[string]any `json:"data"`
	ReadAt    *time.Time     `json:"read_at"`
	CreatedAt time.Time      `json:"created_at"`
}

func toResponse(n *domain.Notification) NotificationResponse {
	return NotificationResponse{
		ID:        n.ID.String(),
		Type:      n.Type,
		Title:     n.Title,
		Body:      n.Body,
		Data:      n.Data,
		ReadAt:    n.ReadAt,
		CreatedAt: n.CreatedAt,
	}
}
