package middleware

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/eduaccess/eduaccess-api/pkg/database"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"

	"github.com/eduaccess/eduaccess-api/internal/shared/response"
)

// Context keys for values injected by the auth middleware.
const (
	ContextKeyUserID   = "user_id"
	ContextKeySchoolID = "school_id"
	ContextKeyRole     = "role"
)

// supabaseClaims maps the JWT payload issued by Supabase Auth.
// app_role is our custom claim (injected via the custom_access_token_hook function),
// distinct from Supabase's built-in 'role' claim which is used for Postgres RLS.
type supabaseClaims struct {
	AppRole  string  `json:"app_role"`
	SchoolID *string `json:"school_id"`
	jwt.RegisteredClaims
}

var (
	jwksMu     sync.RWMutex
	jwksKeys     map[string]*ecdsa.PublicKey
	jwksExpiry   time.Time
	authDBOnce   sync.Once
	authDB       *gorm.DB
	authDBErr    error
	fallbackCache = cache.New(5*time.Minute, 10*time.Minute)
)

type jwksResponse struct {
	Keys []struct {
		Kid string `json:"kid"`
		Kty string `json:"kty"`
		Alg string `json:"alg"`
		X   string `json:"x"`
		Y   string `json:"y"`
	} `json:"keys"`
}

// loadJWKS fetches Supabase JWKS and caches with 1-hour TTL.
// Forces refresh if kid is unknown, or if cache is expired.
func loadJWKS(kid string) (map[string]*ecdsa.PublicKey, error) {
	jwksMu.RLock()
	// Check if cache is still valid and kid exists (or kid is empty, allowing any key)
	if time.Now().Before(jwksExpiry) && (kid == "" || jwksKeys[kid] != nil) && len(jwksKeys) > 0 {
		defer jwksMu.RUnlock()
		return jwksKeys, nil
	}
	jwksMu.RUnlock()

	// Fetch fresh JWKS
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		return nil, errors.New("SUPABASE_URL not configured")
	}

	resp, err := http.Get(supabaseURL + "/auth/v1/.well-known/jwks.json")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}
	if len(jwks.Keys) == 0 {
		return nil, errors.New("no keys in JWKS response")
	}

	keys := make(map[string]*ecdsa.PublicKey, len(jwks.Keys))
	for _, key := range jwks.Keys {
		if key.Kty != "EC" {
			continue
		}

		xBytes, err := base64.RawURLEncoding.DecodeString(key.X)
		if err != nil {
			return nil, fmt.Errorf("invalid JWKS x: %w", err)
		}
		yBytes, err := base64.RawURLEncoding.DecodeString(key.Y)
		if err != nil {
			return nil, fmt.Errorf("invalid JWKS y: %w", err)
		}

		keys[key.Kid] = &ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     new(big.Int).SetBytes(xBytes),
			Y:     new(big.Int).SetBytes(yBytes),
		}
	}

	if len(keys) == 0 {
		return nil, errors.New("no usable EC keys in JWKS response")
	}

	jwksMu.Lock()
	jwksKeys = keys
	jwksExpiry = time.Now().Add(1 * time.Hour) // 1-hour cache TTL
	jwksMu.Unlock()

	return keys, nil
}

func loadAuthDB() (*gorm.DB, error) {
	authDBOnce.Do(func() {
		authDB, authDBErr = database.Connect()
	})
	return authDB, authDBErr
}

type fallbackAuthRow struct {
	RoleName *string `gorm:"column:role_name"`
	SchoolID *string `gorm:"column:school_id"`
}

// RequireAuth validates the Supabase Bearer JWT and injects claims into context.
func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		header := c.Request().Header.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return response.Unauthorized(c, "missing or malformed authorization header")
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")

		// Parse without verification first to extract kid from header
		unverified, _, _ := jwt.NewParser().ParseUnverified(tokenStr, &supabaseClaims{})
		kid, _ := unverified.Header["kid"].(string)

		// Load JWKS (with refresh on missing kid or expired cache)
		signingKeys, err := loadJWKS(kid)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "auth config error: " + err.Error()})
		}

		token, err := jwt.ParseWithClaims(tokenStr, &supabaseClaims{}, func(t *jwt.Token) (interface{}, error) {
			kid, _ := t.Header["kid"].(string)
			alg, _ := t.Header["alg"].(string)

			switch alg {
			case "ES256":
				if kid == "" {
					for _, key := range signingKeys {
						return key, nil
					}
					return nil, errors.New("missing kid in token header")
				}
				key, ok := signingKeys[kid]
				if !ok {
					return nil, fmt.Errorf("unknown signing key: %s", kid)
				}
				return key, nil
			default:
				secret := strings.TrimSpace(os.Getenv("SUPABASE_JWT_SECRET"))
				if secret == "" {
					return nil, errors.New("SUPABASE_JWT_SECRET not configured")
				}
				return []byte(secret), nil
			}
		})
		if err != nil || !token.Valid {
			return response.Unauthorized(c, "invalid or expired token")
		}

		claims, ok := token.Claims.(*supabaseClaims)
		if !ok {
			return response.Unauthorized(c, "malformed token claims")
		}

		// sub is the Supabase auth.users UUID
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return response.Unauthorized(c, "invalid token subject")
		}

		role := claims.AppRole
		var schoolID *uuid.UUID
		if claims.SchoolID != nil && *claims.SchoolID != "" && *claims.SchoolID != "null" {
			if parsedSchoolID, err := uuid.Parse(*claims.SchoolID); err == nil {
				schoolID = &parsedSchoolID
			}
		}

		if role == "" {
			if cached, found := fallbackCache.Get(userID.String()); found {
				row := cached.(fallbackAuthRow)
				if row.RoleName != nil {
					role = *row.RoleName
				}
				if row.SchoolID != nil && *row.SchoolID != "" && *row.SchoolID != "null" {
					if parsedSchoolID, err := uuid.Parse(*row.SchoolID); err == nil {
						schoolID = &parsedSchoolID
					}
				}
			} else if db, err := loadAuthDB(); err == nil {
				var row fallbackAuthRow
				fallbackSQL := `
SELECT
	r.name AS role_name,
	preferred_school.school_id
FROM users u
LEFT JOIN model_has_roles mhr ON mhr.user_id = u.id
LEFT JOIN roles r ON r.id = mhr.role_id
LEFT JOIN LATERAL (
	SELECT su.school_id
	FROM school_users su
	LEFT JOIN schools s ON s.id = su.school_id
	WHERE su.user_id = u.id
		AND su.deleted_at IS NULL
		AND (s.id IS NULL OR s.deleted_at IS NULL)
	ORDER BY
		CASE WHEN s.status = 'active' THEN 0 ELSE 1 END,
		su.created_at DESC,
		su.school_id
	LIMIT 1
) preferred_school ON TRUE
WHERE u.id = ? AND u.deleted_at IS NULL
LIMIT 1`
				if err := db.WithContext(c.Request().Context()).Raw(fallbackSQL, userID).Scan(&row).Error; err == nil {
					fallbackCache.Set(userID.String(), row, cache.DefaultExpiration)
					if row.RoleName != nil {
						role = *row.RoleName
					}
					if row.SchoolID != nil && *row.SchoolID != "" && *row.SchoolID != "null" {
						if parsedSchoolID, err := uuid.Parse(*row.SchoolID); err == nil {
							schoolID = &parsedSchoolID
						}
					}
				}
			}
		}

		c.Set(ContextKeyUserID, userID)
		c.Set(ContextKeyRole, role)
		if schoolID != nil {
			c.Set(ContextKeySchoolID, schoolID)
		}

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

// ValidateToken parses a raw Supabase JWT string and returns the auth.users UUID.
// Used by non-HTTP transports (e.g. WebSocket) that cannot use the RequireAuth middleware.
func ValidateToken(tokenStr string) (uuid.UUID, error) {
	if tokenStr == "" {
		return uuid.Nil, errors.New("empty token")
	}
	unverified, _, _ := jwt.NewParser().ParseUnverified(tokenStr, &supabaseClaims{})
	if unverified == nil {
		return uuid.Nil, errors.New("invalid token format")
	}
	kid, _ := unverified.Header["kid"].(string)

	signingKeys, err := loadJWKS(kid)
	if err != nil {
		return uuid.Nil, err
	}

	token, err := jwt.ParseWithClaims(tokenStr, &supabaseClaims{}, func(t *jwt.Token) (interface{}, error) {
		k, _ := t.Header["kid"].(string)
		alg, _ := t.Header["alg"].(string)
		switch alg {
		case "ES256":
			if k == "" {
				for _, key := range signingKeys {
					return key, nil
				}
				return nil, errors.New("missing kid in token header")
			}
			key, ok := signingKeys[k]
			if !ok {
				return nil, fmt.Errorf("unknown signing key: %s", k)
			}
			return key, nil
		default:
			secret := strings.TrimSpace(os.Getenv("SUPABASE_JWT_SECRET"))
			if secret == "" {
				return nil, errors.New("SUPABASE_JWT_SECRET not configured")
			}
			return []byte(secret), nil
		}
	})
	if err != nil || !token.Valid {
		return uuid.Nil, errors.New("invalid or expired token")
	}

	claims, ok := token.Claims.(*supabaseClaims)
	if !ok {
		return uuid.Nil, errors.New("malformed token claims")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, errors.New("invalid token subject")
	}

	return userID, nil
}
