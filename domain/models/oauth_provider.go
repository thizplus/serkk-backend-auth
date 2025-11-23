package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type OAuthProvider struct {
	ID             uuid.UUID      `gorm:"primaryKey;type:uuid"`
	UserID         uuid.UUID      `gorm:"not null;index"`
	Provider       string         `gorm:"not null;size:50;index"`   // 'google', 'facebook', 'line'
	ProviderID     string         `gorm:"not null;size:255"`        // OAuth provider's user ID
	AccessToken    string         `gorm:"type:text"`
	RefreshToken   string         `gorm:"type:text"`
	TokenExpiresAt *time.Time
	ProfileData    datatypes.JSON `gorm:"type:jsonb"` // Raw profile data from provider
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// Relationship
	User User `gorm:"foreignKey:UserID"`
}

func (OAuthProvider) TableName() string {
	return "oauth_providers"
}

// BeforeCreate hook to generate UUID
func (op *OAuthProvider) BeforeCreate(tx *gorm.DB) error {
	if op.ID == uuid.Nil {
		op.ID = uuid.New()
	}
	return nil
}
