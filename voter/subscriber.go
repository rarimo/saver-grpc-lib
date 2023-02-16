package voter

import (
	"context"
	"fmt"
	"time"

	"github.com/tendermint/tendermint/rpc/client/http"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
	rarimo "gitlab.com/rarimo/rarimo-core/x/rarimocore/types"
	rarimotypes "gitlab.com/rarimo/rarimo-core/x/rarimocore/types"
	"google.golang.org/grpc"
)

const (
	OpServiceName   = "op-subscriber"
	OpQueryTransfer = "tm.event='Tx' AND new_operation.operation_type='TRANSFER'"
	OpPoolSize      = 1000
)

// Subscriber subscribes to the NewOperation events on the tendermint core.
type Subscriber struct {
	voter  *Voter
	client *http.HTTP
	rarimo *grpc.ClientConn
	query  string
	log    *logan.Entry
	cfg    SubscriberConfig
}

func NewTransferSubscriber(voter *Voter, client *http.HTTP, rarimo *grpc.ClientConn, log *logan.Entry, cfg SubscriberConfig) *Subscriber {
	return NewSubscriber(voter, client, rarimo, OpQueryTransfer, log, cfg)
}

// NewSubscriber creates the subscriber instance for listening to new operations
func NewSubscriber(voter *Voter, client *http.HTTP, rarimo *grpc.ClientConn, query string, log *logan.Entry, cfg SubscriberConfig) *Subscriber {
	return &Subscriber{
		voter:  voter,
		client: client,
		rarimo: rarimo,
		query:  query,
		log:    log,
		cfg:    cfg,
	}
}

func (s *Subscriber) Run(ctx context.Context) {
	defer func() {
		if rvr := recover(); rvr != nil {
			s.log.WithRecover(rvr).Error("Subscriber panicked")
		}
	}()

	running.WithBackOff(ctx,
		s.log.WithField("who", "subscriber"),
		"subscriber",
		s.runOnce,
		s.cfg.MinRetryPeriod, s.cfg.MinRetryPeriod, s.cfg.MaxRetryPeriod)
}

func (s *Subscriber) cleanup() {
	if err := s.client.Unsubscribe(context.TODO(), OpServiceName, s.query); err != nil {
		panic(errors.Wrap(err, "Failed to unsubscribe from new operations"))
	}
}

func (s *Subscriber) runOnce(ctx context.Context) error {
	s.log.Infof("Starting subscription for the new unvoted operations")

	out, err := s.client.Subscribe(ctx, OpServiceName, s.query, OpPoolSize)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to the new operations")
	}

	queryClient := rarimotypes.NewQueryClient(s.rarimo)

	for {
		eventData := readOneEvent(out, 10*time.Second)
		if eventData == nil {
			s.log.Debug("no events to process, resubscribing")
			return nil
		}

		for _, index := range eventData.Events[fmt.Sprintf("%s.%s", rarimo.EventTypeNewOperation, rarimo.AttributeKeyOperationId)] {
			s.log.
				WithFields(logan.F{"index": index}).
				Info("New operation found")

			op, err := queryClient.Operation(ctx, &rarimotypes.QueryGetOperationRequest{Index: index})
			if err != nil {
				s.log.
					WithError(err).
					WithFields(logan.F{"index": index}).
					Errorf("failed to fetch operation data")
				continue
			}

			if op.Operation.Status != rarimotypes.OpStatus_INITIALIZED {
				continue
			}

			if err := s.voter.Process(ctx, op.Operation); err != nil {
				s.log.
					WithError(err).
					WithFields(logan.F{"index": index}).
					Errorf("failed to process operation")
			}
		}
	}
}

func readOneEvent(from <-chan coretypes.ResultEvent, timeout time.Duration) *coretypes.ResultEvent {
	select {
	case e := <-from:
		return &e
	case <-time.NewTimer(timeout).C:
		return nil
	default:
		return nil
	}
}
