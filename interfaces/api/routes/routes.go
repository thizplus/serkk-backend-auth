package routes

import (
	"github.com/gofiber/fiber/v2"
	"gofiber-template/interfaces/api/handlers"
)

func SetupRoutes(app *fiber.App, h *handlers.Handlers) {
	// Setup health and root routes
	SetupHealthRoutes(app)

	// API version group
	api := app.Group("/api/v1")

	// Setup all route groups
	SetupAuthRoutes(api, h)
	SetupUserRoutes(api, h)
}