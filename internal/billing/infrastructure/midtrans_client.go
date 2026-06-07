package infrastructure

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	billingapp "github.com/eduaccess/eduaccess-api/internal/billing/application"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
)

type MidtransClient struct {
	serverKey    string
	snapBaseURL  string
	apiBaseURL   string
	expiryMinute int
	client       *http.Client
}

func NewMidtransClientFromEnv() *MidtransClient {
	serverKey := strings.TrimSpace(os.Getenv("MIDTRANS_SERVER_KEY"))
	environment := strings.ToLower(strings.TrimSpace(os.Getenv("MIDTRANS_ENVIRONMENT")))
	snapBaseURL := "https://app.sandbox.midtrans.com"
	apiBaseURL := "https://api.sandbox.midtrans.com"
	if environment == "production" {
		snapBaseURL = "https://app.midtrans.com"
		apiBaseURL = "https://api.midtrans.com"
	}

	expiry := 1440
	if raw := strings.TrimSpace(os.Getenv("MIDTRANS_PAYMENT_EXPIRY_MINUTES")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			expiry = parsed
		}
	}

	return &MidtransClient{
		serverKey:    serverKey,
		snapBaseURL:  snapBaseURL,
		apiBaseURL:   apiBaseURL,
		expiryMinute: expiry,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *MidtransClient) CreateCheckout(ctx context.Context, input billingapp.GatewayCreateCheckoutInput) (*billingapp.GatewayCheckoutSession, error) {
	if strings.TrimSpace(c.serverKey) == "" {
		return nil, apperror.New(apperror.ErrInternal, "MIDTRANS_SERVER_KEY is not configured")
	}
	expiryMinute := input.ExpiryMinute
	if expiryMinute <= 0 {
		expiryMinute = c.expiryMinute
	}

	payload := map[string]any{
		"transaction_details": map[string]any{
			"order_id":     input.OrderID,
			"gross_amount": input.Amount,
		},
		"customer_details": map[string]any{
			"first_name": input.SchoolName,
			"email":      input.SchoolEmail,
			"phone":      input.SchoolPhone,
		},
		"item_details": []map[string]any{
			{
				"id":       strings.ToLower(strings.ReplaceAll(input.PlanName, " ", "-")),
				"price":    input.Amount,
				"quantity": 1,
				"name":     fmt.Sprintf("EduAccess %s (%s)", input.PlanName, input.Cycle),
			},
		},
		"custom_field1": input.PlanName,
		"custom_field2": input.Cycle,
		"expiry": map[string]any{
			"unit":     "minute",
			"duration": expiryMinute,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.snapBaseURL+"/snap/v1/transactions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.serverKey, "")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, apperror.New(apperror.ErrBadRequest, "failed to create Midtrans checkout transaction")
	}

	var parsed struct {
		Token       string `json:"token"`
		RedirectURL string `json:"redirect_url"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(expiryMinute) * time.Minute)
	return &billingapp.GatewayCheckoutSession{
		Token:       parsed.Token,
		RedirectURL: parsed.RedirectURL,
		ExpiresAt:   &expiresAt,
	}, nil
}

func (c *MidtransClient) GetTransactionStatus(ctx context.Context, orderID string) (*billingapp.GatewayTransactionStatus, error) {
	if strings.TrimSpace(c.serverKey) == "" {
		return nil, apperror.New(apperror.ErrInternal, "MIDTRANS_SERVER_KEY is not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBaseURL+"/v2/"+orderID+"/status", nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.serverKey, "")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, apperror.New(apperror.ErrBadRequest, "failed to fetch Midtrans transaction status")
	}

	var parsed struct {
		OrderID           string `json:"order_id"`
		TransactionID     string `json:"transaction_id"`
		TransactionStatus string `json:"transaction_status"`
		StatusCode        string `json:"status_code"`
		GrossAmount       string `json:"gross_amount"`
		PaymentType       string `json:"payment_type"`
		FraudStatus       string `json:"fraud_status"`
		SignatureKey      string `json:"signature_key"`
		TransactionTime   string `json:"transaction_time"`
		SettlementTime    string `json:"settlement_time"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, err
	}

	return &billingapp.GatewayTransactionStatus{
		OrderID:           parsed.OrderID,
		TransactionID:     parsed.TransactionID,
		TransactionStatus: parsed.TransactionStatus,
		StatusCode:        parsed.StatusCode,
		GrossAmount:       parsed.GrossAmount,
		PaymentType:       parsed.PaymentType,
		FraudStatus:       parsed.FraudStatus,
		SignatureKey:      parsed.SignatureKey,
		TransactionTime:   parseMidtransTime(parsed.TransactionTime),
		SettlementTime:    parseMidtransTime(parsed.SettlementTime),
		RawResponse:       string(respBody),
	}, nil
}

func (c *MidtransClient) VerifySignature(orderID, statusCode, grossAmount, signature string) bool {
	if strings.TrimSpace(signature) == "" || strings.TrimSpace(c.serverKey) == "" {
		return false
	}
	hash := sha512.Sum512([]byte(orderID + statusCode + grossAmount + c.serverKey))
	return strings.EqualFold(hex.EncodeToString(hash[:]), signature)
}

func parseMidtransTime(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05 -0700",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return &t
		}
	}
	return nil
}
