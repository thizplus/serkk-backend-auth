package routes

import (
	"github.com/gofiber/fiber/v2"
	"gofiber-template/interfaces/api/handlers"
)

func SetupAuthRoutes(api fiber.Router, h *handlers.Handlers) {
	auth := api.Group("/auth")

	// Standard Auth
	auth.Post("/register", h.UserHandler.Register)
	auth.Post("/login", h.UserHandler.Login)

	// OAuth Code Exchange
	auth.Post("/exchange", h.OAuthHandler.ExchangeCodeForToken)

	// Google OAuth
	auth.Get("/google", h.OAuthHandler.GetGoogleAuthURL)
	auth.Get("/google/callback", h.OAuthHandler.HandleGoogleCallback)

	// Facebook OAuth
	auth.Get("/facebook", h.OAuthHandler.GetFacebookAuthURL)
	auth.Get("/facebook/callback", h.OAuthHandler.HandleFacebookCallback)

	// LINE OAuth
	auth.Get("/line", h.OAuthHandler.GetLINEAuthURL)
	auth.Get("/line/callback", h.OAuthHandler.HandleLINECallback)
}