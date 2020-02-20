package node

import "fmt"

type PulseProposal struct {
	From    string
	Entropy string
}

func (m *PulseProposal) String() string {
	return fmt.Sprintf("[from: %s, data: %s]", m.From, m.Entropy)
}

func NewPulseProposal(from string, data string) *PulseProposal {
	return &PulseProposal{
		from,
		data,
	}
}
