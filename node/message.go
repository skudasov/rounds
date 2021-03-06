package node

//go:generate stringer -type=MsgType

import (
	"fmt"
	"github.com/mr-tron/base58"
)

type MsgType int

const (
	Collect MsgType = iota
	Vector
)

type Messager interface {
	// GetSignature gets message signature
	GetSignature() []byte
	// GetPayload gets message payload
	GetPayload() interface{}
}

type PulseMessagePayload struct {
	Signature     []byte
	Rst           int64          `json:"rst"`
	Epoch         uint64         `json:"epoch"`
	From          string         `json:"from"`
	PulseProposal *PulseProposal `json:"proposal"`
}

type Header struct {
	Type MsgType `json:"type"`
}

type PulseMessage struct {
	Header
	Payload PulseMessagePayload `json:"payload"`
}

func NewPulseMessage(from string, signature []byte, epoch uint64, rst int64) *PulseMessage {
	t := randomBytesString(16)
	t = base58.Encode([]byte(t))
	return &PulseMessage{
		Header: Header{
			Type: Collect,
		},
		Payload: PulseMessagePayload{
			Signature:     signature,
			Epoch:         epoch,
			Rst:           rst,
			From:          from,
			PulseProposal: NewPulseProposal(from, t),
		},
	}
}

func (m PulseMessagePayload) GetEpoch() uint64 {
	return m.Epoch
}

func (m PulseMessagePayload) GetPayload() interface{} {
	return m.PulseProposal
}

func (m PulseMessagePayload) GetSignature() []byte {
	return m.Signature
}

func (m PulseMessagePayload) String() string {
	return fmt.Sprintf(
		"[ rst: %d, from: %s, proposal: %s ]",
		m.Rst,
		m.From,
		m.PulseProposal.String(),
	)
}

type PulseVectorPayload struct {
	Signature       []byte
	Rst             int64        `json:"rst"`
	Epoch           uint64       `json:"epoch"`
	From            string       `json:"from"`
	EntropiesVector *PulseVector `json:"entropies_vector"`
}

type PulseVectorMessage struct {
	Header
	Payload PulseVectorPayload `json:"payload"`
}

func NewPulseVectorMessage(from string, signature []byte, epoch uint64, rst int64, ens *PulseVector) *PulseVectorMessage {
	return &PulseVectorMessage{
		Header: Header{
			Type: Vector,
		},
		Payload: PulseVectorPayload{
			Signature:       signature,
			Rst:             rst,
			Epoch:           epoch,
			From:            from,
			EntropiesVector: ens,
		},
	}
}

func (m PulseVectorPayload) GetEpoch() uint64 {
	return m.Epoch
}

func (m PulseVectorPayload) GetPayload() interface{} {
	return m.EntropiesVector
}

func (m PulseVectorPayload) GetSignature() []byte {
	return m.Signature
}

func (m PulseVectorPayload) String() string {
	return fmt.Sprintf(
		"[ rst: %d, from: %s, vector: %s ]",
		m.Rst,
		m.From,
		m.EntropiesVector,
	)
}
