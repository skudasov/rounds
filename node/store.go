package node

import (
	"context"
	"github.com/mosuka/cete/kvs"
	"google.golang.org/grpc"
	"log"
	testBadgerPb "rounds/ledger/pb"
	"rounds/logger"
)

type Storage interface {
	// Commit commits block to storage
	Commit(context.Context, BlockData) error
	// GetLatestPulse gets latest block from storage
	GetLatestBlock() (*BlockData, error)
	// GetLatestBlockEpoch get latest block epoch for nodes sync
	GetLatestBlockEpoch() uint64
}

type CeteStorage struct {
	client *kvs.GRPCClient
}

func (m *CeteStorage) GetLatestBlock() (*BlockData, error) {
	// no iterator api in cete for now
	return nil, nil
}

func (m *CeteStorage) Commit(ctx context.Context, b BlockData) error {
	return nil
}

func NewCeteStorage(host string) *CeteStorage {
	s, err := kvs.NewGRPCClient(host)
	if err != nil {
		log.Fatal(err)
	}
	return &CeteStorage{
		s,
	}
}

type TestBadgerStorage struct {
	client testBadgerPb.LedgerClient
	log    *logger.Logger
}

func (m *TestBadgerStorage) GetLatestBlockEpoch() uint64 {
	resp, err := m.client.GetLatestBlockEpoch(context.Background(), &testBadgerPb.LatestPNRequest{})
	if err != nil {
		m.log.Error(err)
	}
	if resp.Error != "" {
		m.log.Error(resp.Error)
	}
	m.log.Infof("latest pulse number received: %d", resp.Epoch)
	return resp.Epoch
}

func (m *TestBadgerStorage) Commit(ctx context.Context, b BlockData) error {
	resp, err := m.client.Commit(ctx, &testBadgerPb.CommitPulseRequest{Entropy: b.WinnerEntropy})
	if err != nil {
		return err
	}
	m.log.Debugf("storage response: %s", resp)
	return nil
}

func (m *TestBadgerStorage) GetLatestBlock() (*BlockData, error) {
	return nil, nil
}

func NewBadgerStorage(addr string) *TestBadgerStorage {
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("failed to connect to storage: %s", err)
	}
	client := testBadgerPb.NewLedgerClient(conn)
	return &TestBadgerStorage{
		client,
		logger.NewLogger(),
	}
}
