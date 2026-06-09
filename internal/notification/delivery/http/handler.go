package http

import (
	"errors"

	"github.com/eduaccess/eduaccess-api/internal/notification/application"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	list       *application.ListNotificationsHandler
	markRead   *application.MarkReadHandler
	markAllRead *application.MarkAllReadHandler
}

func NewHandler(
	v1 *echo.Group,
	list *application.ListNotificationsHandler,
	markRead *application.MarkReadHandler,
	markAllRead *application.MarkAllReadHandler,
) *Handler {
	h := &Handler{list: list, markRead: markRead, markAllRead: markAllRead}

	g := v1.Group("/notifications", authmw.RequireAuth)
	g.GET("", h.handleList)
	g.PATCH("/read-all", h.handleMarkAllRead)
	g.PATCH("/:id/read", h.handleMarkRead)
	return h
}

func (h *Handler) handleList(c echo.Context) error {
	userID := authmw.GetUserID(c)
	unreadOnly := c.QueryParam("unread") == "true"

	notifications, err := h.list.Handle(c.Request().Context(), application.ListNotificationsCommand{
		UserID:     userID,
		UnreadOnly: unreadOnly,
	})
	if err != nil {
		return response.InternalError(c, "failed to fetch notifications")
	}

	result := make([]NotificationResponse, len(notifications))
	for i, n := range notifications {
		result[i] = toResponse(n)
	}
	return response.OK(c, "notifications fetched", result)
}

func (h *Handler) handleMarkRead(c echo.Context) error {
	idStr := c.Param("id")
	notifID, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(c, "invalid notification id")
	}

	userID := authmw.GetUserID(c)
	if err := h.markRead.Handle(c.Request().Context(), application.MarkReadCommand{
		NotificationID: notifID,
		UserID:         userID,
	}); err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) && appErr.Err == apperror.ErrNotFound {
			return response.NotFound(c, appErr.Message)
		}
		return response.InternalError(c, "failed to mark notification as read")
	}
	return response.OK(c, "notification marked as read", nil)
}

func (h *Handler) handleMarkAllRead(c echo.Context) error {
	userID := authmw.GetUserID(c)
	if err := h.markAllRead.Handle(c.Request().Context(), application.MarkAllReadCommand{
		UserID: userID,
	}); err != nil {
		return response.InternalError(c, "failed to mark all notifications as read")
	}
	return response.OK(c, "all notifications marked as read", nil)
}
