package node

import (
	"context"
	"crypto/ecdsa"
	"crypto/md5"
	"crypto/rand"
	"encoding/asn1"
	"encoding/json"
	"math/big"
	"net"
	"rounds/logger"
	"time"
)

var (
	DummyHashData = []byte("dummy")
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
	// GetLatestPulse gets latest block, by epoch
	GetLatestPulse() (*BlockData, error)
	// GetLatestPulseNumber
	GetLatestPulseNumber() uint64
	// GetPulseNumber
	GetPulseNumber() uint64
	// SetPulseNumber
	SetPulseNumber(epoch uint64)
	// RouteMsg
	RouteMsg(addr net.Addr, rawMsg map[string]*json.RawMessage)
}

type Node struct {
	privateKey      *ecdsa.PrivateKey
	publicKey       *ecdsa.PublicKey
	publicKeyPem    string
	receiving       bool
	transport       Transport
	Addr            string
	Reconnect       int
	client          Clienter
	Consensus       Consensus
	peersPublicKeys []*ecdsa.PublicKey

	Epoch uint64
	store Storage

	log *logger.Logger
}

func (n *Node) PublicKeyPem() string {
	return n.publicKeyPem
}

func NewNode(c *Config, priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey, pubPem string) *Node {
	tlsClient, tlsSrv := tlsContexts()
	s := NewBadgerStorage(c.Store.Host)
	var transport Transport
	var client Clienter
	switch c.Node.Transport {
	case "tcp":
		transport = NewTCPTransport(tlsSrv)
		client = NewTCPClient(c, tlsClient)
	case "udp":
		transport = NewUDPTransport()
		client = NewUDPClient(c)
	}
	n := &Node{
		priv,
		pub,
		pubPem,
		false,
		transport,
		c.Node.Addr,
		c.Node.Reconnect,
		client,
		NewPulseConsensus(
			c.Node.Rounds.Collect.Duration,
			c.Node.Rounds.Exchange.Duration,
			c.Node.Rounds.Collect.MaxMessages,
			c.Node.Rounds.Exchange.MaxMessages,
		),
		nil,
		0,
		s,
		logger.NewLogger(),
	}
	n.LoadPeerPublicKeys(c.Node.Peers)
	n.Epoch = n.GetLatestPulseNumber()
	return n
}

func (n *Node) SetPulseNumber(epoch uint64) {
	n.Epoch = epoch
}

func (n *Node) GetPulseNumber() uint64 {
	return n.Epoch
}

func (n *Node) GetLatestPulseNumber() uint64 {
	return n.store.GetLatestBlockEpoch()
}

func (n *Node) GetLatestPulse() (*BlockData, error) {
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

func (n *Node) LoadPeerPublicKeys(peers []Peer) {
	keys := make([]*ecdsa.PublicKey, 0)
	for _, pubKeyPem := range peers {
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

//func (n *Node) ConnectPeers() {
//	n.client.ConnectPeers(n.Reconnect)
//}

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
		cons.GetStartChan() <- startTime.Unix()
	}
}

// Processing runs rounds forever, reacts to Schedule() signals to channels
func (n *Node) Processing() {
	for {
		cons := n.Consensus
		startTimeUnix := <-cons.GetStartChan()
		cons.FlushData()
		cons.SetRoundStartTime(startTimeUnix)

		// Send pulses to all
		ctx, cancel1 := context.WithTimeout(context.Background(), time.Duration(cons.GetCollectDuration())*time.Millisecond)
		cons.SendPulses(ctx, n)
		cons.ReceivePulses(ctx, n)

		// After timeout send vectors to all
		ctx2, cancel2 := context.WithTimeout(context.Background(), time.Duration(cons.GetExchangeDuration())*time.Millisecond)
		cons.SendVectors(ctx2, n)
		cons.ReceiveVectors(ctx2, n)

		// Calculate approved data set
		ctx3, cancel3 := context.WithTimeout(context.Background(), time.Duration(cons.GetCollectDuration())*time.Millisecond)
		cons.Commit(ctx3, n)

		bn := n.GetLatestPulseNumber()
		n.SetPulseNumber(bn)
		n.log.Infof("next pulse number: %d", bn+1)
		cancel1()
		cancel2()
		cancel3()
	}
}

// Serve receives all messages, send them to router
func (n *Node) StartTransport() {
	n.transport.Serve(n)
}

// RouteMsg routes messages to channel by type
func (n *Node) RouteMsg(addr net.Addr, rawMsg map[string]*json.RawMessage) {
	var msgType MsgType
	if err := json.Unmarshal(*rawMsg["type"], &msgType); err != nil {
		n.log.Error(err)
	}
	n.log.Debugf("[ %s ] received msg: %s", addr, msgType.String())
	switch msgType {
	case Collect:
		var pulsePayload = PulseMessagePayload{}
		if err := json.Unmarshal(*rawMsg["payload"], &pulsePayload); err != nil {
			n.log.Error(err)
		}
		n.log.Debugf("[ %s ] parsed msg: %s:%s", addr, msgType.String(), pulsePayload.String())
		n.Consensus.GetPulsesChan() <- pulsePayload
	case Vector:
		var vectorPayload = PulseVectorPayload{}
		if err := json.Unmarshal(*rawMsg["payload"], &vectorPayload); err != nil {
			n.log.Error(err)
		}
		n.log.Debugf("[ %s ] parsed msg: %s:%s", addr, msgType.String(), vectorPayload.String())
		n.Consensus.GetVectorsChan() <- vectorPayload
	default:
		n.log.Infof("unknown message type received: %s", msgType)
	}
}
