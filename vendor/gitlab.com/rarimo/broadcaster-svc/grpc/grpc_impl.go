package grpc

import (
	"context"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/rarimo/broadcaster-svc/internal/config"
	"gitlab.com/rarimo/broadcaster-svc/internal/data"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
)

type service struct {
	UnimplementedBroadcasterServer
	log      *logan.Entry
	storage  data.Storage
	listener net.Listener
}

func RunAPI(ctx context.Context, cfg config.Config) {
	cfg.Log().Info("starting grpc api")

	srv := grpc.NewServer()

	RegisterBroadcasterServer(srv, &service{
		log:      cfg.Log(),
		listener: cfg.Listener(),
		storage:  cfg.Storage(),
	})

	serve(ctx, srv, cfg)
}

// gRPC service implementation
var _ BroadcasterServer = &service{}

func (s *service) ScheduleBroadcastTx(ctx context.Context, req *ScheduleBroadcastTxRequest) (*ScheduleBroadcastTxResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if len(req.Tx) == 0 {
		return nil, status.Error(codes.InvalidArgument, "tx is empty")
	}

	err := s.storage.TransactionsQ().InsertCtx(ctx, &data.Transaction{
		Data: req.Tx,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to schedule transaction: %s", err.Error())
	}

	return &ScheduleBroadcastTxResponse{}, nil
}
