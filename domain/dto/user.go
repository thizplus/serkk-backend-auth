package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateUserRequest struct {
	Email       string `json:"email" validate:"required,email,max=255"`
	Username    string `json:"username" validate:"required,min=3,max=20,alphanum"`
	Password    string `json:"password" validate:"required,min=8,max=72"`
	DisplayName string `json:"displayName" validate:"required,min=1,max=100"`
}

type UpdateUserRequest struct {
	DisplayName string `json:"displayName" validate:"omitempty,min=1,max=100"`
	Avatar      string `json:"avatar" validate:"omitempty,url,max=500"`
}

type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	DisplayName string    `json:"displayName"`
	Avatar      string    `json:"avatar"`
	Role        string    `json:"role"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type UserListResponse struct {
	Users []UserResponse `json:"users"`
	Meta  PaginationMeta `json:"meta"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8,max=72"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,eqfield=NewPassword"`
}
