package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
)

func main() {
	// Connect to NATS
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal("Failed to connect to NATS:", err)
	}
	defer nc.Close()

	log.Println("✅ Connected to NATS")

	// Create JetStream context
	js, err := nc.JetStream()
	if err != nil {
		log.Fatal("Failed to create JetStream context:", err)
	}

	log.Println("✅ JetStream context created")

	// Get stream info
	stream, err := js.StreamInfo("USER_EVENTS")
	if err != nil {
		log.Fatal("Failed to get stream info:", err)
	}

	log.Printf("📊 Stream: %s", stream.Config.Name)
	log.Printf("📊 Messages: %d", stream.State.Msgs)
	log.Printf("📊 Bytes: %d", stream.State.Bytes)
	log.Printf("📊 First Seq: %d", stream.State.FirstSeq)
	log.Printf("📊 Last Seq: %d", stream.State.LastSeq)

	// Subscribe to all user events
	sub, err := js.Subscribe("user.events.*", func(msg *nats.Msg) {
		log.Println("\n🔔 Received event:")
		log.Printf("Subject: %s", msg.Subject)
		log.Printf("Size: %d bytes", len(msg.Data))

		// Pretty print JSON
		var payload map[string]interface{}
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			log.Printf("Payload (raw): %s", string(msg.Data))
		} else {
			pretty, _ := json.MarshalIndent(payload, "", "  ")
			log.Printf("Payload:\n%s", string(pretty))
		}

		// Acknowledge the message
		msg.Ack()
		log.Println("✅ Message acknowledged")
	}, nats.Durable("test-subscriber"), nats.ManualAck())

	if err != nil {
		log.Fatal("Failed to subscribe:", err)
	}

	log.Println("👂 Listening for events on user.events.*")
	log.Println("Press Ctrl+C to exit")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("\n👋 Unsubscribing...")
	sub.Unsubscribe()
	log.Println("✅ Done")
}
