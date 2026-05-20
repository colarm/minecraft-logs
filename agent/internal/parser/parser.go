package parser

import "time"

type Event struct {
	ServerID   string                 `json:"server_id"`
	EventType  string                 `json:"event_type"`
	PlayerName string                 `json:"player_name,omitempty"`
	Message    string                 `json:"message,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}
