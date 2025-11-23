package services

import (
	"context"
	"gofiber-template/domain/models"
)

type OAuthService interface {
	// Google OAuth
	GetGoogleAuthURL(state string) string
	HandleGoogleCallback(ctx context.Context, code string) (*models.User, string, bool, error)

	// Facebook OAuth
	GetFacebookAuthURL(state string) string
	HandleFacebookCallback(ctx context.Context, code string) (*models.User, string, bool, error)

	// LINE OAuth
	GetLINEAuthURL(state string) string
	HandleLINECallback(ctx context.Context, code string) (*models.User, string, bool, error)
}
