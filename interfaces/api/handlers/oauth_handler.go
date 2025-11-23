package handlers

import (
	"gofiber-template/domain/dto"
	"gofiber-template/domain/services"
	"gofiber-template/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type OAuthHandler struct {
	oauthService services.OAuthService
}

func NewOAuthHandler(oauthService services.OAuthService) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
	}
}

// ==================== Google OAuth ====================

// GetGoogleAuthURL godoc
// @Summary      Get Google OAuth URL
// @Description  Get Google OAuth authorization URL
// @Tags         OAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  dto.OAuthURLResponse
// @Router       /auth/google [get]
func (h *OAuthHandler) GetGoogleAuthURL(c *fiber.Ctx) error {
	state := uuid.New().String()
	authURL := h.oauthService.GetGoogleAuthURL(state)

	return c.JSON(dto.OAuthURLResponse{
		AuthURL: authURL,
	})
}

// HandleGoogleCallback godoc
// @Summary      Handle Google OAuth callback
// @Description  Handle Google OAuth callback and create/login user
// @Tags         OAuth
// @Accept       json
// @Produce      json
// @Param        code   query     string  true  "Authorization code"
// @Success      200    {object}  dto.OAuthLoginResponse
// @Failure      400    {object}  utils.ErrorResponse
// @Failure      500    {object}  utils.ErrorResponse
// @Router       /auth/google/callback [get]
func (h *OAuthHandler) HandleGoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Missing authorization code", nil)
	}

	user, jwtToken, isNewUser, err := h.oauthService.HandleGoogleCallback(c.Context(), code)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to authenticate with Google", err)
	}

	return c.JSON(dto.OAuthLoginResponse{
		AccessToken: jwtToken,
		TokenType:   "Bearer",
		ExpiresIn:   7 * 24 * 60 * 60, // 7 days in seconds
		User:        *dto.UserToUserResponse(user),
		IsNewUser:   isNewUser,
	})
}

// ==================== Facebook OAuth ====================

// GetFacebookAuthURL godoc
// @Summary      Get Facebook OAuth URL
// @Description  Get Facebook OAuth authorization URL
// @Tags         OAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  dto.OAuthURLResponse
// @Router       /auth/facebook [get]
func (h *OAuthHandler) GetFacebookAuthURL(c *fiber.Ctx) error {
	state := uuid.New().String()
	authURL := h.oauthService.GetFacebookAuthURL(state)

	return c.JSON(dto.OAuthURLResponse{
		AuthURL: authURL,
	})
}

// HandleFacebookCallback godoc
// @Summary      Handle Facebook OAuth callback
// @Description  Handle Facebook OAuth callback and create/login user
// @Tags         OAuth
// @Accept       json
// @Produce      json
// @Param        code   query     string  true  "Authorization code"
// @Success      200    {object}  dto.OAuthLoginResponse
// @Failure      400    {object}  utils.ErrorResponse
// @Failure      500    {object}  utils.ErrorResponse
// @Router       /auth/facebook/callback [get]
func (h *OAuthHandler) HandleFacebookCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Missing authorization code", nil)
	}

	user, jwtToken, isNewUser, err := h.oauthService.HandleFacebookCallback(c.Context(), code)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to authenticate with Facebook", err)
	}

	return c.JSON(dto.OAuthLoginResponse{
		AccessToken: jwtToken,
		TokenType:   "Bearer",
		ExpiresIn:   7 * 24 * 60 * 60,
		User:        *dto.UserToUserResponse(user),
		IsNewUser:   isNewUser,
	})
}

// ==================== LINE OAuth ====================

// GetLINEAuthURL godoc
// @Summary      Get LINE OAuth URL
// @Description  Get LINE OAuth authorization URL
// @Tags         OAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  dto.OAuthURLResponse
// @Router       /auth/line [get]
func (h *OAuthHandler) GetLINEAuthURL(c *fiber.Ctx) error {
	state := uuid.New().String()
	authURL := h.oauthService.GetLINEAuthURL(state)

	return c.JSON(dto.OAuthURLResponse{
		AuthURL: authURL,
	})
}

// HandleLINECallback godoc
// @Summary      Handle LINE OAuth callback
// @Description  Handle LINE OAuth callback and create/login user
// @Tags         OAuth
// @Accept       json
// @Produce      json
// @Param        code   query     string  true  "Authorization code"
// @Success      200    {object}  dto.OAuthLoginResponse
// @Failure      400    {object}  utils.ErrorResponse
// @Failure      500    {object}  utils.ErrorResponse
// @Router       /auth/line/callback [get]
func (h *OAuthHandler) HandleLINECallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Missing authorization code", nil)
	}

	user, jwtToken, isNewUser, err := h.oauthService.HandleLINECallback(c.Context(), code)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to authenticate with LINE", err)
	}

	return c.JSON(dto.OAuthLoginResponse{
		AccessToken: jwtToken,
		TokenType:   "Bearer",
		ExpiresIn:   7 * 24 * 60 * 60,
		User:        *dto.UserToUserResponse(user),
		IsNewUser:   isNewUser,
	})
}
