package serviceimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gofiber-template/domain/models"
	"gofiber-template/domain/services"
	"gofiber-template/pkg/contextutil"
	"gofiber-template/pkg/logger"
)

// SyncService handles user synchronization to backend
// Supports both Event-Driven (NATS) and HTTP sync with fallback
type SyncService struct {
	eventPublisher services.EventPublisher
	backendURL     string
	httpClient     *http.Client
	useEvents      bool // Feature flag
}

// NewSyncService creates a new SyncService with optional EventPublisher
func NewSyncServiceWithPublisher(eventPublisher services.EventPublisher) *SyncService {
	useEvents := os.Getenv("USE_EVENT_SYNC") != "false" // Default: true if publisher available

	return &SyncService{
		eventPublisher: eventPublisher,
		backendURL:     os.Getenv("BACKEND_SYNC_URL"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		useEvents: useEvents && eventPublisher != nil,
	}
}

// UserSyncPayload represents minimal identity event payload
// Auth Service sends only essential identity information.
// Downstream services are responsible for enriching user profiles.
type UserSyncPayload struct {
	// Minimal Identity Data (Primary Key + Identifiers)
	ID       string `json:"id"`       // User ID (Primary Key)
	Email    string `json:"email"`    // Email address (identifier)
	Username string `json:"username"` // Username (identifier)
	Action   string `json:"action"`   // "created" | "updated" | "deleted"

	// Observability Metadata
	RequestID   string `json:"request_id"`   // Correlation ID for distributed tracing
	Timestamp   string `json:"timestamp"`    // ISO 8601 timestamp
	ServiceName string `json:"service_name"` // Source service ("gofiber-auth")
}

// Note: Removed fields (managed by downstream services):
// - displayName, avatar, bio → Social/Profile Service
// - role, isActive, permissions → Auth Service internal only

// SyncUser synchronizes user data using Events or HTTP (with fallback)
func (s *SyncService) SyncUser(ctx context.Context, user *models.User, action string) error {
	requestID := contextutil.GetRequestID(ctx)
	log := logger.GetLogger()

	payload := UserSyncPayload{
		// Minimal Identity Data
		ID:       user.ID.String(),
		Email:    user.Email,
		Username: user.Username,
		Action:   action,

		// Observability Metadata
		RequestID:   requestID,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		ServiceName: "gofiber-auth",
	}

	log.Debug("Synchronizing user", map[string]interface{}{
		"request_id": requestID,
		"user_id":    user.ID.String(),
		"action":     action,
		"method":     getMethod(s.useEvents),
	})

	// Strategy 1: Try Event Publishing (if enabled)
	if s.useEvents {
		err := s.syncViaEvent(&payload)
		if err != nil {
			log.Warn("Event sync failed, falling back to HTTP", map[string]interface{}{
				"request_id": requestID,
				"user_id":    user.ID.String(),
				"action":     action,
				"error":      err.Error(),
			})
			// Fallback to HTTP
			return s.syncViaHTTP(&payload)
		}
		return nil
	}

	// Strategy 2: HTTP Sync (legacy)
	return s.syncViaHTTP(&payload)
}

// syncViaEvent publishes user event to NATS
func (s *SyncService) syncViaEvent(payload *UserSyncPayload) error {
	if s.eventPublisher == nil {
		return fmt.Errorf("event publisher not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Publish to NATS JetStream
	// Topic format: user.events.{action} → user.events.created, user.events.updated, etc.
	topic := payload.Action
	err := s.eventPublisher.Publish(ctx, topic, payload)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("✅ User %s synced via Events (action: %s)", payload.Username, payload.Action)
	return nil
}

// syncViaHTTP syncs user via HTTP POST (legacy method)
func (s *SyncService) syncViaHTTP(payload *UserSyncPayload) error {
	if s.backendURL == "" {
		log.Println("⚠️  BACKEND_SYNC_URL not configured, skipping HTTP sync")
		return nil
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", s.backendURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("❌ Failed to sync user %s to backend via HTTP: %v", payload.Username, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("❌ Backend HTTP sync failed for user %s: HTTP %d", payload.Username, resp.StatusCode)
		return fmt.Errorf("backend returned status %d", resp.StatusCode)
	}

	log.Printf("✅ User %s synced via HTTP (action: %s)", payload.Username, payload.Action)
	return nil
}

// SyncUserWithRetry tries to sync with exponential backoff
func (s *SyncService) SyncUserWithRetry(ctx context.Context, user *models.User, action string) {
	maxRetries := 3
	baseDelay := 1 * time.Second
	requestID := contextutil.GetRequestID(ctx)
	log := logger.GetLogger()

	for i := 0; i < maxRetries; i++ {
		err := s.SyncUser(ctx, user, action)
		if err == nil {
			return // Success
		}

		if i < maxRetries-1 {
			delay := baseDelay * time.Duration(1<<uint(i))
			log.Warn("Retrying sync", map[string]interface{}{
				"request_id": requestID,
				"user_id":    user.ID.String(),
				"action":     action,
				"attempt":    i + 2,
				"max":        maxRetries,
				"delay_ms":   delay.Milliseconds(),
			})
			time.Sleep(delay)
		}
	}

	log.Error("Failed to sync user after all retries", map[string]interface{}{
		"request_id": requestID,
		"user_id":    user.ID.String(),
		"username":   user.Username,
		"action":     action,
		"attempts":   maxRetries,
	})
	// TODO: Send to dead letter queue or alert monitoring
}

// getMethod returns sync method name for logging
func getMethod(useEvents bool) string {
	if useEvents {
		return "events"
	}
	return "http"
}
