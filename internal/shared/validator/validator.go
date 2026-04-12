package validator

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// RequestValidator implements echo.Validator.
type RequestValidator struct {
	v *validator.Validate
}

func New() *RequestValidator {
	return &RequestValidator{v: validator.New()}
}

func (rv *RequestValidator) Validate(i interface{}) error {
	return rv.v.Struct(i)
}

// BindAndValidate binds the request body and validates it.
// Returns an HTTP 422 JSON error on failure so handlers stay clean.
func BindAndValidate(c echo.Context, req interface{}) error {
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(req); err != nil {
		var errs []map[string]string
		for _, fe := range err.(validator.ValidationErrors) {
			errs = append(errs, map[string]string{
				"field":   fe.Field(),
				"message": fe.Tag(),
			})
		}
		return echo.NewHTTPError(http.StatusUnprocessableEntity, map[string]interface{}{
			"message": "validation failed",
			"errors":  errs,
		})
	}
	return nil
}
