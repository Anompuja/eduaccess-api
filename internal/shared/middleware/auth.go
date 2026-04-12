package middleware

import (
	"net/http"
	"strings"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	pkgjwt "github.com/eduaccess/eduaccess-api/pkg/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Context keys for values injected by the auth middleware.
const (
	ContextKeyUserID   = "user_id"
	ContextKeySchoolID = "school_id"
	ContextKeyRole     = "role"
)

// RequireAuth validates the Bearer JWT and injects claims into the Echo context.
func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		header := c.Request().Header.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return response.Unauthorized(c, "missing or malformed authorization header")
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := pkgjwt.Parse(tokenStr)
		if err != nil {
			return response.Unauthorized(c, apperror.ErrInvalidToken.Error())
		}
		if claims.TokenType != pkgjwt.AccessToken {
			return response.Unauthorized(c, "not an access token")
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeySchoolID, claims.SchoolID) // *uuid.UUID, may be nil
		c.Set(ContextKeyRole, claims.Role)

		return next(c)
	}
}

// RequireRoles returns a middleware that allows only the given roles through.
// Must be chained after RequireAuth.
func RequireRoles(roles ...string) echo.MiddlewareFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, _ := c.Get(ContextKeyRole).(string)
			if _, ok := allowed[role]; !ok {
				return c.JSON(http.StatusForbidden, response.Response{
					Success: false,
					Message: "insufficient permissions",
				})
			}
			return next(c)
		}
	}
}

// GetUserID extracts the authenticated user's UUID from context.
func GetUserID(c echo.Context) uuid.UUID {
	id, _ := c.Get(ContextKeyUserID).(uuid.UUID)
	return id
}

// GetSchoolID extracts the authenticated user's school UUID from context (nil for superadmin).
func GetSchoolID(c echo.Context) *uuid.UUID {
	id, _ := c.Get(ContextKeySchoolID).(*uuid.UUID)
	return id
}

// GetRole extracts the authenticated user's role from context.
func GetRole(c echo.Context) string {
	role, _ := c.Get(ContextKeyRole).(string)
	return role
}
