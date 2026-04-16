package application

import (
	"context"
	"log"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	pkgjwt "github.com/eduaccess/eduaccess-api/pkg/jwt"
	"github.com/google/uuid"
)

// RefreshCommand is the input for the token-refresh use-case.
type RefreshCommand struct {
	RefreshToken string
}

// RefreshHandler handles the RefreshCommand.
type RefreshHandler struct {
	users         domain.UserRepository
	refreshTokens domain.RefreshTokenRepository
}

func NewRefreshHandler(users domain.UserRepository, refreshTokens domain.RefreshTokenRepository) *RefreshHandler {
	return &RefreshHandler{users: users, refreshTokens: refreshTokens}
}

func (h *RefreshHandler) Handle(ctx context.Context, cmd RefreshCommand) (*TokenPair, error) {
	claims, err := pkgjwt.Parse(cmd.RefreshToken)
	if err != nil {
		return nil, apperror.New(apperror.ErrInvalidToken, "invalid or expired refresh token")
	}
	if claims.TokenType != pkgjwt.RefreshToken {
		return nil, apperror.New(apperror.ErrInvalidToken, "not a refresh token")
	}

	stored, err := h.refreshTokens.FindByToken(ctx, cmd.RefreshToken)
	if err != nil {
		return nil, apperror.New(apperror.ErrTokenRevoked, "refresh token not found or revoked")
	}
	if time.Now().After(stored.ExpiresAt) {
		_ = h.refreshTokens.DeleteByToken(ctx, cmd.RefreshToken)
		return nil, apperror.New(apperror.ErrInvalidToken, "refresh token expired")
	}

	user, err := h.users.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, apperror.New(apperror.ErrUnauthorized, "user not found")
	}
	if !user.IsActive() {
		return nil, apperror.New(apperror.ErrForbidden, "account is deactivated")
	}
	if user.Role != domain.RoleSuperadmin && user.SchoolID == nil {
		log.Printf("auth refresh: user %s with role %s has no school membership", user.ID.String(), user.Role)
		return nil, apperror.New(apperror.ErrForbidden, "school context required")
	}

	// Rotate: delete old token, issue new pair
	_ = h.refreshTokens.DeleteByToken(ctx, cmd.RefreshToken)

	accessToken, err := pkgjwt.GenerateAccessToken(user.ID, user.SchoolID, user.Role)
	if err != nil {
		return nil, apperror.New(apperror.ErrInternal, "failed to generate access token")
	}
	newRefreshStr, err := pkgjwt.GenerateRefreshToken(user.ID, user.SchoolID, user.Role)
	if err != nil {
		return nil, apperror.New(apperror.ErrInternal, "failed to generate refresh token")
	}

	rt := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     newRefreshStr,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}
	if err := h.refreshTokens.Create(ctx, rt); err != nil {
		return nil, apperror.New(apperror.ErrInternal, "failed to save refresh token")
	}

	return &TokenPair{AccessToken: accessToken, RefreshToken: newRefreshStr}, nil
}
