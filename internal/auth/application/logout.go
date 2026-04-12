package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
)

// LogoutCommand is the input for the logout use-case.
type LogoutCommand struct {
	RefreshToken string
}

// LogoutHandler handles the LogoutCommand.
type LogoutHandler struct {
	refreshTokens domain.RefreshTokenRepository
}

func NewLogoutHandler(refreshTokens domain.RefreshTokenRepository) *LogoutHandler {
	return &LogoutHandler{refreshTokens: refreshTokens}
}

func (h *LogoutHandler) Handle(ctx context.Context, cmd LogoutCommand) error {
	if err := h.refreshTokens.DeleteByToken(ctx, cmd.RefreshToken); err != nil {
		return apperror.New(apperror.ErrInternal, "failed to revoke token")
	}
	return nil
}
