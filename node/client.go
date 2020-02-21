package node

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/prometheus/common/log"
	"go.opencensus.io/trace"
	"net"
	"rounds/logger"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

// Clienter for now tcp client to send data to peers
type Clienter interface {
	// Broadcast sends a message to all peers
	Broadcast(context.Context, interface{}) error
	// GetConns get peers connection
	GetConns() map[string]net.Conn
}

type Client struct {
	cfg   *Config
	Conns map[string]net.Conn

	log *logger.Logger
}

func NewClient(c *Config) *Client {
	return &Client{c, make(map[string]net.Conn), logger.NewLogger()}
}

func (m *Client) GetConns() map[string]net.Conn {
	return m.Conns
}

// Broadcast sends message to all peers
func (m *Client) Broadcast(ctx context.Context, msg interface{}) error {
	tagCtx, err := tag.New(
		context.Background(),
		tag.Insert(KeyLabel, m.cfg.Opencensus.Prometheus.Nodelabel),
		tag.Insert(KeyMethod, "broadcast"),
	)
	if err != nil {
		return err
	}
	startTime := time.Now()
	_, span := trace.StartSpan(context.Background(), "Broadcast")
	span.AddAttributes(
		trace.StringAttribute("method", "BC"),
	)
	defer span.End()

	for ci, c := range m.Conns {
		if c == nil {
			continue
		}
		go func(c net.Conn) {
			m.log.Infof("sending msg to %s: %s", c.RemoteAddr(), msg)
			err := json.NewEncoder(c).Encode(msg)
			if err != nil {
				if err := c.Close(); err != nil {
					m.log.Errorf("failed to close peer connection: %s", c.LocalAddr())
				}
				m.Conns[ci] = nil
			}
		}(c)
	}
	stats.Record(tagCtx, BroadcastMs.M(SinceInMilliseconds(startTime)))
	return nil
}

func tlsContexts() (*tls.Config, *tls.Config) {
	certSender, err := tls.LoadX509KeyPair("certs/client.pem", "certs/client.key")
	if err != nil {
		log.Fatalf("client: load keys: %s", err)
	}
	configSender := tls.Config{Certificates: []tls.Certificate{certSender}, InsecureSkipVerify: true}

	certListener, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	if err != nil {
		log.Fatalf("server: load keys: %s", err)
	}
	configListener := tls.Config{Certificates: []tls.Certificate{certListener}}
	return &configSender, &configListener
}
