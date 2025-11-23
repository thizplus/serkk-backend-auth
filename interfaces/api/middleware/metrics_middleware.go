package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gofiber-template/pkg/metrics"
)

// MetricsMiddleware collects Prometheus metrics for HTTP requests
func MetricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Collect metrics
		duration := time.Since(start).Seconds()
		method := c.Method()
		path := c.Route().Path
		if path == "" {
			path = c.Path() // Fallback for routes without patterns
		}
		status := strconv.Itoa(c.Response().StatusCode())

		// Record request count
		metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()

		// Record request duration
		metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)

		return err
	}
}
