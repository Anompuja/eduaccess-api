package supabase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

// Client wraps the Supabase REST APIs using the service role key.
type Client struct {
	BaseURL    string
	ServiceKey string
	AnonKey    string
	http       *http.Client
}

// NewClient reads credentials from environment and returns a ready-to-use client.
func NewClient() *Client {
	return &Client{
		BaseURL:    os.Getenv("SUPABASE_URL"),
		ServiceKey: os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		AnonKey:    os.Getenv("SUPABASE_ANON_KEY"),
		http:       &http.Client{},
	}
}

// AdminUser is the relevant subset of a Supabase auth.users row.
type AdminUser struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

// CreateUser creates a new user in Supabase Auth via the Admin API.
// email_confirm is set to true so admin-created accounts are immediately usable.
func (c *Client) CreateUser(ctx context.Context, email, password string) (*AdminUser, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"email":         email,
		"password":      password,
		"email_confirm": true,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.BaseURL+"/auth/v1/admin/users", bytes.NewReader(body))
	if err != nil {
		return nil, apperror.New(apperror.ErrInternal, "failed to build request")
	}
	c.setAdminHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, apperror.New(apperror.ErrInternal, "supabase auth unavailable")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, apperror.New(apperror.ErrInternal, fmt.Sprintf("create user failed: %s", string(b)))
	}

	var user AdminUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, apperror.New(apperror.ErrInternal, "failed to parse auth response")
	}
	return &user, nil
}

// UpdateUserEmail changes a user's email via the Admin API.
func (c *Client) UpdateUserEmail(ctx context.Context, userID uuid.UUID, newEmail string) error {
	body, _ := json.Marshal(map[string]string{"email": newEmail})

	req, err := http.NewRequestWithContext(ctx, http.MethodPut,
		fmt.Sprintf("%s/auth/v1/admin/users/%s", c.BaseURL, userID), bytes.NewReader(body))
	if err != nil {
		return apperror.New(apperror.ErrInternal, "failed to build request")
	}
	c.setAdminHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return apperror.New(apperror.ErrInternal, "supabase auth unavailable")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return apperror.New(apperror.ErrInternal, fmt.Sprintf("failed to update email: %s", string(b)))
	}
	return nil
}

// UpdateUserPassword changes a user's password via the Admin API.
func (c *Client) UpdateUserPassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	body, _ := json.Marshal(map[string]string{"password": newPassword})

	req, err := http.NewRequestWithContext(ctx, http.MethodPut,
		fmt.Sprintf("%s/auth/v1/admin/users/%s", c.BaseURL, userID), bytes.NewReader(body))
	if err != nil {
		return apperror.New(apperror.ErrInternal, "failed to build request")
	}
	c.setAdminHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return apperror.New(apperror.ErrInternal, "supabase auth unavailable")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return apperror.New(apperror.ErrInternal, "failed to update password")
	}
	return nil
}

// VerifyPassword confirms credentials by attempting a sign-in.
// Returns ErrWrongPassword if credentials are invalid.
func (c *Client) VerifyPassword(ctx context.Context, email, password string) error {
	body, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.BaseURL+"/auth/v1/token?grant_type=password", bytes.NewReader(body))
	if err != nil {
		return apperror.New(apperror.ErrInternal, "failed to build request")
	}
	req.Header.Set("apikey", c.AnonKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return apperror.New(apperror.ErrInternal, "supabase auth unavailable")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return apperror.New(apperror.ErrWrongPassword, "current password is incorrect")
	}
	return nil
}

// TokenResponse is the successful response from Supabase sign-in.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// SignIn authenticates a user via email+password and returns the Supabase JWT.
func (c *Client) SignIn(ctx context.Context, email, password string) (*TokenResponse, error) {
	body, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.BaseURL+"/auth/v1/token?grant_type=password", bytes.NewReader(body))
	if err != nil {
		return nil, apperror.New(apperror.ErrInternal, "failed to build request")
	}
	req.Header.Set("apikey", c.AnonKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, apperror.New(apperror.ErrInternal, "supabase auth unavailable")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, apperror.New(apperror.ErrUnauthorized, "invalid email or password")
	}

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, apperror.New(apperror.ErrInternal, "failed to parse token response")
	}
	return &token, nil
}

// RefreshToken exchanges a refresh token for a new access token.
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	body, _ := json.Marshal(map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.BaseURL+"/auth/v1/token?grant_type=refresh_token", bytes.NewReader(body))
	if err != nil {
		return nil, apperror.New(apperror.ErrInternal, "failed to build request")
	}
	req.Header.Set("apikey", c.AnonKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, apperror.New(apperror.ErrInternal, "supabase auth unavailable")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, apperror.New(apperror.ErrUnauthorized, "invalid or expired refresh token")
	}

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, apperror.New(apperror.ErrInternal, "failed to parse token response")
	}
	return &token, nil
}

// DeleteUser removes a user from Supabase Auth (used for rollback on failed profile creation).
func (c *Client) DeleteUser(ctx context.Context, userID uuid.UUID) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("%s/auth/v1/admin/users/%s", c.BaseURL, userID), nil)
	if err != nil {
		return
	}
	c.setAdminHeaders(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

func (c *Client) setAdminHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.ServiceKey)
	req.Header.Set("apikey", c.ServiceKey)
	req.Header.Set("Content-Type", "application/json")
}
