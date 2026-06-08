// Package httpcache provides an Echo middleware that adds HTTP caching
// semantics (Cache-Control, ETag, conditional revalidation) to safe,
// authenticated GET responses.
//
// All presets use the "private" directive: every endpoint these protect
// returns user- or tenant-scoped data behind an Authorization header, so a
// shared cache (CDN/proxy) must never store them. "Vary: Authorization" is
// always emitted as defense-in-depth so a cache keyed only by URL can never
// serve one principal's response to another.
package httpcache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// Cache-Control presets, tuned per endpoint volatility. ETag revalidation
// backs every preset, so even a zero freshness window stays cheap.
const (
	// Profile identity is derived from the JWT and changes only on re-login.
	Profile = "private, max-age=300"
	// Stats are aggregates that tolerate a short staleness window.
	Stats = "private, max-age=60, stale-while-revalidate=30"
	// Reference covers academic config data that changes infrequently.
	Reference = "private, max-age=120"
	// ShortLived covers mutable CRUD lists/details (e.g. parents).
	ShortLived = "private, max-age=30"
	// AlwaysRevalidate forces a conditional request every time: zero
	// staleness for live operational data, but 304s keep it bandwidth-cheap.
	AlwaysRevalidate = "private, no-cache"
)

// Middleware returns an Echo middleware that, for successful (200) GET/HEAD
// responses, attaches the given Cache-Control directive plus a strong ETag,
// and answers matching If-None-Match requests with 304 Not Modified.
//
// Non-GET methods and non-200 responses pass through untouched, so the
// middleware is safe to attach at the group level alongside mutating routes.
func Middleware(cacheControl string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			if req.Method != http.MethodGet && req.Method != http.MethodHead {
				return next(c)
			}

			res := c.Response()
			orig := res.Writer
			cw := &captureWriter{ResponseWriter: orig, buf: new(bytes.Buffer)}
			res.Writer = cw

			err := next(c)
			res.Writer = orig

			// On error, or any response the handler already streamed through
			// (non-200), there is nothing buffered to revalidate.
			if err != nil || cw.passthrough || cw.status != http.StatusOK {
				return err
			}

			body := cw.buf.Bytes()
			etag := makeETag(body)
			header := orig.Header()
			if header.Get("Cache-Control") == "" {
				header.Set("Cache-Control", cacheControl)
			}
			header.Set("ETag", etag)
			addVary(header, "Authorization")

			if ifNoneMatch(req.Header.Get("If-None-Match"), etag) {
				header.Del("Content-Length")
				res.Status = http.StatusNotModified
				res.Size = 0
				orig.WriteHeader(http.StatusNotModified)
				return nil
			}

			header.Set("Content-Length", strconv.Itoa(len(body)))
			orig.WriteHeader(http.StatusOK)
			_, werr := orig.Write(body)
			return werr
		}
	}
}

// captureWriter buffers a 200 response so its body can be hashed into an ETag.
// Any non-200 status switches to pass-through, writing straight to the
// underlying writer (used by error responses such as 403/404).
type captureWriter struct {
	http.ResponseWriter
	buf         *bytes.Buffer
	status      int
	passthrough bool
	wroteHeader bool
}

func (w *captureWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.status = code
	if code != http.StatusOK {
		w.passthrough = true
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *captureWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	if w.passthrough {
		return w.ResponseWriter.Write(b)
	}
	return w.buf.Write(b)
}

func makeETag(body []byte) string {
	sum := sha256.Sum256(body)
	return `"` + hex.EncodeToString(sum[:16]) + `"`
}

// ifNoneMatch reports whether the If-None-Match header satisfies the current
// ETag. It accepts "*", a comma-separated list, and weak ("W/") validators.
func ifNoneMatch(header, etag string) bool {
	if header == "" {
		return false
	}
	want := strings.TrimPrefix(etag, "W/")
	for tok := range strings.SplitSeq(header, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "*" {
			return true
		}
		if strings.TrimPrefix(tok, "W/") == want {
			return true
		}
	}
	return false
}

func addVary(header http.Header, value string) {
	for _, existing := range header.Values("Vary") {
		for field := range strings.SplitSeq(existing, ",") {
			if strings.EqualFold(strings.TrimSpace(field), value) {
				return
			}
		}
	}
	header.Add("Vary", value)
}
