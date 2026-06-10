package parser

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Docker log prefix: 2024-06-15T14:30:00.123456789Z
var dockerTSRe = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z)\s+(.*)$`)

// Minecraft log line: [14:30:00] [Server thread/INFO]: message
var logLineRe = regexp.MustCompile(`^\[(\d{2}:\d{2}:\d{2})\]\s+\[([^/]+)/(\w+)\]:\s+(.*)$`)

type Matcher interface {
	Match(serverID string, ts time.Time, level string, message string) (*Event, bool)
}

type matcherFunc func(serverID string, ts time.Time, level string, message string) (*Event, bool)

func (f matcherFunc) Match(serverID string, ts time.Time, level string, message string) (*Event, bool) {
	return f(serverID, ts, level, message)
}

var matchers = []Matcher{
	matchPlayerJoin(),
	matchPlayerLeave(),
	matchChat(),
	matchDeath(),
	matchServerStart(),
	matchServerStop(),
	matchTps(),
	matchWhitelistReject(),
	matchLoginFail(),
	matchConnect(),
	matchBan(),
	matchCommand(),
	matchRcon(),
	matchOp(),
	matchAdvancement(),
	matchGameMode(),
	matchGameRule(),
	matchError(),
}

func matchPlayerJoin() Matcher {
	re := regexp.MustCompile(`^(\S+) joined the game$`)
	return matcherFunc(func(serverID string, ts time.Time, level string, msg string) (*Event, bool) {
		m := re.FindStringSubmatch(msg)
		if m == nil {
			return nil, false
		}
		return &Event{
			ServerID:   serverID,
			EventType:  "player_join",
			PlayerName: m[1],
			Timestamp:  ts,
		}, true
	})
}

func matchPlayerLeave() Matcher {
	re1 := regexp.MustCompile(`^(\S+) left the game$`)
	re2 := regexp.MustCompile(`^(\S+) lost connection: (.+)$`)
	return matcherFunc(func(serverID string, ts time.Time, level string, msg string) (*Event, bool) {
		if m := re1.FindStringSubmatch(msg); m != nil {
			return &Event{
				ServerID:   serverID,
				EventType:  "player_leave",
				PlayerName: m[1],
				Timestamp:  ts,
			}, true
		}
		if m := re2.FindStringSubmatch(msg); m != nil {
			return &Event{
				ServerID:   serverID,
				EventType:  "player_leave",
				PlayerName: m[1],
				Message:    m[2],
				Timestamp:  ts,
			}, true
		}
		return nil, false
	})
}

func matchChat() Matcher {
	re := regexp.MustCompile(`^<([^>]+?)>\s+(.+)$`)
	return matcherFunc(func(serverID string, ts time.Time, level string, msg string) (*Event, bool) {
		if m := re.FindStringSubmatch(msg); m != nil {
			name := m[1]
			if idx := strings.LastIndex(name, " "); idx > 0 && strings.HasPrefix(name, "[") {
				name = name[idx+1:]
			} else if strings.HasPrefix(name, "[") && strings.HasSuffix(name, "]") {
				name = name[1 : len(name)-1]
			}
			return &Event{
				ServerID:   serverID,
				EventType:  "chat",
				PlayerName: name,
				Message:    m[2],
				Timestamp:  ts,
			}, true
		}
		return nil, false
	})
}

func matchDeath() Matcher {
	re := regexp.MustCompile(`^(\S+)\s+(?:was slain by|was killed by|drowned|burned to death|fell from|suffocated in|froze to death|starved|died|was pricked to death|was doomed to fall|was squashed|was crushed|tried to swim in lava|burned to death|went up in flames|burned|fell out of the world|fell off a ladder|fell from a high place)(.*)?$`)
	return matcherFunc(func(serverID string, ts time.Time, level string, msg string) (*Event, bool) {
		if m := re.FindStringSubmatch(msg); m != nil {
			meta := map[string]interface{}{"death_message": msg}
			if m[2] != "" {
				meta["killer"] = strings.TrimSpace(m[2])
			}
			return &Event{
				ServerID:   serverID,
				EventType:  "death",
				PlayerName: m[1],
				Message:    msg,
				Metadata:   meta,
				Timestamp:  ts,
			}, true
		}
		return nil, false
	})
}

func matchServerStart() Matcher {
	re := regexp.MustCompile(`^(Done \([^)]+\)!|Starting minecraft server version)`)
	return matcherFunc(func(serverID string, ts time.Time, level string, msg string) (*Event, bool) {
		if re.MatchString(msg) {
			return &Event{
				ServerID:  serverID,
				EventType: "server_start",
				Message:   msg,
				Timestamp: ts,
			}, true
		}
		return nil, false
	})
}

func matchServerStop() Matcher {
	re := regexp.MustCompile(`^Stopping server$`)
	return matcherFunc(func(serverID string, ts time.Time, level string, msg string) (*Event, bool) {
		if re.MatchString(msg) {
			return &Event{
				ServerID:  serverID,
				EventType: "server_stop",
				Message:   msg,
				Timestamp: ts,
			}, true
		}
		return nil, false
	})
}

func matchTps() Matcher {
	re1 := regexp.MustCompile(`^TPS from last \d+ seconds: ([\d.]+)$`)
	re2 := regexp.MustCompile(`^MSPT: ([\d.]+) TPS: ([\d.]+)$`)
	return matcherFunc(func(serverID string, ts time.Time, level string, msg string) (*Event, bool) {
		if m := re1.FindStringSubmatch(msg); m != nil {
			tps, _ := strconv.ParseFloat(m[1], 64)
			return &Event{
				ServerID:  serverID,
				EventType: "tps",
				Metadata:  map[string]interface{}{"tps": tps, "source": "carpet"},
				Timestamp: ts,
			}, true
		}
		if m := re2.FindStringSubmatch(msg); m != nil {
			mspt, _ := strconv.ParseFloat(m[1], 64)
			tps, _ := strconv.ParseFloat(m[2], 64)
			return &Event{
				ServerID:  serverID,
				EventType: "tps",
				Metadata:  map[string]interface{}{"tps": tps, "mspt": mspt, "source": "spark"},
				Timestamp: ts,
			}, true
		}
		return nil, false
	})
}

func matchWhitelistReject() Matcher {
	re := regexp.MustCompile(`^Disconnecting (\S+): You are not white-listed on this server!$`)
	return matcherFunc(func(serverID string, ts time.Time, level string, msg string) (*Event, bool) {
		m := re.FindStringSubmatch(msg)
		if m == nil {
			return nil, false
		}
		return &Event{
			ServerID:   serverID,
			EventType:  "whitelist_reject",
			PlayerName: m[1],
			Message:    msg,
			Timestamp:  ts,
		}, true
	})
}

func matchLoginFail() Matcher {
	re := regexp.MustCompile(`^Disconnecting (\S+): Failed to verify username!$`)
	return matcherFunc(func(serverID string, ts time.Time, level string, msg string) (*Event, bool) {
		m := re.FindStringSubmatch(msg)
		if m == nil {
			return nil, false
		}
		return &Event{
			ServerID:   serverID,
			EventType:  "login_fail",
			PlayerName: m[1],
			Message:    "Failed to verify username",
			Timestamp:  ts,
		}, true
	})
}

func matchConnect() Matcher {
	re := regexp.MustCompile(`^\[/([\d.]+):(\d+)\] logged in with entity id \d+ at`)
	return matcherFunc(func(serverID string, ts time.Time, level string, msg string) (*Event, bool) {
		m := re.FindStringSubmatch(msg)
		if m == nil {
			return nil, false
		}
		return &Event{
			ServerID:  serverID,
			EventType: "connect",
			Metadata:  map[string]interface{}{"ip": m[1], "port": m[2]},
			Message:   msg,
			Timestamp: ts,
		}, true
	})
}

func matchBan() Matcher {
	re1 := regexp.MustCompile(`^Disconnecting (\S+): You are banned from this server\.?(.*)$`)
	re2 := regexp.MustCompile(`^Banned (\S+): (.+)$`)
	return matcherFunc(func(serverID string, ts time.Time, level string, msg string) (*Event, bool) {
		if m := re1.FindStringSubmatch(msg); m != nil {
			meta := map[string]interface{}{"reason": "banned"}
			if strings.TrimSpace(m[2]) != "" {
				meta["detail"] = strings.TrimPrefix(strings.TrimSpace(m[2]), "Reason: ")
			}
			return &Event{
				ServerID:   serverID,
				EventType:  "ban",
				PlayerName: m[1],
				Metadata:   meta,
				Message:    msg,
				Timestamp:  ts,
			}, true
		}
		if m := re2.FindStringSubmatch(msg); m != nil {
			return &Event{
				ServerID:   serverID,
				EventType:  "ban",
				PlayerName: m[1],
				Message:    m[2],
				Timestamp:  ts,
			}, true
		}
		return nil, false
	})
}

func matchCommand() Matcher {
	re := regexp.MustCompile(`^(\S+) issued server command: (.+)$`)
	return matcherFunc(func(serverID string, ts time.Time, level string, msg string) (*Event, bool) {
		m := re.FindStringSubmatch(msg)
		if m == nil {
			return nil, false
		}
		return &Event{
			ServerID:   serverID,
			EventType:  "command",
			PlayerName: m[1],
			Message:    m[2],
			Timestamp:  ts,
		}, true
	})
}

func matchRcon() Matcher {
	// RCON/Console commands show as bare /command in the message.
	// "issued by IP:port" suffix = came from RCON; bare = console or RCON without IP.
	reWithSource := regexp.MustCompile(`^/(.+)\s+issued by\s+(.+)$`)
	reBare := regexp.MustCompile(`^/(.+)$`)
	return matcherFunc(func(serverID string, ts time.Time, level string, msg string) (*Event, bool) {
		if m := reWithSource.FindStringSubmatch(msg); m != nil {
			return &Event{
				ServerID:  serverID,
				EventType: "rcon",
				Message:   m[1],
				Metadata:  map[string]interface{}{"source": m[2]},
				Timestamp: ts,
			}, true
		}
		if m := reBare.FindStringSubmatch(msg); m != nil {
			return &Event{
				ServerID:  serverID,
				EventType: "rcon",
				Message:   m[1],
				Timestamp: ts,
			}, true
		}
		return nil, false
	})
}

func matchOp() Matcher {
	reGive := regexp.MustCompile(`^Made (\S+) a server operator$`)
	reTake := regexp.MustCompile(`^(\S+) is no longer a server operator$`)
	return matcherFunc(func(serverID string, ts time.Time, level string, msg string) (*Event, bool) {
		if m := reGive.FindStringSubmatch(msg); m != nil {
			return &Event{
				ServerID:   serverID,
				EventType:  "op",
				PlayerName: m[1],
				Metadata:   map[string]interface{}{"action": "grant"},
				Timestamp:  ts,
			}, true
		}
		if m := reTake.FindStringSubmatch(msg); m != nil {
			return &Event{
				ServerID:   serverID,
				EventType:  "op",
				PlayerName: m[1],
				Metadata:   map[string]interface{}{"action": "revoke"},
				Timestamp:  ts,
			}, true
		}
		return nil, false
	})
}

func matchAdvancement() Matcher {
	re := regexp.MustCompile(`^(\S+) has (made the advancement|completed the challenge|completed the goal) \[(.+)\]$`)
	return matcherFunc(func(serverID string, ts time.Time, level string, msg string) (*Event, bool) {
		m := re.FindStringSubmatch(msg)
		if m == nil {
			return nil, false
		}
		return &Event{
			ServerID:   serverID,
			EventType:  "advancement",
			PlayerName: m[1],
			Message:    m[3],
			Metadata:   map[string]interface{}{"type": m[2]},
			Timestamp:  ts,
		}, true
	})
}

func matchGameMode() Matcher {
	re1 := regexp.MustCompile(`^(\S+) changed game mode to (\S+)$`)
	re2 := regexp.MustCompile(`^(\S+)'s game mode has been changed to (\S+)$`)
	re3 := regexp.MustCompile(`^Set (\S+)'s game mode to (\S+)$`)
	return matcherFunc(func(serverID string, ts time.Time, level string, msg string) (*Event, bool) {
		if m := re1.FindStringSubmatch(msg); m != nil {
			return &Event{
				ServerID:   serverID,
				EventType:  "game_mode",
				PlayerName: m[1],
				Metadata:   map[string]interface{}{"mode": m[2]},
				Timestamp:  ts,
			}, true
		}
		if m := re2.FindStringSubmatch(msg); m != nil {
			return &Event{
				ServerID:   serverID,
				EventType:  "game_mode",
				PlayerName: m[1],
				Metadata:   map[string]interface{}{"mode": m[2]},
				Timestamp:  ts,
			}, true
		}
		if m := re3.FindStringSubmatch(msg); m != nil {
			return &Event{
				ServerID:   serverID,
				EventType:  "game_mode",
				PlayerName: m[1],
				Metadata:   map[string]interface{}{"mode": m[2]},
				Timestamp:  ts,
			}, true
		}
		return nil, false
	})
}

func matchGameRule() Matcher {
	re1 := regexp.MustCompile(`^(\S+) changed the game rule: (\S+) = (.+)$`)
	re2 := regexp.MustCompile(`^Game rule (\S+) has been changed to (.+)$`)
	return matcherFunc(func(serverID string, ts time.Time, level string, msg string) (*Event, bool) {
		if m := re1.FindStringSubmatch(msg); m != nil {
			return &Event{
				ServerID:   serverID,
				EventType:  "game_rule",
				PlayerName: m[1],
				Metadata:   map[string]interface{}{"rule": m[2], "value": m[3]},
				Timestamp:  ts,
			}, true
		}
		if m := re2.FindStringSubmatch(msg); m != nil {
			return &Event{
				ServerID:  serverID,
				EventType: "game_rule",
				Metadata:  map[string]interface{}{"rule": m[1], "value": m[2]},
				Timestamp: ts,
			}, true
		}
		return nil, false
	})
}

func matchError() Matcher {
	return matcherFunc(func(serverID string, ts time.Time, level string, msg string) (*Event, bool) {
		if level == "ERROR" || level == "WARN" {
			return &Event{
				ServerID:  serverID,
				EventType: "error",
				Message:   msg,
				Metadata:  map[string]interface{}{"level": level},
				Timestamp: ts,
			}, true
		}
		return nil, false
	})
}

func ParseLine(serverID string, line string) *Event {
	// Strip docker timestamp prefix if present (docker logs --timestamps)
	var dockerTS time.Time
	if m := dockerTSRe.FindStringSubmatch(line); m != nil {
		dockerTS, _ = time.Parse(time.RFC3339Nano, m[1])
		line = m[2]
	}

	matches := logLineRe.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	level := matches[3]
	message := matches[4]

	// Use docker timestamp if available, otherwise fall back to Minecraft log time
	ts := dockerTS
	if ts.IsZero() {
		ts = parseMCTime(matches[1])
	}

	for _, m := range matchers {
		if event, ok := m.Match(serverID, ts, level, message); ok {
			return event
		}
	}

	return nil
}

func parseMCTime(timePart string) time.Time {
	now := time.Now()
	parts := strings.Split(timePart, ":")
	if len(parts) != 3 {
		return now
	}
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	s, _ := strconv.Atoi(parts[2])
	return time.Date(now.Year(), now.Month(), now.Day(), h, m, s, 0, now.Location())
}
