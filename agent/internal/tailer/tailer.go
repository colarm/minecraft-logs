package tailer

import (
	"bufio"
	"context"
	"log"
	"os/exec"
	"strings"
	"time"
)

type Tailer struct {
	containerName string
	lines         chan string
	cmd           *exec.Cmd
}

func New(containerName string, bufferSize int) *Tailer {
	return &Tailer{
		containerName: containerName,
		lines:         make(chan string, bufferSize),
	}
}

func (t *Tailer) Lines() <-chan string {
	return t.lines
}

func (t *Tailer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		cmd := exec.CommandContext(ctx, "docker", "logs",
			"--timestamps",
			"--since", "now",
			"--follow",
			t.containerName,
		)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("tailer: docker logs failed: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if err := cmd.Start(); err != nil {
			log.Printf("tailer: docker logs start failed: %v, retrying in 5s", err)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
			}
			continue
		}

		t.cmd = cmd
		go t.readLines(stdout, ctx)
		go t.waitExit(cmd, ctx)
		return nil
	}
}

func (t *Tailer) Stop() error {
	if t.cmd != nil && t.cmd.Process != nil {
		return t.cmd.Process.Kill()
	}
	return nil
}

func (t *Tailer) readLines(stdout interface{ Read(p []byte) (n int, err error) }, ctx context.Context) {
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 256*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		select {
		case t.lines <- text:
		case <-ctx.Done():
			return
		}
	}
}

func (t *Tailer) waitExit(cmd *exec.Cmd, ctx context.Context) {
	cmd.Wait()
	select {
	case <-ctx.Done():
		return
	default:
	}
	log.Printf("tailer: docker logs exited, reconnecting in 5s")
	time.Sleep(5 * time.Second)
	t.Start(ctx)
}
