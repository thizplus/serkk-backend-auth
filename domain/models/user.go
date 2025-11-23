package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID          uuid.UUID  `gorm:"primaryKey;type:uuid"`
	Email       string     `gorm:"uniqueIndex;not null"`
	Username    string     `gorm:"uniqueIndex;not null"`
	Password    *string    `gorm:"default:null"` // Nullable for OAuth users
	FirstName   string
	LastName    string
	DisplayName string     // For OAuth users (LINE, Facebook)
	Avatar      string
	Role        string     `gorm:"default:'user'"`
	IsActive    bool       `gorm:"default:true"`

	// OAuth fields
	IsOAuthUser   bool       `gorm:"default:false"`
	OAuthProvider string     `gorm:"index"` // 'google', 'facebook', 'line'
	OAuthID       string     `gorm:"index"`
	EmailVerified bool       `gorm:"default:false;index"`
	LastLoginAt   *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (User) TableName() string {
	return "users"
}

// BeforeCreate hook to generate UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}