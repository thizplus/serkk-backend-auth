package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"gofiber-template/pkg/config"
)

// NATSPublisher implements the EventPublisher interface using NATS JetStream
type NATSPublisher struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	config *config.NATSConfig
}

// NewNATSPublisher creates a new NATS publisher with JetStream enabled
func NewNATSPublisher(cfg *config.NATSConfig) (*NATSPublisher, error) {
	// Connect to NATS
	nc, err := nats.Connect(cfg.URL,
		nats.MaxReconnects(-1),                    // Unlimited reconnects
		nats.ReconnectWait(2*time.Second),         // Wait 2s between reconnects
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				log.Printf("⚠️  NATS disconnected: %v", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("✅ NATS reconnected to %s", nc.ConnectedUrl())
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	log.Printf("✅ NATS connected to %s", cfg.URL)

	var js nats.JetStreamContext
	if cfg.EnableJetStream {
		// Create JetStream context
		js, err = nc.JetStream()
		if err != nil {
			nc.Close()
			return nil, fmt.Errorf("failed to create JetStream context: %w", err)
		}

		// Create or update stream
		if err := createOrUpdateStream(js, cfg); err != nil {
			nc.Close()
			return nil, fmt.Errorf("failed to setup stream: %w", err)
		}

		log.Printf("✅ NATS JetStream enabled (stream: %s)", cfg.StreamName)
	}

	return &NATSPublisher{
		conn:   nc,
		js:     js,
		config: cfg,
	}, nil
}

// createOrUpdateStream creates or updates the JetStream stream
func createOrUpdateStream(js nats.JetStreamContext, cfg *config.NATSConfig) error {
	streamConfig := &nats.StreamConfig{
		Name:        cfg.StreamName,
		Subjects:    []string{cfg.Subject + ".*"}, // Match all subtopics
		Storage:     nats.FileStorage,              // Persistent storage
		Retention:   nats.WorkQueuePolicy,          // Messages deleted after ack
		MaxAge:      7 * 24 * time.Hour,            // Keep messages for 7 days max
		Duplicates:  5 * time.Minute,               // Duplicate detection window
		Replicas:    1,                             // Single replica (change for HA)
		Description: "User events stream for Auth Service",
	}

	// Try to get existing stream
	stream, err := js.StreamInfo(cfg.StreamName)
	if err != nil {
		// Stream doesn't exist, create it
		_, err = js.AddStream(streamConfig)
		if err != nil {
			return fmt.Errorf("failed to create stream: %w", err)
		}
		log.Printf("✅ Created NATS stream: %s", cfg.StreamName)
	} else {
		// Stream exists, update it
		_, err = js.UpdateStream(streamConfig)
		if err != nil {
			return fmt.Errorf("failed to update stream: %w", err)
		}
		log.Printf("✅ Updated NATS stream: %s (msgs: %d)", cfg.StreamName, stream.State.Msgs)
	}

	return nil
}

// Publish sends an event synchronously to NATS
func (n *NATSPublisher) Publish(ctx context.Context, topic string, payload interface{}) error {
	// Marshal payload to JSON
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Use full subject path
	subject := fmt.Sprintf("%s.%s", n.config.Subject, topic)

	// Publish with JetStream if enabled
	if n.js != nil {
		// Publish with acknowledgment
		pubAck, err := n.js.Publish(subject, data)
		if err != nil {
			log.Printf("❌ Failed to publish to NATS JetStream: %v", err)
			return fmt.Errorf("failed to publish: %w", err)
		}
		log.Printf("✅ Published to %s (stream: %s, seq: %d)", subject, pubAck.Stream, pubAck.Sequence)
	} else {
		// Publish without JetStream (core NATS)
		err := n.conn.Publish(subject, data)
		if err != nil {
			log.Printf("❌ Failed to publish to NATS: %v", err)
			return fmt.Errorf("failed to publish: %w", err)
		}
		log.Printf("✅ Published to %s", subject)
	}

	return nil
}

// PublishAsync sends an event asynchronously (fire-and-forget)
func (n *NATSPublisher) PublishAsync(topic string, payload interface{}) {
	go func() {
		ctx := context.Background()
		if err := n.Publish(ctx, topic, payload); err != nil {
			log.Printf("❌ Async publish failed: %v", err)
		}
	}()
}

// Close gracefully shuts down the NATS connection
func (n *NATSPublisher) Close() error {
	if n.conn != nil {
		// Drain connection (flush pending messages)
		if err := n.conn.Drain(); err != nil {
			log.Printf("⚠️  Failed to drain NATS connection: %v", err)
		}

		// Close connection
		n.conn.Close()
		log.Println("✅ NATS connection closed")
	}
	return nil
}

// IsConnected checks if NATS connection is active
func (n *NATSPublisher) IsConnected() bool {
	return n.conn != nil && n.conn.IsConnected()
}

// Stats returns NATS connection statistics
func (n *NATSPublisher) Stats() nats.Statistics {
	if n.conn != nil {
		return n.conn.Stats()
	}
	return nats.Statistics{}
}
