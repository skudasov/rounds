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
}

type TCPClient struct {
	cfg          *Config
	Addr         string
	clientTlsCtx *tls.Config
	Peers        []Peer
	Conns        map[string]net.Conn

	log *logger.Logger
}

func NewTCPClient(c *Config, clientTlsCtx *tls.Config) *TCPClient {
	client := &TCPClient{c, c.Node.Addr, clientTlsCtx, c.Node.Peers, make(map[string]net.Conn), logger.NewLogger()}
	go client.ConnectPeers(c.Node.Reconnect)
	return client
}

// ConnectPeer connects to peer, if peer is offline nil the connection so it can be reconnected later
func (m *TCPClient) ConnectPeer(addr string) {
	m.log.Infof("connecting to peer: %s", addr)
	conn, err := tls.Dial("tcp", addr, m.clientTlsCtx)
	if err != nil {
		m.log.Errorf("failed to connect peer: %s", addr)
		m.Conns[addr] = nil
		return
	}
	m.Conns[addr] = conn
	m.log.Infof("[ %s ] connected to %s", m.Addr, addr)
}

// ConnectPeers connect to peers from config, reconnect if connection is nil
func (m *TCPClient) ConnectPeers(reconnect int) {
	for _, p := range m.Peers {
		m.ConnectPeer(p.Addr)
	}
	for {
		time.Sleep(time.Duration(reconnect) * time.Second)
		for addr, c := range m.Conns {
			if c == nil {
				m.log.Infof("reconnecting to peer: %s", addr)
				m.ConnectPeer(addr)
			}
		}
	}
}

// Broadcast send messages to all peers
func (m *TCPClient) Broadcast(ctx context.Context, msg interface{}) error {
	tagCtx, err := tag.New(
		context.Background(),
		tag.Insert(KeyLabel, m.cfg.Opencensus.Prometheus.Nodelabel),
		tag.Insert(KeyMethod, "broadcast_TCP"),
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
		go func(ci string, c net.Conn) {
			m.log.Debugf("sending msg to %s: %s", c.RemoteAddr(), msg)
			err := json.NewEncoder(c).Encode(msg)
			if err != nil {
				if err := c.Close(); err != nil {
					m.log.Errorf("failed to close peer connection: %s", c.LocalAddr())
				}
				m.log.Errorf("failed to send msg to: %s", c.RemoteAddr())
				m.Conns[ci] = nil
			}
		}(ci, c)
	}
	stats.Record(tagCtx, BroadcastMs.M(SinceInMilliseconds(startTime)))
	return nil
}

type UDPClient struct {
	cfg   *Config
	Addr  string
	Peers []Peer
	Conns map[string]net.Conn

	log *logger.Logger
}

func NewUDPClient(c *Config) *UDPClient {
	client := &UDPClient{c, c.Node.Addr, c.Node.Peers, make(map[string]net.Conn), logger.NewLogger()}
	go client.ConnectPeers(c.Node.Reconnect)
	return client
}

// ConnectPeer connects to peer, if peer is offline nil the connection so it can be reconnected later
func (m *UDPClient) ConnectPeer(addr string) {
	m.log.Infof("connecting to peer: %s", addr)
	remoteAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		m.log.Errorf("failed to resolve udp addr: %s", addr)
	}
	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		m.log.Errorf("failed to connect peer: %s", addr)
		m.Conns[addr] = nil
		return
	}
	m.Conns[addr] = conn
	m.log.Infof("[ %s ] connected to %s", m.Addr, addr)
}

// ConnectPeers connect to peers from config, reconnect if connection is nil
func (m *UDPClient) ConnectPeers(reconnect int) {
	for _, p := range m.Peers {
		m.ConnectPeer(p.Addr)
	}
	for {
		time.Sleep(time.Duration(reconnect) * time.Second)
		for addr, c := range m.Conns {
			if c == nil {
				m.log.Infof("reconnecting to peer: %s", addr)
				m.ConnectPeer(addr)
			}
		}
	}
}

// Broadcast send messages to all peers
func (m *UDPClient) Broadcast(ctx context.Context, msg interface{}) error {
	tagCtx, err := tag.New(
		context.Background(),
		tag.Insert(KeyLabel, m.cfg.Opencensus.Prometheus.Nodelabel),
		tag.Insert(KeyMethod, "broadcast_UDP"),
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
		go func(ci string, c net.Conn) {
			m.log.Debugf("sending msg to %s: %s", c.RemoteAddr(), msg)
			err = json.NewEncoder(c).Encode(msg)
			if err != nil {
				if err := c.Close(); err != nil {
					m.log.Errorf("failed to close peer connection: %s", c.LocalAddr())
				}
				m.log.Errorf("failed to send msg to: %s", c.RemoteAddr())
				m.Conns[ci] = nil
			}
		}(ci, c)
	}
	stats.Record(tagCtx, BroadcastMs.M(SinceInMilliseconds(startTime)))
	return nil
}

func tlsContexts() (*tls.Config, *tls.Config) {
	certClient, err := tls.LoadX509KeyPair("certs/client.pem", "certs/client.key")
	if err != nil {
		log.Fatalf("client: load keys: %s", err)
	}
	configClient := tls.Config{Certificates: []tls.Certificate{certClient}, InsecureSkipVerify: true}

	certServer, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	if err != nil {
		log.Fatalf("server: load keys: %s", err)
	}
	configServer := tls.Config{Certificates: []tls.Certificate{certServer}}
	return &configClient, &configServer
}
