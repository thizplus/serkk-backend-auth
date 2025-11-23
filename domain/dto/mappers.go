package dto

import (
	"gofiber-template/domain/models"
	"strings"
)

func UserToUserResponse(user *models.User) *UserResponse {
	if user == nil {
		return nil
	}

	// Generate displayName from firstName + lastName
	displayName := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if displayName == "" {
		// Fallback to displayName from model, or username if empty
		displayName = user.DisplayName
		if displayName == "" {
			displayName = user.Username
		}
	}

	return &UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		Username:    user.Username,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
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
		Email:     req.Email,
		Username:  req.Username,
		Password:  &password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}
}

func UpdateUserRequestToUser(req *UpdateUserRequest) *models.User {
	return &models.User{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DisplayName: req.DisplayName,
		Avatar:      req.Avatar,
	}
}