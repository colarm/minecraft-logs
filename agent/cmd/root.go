package cmd

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"minecraft-log-agent/config"
	"minecraft-log-agent/internal/mock"
	"minecraft-log-agent/internal/nats"
	"minecraft-log-agent/internal/parser"
	"minecraft-log-agent/internal/tailer"

	"github.com/spf13/cobra"
)

// LogLineReader is the interface for consuming log lines.
// The real implementation runs docker logs; the mock reads from a file.
type LogLineReader interface {
	Lines() <-chan string
	Start(ctx context.Context) error
	Stop() error
}

// EventPublisher is the interface for publishing parsed events.
// The real implementation uses NATS; the mock prints to stdout.
type EventPublisher interface {
	Connect() error
	Publish(event *parser.Event)
	Close()
}

var (
	cfgFile     string
	containerID string
	serverID    string
	natsURL     string
	natsToken   string

	// mock mode flags
	mockMode    bool
	mockLogFile string
)

var rootCmd = &cobra.Command{
	Use:   "minecraft-log-agent",
	Short: "Ultra-lightweight Minecraft log agent",
	Long:  "Streams Docker container logs, parses events, and publishes to NATS.",
	RunE:  runAgent,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&cfgFile, "config", "c", "./config.yaml", "config file path")
	rootCmd.Flags().StringVar(&containerID, "container", "", "override container name/ID")
	rootCmd.Flags().StringVar(&serverID, "server-id", "", "override server ID")
	rootCmd.Flags().StringVar(&natsURL, "nats-url", "", "override NATS URL")
	rootCmd.Flags().StringVar(&natsToken, "nats-token", "", "override NATS token")

	// Mock mode: run without Docker / NATS
	rootCmd.Flags().BoolVar(&mockMode, "mock", false, "run in mock mode (read local file, print to stdout)")
	rootCmd.Flags().StringVar(&mockLogFile, "mock-log", "./testdata/sample.log", "log file to read in mock mode")
}

func loadConfig() *config.Config {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		if mockMode {
			log.Printf("config: %v (using defaults in mock mode)", err)
			return &config.Config{
				ServerID: "mock-server",
				Nats: config.NatsConfig{
					URL: "nats://localhost:4222",
				},
			}
		}
		log.Fatalf("config: %v", err)
	}
	return cfg
}

func runAgent(cmd *cobra.Command, args []string) error {
	cfg := loadConfig()
	if cfg == nil {
		return nil
	}

	if containerID != "" {
		cfg.ContainerName = containerID
	}
	if serverID != "" {
		cfg.ServerID = serverID
	}
	if natsURL != "" {
		cfg.Nats.URL = natsURL
	}
	if natsToken != "" {
		cfg.Nats.Token = natsToken
	}

	// In mock mode, ensure we always have a server ID even without config
	if mockMode && cfg.ServerID == "" {
		cfg.ServerID = "mock-server"
	}

	log.Printf("agent: starting for server %q, tailing container %q", cfg.ServerID, cfg.ContainerName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("agent: received %v, shutting down", sig)
		cancel()
	}()

	// ── Choose implementation: mock or real ──────────────────────
	var lineSrc LogLineReader
	var eventPub EventPublisher

	if mockMode {
		log.Printf("agent: MOCK MODE — reading from %s, printing to stdout", mockLogFile)
		lineSrc = mock.NewFileLineReader(mockLogFile)
		eventPub = mock.NewConsolePublisher(cfg.ServerID)
		if err := eventPub.Connect(); err != nil {
			return err
		}
		defer eventPub.Close()
	} else {
		// Real NATS publisher
		pub := nats.New(cfg.Nats.URL, cfg.Nats.Token, cfg.ServerID)
		if err := pub.Connect(); err != nil {
			log.Printf("nats: initial connect failed, will retry: %v", err)
			go func() {
				for {
					if err := pub.Connect(); err == nil {
						break
					}
					select {
					case <-ctx.Done():
						return
					default:
					}
				}
			}()
		}
		eventPub = pub
		defer eventPub.Close()

		// Real Docker tailer
		lineSrc = tailer.New(cfg.ContainerName, 64)
	}

	if err := lineSrc.Start(ctx); err != nil {
		return err
	}
	defer lineSrc.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case line := <-lineSrc.Lines():
			eventID := fmt.Sprintf("%x", sha256.Sum256([]byte(line)))
			event := parser.ParseLine(cfg.ServerID, line)
			if event != nil {
				event.ID = eventID
				eventPub.Publish(event)
			}
		}
	}
}
