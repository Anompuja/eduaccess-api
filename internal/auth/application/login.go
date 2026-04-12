package application

import (
	"context"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	pkgjwt "github.com/eduaccess/eduaccess-api/pkg/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// LoginCommand is the input for the login use-case.
type LoginCommand struct {
	Email    string
	Password string
}

// TokenPair is returned on successful login or refresh.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// LoginHandler handles the LoginCommand.
type LoginHandler struct {
	users         domain.UserRepository
	refreshTokens domain.RefreshTokenRepository
}

func NewLoginHandler(users domain.UserRepository, refreshTokens domain.RefreshTokenRepository) *LoginHandler {
	return &LoginHandler{users: users, refreshTokens: refreshTokens}
}

func (h *LoginHandler) Handle(ctx context.Context, cmd LoginCommand) (*TokenPair, error) {
	user, err := h.users.FindByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, apperror.New(apperror.ErrUnauthorized, "invalid credentials")
	}

	if !user.IsActive() {
		return nil, apperror.New(apperror.ErrForbidden, "account is deactivated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(cmd.Password)); err != nil {
		return nil, apperror.New(apperror.ErrWrongPassword, "invalid credentials")
	}

	accessToken, err := pkgjwt.GenerateAccessToken(user.ID, user.SchoolID, user.Role)
	if err != nil {
		return nil, apperror.New(apperror.ErrInternal, "failed to generate access token")
	}

	refreshTokenStr, err := pkgjwt.GenerateRefreshToken(user.ID, user.SchoolID, user.Role)
	if err != nil {
		return nil, apperror.New(apperror.ErrInternal, "failed to generate refresh token")
	}

	rt := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refreshTokenStr,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}
	if err := h.refreshTokens.Create(ctx, rt); err != nil {
		return nil, apperror.New(apperror.ErrInternal, "failed to save refresh token")
	}

	return &TokenPair{AccessToken: accessToken, RefreshToken: refreshTokenStr}, nil
}
