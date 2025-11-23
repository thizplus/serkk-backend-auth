package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler exposes Prometheus metrics
type MetricsHandler struct{}

func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{}
}

// GetMetrics returns Prometheus metrics in text format
func (h *MetricsHandler) GetMetrics(c *fiber.Ctx) error {
	// Adapt Prometheus HTTP handler to Fiber
	handler := promhttp.Handler()
	return adaptor.HTTPHandler(handler)(c)
}
