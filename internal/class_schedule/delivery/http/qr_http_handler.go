package http

import (
	"net/http"

	"github.com/eduaccess/eduaccess-api/internal/class_schedule/application"
	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	"github.com/labstack/echo/v4"
	qrcode "github.com/skip2/go-qrcode"
)

// GetQRToken godoc
//
//	@Summary      Get rotating QR token for class attendance
//	@Description  Returns a signed JWT token (30-second TTL) that encodes the class session. Teachers display this as a QR code; students scan it to mark attendance. Class must be in "ongoing" status.
//	@Tags         class-schedules
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "Class schedule UUID"
//	@Success      200  {object}  response.Response{data=QRTokenResponse}
//	@Failure      400  {object}  response.Response
//	@Failure      403  {object}  response.Response
//	@Failure      404  {object}  response.Response
//	@Router       /class-schedules/{id}/qr [get]
func (h *Handler) GetQRToken(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return handleAppError(c, err)
	}
	result, err := h.generateQR.Handle(c.Request().Context(), application.GenerateQRCommand{
		RequesterSchoolID: getSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		ScheduleID:        id,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "QR token generated", QRTokenResponse{
		Token:     result.Token,
		ExpiresIn: result.ExpiresIn,
	})
}

// GetQRImage godoc
//
//	@Summary      Get QR code image for class attendance
//	@Description  Returns a 256×256 PNG QR code image containing the current attendance token. Auto-refresh this endpoint every 30 seconds. Class must be in "ongoing" status.
//	@Tags         class-schedules
//	@Produce      image/png
//	@Security     BearerAuth
//	@Param        id   path  string  true  "Class schedule UUID"
//	@Success      200  {file}    binary
//	@Failure      400  {object}  response.Response
//	@Failure      403  {object}  response.Response
//	@Router       /class-schedules/{id}/qr/image [get]
func (h *Handler) GetQRImage(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return handleAppError(c, err)
	}
	result, err := h.generateQR.Handle(c.Request().Context(), application.GenerateQRCommand{
		RequesterSchoolID: getSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		ScheduleID:        id,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	png, err := qrcode.Encode(result.Token, qrcode.High, 256)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Response{
			Success: false,
			Message: "failed to generate QR image",
		})
	}

	c.Response().Header().Set("Cache-Control", "no-store")
	return c.Blob(http.StatusOK, "image/png", png)
}

// ScanQR godoc
//
//	@Summary      Scan QR code to mark attendance
//	@Description  Called by a student after their camera decodes the QR code. Verifies the token, checks enrollment, and marks the student as "present" or "late". A second scan by the same student returns 409.
//	@Tags         attendance
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      ScanQRRequest  true  "QR token from the code"
//	@Success      200   {object}  response.Response{data=ScanQRResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      401   {object}  response.Response
//	@Failure      403   {object}  response.Response
//	@Failure      409   {object}  response.Response
//	@Router       /attendance/scan [post]
func (h *Handler) ScanQR(c echo.Context) error {
	var req ScanQRRequest
	if err := c.Bind(&req); err != nil || req.Token == "" {
		return response.BadRequest(c, "token is required")
	}
	result, err := h.scanQR.Handle(c.Request().Context(), application.ScanQRCommand{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		StudentUserID:     authmw.GetUserID(c),
		Token:             req.Token,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, result.Message, ScanQRResponse{
		Status:  result.Status,
		Message: result.Message,
	})
}

// ── DTOs ─────────────────────────────────────────────────────────────────────

type QRTokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}

type ScanQRRequest struct {
	Token string `json:"token"`
}

type ScanQRResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

