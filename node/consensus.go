package node

import (
	"bytes"
	"context"
	"encoding/gob"
	"github.com/prometheus/common/log"
	"rounds/logger"
	"sort"
	"strings"
	"time"
)

const (
	NoConsensusStatus = "no_consensus"
)

// Consensus describes abstract rounds of consensus
type Consensus interface {
	// Flushes pulses and vectors data
	FlushData()
	// SendPulses sends pulse data proposals
	SendPulses(ctx context.Context, n Noder) error
	// SendVectors sends acquired pulse vector
	SendVectors(ctx context.Context, n Noder) error
	// ReceivePulses collects data
	ReceivePulses(ctx context.Context, n Noder)
	// ReceiveVectors receives all vectors from peers
	ReceiveVectors(ctx context.Context, n Noder)
	// Commit commits matching set of consensus
	Commit(ctx context.Context, n Noder)
	// GetCollectDuration collect round duration
	GetCollectDuration() int
	// GetExchangeDuration exchange round duration
	GetExchangeDuration() int
	// GetStartChan round start channel
	GetStartChan() chan struct{}
	// GetPulsesChan pulse messages
	GetPulsesChan() chan Messager
	// GetVectorsChan vector messages
	GetVectorsChan() chan Messager
}

type PulseConsensus struct {
	TotalNodes       int
	CollectDuration  int
	ExchangeDuration int
	StartChan        chan struct{}
	PulsesChan       chan Messager
	VectorChan       chan Messager
	SelfProposal     *PulseProposal
	PulseProposals   []*PulseProposal
	PulseVectors     []*PulseVector
	MajorityData     []string

	log *logger.Logger
}

type PulseVector struct {
	From   string
	Vector []*PulseProposal
}

func (p *PulseVector) String() string {
	vecs := make([]string, 0)
	for _, v := range p.Vector {
		vecs = append(vecs, v.String())
	}
	return strings.Join(vecs, " ")
}

func NewPulseConsensus(collectDuration int, exchangeDuration int, maxPulsesChan int, maxVectorsChan int) *PulseConsensus {
	return &PulseConsensus{
		4,
		collectDuration,
		exchangeDuration,
		make(chan struct{}),
		make(chan Messager, maxPulsesChan),
		make(chan Messager, maxVectorsChan),
		nil,
		make([]*PulseProposal, 0),
		make([]*PulseVector, 0),
		make([]string, 0),
		logger.NewLogger(),
	}
}

func (r *PulseConsensus) GetStartChan() chan struct{} {
	return r.StartChan
}

func (r *PulseConsensus) GetPulsesChan() chan Messager {
	return r.PulsesChan
}

func (r *PulseConsensus) GetVectorsChan() chan Messager {
	return r.VectorChan
}

func (r *PulseConsensus) GetCollectDuration() int {
	return r.CollectDuration
}

func (r *PulseConsensus) GetExchangeDuration() int {
	return r.ExchangeDuration
}

func (r *PulseConsensus) SendPulses(ctx context.Context, n Noder) error {
	log.Infof("collect round started")
	s := n.Sign(DummyHashData)
	pm := NewPulseMessage(n.GetAddr(), s, n.GetEpoch())
	// add self entropy too
	selfProposal := pm.Payload.PulseProposal
	r.PulseProposals = append(r.PulseProposals, selfProposal)
	r.SelfProposal = selfProposal
	if err := n.GetClient().Broadcast(ctx, pm); err != nil {
		return err
	}
	return nil
}

func (r *PulseConsensus) FlushData() {
	log.Infof("flushing pulses data")
	r.PulseProposals = make([]*PulseProposal, 0)
	log.Infof("flushing vector data")
	r.PulseVectors = make([]*PulseVector, 0)
}

func (r *PulseConsensus) ReceivePulses(ctx context.Context, n Noder) {
	log.Infof("finalizing collect round #%d", n.GetEpoch())
	for {
		select {
		case <-ctx.Done():
			log.Infof("collect round #%d ended", n.GetEpoch())
			log.Debugf("proposals for round #%d: %s", n.GetEpoch(), r.PulseProposals)
			return
		case msg := <-r.PulsesChan:
			signature := msg.GetSignature()
			if n.VerifyMessageTrusted(signature) {
				log.Infof("message verified")
				r.PulseProposals = append(r.PulseProposals, msg.GetPayload().(*PulseProposal))
				continue
			}
			log.Error("message verification failed, signature is not from known public keys")
		}
	}
}

func (r *PulseConsensus) SendVectors(ctx context.Context, n Noder) error {
	log.Infof("exchange round started")
	s := n.Sign(DummyHashData)
	// add self vectors too
	pv := &PulseVector{n.GetAddr(), r.PulseProposals}
	r.PulseVectors = append(r.PulseVectors, &PulseVector{n.GetAddr(), r.PulseProposals})
	log.Debugf("pulse proposals: %s", r.PulseProposals)
	if err := n.GetClient().Broadcast(ctx, NewPulseVectorMessage(n.GetAddr(), s, n.GetEpoch(), pv)); err != nil {
		return err
	}
	return nil
}

func (r *PulseConsensus) ReceiveVectors(ctx context.Context, n Noder) {
	log.Infof("finalizing exchange round #%d", n.GetEpoch())
	for {
		select {
		case <-ctx.Done():
			log.Infof("exchange round #%d ended", n.GetEpoch())
			log.Debugf("vectors for round #%d: %s", n.GetEpoch(), r.PulseVectors)
			return
		case msg := <-r.VectorChan:
			signature := msg.GetSignature()
			if n.VerifyMessageTrusted(signature) {
				log.Infof("message verified")
				r.PulseVectors = append(r.PulseVectors, msg.GetPayload().(*PulseVector))
				continue
			}
			log.Error("message verification failed, signature is not from known public keys")
		}
	}
}

func (r *PulseConsensus) Commit(ctx context.Context, n Noder) {
	log.Infof("committing consensus data round #%d", n.GetEpoch())
	for {
		select {
		case <-ctx.Done():
			log.Infof("commit round ended")
			return
		default:
			winner := r.DecideWinner()
			log.Infof("winner: %s, me: %s", winner, r.SelfProposal.Entropy)
			if winner == r.SelfProposal.Entropy {
				log.Infof("committing winner pulse")
				var buf bytes.Buffer
				enc := gob.NewEncoder(&buf)
				if err := enc.Encode(winner); err != nil {
					log.Errorf("failed to encode pulse proposals: %s", err)
					continue
				}
				b := BlockData{
					time.Now().Unix(),
					buf.Bytes(),
				}
				log.Debugf("committing block: %v", b)
				if err := n.Commit(context.Background(), b); err != nil {
					log.Error(ErrStorageConnection(err))
				}
				return
			}
			return
		}
	}
}

// DecideWinner counts BFT data versions and select random pulsar as a winner
// https://ru.wikipedia.org/wiki/%D0%97%D0%B0%D0%B4%D0%B0%D1%87%D0%B0_%D0%B2%D0%B8%D0%B7%D0%B0%D0%BD%D1%82%D0%B8%D0%B9%D1%81%D0%BA%D0%B8%D1%85_%D0%B3%D0%B5%D0%BD%D0%B5%D1%80%D0%B0%D0%BB%D0%BE%D0%B2
func (r *PulseConsensus) DecideWinner() string {
	versions := make(map[string]int)
	for _, ver := range r.PulseVectors {
		for _, proposal := range ver.Vector {
			// skip version of node we are counting gossips for
			if strings.Contains(proposal.String(), ver.From) {
				continue
			}
			versions[proposal.Entropy] += 1
		}
	}
	log.Infof("versions: %v", versions)
	r.MajorityData = r.AgreeSet(versions)
	log.Infof("majority data: %v", r.MajorityData)
	if len(r.MajorityData) == 0 {
		log.Infof("failed to establish consensus")
		return NoConsensusStatus
	} else {
		return r.Winner(r.MajorityData)
	}
}

// AgreeSet forms set from pulses if 2/3 of nodes agree on data
func (r *PulseConsensus) AgreeSet(versions map[string]int) []string {
	majorityData := make([]string, 0)
	for entropy, versionCount := range versions {
		if versionCount >= (r.TotalNodes-1)*2/3 {
			majorityData = append(majorityData, entropy)
		}
	}
	sort.Strings(majorityData)
	return majorityData
}

// Winner select random winning entropy, random is the same across all nodes
func (r *PulseConsensus) Winner(ents []string) string {
	return ents[hashFnv64(ents)%uint64(len(ents))]
}
