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

// Start 执行两阶段初始化：
//
// Phase 1 — 全量 dump：执行 docker logs（无 --follow、无 --since），
// 读取容器全部历史日志，等读完后才进入 Phase 2。
//
// Phase 2 — 实时 follow：启动 goroutine 执行 docker logs --since now --follow，
// 不断获取新日志，断线后自动重连。
func (t *Tailer) Start(ctx context.Context) error {
	// Phase 1: dump all existing logs (one-shot, no follow)
	dumpCmd := exec.CommandContext(ctx, "docker", "logs", "--timestamps", t.containerName)
	stdout, err := dumpCmd.StdoutPipe()
	if err != nil {
		log.Printf("tailer: dump pipe error: %v", err)
	} else {
		if err := dumpCmd.Start(); err != nil {
			log.Printf("tailer: dump start error: %v", err)
		} else {
			t.readLines(stdout, ctx)
			dumpCmd.Wait()
			log.Printf("tailer: initial log dump complete")
		}
	}

	// Phase 2: continuously follow new logs (reconnects automatically)
	go t.followLoop(ctx)
	return nil
}

func (t *Tailer) Stop() error {
	if t.cmd != nil && t.cmd.Process != nil {
		return t.cmd.Process.Kill()
	}
	return nil
}

// followLoop 持续执行 docker logs --since now --follow，进程退出后自动重连。
func (t *Tailer) followLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
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
			log.Printf("tailer: docker logs pipe: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if err := cmd.Start(); err != nil {
			log.Printf("tailer: docker logs start: %v, retrying in 5s", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}

		t.cmd = cmd
		go t.readLines(stdout, ctx)
		t.waitExit(cmd, ctx)
		// waitExit 返回后继续 for 循环，自动重连
	}
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

	if err := scanner.Err(); err != nil {
		log.Printf("tailer: scanner error: %v", err)
	}
}

// waitExit 阻塞等待 docker logs 进程退出，如有需要给 followLoop 的下一轮让出 5s。
// 不再递归调用 Start，由 followLoop 的 for 循环处理重试。
func (t *Tailer) waitExit(cmd *exec.Cmd, ctx context.Context) {
	cmd.Wait()
	select {
	case <-ctx.Done():
		return
	default:
	}
	log.Printf("tailer: docker logs exited, reconnecting in 5s")
	time.Sleep(5 * time.Second)
}
