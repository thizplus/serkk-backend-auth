package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Event Publishing Metrics
	EventsPublishedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "events_published_total",
			Help: "Total number of events published",
		},
		[]string{"topic", "status"}, // status: success, failure
	)

	EventPublishDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "event_publish_duration_seconds",
			Help:    "Event publish duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"topic"},
	)

	// User Operations Metrics
	UserRegistrationsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "user_registrations_total",
			Help: "Total number of user registrations",
		},
	)

	UserLoginsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_logins_total",
			Help: "Total number of user logins",
		},
		[]string{"status"}, // status: success, failure
	)

	OAuthLoginsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "oauth_logins_total",
			Help: "Total number of OAuth logins",
		},
		[]string{"provider", "status"}, // provider: google, facebook, line
	)

	// NATS Connection Status
	NATSConnectionStatus = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "nats_connection_status",
			Help: "NATS connection status (1 = connected, 0 = disconnected)",
		},
	)

	// Active Users Gauge
	ActiveUsersTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users_total",
			Help: "Total number of active users in the system",
		},
	)
)
