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

func TestParseWhitelistReject(t *testing.T) {
	e := ParseLine("test-server", "[14:30:00] [Server thread/INFO]: Disconnecting Steve: You are not white-listed on this server!")
	if e == nil {
		t.Fatal("expected event, got nil")
	}
	if e.EventType != "whitelist_reject" {
		t.Errorf("expected whitelist_reject, got %s", e.EventType)
	}
	if e.PlayerName != "Steve" {
		t.Errorf("expected Steve, got %s", e.PlayerName)
	}
}

func TestParseLoginFail(t *testing.T) {
	e := ParseLine("test-server", "[14:30:00] [User Authenticator #1/INFO]: Disconnecting Steve: Failed to verify username!")
	if e == nil {
		t.Fatal("expected event, got nil")
	}
	if e.EventType != "login_fail" {
		t.Errorf("expected login_fail, got %s", e.EventType)
	}
	if e.PlayerName != "Steve" {
		t.Errorf("expected Steve, got %s", e.PlayerName)
	}
}

func TestParseConnect(t *testing.T) {
	e := ParseLine("test-server", "[14:30:00] [Server thread/INFO]: [/1.2.3.4:54321] logged in with entity id 1234 at ([world]-47.5, 64.0, 212.5)")
	if e == nil {
		t.Fatal("expected event, got nil")
	}
	if e.EventType != "connect" {
		t.Errorf("expected connect, got %s", e.EventType)
	}
	ip, ok := e.Metadata["ip"]
	if !ok {
		t.Fatal("expected ip in metadata")
	}
	if ip != "1.2.3.4" {
		t.Errorf("expected ip 1.2.3.4, got %v", ip)
	}
}

func TestParseBan(t *testing.T) {
	e := ParseLine("test-server", "[14:30:00] [Server thread/INFO]: Disconnecting Alex: You are banned from this server. Reason: Griefing")
	if e == nil {
		t.Fatal("expected event, got nil")
	}
	if e.EventType != "ban" {
		t.Errorf("expected ban, got %s", e.EventType)
	}
	if e.PlayerName != "Alex" {
		t.Errorf("expected Alex, got %s", e.PlayerName)
	}
}

func TestParseCommand(t *testing.T) {
	e := ParseLine("test-server", "[14:30:00] [Server thread/INFO]: Steve issued server command: /gamemode creative")
	if e == nil {
		t.Fatal("expected event, got nil")
	}
	if e.EventType != "command" {
		t.Errorf("expected command, got %s", e.EventType)
	}
	if e.PlayerName != "Steve" {
		t.Errorf("expected Steve, got %s", e.PlayerName)
	}
	if e.Message != "/gamemode creative" {
		t.Errorf("expected '/gamemode creative', got %s", e.Message)
	}
}

func TestParseRcon(t *testing.T) {
	e := ParseLine("test-server", "[14:30:00] [RCON Listener #1/INFO]: /op Steve")
	if e == nil {
		t.Fatal("expected event, got nil")
	}
	if e.EventType != "rcon" {
		t.Errorf("expected rcon, got %s", e.EventType)
	}
	if e.Message != "op Steve" {
		t.Errorf("expected 'op Steve', got %s", e.Message)
	}

	e2 := ParseLine("test-server", "[14:31:00] [Rcon/INFO]: /say hello issued by 127.0.0.1:25575")
	if e2 == nil {
		t.Fatal("expected event for RCON with source, got nil")
	}
	if e2.EventType != "rcon" {
		t.Errorf("expected rcon, got %s", e2.EventType)
	}
	src, ok := e2.Metadata["source"]
	if !ok {
		t.Fatal("expected source in metadata")
	}
	if src != "127.0.0.1:25575" {
		t.Errorf("expected '127.0.0.1:25575', got %v", src)
	}
}

func TestParseOp(t *testing.T) {
	e := ParseLine("test-server", "[14:30:00] [Server thread/INFO]: Made Steve a server operator")
	if e == nil {
		t.Fatal("expected event, got nil")
	}
	if e.EventType != "op" {
		t.Errorf("expected op, got %s", e.EventType)
	}
	if e.PlayerName != "Steve" {
		t.Errorf("expected Steve, got %s", e.PlayerName)
	}
	if e.Metadata["action"] != "grant" {
		t.Errorf("expected grant, got %v", e.Metadata["action"])
	}

	e2 := ParseLine("test-server", "[14:31:00] [Server thread/INFO]: Steve is no longer a server operator")
	if e2 == nil {
		t.Fatal("expected event for deop, got nil")
	}
	if e2.Metadata["action"] != "revoke" {
		t.Errorf("expected revoke, got %v", e2.Metadata["action"])
	}
}

func TestParseAdvancement(t *testing.T) {
	e := ParseLine("test-server", "[14:30:00] [Server thread/INFO]: Steve has made the advancement [Monsters Hunted]")
	if e == nil {
		t.Fatal("expected event, got nil")
	}
	if e.EventType != "advancement" {
		t.Errorf("expected advancement, got %s", e.EventType)
	}
	if e.PlayerName != "Steve" {
		t.Errorf("expected Steve, got %s", e.PlayerName)
	}
	if e.Message != "Monsters Hunted" {
		t.Errorf("expected 'Monsters Hunted', got %s", e.Message)
	}
	if e.Metadata["type"] != "made the advancement" {
		t.Errorf("expected 'made the advancement', got %v", e.Metadata["type"])
	}
}

func TestParseGameMode(t *testing.T) {
	e := ParseLine("test-server", "[14:30:00] [Server thread/INFO]: Steve changed game mode to Creative")
	if e == nil {
		t.Fatal("expected event, got nil")
	}
	if e.EventType != "game_mode" {
		t.Errorf("expected game_mode, got %s", e.EventType)
	}
	if e.PlayerName != "Steve" {
		t.Errorf("expected Steve, got %s", e.PlayerName)
	}
	if e.Metadata["mode"] != "Creative" {
		t.Errorf("expected Creative, got %v", e.Metadata["mode"])
	}

	e2 := ParseLine("test-server", "[14:31:00] [Server thread/INFO]: Alex's game mode has been changed to Spectator")
	if e2 == nil {
		t.Fatal("expected event for passive format, got nil")
	}
	if e2.PlayerName != "Alex" || e2.Metadata["mode"] != "Spectator" {
		t.Errorf("expected Alex/Spectator, got %s/%v", e2.PlayerName, e2.Metadata["mode"])
	}

	e3 := ParseLine("test-server", "[14:32:00] [Server thread/INFO]: Set Steve's game mode to Adventure")
	if e3 == nil {
		t.Fatal("expected event for set format, got nil")
	}
	if e3.PlayerName != "Steve" || e3.Metadata["mode"] != "Adventure" {
		t.Errorf("expected Steve/Adventure, got %s/%v", e3.PlayerName, e3.Metadata["mode"])
	}
}

func TestParseGameRule(t *testing.T) {
	e := ParseLine("test-server", "[14:30:00] [Server thread/INFO]: Steve changed the game rule: doDaylightCycle = false")
	if e == nil {
		t.Fatal("expected event, got nil")
	}
	if e.EventType != "game_rule" {
		t.Errorf("expected game_rule, got %s", e.EventType)
	}
	if e.PlayerName != "Steve" {
		t.Errorf("expected Steve, got %s", e.PlayerName)
	}
	if e.Metadata["rule"] != "doDaylightCycle" {
		t.Errorf("expected doDaylightCycle, got %v", e.Metadata["rule"])
	}
	if e.Metadata["value"] != "false" {
		t.Errorf("expected false, got %v", e.Metadata["value"])
	}

	e2 := ParseLine("test-server", "[14:31:00] [Server thread/INFO]: Game rule doDaylightCycle has been changed to true")
	if e2 == nil {
		t.Fatal("expected event for passive format, got nil")
	}
	if e2.Metadata["rule"] != "doDaylightCycle" || e2.Metadata["value"] != "true" {
		t.Errorf("expected doDaylightCycle/true, got %v/%v", e2.Metadata["rule"], e2.Metadata["value"])
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
	result := parseMCTime("14:30:00")
	if result.Hour() != 14 || result.Minute() != 30 || result.Second() != 0 {
		t.Errorf("expected 14:30:00, got %02d:%02d:%02d", result.Hour(), result.Minute(), result.Second())
	}
	if result.Year() != now.Year() || result.Month() != now.Month() || result.Day() != now.Day() {
		t.Errorf("expected today's date, got %s", result.Format("2006-01-02"))
	}
}
