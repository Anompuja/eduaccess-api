package http

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	supabasePkg "github.com/eduaccess/eduaccess-api/pkg/supabase"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var allowedBuckets = map[string]bool{
	"avatars":        true,
	"school-images":  true,
	"documents":      true,
	"profile-photos": true,
}

// Handler exposes Supabase Storage operations over HTTP.
type Handler struct {
	supabase *supabasePkg.Client
}

// NewHandler registers storage routes on the given group.
func NewHandler(v1 *echo.Group, supabase *supabasePkg.Client) *Handler {
	h := &Handler{supabase: supabase}

	storage := v1.Group("/storage", middleware.RequireAuth)
	storage.POST("/upload", h.Upload)
	storage.GET("/signed-url", h.SignedURL)

	return h
}

// Upload godoc
//
//	@Summary      Upload a file
//	@Description  Uploads a multipart file to a Supabase Storage bucket and returns the public URL.
//	@Tags         storage
//	@Accept       multipart/form-data
//	@Produce      json
//	@Security     BearerAuth
//	@Param        bucket  query     string  false  "Bucket name (avatars|school-images|documents|profile-photos)"
//	@Param        file    formData  file    true   "File to upload"
//	@Success      200     {object}  response.Response{data=UploadResponse}
//	@Failure      400     {object}  response.Response
//	@Router       /storage/upload [post]
func (h *Handler) Upload(c echo.Context) error {
	bucket := c.QueryParam("bucket")
	if bucket == "" {
		bucket = "avatars"
	}
	if !allowedBuckets[bucket] {
		return response.BadRequest(c, "invalid bucket; allowed: avatars, school-images, documents, profile-photos")
	}

	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "file is required")
	}

	src, err := file.Open()
	if err != nil {
		return response.BadRequest(c, "failed to open file")
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return response.BadRequest(c, "failed to read file")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	userID := middleware.GetUserID(c)
	// Generate unique path using UUID to prevent overwrites on repeat uploads
	uploadID := uuid.New().String()
	path := fmt.Sprintf("%s/%s%s", userID, uploadID, ext)

	contentType := supabasePkg.ContentTypeFromFilename(file.Filename)
	url, err := h.supabase.UploadFile(c.Request().Context(), bucket, path, data, contentType)
	if err != nil {
		return response.InternalError(c, "upload failed")
	}

	return response.OK(c, "file uploaded", UploadResponse{URL: url, Path: path, Bucket: bucket})
}

// SignedURL godoc
//
//	@Summary      Get signed URL
//	@Description  Generates a time-limited signed URL for a private bucket file.
//	@Tags         storage
//	@Produce      json
//	@Security     BearerAuth
//	@Param        bucket  query  string  true  "Bucket name"
//	@Param        path    query  string  true  "File path within bucket"
//	@Success      200     {object}  response.Response{data=SignedURLResponse}
//	@Failure      400     {object}  response.Response
//	@Router       /storage/signed-url [get]
func (h *Handler) SignedURL(c echo.Context) error {
	bucket := c.QueryParam("bucket")
	path := c.QueryParam("path")

	if bucket == "" || path == "" {
		return response.BadRequest(c, "bucket and path are required")
	}
	if !allowedBuckets[bucket] {
		return response.BadRequest(c, "invalid bucket")
	}

	url, err := h.supabase.GetSignedURL(c.Request().Context(), bucket, path, 3600)
	if err != nil {
		return response.InternalError(c, "failed to generate signed URL")
	}

	return response.OK(c, "signed URL generated", SignedURLResponse{URL: url})
}

type UploadResponse struct {
	URL    string `json:"url"`
	Path   string `json:"path"`
	Bucket string `json:"bucket"`
}

type SignedURLResponse struct {
	URL string `json:"url"`
}
