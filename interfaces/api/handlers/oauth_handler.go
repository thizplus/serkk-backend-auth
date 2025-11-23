package handlers

import (
	"gofiber-template/domain/dto"
	"gofiber-template/domain/services"
	"gofiber-template/pkg/auth_code_store"
	"gofiber-template/pkg/config"
	"gofiber-template/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type OAuthHandler struct {
	oauthService services.OAuthService
	config       *config.Config
}

func NewOAuthHandler(oauthService services.OAuthService, cfg *config.Config) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
		config:       cfg,
	}
}

// ==================== Google OAuth ====================

// GetGoogleAuthURL godoc
// @Summary      Get Google OAuth URL
// @Description  Get Google OAuth authorization URL with CSRF protection
// @Tags         OAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  dto.OAuthURLResponse
// @Router       /auth/google [get]
func (h *OAuthHandler) GetGoogleAuthURL(c *fiber.Ctx) error {
	// Generate state for CSRF protection
	state := uuid.New().String()

	// Store state in HTTPOnly cookie
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HTTPOnly: true,
		Secure:   h.config.App.Env == "production",
		SameSite: "Lax",
		Path:     "/",
		MaxAge:   300, // 5 minutes
	})

	authURL := h.oauthService.GetGoogleAuthURL(state)

	return utils.SuccessResponse(c, "Google OAuth URL generated", map[string]string{
		"url": authURL,
	})
}

// HandleGoogleCallback godoc
// @Summary      Handle Google OAuth callback
// @Description  Handle Google OAuth callback and generate authorization code
// @Tags         OAuth
// @Accept       json
// @Produce      json
// @Param        code   query     string  true  "Authorization code from Google"
// @Param        state  query     string  true  "State parameter for CSRF protection"
// @Success      302    {string}  string  "Redirect to frontend with authorization code"
// @Failure      400    {string}  string  "Redirect to frontend with error"
// @Router       /auth/google/callback [get]
func (h *OAuthHandler) HandleGoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")

	// Validate code parameter
	if code == "" {
		return c.Redirect(h.config.App.FrontendURL + "/auth/callback?error=missing_code")
	}

	// Validate state parameter (CSRF protection)
	// Note: In development, we skip state validation because cookies might not be present
	// when Google redirects back. Google OAuth itself validates the state parameter.
	storedState := c.Cookies("oauth_state")
	if storedState != "" {
		// Only validate if cookie exists
		if storedState != state {
			return c.Redirect(h.config.App.FrontendURL + "/auth/callback?error=invalid_state")
		}
		// Clear state cookie
		c.ClearCookie("oauth_state")
	}

	// Handle OAuth callback
	user, jwtToken, isNewUser, err := h.oauthService.HandleGoogleCallback(c.Context(), code)
	if err != nil {
		return c.Redirect(h.config.App.FrontendURL + "/auth/callback?error=oauth_failed")
	}

	// Generate temporary authorization code
	store := auth_code_store.GetInstance()
	authCode, err := store.GenerateCode(jwtToken, *dto.UserToUserResponse(user), isNewUser, state)
	if err != nil {
		return c.Redirect(h.config.App.FrontendURL + "/auth/callback?error=code_generation_failed")
	}

	// Redirect to frontend with authorization code
	return c.Redirect(h.config.App.FrontendURL + "/auth/callback?code=" + authCode + "&state=" + state)
}

// ExchangeCodeForToken godoc
// @Summary      Exchange authorization code for token
// @Description  Exchange temporary authorization code for JWT token
// @Tags         OAuth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ExchangeCodeRequest  true  "Exchange code request"
// @Success      200      {object}  dto.ExchangeCodeResponse
// @Failure      400      {object}  utils.ErrorResponse
// @Router       /auth/exchange [post]
func (h *OAuthHandler) ExchangeCodeForToken(c *fiber.Ctx) error {
	var req dto.ExchangeCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Validation failed", err)
	}

	// Exchange code for token
	store := auth_code_store.GetInstance()
	data, ok := store.ExchangeCode(req.Code, req.State)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid or expired authorization code", nil)
	}

	// Return token and user info
	return utils.SuccessResponse(c, "Authentication successful", dto.ExchangeCodeResponse{
		Token:        data.Token,
		User:         data.User,
		IsNewUser:    data.IsNewUser,
		NeedsProfile: false, // Google provides all necessary info
	})
}

// ==================== Facebook OAuth ====================

// GetFacebookAuthURL godoc
// @Summary      Get Facebook OAuth URL
// @Description  Get Facebook OAuth authorization URL with CSRF protection
// @Tags         OAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  dto.OAuthURLResponse
// @Router       /auth/facebook [get]
func (h *OAuthHandler) GetFacebookAuthURL(c *fiber.Ctx) error {
	state := uuid.New().String()

	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HTTPOnly: true,
		Secure:   h.config.App.Env == "production",
		SameSite: "Lax",
		Path:     "/",
		MaxAge:   300,
	})

	authURL := h.oauthService.GetFacebookAuthURL(state)

	return utils.SuccessResponse(c, "Facebook OAuth URL generated", map[string]string{
		"url": authURL,
	})
}

// HandleFacebookCallback godoc
// @Summary      Handle Facebook OAuth callback
// @Description  Handle Facebook OAuth callback and generate authorization code
// @Tags         OAuth
// @Accept       json
// @Produce      json
// @Param        code   query     string  true  "Authorization code from Facebook"
// @Param        state  query     string  true  "State parameter for CSRF protection"
// @Success      302    {string}  string  "Redirect to frontend with authorization code"
// @Failure      400    {string}  string  "Redirect to frontend with error"
// @Router       /auth/facebook/callback [get]
func (h *OAuthHandler) HandleFacebookCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		return c.Redirect(h.config.App.FrontendURL + "/auth/callback?error=missing_code")
	}

	// Validate state parameter (CSRF protection)
	// Skip validation if cookie not present (development mode)
	storedState := c.Cookies("oauth_state")
	if storedState != "" {
		if storedState != state {
			return c.Redirect(h.config.App.FrontendURL + "/auth/callback?error=invalid_state")
		}
		c.ClearCookie("oauth_state")
	}

	user, jwtToken, isNewUser, err := h.oauthService.HandleFacebookCallback(c.Context(), code)
	if err != nil {
		return c.Redirect(h.config.App.FrontendURL + "/auth/callback?error=oauth_failed")
	}

	store := auth_code_store.GetInstance()
	authCode, err := store.GenerateCode(jwtToken, *dto.UserToUserResponse(user), isNewUser, state)
	if err != nil {
		return c.Redirect(h.config.App.FrontendURL + "/auth/callback?error=code_generation_failed")
	}

	return c.Redirect(h.config.App.FrontendURL + "/auth/callback?code=" + authCode + "&state=" + state)
}

// ==================== LINE OAuth ====================

// GetLINEAuthURL godoc
// @Summary      Get LINE OAuth URL
// @Description  Get LINE OAuth authorization URL with CSRF protection
// @Tags         OAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  dto.OAuthURLResponse
// @Router       /auth/line [get]
func (h *OAuthHandler) GetLINEAuthURL(c *fiber.Ctx) error {
	state := uuid.New().String()

	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HTTPOnly: true,
		Secure:   h.config.App.Env == "production",
		SameSite: "Lax",
		Path:     "/",
		MaxAge:   300,
	})

	authURL := h.oauthService.GetLINEAuthURL(state)

	return utils.SuccessResponse(c, "LINE OAuth URL generated", map[string]string{
		"url": authURL,
	})
}

// HandleLINECallback godoc
// @Summary      Handle LINE OAuth callback
// @Description  Handle LINE OAuth callback and generate authorization code
// @Tags         OAuth
// @Accept       json
// @Produce      json
// @Param        code   query     string  true  "Authorization code from LINE"
// @Param        state  query     string  true  "State parameter for CSRF protection"
// @Success      302    {string}  string  "Redirect to frontend with authorization code"
// @Failure      400    {string}  string  "Redirect to frontend with error"
// @Router       /auth/line/callback [get]
func (h *OAuthHandler) HandleLINECallback(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		return c.Redirect(h.config.App.FrontendURL + "/auth/callback?error=missing_code")
	}

	// Validate state parameter (CSRF protection)
	// Skip validation if cookie not present (development mode)
	storedState := c.Cookies("oauth_state")
	if storedState != "" {
		if storedState != state {
			return c.Redirect(h.config.App.FrontendURL + "/auth/callback?error=invalid_state")
		}
		c.ClearCookie("oauth_state")
	}

	user, jwtToken, isNewUser, err := h.oauthService.HandleLINECallback(c.Context(), code)
	if err != nil {
		return c.Redirect(h.config.App.FrontendURL + "/auth/callback?error=oauth_failed")
	}

	store := auth_code_store.GetInstance()
	authCode, err := store.GenerateCode(jwtToken, *dto.UserToUserResponse(user), isNewUser, state)
	if err != nil {
		return c.Redirect(h.config.App.FrontendURL + "/auth/callback?error=code_generation_failed")
	}

	return c.Redirect(h.config.App.FrontendURL + "/auth/callback?code=" + authCode + "&state=" + state)
}
