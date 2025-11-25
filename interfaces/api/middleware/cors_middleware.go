package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func CorsMiddleware() fiber.Handler {
	// Get frontend URL from environment
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	// Build allowed origins (support multiple origins)
	allowedOrigins := []string{
		frontendURL,
		"http://localhost:3000", // Development
		"http://localhost:3030", // Alternative dev port
	}

	// In production, only allow configured frontend URL
	env := os.Getenv("APP_ENV")
	if env == "production" {
		allowedOrigins = []string{frontendURL}
	}

	return cors.New(cors.Config{
		AllowOrigins:     strings.Join(allowedOrigins, ","),
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
		ExposeHeaders:    "Content-Length,Authorization",
	})
}