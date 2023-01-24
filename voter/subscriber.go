package voter

import (
	"context"
	"fmt"

	"github.com/tendermint/tendermint/rpc/client/http"
	"gitlab.com/distributed_lab/logan/v3"
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
}

func NewTransferSubscriber(voter *Voter, client *http.HTTP, rarimo *grpc.ClientConn, log *logan.Entry) *Subscriber {
	return NewSubscriber(voter, client, rarimo, OpQueryTransfer, log)
}

// NewSubscriber creates the subscriber instance for listening to new operations
func NewSubscriber(voter *Voter, client *http.HTTP, rarimo *grpc.ClientConn, query string, log *logan.Entry) *Subscriber {
	return &Subscriber{
		voter:  voter,
		client: client,
		rarimo: rarimo,
		query:  query,
		log:    log,
	}
}

func (o *Subscriber) Run() {
	go func() {
		for {
			o.runner()
			o.log.Info("Resubscribing to the pool...")
		}
	}()
}

func (o *Subscriber) runner() {
	out, err := o.client.Subscribe(context.Background(), OpServiceName, o.query, OpPoolSize)
	if err != nil {
		panic(err)
	}

	for {
		c, ok := <-out
		if !ok {
			if err := o.client.Unsubscribe(context.Background(), OpServiceName, o.query); err != nil {
				o.log.WithError(err).Error("Failed to unsubscribe from new operations")
			}
			break
		}

		for _, index := range c.Events[fmt.Sprintf("%s.%s", rarimo.EventTypeNewOperation, rarimo.AttributeKeyOperationId)] {
			o.log.Infof("New operation found index=%s", index)

			op, err := rarimotypes.NewQueryClient(o.rarimo).Operation(context.TODO(), &rarimotypes.QueryGetOperationRequest{Index: index})
			if err != nil {
				o.log.WithError(err).Errorf("failed to fetch operation data, index = %s", index)
				continue
			}

			if err := o.voter.Process(context.TODO(), op.Operation); err != nil {
				o.log.WithError(err).Errorf("failed to process operation, index = %s", index)
			}
		}
	}
}
