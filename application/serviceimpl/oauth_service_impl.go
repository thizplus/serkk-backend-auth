package serviceimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gofiber-template/domain/dto"
	"gofiber-template/domain/models"
	"gofiber-template/domain/repositories"
	"gofiber-template/domain/services"
	"gofiber-template/pkg/config"
	"gofiber-template/pkg/contextutil"
	"gofiber-template/pkg/logger"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
	googleOAuth2 "google.golang.org/api/oauth2/v2"
	"gorm.io/datatypes"
)

type oauthService struct {
	userRepo        repositories.UserRepository
	oauthRepo       repositories.OAuthRepository
	userService     services.UserService
	syncService     *SyncService
	googleConfig    *oauth2.Config
	facebookConfig  *oauth2.Config
	lineConfig      *oauth2.Config
}

func NewOAuthService(
	userRepo repositories.UserRepository,
	oauthRepo repositories.OAuthRepository,
	userService services.UserService,
	syncService *SyncService,
	cfg *config.Config,
) services.OAuthService {
	// Google OAuth Config
	googleConfig := &oauth2.Config{
		ClientID:     cfg.OAuth.GoogleClientID,
		ClientSecret: cfg.OAuth.GoogleClientSecret,
		RedirectURL:  cfg.OAuth.GoogleRedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint, // Use official Google OAuth2 endpoints
	}

	// Facebook OAuth Config
	facebookConfig := &oauth2.Config{
		ClientID:     cfg.OAuth.FacebookClientID,
		ClientSecret: cfg.OAuth.FacebookClientSecret,
		RedirectURL:  cfg.OAuth.FacebookRedirectURL,
		Scopes:       []string{"email", "public_profile"},
		Endpoint:     facebook.Endpoint,
	}

	// LINE OAuth Config
	lineConfig := &oauth2.Config{
		ClientID:     cfg.OAuth.LINEClientID,
		ClientSecret: cfg.OAuth.LINEClientSecret,
		RedirectURL:  cfg.OAuth.LINERedirectURL,
		Scopes:       []string{"profile", "openid", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://access.line.me/oauth2/v2.1/authorize",
			TokenURL: "https://api.line.me/oauth2/v2.1/token",
		},
	}

	return &oauthService{
		userRepo:       userRepo,
		oauthRepo:      oauthRepo,
		userService:    userService,
		syncService:    syncService,
		googleConfig:   googleConfig,
		facebookConfig: facebookConfig,
		lineConfig:     lineConfig,
	}
}

// ==================== Google OAuth ====================

func (s *oauthService) GetGoogleAuthURL(state string) string {
	return s.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (s *oauthService) HandleGoogleCallback(ctx context.Context, code string) (*models.User, string, bool, error) {
	startTime := time.Now()
	requestID := contextutil.GetRequestID(ctx)
	log := logger.GetLogger()

	log.Info("Google OAuth callback started", map[string]interface{}{
		"request_id": requestID,
		"action":     "oauth_google",
		"provider":   "google",
	})

	// Exchange code for token
	token, err := s.googleConfig.Exchange(ctx, code)
	if err != nil {
		log.Error("Google OAuth code exchange failed", map[string]interface{}{
			"request_id": requestID,
			"action":     "oauth_google",
			"error":      err.Error(),
		})
		return nil, "", false, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info from Google
	client := s.googleConfig.Client(ctx, token)
	oauth2Service, err := googleOAuth2.New(client)
	if err != nil {
		log.Error("Google OAuth2 service creation failed", map[string]interface{}{
			"request_id": requestID,
			"action":     "oauth_google",
			"error":      err.Error(),
		})
		return nil, "", false, fmt.Errorf("failed to create oauth2 service: %w", err)
	}

	userInfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		log.Error("Failed to get Google user info", map[string]interface{}{
			"request_id": requestID,
			"action":     "oauth_google",
			"error":      err.Error(),
		})
		return nil, "", false, fmt.Errorf("failed to get user info: %w", err)
	}

	googleUserInfo := &dto.GoogleUserInfo{
		ID:            userInfo.Id,
		Email:         userInfo.Email,
		VerifiedEmail: userInfo.VerifiedEmail != nil && *userInfo.VerifiedEmail,
		Name:          userInfo.Name,
		GivenName:     userInfo.GivenName,
		FamilyName:    userInfo.FamilyName,
		Picture:       userInfo.Picture,
	}

	log.Info("Google user info retrieved", map[string]interface{}{
		"request_id":   requestID,
		"action":       "oauth_google",
		"provider_id":  googleUserInfo.ID,
		"email":        googleUserInfo.Email,
	})

	// Find or create user
	user, isNewUser, err := s.findOrCreateOAuthUser(ctx, "google", googleUserInfo.ID, googleUserInfo, token)
	if err != nil {
		log.Error("Failed to find or create OAuth user", map[string]interface{}{
			"request_id": requestID,
			"action":     "oauth_google",
			"provider":   "google",
			"error":      err.Error(),
		})
		return nil, "", false, err
	}

	// Generate JWT
	jwtToken, err := s.userService.GenerateJWT(user)
	if err != nil {
		log.Error("JWT generation failed", map[string]interface{}{
			"request_id": requestID,
			"action":     "oauth_google",
			"user_id":    user.ID.String(),
			"error":      err.Error(),
		})
		return nil, "", false, fmt.Errorf("failed to generate JWT: %w", err)
	}

	duration := time.Since(startTime).Milliseconds()
	log.Info("Google OAuth completed successfully", map[string]interface{}{
		"request_id":  requestID,
		"action":      "oauth_google",
		"user_id":     user.ID.String(),
		"is_new_user": isNewUser,
		"duration_ms": duration,
	})

	return user, jwtToken, isNewUser, nil
}

// ==================== Facebook OAuth ====================

func (s *oauthService) GetFacebookAuthURL(state string) string {
	return s.facebookConfig.AuthCodeURL(state)
}

func (s *oauthService) HandleFacebookCallback(ctx context.Context, code string) (*models.User, string, bool, error) {
	// Exchange code for token
	token, err := s.facebookConfig.Exchange(ctx, code)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info from Facebook
	client := s.facebookConfig.Client(ctx, token)
	resp, err := client.Get("https://graph.facebook.com/me?fields=id,name,email,picture")
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	var fbUserInfo dto.FacebookUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&fbUserInfo); err != nil {
		return nil, "", false, fmt.Errorf("failed to decode user info: %w", err)
	}

	// Find or create user
	user, isNewUser, err := s.findOrCreateOAuthUser(ctx, "facebook", fbUserInfo.ID, &fbUserInfo, token)
	if err != nil {
		return nil, "", false, err
	}

	// Generate JWT
	jwtToken, err := s.userService.GenerateJWT(user)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to generate JWT: %w", err)
	}

	return user, jwtToken, isNewUser, nil
}

// ==================== LINE OAuth ====================

func (s *oauthService) GetLINEAuthURL(state string) string {
	return s.lineConfig.AuthCodeURL(state)
}

func (s *oauthService) HandleLINECallback(ctx context.Context, code string) (*models.User, string, bool, error) {
	// Exchange code for token
	token, err := s.lineConfig.Exchange(ctx, code)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user profile from LINE
	client := s.lineConfig.Client(ctx, token)
	resp, err := client.Get("https://api.line.me/v2/profile")
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to get user profile: %w", err)
	}
	defer resp.Body.Close()

	var lineUserInfo dto.LINEUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&lineUserInfo); err != nil {
		return nil, "", false, fmt.Errorf("failed to decode user profile: %w", err)
	}

	// Get email from ID token (if available)
	var email string
	if idToken, ok := token.Extra("id_token").(string); ok {
		email = s.extractEmailFromLINEIDToken(idToken)
	}

	// If no email, generate one
	if email == "" {
		email = fmt.Sprintf("line_%s@oauth.local", lineUserInfo.UserID)
	}

	// Add email to user info
	lineUserInfo.UserID = lineUserInfo.UserID

	// Find or create user
	user, isNewUser, err := s.findOrCreateOAuthUser(ctx, "line", lineUserInfo.UserID, &lineUserInfo, token)
	if err != nil {
		return nil, "", false, err
	}

	// Generate JWT
	jwtToken, err := s.userService.GenerateJWT(user)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to generate JWT: %w", err)
	}

	return user, jwtToken, isNewUser, nil
}

// ==================== Helper Methods ====================

func (s *oauthService) findOrCreateOAuthUser(
	ctx context.Context,
	provider string,
	providerID string,
	userInfo interface{},
	token *oauth2.Token,
) (*models.User, bool, error) {
	// Check if OAuth provider exists
	oauthProvider, err := s.oauthRepo.FindByProviderAndProviderID(ctx, provider, providerID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to find oauth provider: %w", err)
	}

	// If exists, return existing user
	if oauthProvider != nil {
		// Update token
		oauthProvider.AccessToken = token.AccessToken
		oauthProvider.RefreshToken = token.RefreshToken
		if !token.Expiry.IsZero() {
			oauthProvider.TokenExpiresAt = &token.Expiry
		}
		if err := s.oauthRepo.Update(ctx, oauthProvider); err != nil {
			return nil, false, fmt.Errorf("failed to update oauth provider: %w", err)
		}

		// Update last login
		now := time.Now()
		oauthProvider.User.LastLoginAt = &now
		if err := s.userRepo.Update(ctx, oauthProvider.User.ID, &oauthProvider.User); err != nil {
			return nil, false, fmt.Errorf("failed to update user: %w", err)
		}

		return &oauthProvider.User, false, nil
	}

	// Extract user data based on provider
	var email, displayName, avatar string

	switch provider {
	case "google":
		info := userInfo.(*dto.GoogleUserInfo)
		email = info.Email
		displayName = info.Name
		avatar = info.Picture
	case "facebook":
		info := userInfo.(*dto.FacebookUserInfo)
		email = info.Email
		displayName = info.Name
		avatar = info.Picture.Data.URL
	case "line":
		info := userInfo.(*dto.LINEUserInfo)
		email = fmt.Sprintf("line_%s@oauth.local", info.UserID)
		displayName = info.DisplayName
		avatar = info.PictureURL
	}

	// Generate username from email or display name
	username := s.generateUsername(email, displayName)

	// Check if email already exists
	existingUser, _ := s.userRepo.GetByEmail(ctx, email)

	var user *models.User
	isNewUser := false

	if existingUser != nil {
		// Link OAuth to existing user
		user = existingUser
	} else {
		// Create new user
		user = &models.User{
			Email:         email,
			Username:      username,
			Password:      nil, // OAuth users don't have password
			DisplayName:   displayName,
			Avatar:        avatar,
			IsOAuthUser:   true,
			OAuthProvider: provider,
			OAuthID:       providerID,
			EmailVerified: true, // OAuth providers verify email
			Role:          "user",
			IsActive:      true,
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, false, fmt.Errorf("failed to create user: %w", err)
		}
		isNewUser = true

		// Sync new OAuth user to backend (async with context)
		go s.syncService.SyncUserWithRetry(ctx, user, "created")
	}

	// Create OAuth provider record
	profileData, _ := json.Marshal(userInfo)
	oauthProviderModel := &models.OAuthProvider{
		UserID:         user.ID,
		Provider:       provider,
		ProviderID:     providerID,
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
		TokenExpiresAt: nil,
		ProfileData:    datatypes.JSON(profileData),
	}
	if !token.Expiry.IsZero() {
		oauthProviderModel.TokenExpiresAt = &token.Expiry
	}

	if err := s.oauthRepo.Create(ctx, oauthProviderModel); err != nil {
		return nil, false, fmt.Errorf("failed to create oauth provider: %w", err)
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	if err := s.userRepo.Update(ctx, user.ID, user); err != nil {
		return nil, false, fmt.Errorf("failed to update user: %w", err)
	}

	return user, isNewUser, nil
}

func (s *oauthService) generateUsername(email, displayName string) string {
	// Try email username first
	if email != "" && !strings.HasPrefix(email, "line_") {
		parts := strings.Split(email, "@")
		if len(parts) > 0 {
			return parts[0] + "_" + uuid.New().String()[:8]
		}
	}

	// Use display name
	if displayName != "" {
		username := strings.ToLower(strings.ReplaceAll(displayName, " ", "_"))
		return username + "_" + uuid.New().String()[:8]
	}

	// Fallback to random
	return "user_" + uuid.New().String()[:12]
}

func (s *oauthService) extractEmailFromLINEIDToken(idToken string) string {
	// This is a simplified version
	// In production, you should properly verify and decode the JWT
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return ""
	}

	// Decode payload (base64)
	payload := parts[1]
	// Add padding if needed
	if len(payload)%4 != 0 {
		payload += strings.Repeat("=", 4-len(payload)%4)
	}

	// This is just a placeholder - properly implement JWT decoding
	return ""
}
