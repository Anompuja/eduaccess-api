package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/eduaccess/eduaccess-api/internal/billing/application"
	billingdomain "github.com/eduaccess/eduaccess-api/internal/billing/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	"github.com/eduaccess/eduaccess-api/internal/shared/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	createCheckout             *application.CreateCheckoutHandler
	getPayment                 *application.GetPaymentHandler
	handleMidtransNotification *application.HandleMidtransNotificationHandler
}

func NewHandler(
	v1 *echo.Group,
	createCheckout *application.CreateCheckoutHandler,
	getPayment *application.GetPaymentHandler,
	handleMidtransNotification *application.HandleMidtransNotificationHandler,
) *Handler {
	h := &Handler{
		createCheckout:             createCheckout,
		getPayment:                 getPayment,
		handleMidtransNotification: handleMidtransNotification,
	}

	schools := v1.Group("/schools", authmw.RequireAuth)
	schools.POST("/:id/subscription/checkout", h.CreateCheckout)
	schools.GET("/:id/subscription/payments/:payment_id", h.GetPayment)

	v1.POST("/billing/webhooks/midtrans", h.HandleMidtransNotification)

	return h
}

// CreateCheckout godoc
//
//	@Summary      Create subscription checkout
//	@Description  Creates a Midtrans Snap checkout transaction for upgrading or purchasing a paid school subscription. Only admin_sekolah for the target school or superadmin may call this endpoint.
//	@Tags         billing
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id    path      string                 true  "School UUID"
//	@Param        body  body      CreateCheckoutRequest  true  "Checkout request"
//	@Success      201   {object}  response.Response{data=PaymentResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      401   {object}  response.Response
//	@Failure      403   {object}  response.Response
//	@Failure      404   {object}  response.Response
//	@Failure      409   {object}  response.Response
//	@Router       /schools/{id}/subscription/checkout [post]
func (h *Handler) CreateCheckout(c echo.Context) error {
	schoolID, err := parseUUID(c, "id")
	if err != nil {
		return err
	}

	var req CreateCheckoutRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	planID, err := uuid.Parse(req.PlanID)
	if err != nil {
		return response.BadRequest(c, "invalid plan_id")
	}

	payment, err := h.createCheckout.Handle(c.Request().Context(), application.CreateCheckoutCommand{
		RequesterRole:     authmw.GetRole(c),
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterUserID:   authmw.GetUserID(c),
		SchoolID:          schoolID,
		PlanID:            planID,
		Cycle:             req.Cycle,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return response.Created(c, "checkout created", toPaymentResponse(payment))
}

// GetPayment godoc
//
//	@Summary      Get subscription payment
//	@Description  Returns the current state of a Midtrans-backed payment transaction for a school's subscription purchase or upgrade.
//	@Tags         billing
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id          path      string  true  "School UUID"
//	@Param        payment_id  path      string  true  "Payment transaction UUID"
//	@Success      200         {object}  response.Response{data=PaymentResponse}
//	@Failure      400         {object}  response.Response
//	@Failure      401         {object}  response.Response
//	@Failure      403         {object}  response.Response
//	@Failure      404         {object}  response.Response
//	@Router       /schools/{id}/subscription/payments/{payment_id} [get]
func (h *Handler) GetPayment(c echo.Context) error {
	schoolID, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	paymentID, err := parseUUID(c, "payment_id")
	if err != nil {
		return err
	}

	payment, err := h.getPayment.Handle(c.Request().Context(), application.GetPaymentQuery{
		RequesterRole:     authmw.GetRole(c),
		RequesterSchoolID: authmw.GetSchoolID(c),
		SchoolID:          schoolID,
		PaymentID:         paymentID,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "payment retrieved", toPaymentResponse(payment))
}

// HandleMidtransNotification godoc
//
//	@Summary      Handle Midtrans webhook
//	@Description  Receives Midtrans HTTP notifications. Real notifications with complete fields are verified and processed. Probe/test requests without a complete notification payload return HTTP 200 without changing payment state.
//	@Tags         billing
//	@Accept       json
//	@Produce      json
//	@Param        body  body      MidtransNotificationRequest  false  "Midtrans notification payload"
//	@Success      200   {object}  response.Response
//	@Failure      400   {object}  response.Response
//	@Failure      401   {object}  response.Response
//	@Failure      404   {object}  response.Response
//	@Failure      409   {object}  response.Response
//	@Router       /billing/webhooks/midtrans [post]
func (h *Handler) HandleMidtransNotification(c echo.Context) error {
	var req MidtransNotificationRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if strings.TrimSpace(req.OrderID) == "" ||
		strings.TrimSpace(req.StatusCode) == "" ||
		strings.TrimSpace(req.GrossAmount) == "" ||
		strings.TrimSpace(req.SignatureKey) == "" {
		return response.OK(c, "midtrans webhook endpoint reachable", nil)
	}

	rawPayload, _ := json.Marshal(req)

	payment, err := h.handleMidtransNotification.Handle(c.Request().Context(), application.HandleMidtransNotificationCommand{
		OrderID:           req.OrderID,
		StatusCode:        req.StatusCode,
		GrossAmount:       req.GrossAmount,
		SignatureKey:      req.SignatureKey,
		TransactionID:     req.TransactionID,
		TransactionStatus: req.TransactionStatus,
		PaymentType:       req.PaymentType,
		FraudStatus:       req.FraudStatus,
		RawNotification:   string(rawPayload),
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "midtrans notification processed", toPaymentResponse(payment))
}

func toPaymentResponse(payment *billingdomain.PaymentTransaction) PaymentResponse {
	dto := PaymentResponse{
		ID:                    payment.ID.String(),
		SchoolID:              payment.SchoolID.String(),
		PlanID:                payment.PlanID.String(),
		CreatedByUserID:       payment.CreatedByUserID.String(),
		Status:                payment.Status,
		Cycle:                 payment.Cycle,
		Amount:                payment.Amount,
		Currency:              payment.Currency,
		Provider:              payment.Provider,
		ProviderOrderID:       payment.ProviderOrderID,
		ProviderTransactionID: payment.ProviderTransactionID,
		ProviderSnapToken:     payment.ProviderSnapToken,
		ProviderRedirectURL:   payment.ProviderRedirectURL,
		PaymentType:           payment.PaymentType,
		TransactionStatus:     payment.TransactionStatus,
		FraudStatus:           payment.FraudStatus,
		PaidAt:                payment.PaidAt,
		ExpiresAt:             payment.ExpiresAt,
		CreatedAt:             payment.CreatedAt,
		UpdatedAt:             payment.UpdatedAt,
	}
	if payment.ActivatedSubscriptionID != nil {
		id := payment.ActivatedSubscriptionID.String()
		dto.ActivatedSubscriptionID = &id
	}
	return dto
}

func parseUUID(c echo.Context, param string) (uuid.UUID, error) {
	raw := c.Param(param)
	id, err := uuid.Parse(raw)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response.Response{
			Success: false,
			Message: "invalid UUID: " + param,
		})
		return uuid.UUID{}, echo.ErrBadRequest
	}
	return id, nil
}

func handleAppError(c echo.Context, err error) error {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		switch appErr.Err {
		case apperror.ErrNotFound:
			return response.NotFound(c, appErr.Message)
		case apperror.ErrUnauthorized, apperror.ErrInvalidToken, apperror.ErrTokenRevoked:
			return response.Unauthorized(c, appErr.Message)
		case apperror.ErrForbidden:
			return response.Forbidden(c, appErr.Message)
		case apperror.ErrConflict:
			return response.Conflict(c, appErr.Message)
		case apperror.ErrBadRequest:
			return response.BadRequest(c, appErr.Message)
		}
	}
	return c.JSON(http.StatusInternalServerError, response.Response{
		Success: false,
		Message: "internal server error",
	})
}
