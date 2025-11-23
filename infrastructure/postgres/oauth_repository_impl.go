package postgres

import (
	"context"
	"errors"
	"gofiber-template/domain/models"
	"gofiber-template/domain/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type oauthRepository struct {
	db *gorm.DB
}

func NewOAuthRepository(db *gorm.DB) repositories.OAuthRepository {
	return &oauthRepository{db: db}
}

func (r *oauthRepository) Create(ctx context.Context, oauth *models.OAuthProvider) error {
	return r.db.WithContext(ctx).Create(oauth).Error
}

func (r *oauthRepository) FindByProviderAndProviderID(ctx context.Context, provider, providerID string) (*models.OAuthProvider, error) {
	var oauth models.OAuthProvider
	err := r.db.WithContext(ctx).
		Where("provider = ? AND provider_id = ?", provider, providerID).
		Preload("User").
		First(&oauth).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &oauth, nil
}

func (r *oauthRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*models.OAuthProvider, error) {
	var oauths []*models.OAuthProvider
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&oauths).Error

	if err != nil {
		return nil, err
	}

	return oauths, nil
}

func (r *oauthRepository) Update(ctx context.Context, oauth *models.OAuthProvider) error {
	return r.db.WithContext(ctx).Save(oauth).Error
}

func (r *oauthRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.OAuthProvider{}, id).Error
}
