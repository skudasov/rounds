package ledger

import (
	"encoding/binary"
	"github.com/dgraph-io/badger"
	"log"
	"rounds/logger"
	"rounds/node"
)

type Storer interface {
	GetLatestBlockEpoch() (uint64, error)
	CommitPulse(block *node.Block) error
}

type BadgerStore struct {
	db *badger.DB

	log *logger.Logger
}

func (m *BadgerStore) CommitPulse(b *node.Block) error {
	m.log.Infof("committing pulse: %s", b.String())
	epoch := make([]byte, 64)
	binary.BigEndian.PutUint64(epoch, b.Epoch)
	if err := m.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(epoch, b.WinnerEntropy)
		return err
	}); err != nil {
		return err
	}
	return nil
}

func (m *BadgerStore) GetLatestBlockEpoch() (uint64, error) {
	var latestPN []byte
	if err := m.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Reverse = true
		opts.AllVersions = false
		it := txn.NewIterator(opts)
		defer it.Close()
		it.Rewind()
		if it.Valid() {
			latestPN = it.Item().Key()
		}
		return nil
	}); err != nil {
		return 0, err
	}
	if len(latestPN) == 0 {
		return 0, nil
	}
	l := binary.BigEndian.Uint64(latestPN)
	return l, nil
}

func NewBadgerStore(c *Config) *BadgerStore {
	log.Printf("opening db by path: %s", c.DB.Path)
	badger.DefaultOptions.Dir = c.DB.Path
	badger.DefaultOptions.ValueDir = c.DB.Path
	db, err := badger.Open(badger.DefaultOptions)
	if err != nil {
		log.Fatal(err)
	}
	return &BadgerStore{
		db,
		logger.NewLogger(),
	}
}
