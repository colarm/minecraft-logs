package nats

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"minecraft-log-agent/internal/parser"

	"github.com/nats-io/nats.go"
)

type Publisher struct {
	url      string
	token    string
	nc       *nats.Conn
	buf      []*parser.Event
	mu       sync.Mutex
	serverID string
}

func New(url, token, serverID string) *Publisher {
	return &Publisher{
		url:      url,
		token:    token,
		serverID: serverID,
		buf:      make([]*parser.Event, 0, 100),
	}
}

func (p *Publisher) Connect() error {
	opts := []nats.Option{
		nats.Token(p.token),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(5 * time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			log.Printf("nats: disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("nats: reconnected to %s", nc.ConnectedUrl())
			p.flushBuffer()
		}),
	}

	nc, err := nats.Connect(p.url, opts...)
	if err != nil {
		return err
	}
	p.nc = nc
	log.Printf("nats: connected to %s", p.url)
	return nil
}

func (p *Publisher) Publish(event *parser.Event) {
	if p.nc == nil || !p.nc.IsConnected() {
		p.buffer(event)
		return
	}

	subject := "minecraft.events." + p.serverID + "." + event.EventType
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("nats: marshal error: %v", err)
		return
	}

	if err := p.nc.Publish(subject, data); err != nil {
		log.Printf("nats: publish error: %v", err)
		p.buffer(event)
	}
}

func (p *Publisher) buffer(event *parser.Event) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.buf) < cap(p.buf) {
		p.buf = append(p.buf, event)
	} else {
		log.Printf("nats: buffer full, dropping oldest event")
		p.buf = append(p.buf[1:], event)
	}
}

func (p *Publisher) flushBuffer() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.buf) == 0 {
		return
	}

	log.Printf("nats: flushing %d buffered events", len(p.buf))
	for _, event := range p.buf {
		subject := "minecraft.events." + p.serverID + "." + event.EventType
		data, err := json.Marshal(event)
		if err != nil {
			continue
		}
		if err := p.nc.Publish(subject, data); err != nil {
			log.Printf("nats: flush publish error: %v", err)
		}
	}
	if err := p.nc.Flush(); err != nil {
		log.Printf("nats: flush error: %v", err)
	}
	p.buf = p.buf[:0]
}

func (p *Publisher) Close() {
	if p.nc != nil {
		p.nc.Close()
	}
}
