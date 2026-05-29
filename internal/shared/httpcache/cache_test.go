package httpcache

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func okHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"hello": "world"})
}

// newServer wires the middleware around a fixed-body 200 handler.
func newServer(directive string) *echo.Echo {
	e := echo.New()
	e.GET("/r", okHandler, Middleware(directive))
	e.POST("/r", okHandler, Middleware(directive))
	e.GET("/missing", func(c echo.Context) error {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "nope"})
	}, Middleware(directive))
	return e
}

func TestGetSetsCacheHeadersAndBody(t *testing.T) {
	e := newServer(Reference)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/r", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Cache-Control"); got != Reference {
		t.Errorf("Cache-Control = %q, want %q", got, Reference)
	}
	if rec.Header().Get("ETag") == "" {
		t.Error("ETag header missing")
	}
	if got := rec.Header().Get("Vary"); got != "Authorization" {
		t.Errorf("Vary = %q, want Authorization", got)
	}
	if rec.Body.Len() == 0 {
		t.Error("body should be present on a fresh request")
	}
}

func TestIfNoneMatchReturns304(t *testing.T) {
	e := newServer(Reference)

	first := httptest.NewRecorder()
	e.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/r", nil))
	etag := first.Header().Get("ETag")
	if etag == "" {
		t.Fatal("no ETag from first response")
	}

	second := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/r", nil)
	req.Header.Set("If-None-Match", etag)
	e.ServeHTTP(second, req)

	if second.Code != http.StatusNotModified {
		t.Fatalf("status = %d, want 304", second.Code)
	}
	if second.Body.Len() != 0 {
		t.Errorf("304 body should be empty, got %q", second.Body.String())
	}
	if second.Header().Get("ETag") != etag {
		t.Error("304 should echo the ETag")
	}
}

func TestWildcardIfNoneMatchReturns304(t *testing.T) {
	e := newServer(Reference)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/r", nil)
	req.Header.Set("If-None-Match", "*")
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotModified {
		t.Fatalf("status = %d, want 304 for If-None-Match: *", rec.Code)
	}
}

func TestStaleEtagReturnsFullBody(t *testing.T) {
	e := newServer(Reference)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/r", nil)
	req.Header.Set("If-None-Match", `"deadbeef"`)
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 for non-matching ETag", rec.Code)
	}
	if rec.Body.Len() == 0 {
		t.Error("non-matching ETag should yield full body")
	}
}

func TestNonGetIsNotCached(t *testing.T) {
	e := newServer(Reference)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/r", nil))

	if rec.Header().Get("ETag") != "" {
		t.Error("POST responses must not get an ETag")
	}
	if rec.Header().Get("Cache-Control") != "" {
		t.Error("POST responses must not get Cache-Control")
	}
}

func TestNon200IsNotCached(t *testing.T) {
	e := newServer(Reference)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/missing", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if rec.Header().Get("ETag") != "" {
		t.Error("404 responses must not get an ETag")
	}
}
