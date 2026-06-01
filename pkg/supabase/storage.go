package supabase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
)

// UploadFile uploads raw bytes to a Supabase Storage bucket.
// Returns the public URL on success. Use x-upsert=true to overwrite existing files.
func (c *Client) UploadFile(ctx context.Context, bucket, path string, data []byte, contentType string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/storage/v1/object/%s/%s", c.BaseURL, bucket, path),
		bytes.NewReader(data))
	if err != nil {
		return "", apperror.New(apperror.ErrInternal, "failed to build storage request")
	}
	req.Header.Set("Authorization", "Bearer "+c.ServiceKey)
	req.Header.Set("apikey", c.ServiceKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", apperror.New(apperror.ErrInternal, "storage service unavailable")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return "", apperror.New(apperror.ErrInternal, fmt.Sprintf("upload failed: %s", string(b)))
	}

	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", c.BaseURL, bucket, path), nil
}

// GetSignedURL generates a time-limited signed URL for a private bucket file.
// expiresIn is in seconds.
func (c *Client) GetSignedURL(ctx context.Context, bucket, path string, expiresIn int) (string, error) {
	body, _ := json.Marshal(map[string]interface{}{"expiresIn": expiresIn})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/storage/v1/object/sign/%s/%s", c.BaseURL, bucket, path),
		bytes.NewReader(body))
	if err != nil {
		return "", apperror.New(apperror.ErrInternal, "failed to build request")
	}
	req.Header.Set("Authorization", "Bearer "+c.ServiceKey)
	req.Header.Set("apikey", c.ServiceKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", apperror.New(apperror.ErrInternal, "storage service unavailable")
	}
	defer resp.Body.Close()

	var result struct {
		SignedURL string `json:"signedURL"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || result.SignedURL == "" {
		return "", apperror.New(apperror.ErrInternal, "failed to generate signed URL")
	}

	return c.BaseURL + result.SignedURL, nil
}

// ContentTypeFromFilename guesses the MIME type from a filename extension.
// Falls back to application/octet-stream when unknown.
func ContentTypeFromFilename(filename string) string {
	ext := filepath.Ext(filename)
	ct := mime.TypeByExtension(ext)
	if ct == "" {
		return "application/octet-stream"
	}
	return ct
}
