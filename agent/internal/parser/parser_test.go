package parser

import (
	"testing"
	"time"
)

func TestParsePlayerJoin(t *testing.T) {
	e := ParseLine("test-server", "[14:30:00] [Server thread/INFO]: Steve joined the game")
	if e == nil {
		t.Fatal("expected event, got nil")
	}
	if e.EventType != "player_join" {
		t.Errorf("expected player_join, got %s", e.EventType)
	}
	if e.PlayerName != "Steve" {
		t.Errorf("expected Steve, got %s", e.PlayerName)
	}
	if e.ServerID != "test-server" {
		t.Errorf("expected test-server, got %s", e.ServerID)
	}
}

func TestParsePlayerLeave(t *testing.T) {
	e := ParseLine("test-server", "[14:35:00] [Server thread/INFO]: Steve left the game")
	if e == nil {
		t.Fatal("expected event, got nil")
	}
	if e.EventType != "player_leave" {
		t.Errorf("expected player_leave, got %s", e.EventType)
	}
	if e.PlayerName != "Steve" {
		t.Errorf("expected Steve, got %s", e.PlayerName)
	}

	e2 := ParseLine("test-server", "[14:36:00] [Server thread/INFO]: Alex lost connection: Internal Exception")
	if e2 == nil || e2.EventType != "player_leave" {
		t.Errorf("expected player_leave for lost connection")
	}
}

func TestParseChat(t *testing.T) {
	e := ParseLine("test-server", "[14:31:00] [Server thread/INFO]: <Steve> hello everyone")
	if e == nil {
		t.Fatal("expected event, got nil")
	}
	if e.EventType != "chat" {
		t.Errorf("expected chat, got %s", e.EventType)
	}
	if e.PlayerName != "Steve" {
		t.Errorf("expected Steve, got %s", e.PlayerName)
	}
	if e.Message != "hello everyone" {
		t.Errorf("expected 'hello everyone', got %s", e.Message)
	}
}

func TestParseDeath(t *testing.T) {
	e := ParseLine("test-server", "[14:32:00] [Server thread/INFO]: Steve was slain by Zombie")
	if e == nil {
		t.Fatal("expected event, got nil")
	}
	if e.EventType != "death" {
		t.Errorf("expected death, got %s", e.EventType)
	}
	if e.PlayerName != "Steve" {
		t.Errorf("expected Steve, got %s", e.PlayerName)
	}
}

func TestParseServerStart(t *testing.T) {
	e := ParseLine("test-server", "[14:00:00] [Server thread/INFO]: Done (3.215s)! For help, type \"help\"")
	if e == nil {
		t.Fatal("expected event, got nil")
	}
	if e.EventType != "server_start" {
		t.Errorf("expected server_start, got %s", e.EventType)
	}
}

func TestParseServerStop(t *testing.T) {
	e := ParseLine("test-server", "[15:00:00] [Server thread/INFO]: Stopping server")
	if e == nil {
		t.Fatal("expected event, got nil")
	}
	if e.EventType != "server_stop" {
		t.Errorf("expected server_stop, got %s", e.EventType)
	}
}

func TestParseTps(t *testing.T) {
	e := ParseLine("test-server", "[14:33:00] [Server thread/INFO]: MSPT: 45.2 TPS: 19.85")
	if e == nil {
		t.Fatal("expected event, got nil")
	}
	if e.EventType != "tps" {
		t.Errorf("expected tps, got %s", e.EventType)
	}
	if e.Metadata["tps"] != 19.85 {
		t.Errorf("expected tps 19.85, got %v", e.Metadata["tps"])
	}
}

func TestParseError(t *testing.T) {
	e := ParseLine("test-server", "[14:34:00] [Server thread/ERROR]: Encountered an unexpected exception")
	if e == nil {
		t.Fatal("expected event, got nil")
	}
	if e.EventType != "error" {
		t.Errorf("expected error, got %s", e.EventType)
	}
	if e.Metadata["level"] != "ERROR" {
		t.Errorf("expected ERROR level, got %v", e.Metadata["level"])
	}
}

func TestParseUnknown(t *testing.T) {
	e := ParseLine("test-server", "[14:35:00] [Server thread/INFO]: Some random message")
	if e != nil {
		t.Errorf("expected nil for unknown message, got %+v", e)
	}
}

func TestParseTime(t *testing.T) {
	now := time.Now()
	result := parseTime("14:30:00")
	if result.Hour() != 14 || result.Minute() != 30 || result.Second() != 0 {
		t.Errorf("expected 14:30:00, got %02d:%02d:%02d", result.Hour(), result.Minute(), result.Second())
	}
	if result.Year() != now.Year() || result.Month() != now.Month() || result.Day() != now.Day() {
		t.Errorf("expected today's date, got %s", result.Format("2006-01-02"))
	}
}
