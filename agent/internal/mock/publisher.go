package mock

import (
	"encoding/json"
	"log"

	"minecraft-log-agent/internal/parser"
)

// ConsolePublisher prints parsed events to stdout as JSON instead of
// publishing to NATS. Implements the same subset of methods root.go
// calls on the real publisher.
type ConsolePublisher struct {
	serverID string
}

// NewConsolePublisher creates a ConsolePublisher for the given serverID.
func NewConsolePublisher(serverID string) *ConsolePublisher {
	return &ConsolePublisher{serverID: serverID}
}

// Connect always succeeds for the mock.
func (p *ConsolePublisher) Connect() error {
	log.Printf("mock: publisher ready (server=%s)", p.serverID)
	return nil
}

// Publish JSON-encodes the event and writes it to stdout.
func (p *ConsolePublisher) Publish(event *parser.Event) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("mock: marshal error: %v", err)
		return
	}
	log.Printf("mock: event -> %s", string(data))
}

// Close is a no-op for the mock.
func (p *ConsolePublisher) Close() {}
