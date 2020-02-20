//go:generate protoc -I ./pb --go_out=plugins=grpc:./pb ./pb/ledger.proto

package ledger

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
	pb "rounds/ledger/pb"
	"rounds/logger"
	"rounds/node"
	"time"
)

type server struct {
	pb.UnimplementedLedgerServer
	store Storer

	log *logger.Logger
}

func (s *server) Commit(ctx context.Context, in *pb.CommitRequest) (*pb.CommitResponse, error) {
	s.log.Infof("Received block data: %s", in.GetEntropy())
	latestEpoch, err := s.store.GetLatestBlockEpoch()
	if err != nil {
		return &pb.CommitResponse{Error: "failed to get latest block epoch"}, nil
	}
	b := &node.Block{
		Epoch:         uint64(latestEpoch + 1),
		Timestamp:     time.Now().Unix(),
		WinnerEntropy: in.GetEntropy(),
	}
	if err := s.store.CommitBlock(b); err != nil {
		return &pb.CommitResponse{Error: "error"}, nil
	}
	return &pb.CommitResponse{Error: "no error"}, nil
}

func (s *server) GetLatestBlockEpoch(ctx context.Context, in *pb.LatestBlockEpochRequest) (*pb.LatestBlockEpochResponse, error) {
	s.log.Infof("Received latest PN request")
	epoch, err := s.store.GetLatestBlockEpoch()
	if err != nil {
		return &pb.LatestBlockEpochResponse{Error: err.Error()}, nil
	}
	return &pb.LatestBlockEpochResponse{Epoch: epoch}, nil
}

func Serve(c *Config) {
	host := c.Ledger.Host
	lis, err := net.Listen("tcp", host)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterLedgerServer(s, &server{
		store: NewBadgerStore(c),
		log:   logger.NewLogger(),
	})
	log.Printf("starting ledger: %s", host)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
