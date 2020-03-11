package node

import (
	"crypto/tls"
	"encoding/json"
	"net"
	"rounds/logger"
)

type Transport interface {
	Serve(n Noder)
}

type UDPTransport struct {
	log *logger.Logger
}

func NewUDPTransport() *UDPTransport {
	return &UDPTransport{logger.NewLogger()}
}

func (m *UDPTransport) Serve(node Noder) {
	udpAddr, err := net.ResolveUDPAddr("udp", node.GetAddr())
	if err != nil {
		m.log.Fatal(err)
	}
	ln, err := net.ListenUDP("udp", udpAddr)
	defer ln.Close()
	if err != nil {
		m.log.Fatal(err)
	}
	m.log.Infof("UDP server up and listening on %s", node.GetAddr())
	for {
		var rawMsg map[string]*json.RawMessage
		err := json.NewDecoder(ln).Decode(&rawMsg)
		if err != nil {
			m.log.Infof("failed to unmarshal udp message, dropping: %s", err.Error())
			continue
		}
		node.RouteMsg(ln.LocalAddr(), rawMsg)
	}
}

type TCPTransport struct {
	log    *logger.Logger
	tlsCtx *tls.Config
}

func NewTCPTransport(tlsSrv *tls.Config) *TCPTransport {
	return &TCPTransport{logger.NewLogger(), tlsSrv}
}

func (m *TCPTransport) Serve(node Noder) {
	ln, err := tls.Listen("tcp", node.GetAddr(), m.tlsCtx)
	if err != nil {
		m.log.Fatal(err)
	}

	defer ln.Close()
	m.log.Infof("Node started on %s", node.GetAddr())
	for {
		m.log.Infof("entering receiving loop")
		conn, err := ln.Accept()
		if err != nil {
			m.log.Errorf("server: accept: %s", err)
			break
		}
		defer conn.Close()
		m.log.Infof("server: accepted from %s", conn.RemoteAddr())
		go func() {
			for {
				var rawMsg map[string]*json.RawMessage
				err := json.NewDecoder(conn).Decode(&rawMsg)
				if err != nil {
					// TODO: for now just reconnect if any error, get parsing errors and connect errors aside
					m.log.Infof("error reading stream or decoding")
					break
				}
				node.RouteMsg(conn.LocalAddr(), rawMsg)
			}
		}()
	}
}
