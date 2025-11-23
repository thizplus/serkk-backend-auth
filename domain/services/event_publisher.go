package services

import "context"

// EventPublisher defines the interface for publishing events to message brokers
// This abstraction allows switching between different event bus implementations
// (NATS, Kafka, RabbitMQ, etc.) without changing business logic
type EventPublisher interface {
	// Publish sends an event synchronously to the specified topic
	// Returns error if publishing fails
	Publish(ctx context.Context, topic string, payload interface{}) error

	// PublishAsync sends an event asynchronously (fire-and-forget)
	// Errors are logged but not returned
	PublishAsync(topic string, payload interface{})

	// Close gracefully shuts down the publisher and closes connections
	Close() error
}

// Event represents a generic event structure
type Event struct {
	ID        string                 `json:"id"`        // Unique event ID
	Type      string                 `json:"type"`      // Event type (e.g., "user.created")
	Timestamp int64                  `json:"timestamp"` // Unix timestamp
	Data      map[string]interface{} `json:"data"`      // Event payload
}

// UserEventData represents minimal identity event data
// Auth Service sends only essential identity information.
// Downstream services (Social, Profile) are responsible for enriching user data.
type UserEventData struct {
	// Minimal Identity Data (Primary Key + Identifiers)
	ID       string `json:"id"`       // User ID (Primary Key)
	Email    string `json:"email"`    // Email address (identifier)
	Username string `json:"username"` // Username (identifier)
	Action   string `json:"action"`   // "created" | "updated" | "deleted"

	// Observability Metadata
	RequestID   string `json:"request_id"`           // Correlation ID for distributed tracing
	Timestamp   string `json:"timestamp"`            // ISO 8601 timestamp
	Sequence    uint64 `json:"sequence,omitempty"`   // NATS JetStream sequence number
	ServiceName string `json:"service_name"`         // Source service (e.g., "gofiber-auth")
}

// Note: Fields removed from events (managed by downstream services):
// - displayName, avatar, bio → Managed by Social/Profile Service
// - role, isActive, permissions → Internal to Auth Service, not needed by downstream
