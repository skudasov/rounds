package node

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func basicCons() *PulseConsensus {
	return NewPulseConsensus(
		500,
		500,
		200,
		200,
	)
}

func TestDecideWinnerFound(t *testing.T) {
	cons := basicCons()
	versions := []*PulseVector{
		{
			From: "A",
			Vector: []*PulseProposal{
				{"A", "1"},
				{"B", "2"},
				{"C", "3"},
				{"D", "4"},
			},
		},
		{
			From: "B",
			Vector: []*PulseProposal{
				{"A", "1"},
				{"B", "2"},
				{"C", "3"},
				{"D", "4"},
			},
		},
		{
			From: "C",
			Vector: []*PulseProposal{
				{"A", "1"},
				{"B", "2"},
				{"C", "3"},
				{"D", "4"},
			},
		},
		{
			From: "D",
			Vector: []*PulseProposal{
				{"A", "1"},
				{"B", "2"},
				{"C", "3"},
				{"D", "4"},
			},
		},
	}
	cons.PulseVectors = versions
	winner := cons.DecideWinner()
	require.Equal(t, "2", winner)
}

func TestDecideWinnerStrictMajority(t *testing.T) {
	cons := basicCons()
	versions := []*PulseVector{
		{
			From: "A",
			Vector: []*PulseProposal{
				{"A", "1"},
				{"B", "2"},
				{"C", "3"},
				{"D", "4"},
			},
		},
		{
			From: "B",
			Vector: []*PulseProposal{
				{"A", "1"},
				{"B", "2"},
				{"C", "3"},
				{"D", "4"},
			},
		},
		{
			From: "C",
			Vector: []*PulseProposal{
				{"A", "1"},
				{"B", "2"},
				{"C", "3"},
				{"D", "4"},
			},
		},
	}
	cons.PulseVectors = versions
	winner := cons.DecideWinner()
	require.Equal(t, "2", winner)
}

func TestDecideWinnerNotEnoughNodesNoConsensus(t *testing.T) {
	cons := basicCons()
	versions := []*PulseVector{
		{
			From: "A",
			Vector: []*PulseProposal{
				{"A", "1"},
				{"B", "2"},
			},
		},
		{
			From: "B",
			Vector: []*PulseProposal{
				{"A", "1"},
				{"B", "2"},
			},
		},
	}
	cons.PulseVectors = versions
	winner := cons.DecideWinner()
	require.Equal(t, NoConsensusStatus, winner)
}

func TestDecideWinnerAdditionalDataNoConsensus(t *testing.T) {
	cons := basicCons()
	versions := []*PulseVector{
		{
			From: "A",
			Vector: []*PulseProposal{
				{"A", "1"},
				{"B", "2"},
				{"C", "3"},
				{"D", "4"},
				{"E", "5"},
			},
		},
		{
			From: "B",
			Vector: []*PulseProposal{
				{"A", "1"},
				{"B", "2"},
			},
		},
	}
	cons.PulseVectors = versions
	winner := cons.DecideWinner()
	require.Equal(t, NoConsensusStatus, winner)
}

func TestDecideWinnerEmptyDataNoConsensus(t *testing.T) {
	cons := basicCons()
	versions := make([]*PulseVector, 0)
	cons.PulseVectors = versions
	winner := cons.DecideWinner()
	require.Equal(t, NoConsensusStatus, winner)
}
