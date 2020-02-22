package node

import (
	"context"
	"crypto/ecdsa"
	"crypto/md5"
	"crypto/rand"
	"crypto/tls"
	"encoding/asn1"
	"encoding/json"
	"math/big"
	"rounds/logger"
	"time"
)

var (
	DefaultNetworkProto = "tcp"
	DummyHashData       = []byte("dummy")
)

// Noder describes pulse consensus node
type Noder interface {
	// VerifyMessageTrusted verifies that message is from known peer
	VerifyMessageTrusted([]byte) bool
	// GetClient gets node client
	GetClient() Clienter
	// GetAddr gets current receiving network port
	GetAddr() string
	// Sign signs new string
	Sign(data []byte) []byte
	// Commit stores block
	Commit(context.Context, BlockData) error
	// GetLatestBlock gets latest block, by epoch
	GetLatestBlock() (*BlockData, error)
	// GetLatestBlockNumber
	GetLatestBlockNumber() uint64
	// GetEpoch
	GetEpoch() uint64
	//SetEpoch
	SetEpoch(epoch uint64)
}

type Node struct {
	privateKey      *ecdsa.PrivateKey
	publicKey       *ecdsa.PublicKey
	publicKeyPem    string
	srvTlsCtx       *tls.Config
	clientTlsCtx    *tls.Config
	Addr            string
	Reconnect       int
	client          *Client
	Consensus       Consensus
	peers           []Peer
	peersPublicKeys []*ecdsa.PublicKey

	Epoch uint64
	store Storage

	log *logger.Logger
}

func NewNode(c *Config, priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey, pubPem string) *Node {
	tlsSrv, tlsClient := tlsContexts()
	s := NewBadgerStorage(c.Store.Host)
	n := &Node{
		priv,
		pub,
		pubPem,
		tlsClient,
		tlsSrv,
		c.Node.Addr,
		c.Node.Reconnect,
		NewClient(c),
		NewPulseConsensus(
			c.Node.Rounds.Collect.Duration,
			c.Node.Rounds.Exchange.Duration,
			c.Node.Rounds.Collect.MaxMessages,
			c.Node.Rounds.Exchange.MaxMessages,
		),
		c.Node.Peers,
		nil,
		0,
		s,
		logger.NewLogger(),
	}
	n.LoadPeerPublicKeys()
	n.Epoch = n.GetLatestBlockNumber()
	return n
}

func (n *Node) SetEpoch(epoch uint64) {
	n.Epoch = epoch
}

func (n *Node) GetEpoch() uint64 {
	return n.Epoch
}

func (n *Node) GetLatestBlockNumber() uint64 {
	return n.store.GetLatestBlockEpoch()
}

func (n *Node) GetLatestBlock() (*BlockData, error) {
	// no iterator api in cete for now
	b, err := n.store.GetLatestBlock()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (n *Node) Commit(ctx context.Context, b BlockData) error {
	if err := n.store.Commit(ctx, b); err != nil {
		return err
	}
	return nil
}

func (n *Node) LoadPeerPublicKeys() {
	keys := make([]*ecdsa.PublicKey, 0)
	for _, pubKeyPem := range n.peers {
		n.log.Infof("pub key loaded from: %s", pubKeyPem.PubKeyDir)
		keys = append(keys, LoadPublicKey(pubKeyPem.PubKeyDir))
	}
	n.peersPublicKeys = keys
}

func (n *Node) GetClient() Clienter {
	return n.client
}

func (n *Node) GetAddr() string {
	return n.Addr
}

// VerifyMessageTrusted gets signature from message and check if it's any known peer message
func (n *Node) VerifyMessageTrusted(signature []byte) bool {
	for _, pubKey := range n.peersPublicKeys {
		hasher := md5.New()
		if _, err := hasher.Write(DummyHashData); err != nil {
			n.log.Error(err)
		}
		// by default ecdsa marshal signature in asn1 format
		var esig struct {
			R, S *big.Int
		}
		if _, err := asn1.Unmarshal(signature, &esig); err != nil {
			n.log.Error(err)
			continue
		}
		verified := ecdsa.Verify(pubKey, hasher.Sum(nil), esig.R, esig.S)
		if verified {
			return true
		}
	}
	return false
}

// Sign signs dummy hash with private key
func (n *Node) Sign(data []byte) []byte {
	h := md5.New()
	if _, err := h.Write(data); err != nil {
		n.log.Fatal(err)
	}
	signhash := h.Sum(nil)
	sign, err := n.privateKey.Sign(rand.Reader, signhash, nil)
	if err != nil {
		n.log.Fatal(err)
	}
	return sign
}

// ConnectPeer connects to peer, if peer is offline nil the connection so it can be reconnected by other goroutine
func (n *Node) ConnectPeer(addr string) {
	n.log.Infof("connecting to peer: %s", addr)
	conn, err := tls.Dial(DefaultNetworkProto, addr, n.clientTlsCtx)
	if err != nil {
		n.log.Errorf("failed to connect peer: %s", addr)
		n.client.Conns[addr] = nil
		return
	}
	n.client.Conns[addr] = conn
	n.log.Infof("[ %s ] connected to %s", n.Addr, addr)
}

// ConnectPeers connect to peers from config, reconnect if connection is nil
func (n *Node) ConnectPeers() {
	for _, p := range n.peers {
		n.ConnectPeer(p.Addr)
	}
	for {
		time.Sleep(time.Duration(n.Reconnect) * time.Second)
		for addr, c := range n.client.Conns {
			if c == nil {
				n.log.Infof("reconnecting to peer: %s", addr)
				n.ConnectPeer(addr)
			}
		}
	}
}

// Schedule schedules round timings so it can be synced between nodes
func (n *Node) Schedule(cfg *Config) {
	for {
		cons := n.Consensus
		ts := time.Now()
		wait := time.Duration(cfg.Node.Rounds.PaceMs) * time.Millisecond
		startTime := ts.Truncate(wait).Add(wait)
		syncNodeWaitBeforeRoundStart := startTime.Sub(ts)
		n.log.Infof("round start time: %s", startTime.String())
		n.log.Infof("sync wait for: %.2f ms", float64(syncNodeWaitBeforeRoundStart/time.Millisecond))
		time.Sleep(syncNodeWaitBeforeRoundStart)
		// event will be dispatched every N seconds on every node if ntpd is working
		cons.GetStartChan() <- struct{}{}
	}
}

// Processing runs rounds forever, reacts to Schedule() signals to channels
func (n *Node) Processing() {
	for {
		cons := n.Consensus
		<-cons.GetStartChan()
		cons.FlushData()

		// Send pulses to all
		ctx, cancel1 := context.WithTimeout(context.Background(), time.Duration(cons.GetCollectDuration())*time.Millisecond)
		cons.SendPulses(ctx, n)
		cons.ReceivePulses(ctx, n)

		// After timeout send vectors to all
		ctx2, cancel2 := context.WithTimeout(context.Background(), time.Duration(cons.GetExchangeDuration())*time.Millisecond)
		cons.SendVectors(ctx2, n)
		cons.ReceiveVectors(ctx2, n)

		// Calculate data set and blame malicious nodes
		ctx3, cancel3 := context.WithTimeout(context.Background(), time.Duration(cons.GetCollectDuration())*time.Millisecond)
		cons.Commit(ctx3, n)

		bn := n.GetLatestBlockNumber()
		n.SetEpoch(bn)
		n.log.Infof("next pulse number: %d", bn+1)
		cancel1()
		cancel2()
		cancel3()
	}
}

// ReceiveLoop receives all messages, send them to round input channel switched by type
func (n *Node) ReceiveLoop() {
	ln, err := tls.Listen(DefaultNetworkProto, n.Addr, n.srvTlsCtx)
	if err != nil {
		n.log.Fatal(err)
	}

	defer ln.Close()
	n.log.Infof("Node started on %s", n.Addr)
	for {
		n.log.Infof("entering receiving loop")
		conn, err := ln.Accept()
		if err != nil {
			n.log.Errorf("server: accept: %s", err)
			break
		}
		defer conn.Close()
		n.log.Infof("server: accepted from %s", conn.RemoteAddr())
		// TODO: verify peer certificates or check signature?
		//tlscon, ok := conn.(*tls.Conn)
		//if !ok {
		//	n.log.Error("tls handshake error")
		//}
		//n.log.Infof("tls handshake success")
		//state := tlscon.ConnectionState()
		//for _, v := range state.PeerCertificates {
		//	b, err := x509.MarshalPKIXPublicKey(v.PublicKey)
		//	if err != nil {
		//		n.log.Error(err)
		//	}
		//	n.log.Debugf("peer certificates found, pubKey: %s", string(b))
		//}
		go func() {
			for {
				var rawMsg map[string]*json.RawMessage
				err := json.NewDecoder(conn).Decode(&rawMsg)
				if err != nil {
					// TODO: for now just reconnect if any error, get parsing errors and connect errors aside
					n.log.Infof("error reading stream or decoding")
					break
				}
				var msgType MsgType
				if err := json.Unmarshal(*rawMsg["type"], &msgType); err != nil {
					n.log.Error(err)
				}
				n.log.Debugf("[ %s ] received msg: %s", conn.LocalAddr(), msgType.String())
				switch msgType {
				case Collect:
					var pulsePayload = PulseMessagePayload{}
					if err := json.Unmarshal(*rawMsg["payload"], &pulsePayload); err != nil {
						n.log.Error(err)
					}
					n.log.Debugf("[ %s ] parsed msg: %s:%s", conn.LocalAddr(), msgType.String(), pulsePayload.String())
					n.Consensus.GetPulsesChan() <- pulsePayload
				case Vector:
					var vectorPayload = PulseVectorPayload{}
					if err := json.Unmarshal(*rawMsg["payload"], &vectorPayload); err != nil {
						n.log.Error(err)
					}
					n.log.Debugf("[ %s ] parsed msg: %s:%s", conn.LocalAddr(), msgType.String(), vectorPayload.String())
					n.Consensus.GetVectorsChan() <- vectorPayload
				default:
					panic("unknown message type")
				}
			}
		}()
	}
}
