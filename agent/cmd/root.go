package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"minecraft-log-agent/config"
	"minecraft-log-agent/internal/nats"
	"minecraft-log-agent/internal/parser"
	"minecraft-log-agent/internal/tailer"

	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	containerID string
	serverID    string
	natsURL     string
	natsToken   string
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
}

func runAgent(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
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
	defer pub.Close()

	t := tailer.New(cfg.ContainerName, 64)
	if err := t.Start(ctx); err != nil {
		return err
	}
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case line := <-t.Lines():
			event := parser.ParseLine(cfg.ServerID, line)
			if event != nil {
				pub.Publish(event)
			}
		}
	}
}
