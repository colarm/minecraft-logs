package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type NatsConfig struct {
	URL           string        `yaml:"url"`
	Token         string        `yaml:"token"`
	ReconnectWait time.Duration `yaml:"reconnect_wait"`
}

type Config struct {
	ServerID      string     `yaml:"server_id"`
	ContainerName string     `yaml:"container_name"`
	Nats          NatsConfig `yaml:"nats"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Nats.ReconnectWait == 0 {
		cfg.Nats.ReconnectWait = 5 * time.Second
	}

	return &cfg, nil
}
