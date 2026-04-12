package jwt

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims is the JWT payload stored in every token.
type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	SchoolID  *uuid.UUID `json:"school_id,omitempty"` // nil for superadmin
	Role      string    `json:"role"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a 15-minute access token.
func GenerateAccessToken(userID uuid.UUID, schoolID *uuid.UUID, role string) (string, error) {
	return generate(userID, schoolID, role, AccessToken, 15*time.Minute)
}

// GenerateRefreshToken creates a 7-day refresh token.
func GenerateRefreshToken(userID uuid.UUID, schoolID *uuid.UUID, role string) (string, error) {
	return generate(userID, schoolID, role, RefreshToken, 7*24*time.Hour)
}

func generate(userID uuid.UUID, schoolID *uuid.UUID, role string, tokenType TokenType, ttl time.Duration) (string, error) {
	secret := secret()
	now := time.Now()

	claims := Claims{
		UserID:    userID,
		SchoolID:  schoolID,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Issuer:    "eduaccess-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// Parse validates the token and returns its claims.
func Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret()), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func secret() string {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		panic("JWT_SECRET environment variable is not set")
	}
	return s
}
