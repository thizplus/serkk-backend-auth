package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gofiber-template/pkg/contextutil"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if request ID already exists (from client)
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			// Generate new request ID
			requestID = uuid.New().String()
		}

		// Store in Fiber context
		c.Locals("requestID", requestID)

		// Store in Go context (for service layer access)
		ctx := contextutil.WithRequestID(c.Context(), requestID)
		c.SetUserContext(ctx)

		// Add to response headers
		c.Set("X-Request-ID", requestID)

		return c.Next()
	}
}

// GetRequestID retrieves request ID from context
func GetRequestID(c *fiber.Ctx) string {
	if requestID, ok := c.Locals("requestID").(string); ok {
		return requestID
	}
	return ""
}
