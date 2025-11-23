package repositories

import (
	"context"
	"gofiber-template/domain/models"
	"github.com/google/uuid"
)

type OAuthRepository interface {
	Create(ctx context.Context, oauth *models.OAuthProvider) error
	FindByProviderAndProviderID(ctx context.Context, provider, providerID string) (*models.OAuthProvider, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*models.OAuthProvider, error)
	Update(ctx context.Context, oauth *models.OAuthProvider) error
	Delete(ctx context.Context, id uuid.UUID) error
}
