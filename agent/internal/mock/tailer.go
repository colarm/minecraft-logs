// Package mock provides mock implementations for local testing without
// a Docker/Minecraft container or NATS server.
package mock

import (
	"bufio"
	"context"
	"log"
	"os"
	"sync"
)

// FileLineReader reads log lines from a local file instead of running
// docker logs. It sends all existing lines on Start and optionally
// watches the file for appends (follow mode).
type FileLineReader struct {
	path  string
	lines chan string
	mu    sync.Mutex
}

// NewFileLineReader creates a FileLineReader that reads from path.
func NewFileLineReader(path string) *FileLineReader {
	return &FileLineReader{
		path:  path,
		lines: make(chan string, 256),
	}
}

// Lines returns the read-only channel of log line strings.
func (r *FileLineReader) Lines() <-chan string {
	return r.lines
}

// Start opens the file and reads all existing lines into the channel in a
// background goroutine. After reading, the channel is closed so the agent's
// main loop receives a zero-value read and can detect completion via comma-ok.
//
// In mock mode the agent reads a static file and exits; there is no follow
// loop. Caller must cancel ctx if the readLoop needs to be interrupted early.
func (r *FileLineReader) Start(ctx context.Context) error {
	if _, err := os.Stat(r.path); err != nil {
		return err
	}

	go r.readLoop(ctx)
	return nil
}

// readLoop runs in a goroutine, reads the file in full, closes the lines
// channel, and returns. Close signals the agent's main loop to exit.
func (r *FileLineReader) readLoop(ctx context.Context) {
	defer close(r.lines)

	f, err := os.Open(r.path)
	if err != nil {
		log.Printf("mock: open error: %v", err)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 256*1024)

	var lineCount int
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		lineCount++
		select {
		case r.lines <- text:
		case <-ctx.Done():
			return
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("mock: file scanner error: %v", err)
	}

	log.Printf("mock: read %d lines from %s", lineCount, r.path)
}

// Stop is a no-op for the mock; cleanup happens via ctx cancel.
func (r *FileLineReader) Stop() error {
	return nil
}
