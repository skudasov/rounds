package node

import "fmt"

type BlockData struct {
	Timestamp     int64
	WinnerEntropy []byte
}

type Block struct {
	Epoch         uint64
	Timestamp     int64
	WinnerEntropy []byte
}

func (b *Block) String() string {
	return fmt.Sprintf(
		"[epoch: %d, ts: %d, winner_entropy: %s]",
		b.Epoch,
		b.Timestamp,
		b.WinnerEntropy,
	)
}
