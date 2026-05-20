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
var logLineRe = regexp.MustCompile(`^\[(\d{2}:\d{2}:\d{2})\]\s+\[([\w\-]+)/(\w+)\]:\s+(.*)$`)

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
	re := regexp.MustCompile(`^(Done \(.+!\)|Starting minecraft server version)`)
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
