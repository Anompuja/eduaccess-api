package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

type Pagination struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func OK(c echo.Context, message string, data interface{}) error {
	return c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Created(c echo.Context, message string, data interface{}) error {
	return c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func NoContent(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

func BadRequest(c echo.Context, message string) error {
	return c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Message: message,
	})
}

func Unauthorized(c echo.Context, message string) error {
	return c.JSON(http.StatusUnauthorized, Response{
		Success: false,
		Message: message,
	})
}

func Forbidden(c echo.Context, message string) error {
	return c.JSON(http.StatusForbidden, Response{
		Success: false,
		Message: message,
	})
}

func NotFound(c echo.Context, message string) error {
	return c.JSON(http.StatusNotFound, Response{
		Success: false,
		Message: message,
	})
}

func Conflict(c echo.Context, message string) error {
	return c.JSON(http.StatusConflict, Response{
		Success: false,
		Message: message,
	})
}

func InternalError(c echo.Context, message string) error {
	return c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Message: message,
	})
}

func Paginated(c echo.Context, message string, data interface{}, page, perPage int, total int64) error {
	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}
	return c.JSON(http.StatusOK, PaginatedResponse{
		Success: true,
		Message: message,
		Data:    data,
		Pagination: Pagination{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}
