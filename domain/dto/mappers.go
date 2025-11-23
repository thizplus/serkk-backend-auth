package dto

import (
	"gofiber-template/domain/models"
)

func UserToUserResponse(user *models.User) *UserResponse {
	if user == nil {
		return nil
	}

	// Use displayName from database, fallback to username if empty
	displayName := user.DisplayName
	if displayName == "" {
		displayName = user.Username
	}

	return &UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		Username:    user.Username,
		DisplayName: displayName,
		Avatar:      user.Avatar,
		Role:        user.Role,
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

func CreateUserRequestToUser(req *CreateUserRequest) *models.User {
	password := req.Password
	return &models.User{
		Email:       req.Email,
		Username:    req.Username,
		Password:    &password,
		DisplayName: req.DisplayName,
	}
}

func UpdateUserRequestToUser(req *UpdateUserRequest) *models.User {
	return &models.User{
		DisplayName: req.DisplayName,
		Avatar:      req.Avatar,
	}
}