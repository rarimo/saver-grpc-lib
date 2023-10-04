package voter

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types/query"
	rarimotypes "github.com/rarimo/rarimo-core/x/rarimocore/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"google.golang.org/grpc"
)

// Catchupper catches up old unsigned operations from core.
type Catchupper struct {
	rarimoClient rarimotypes.QueryClient
	voter        *Voter
	log          *logan.Entry
}

// NewCatchupper creates the catchup instance for adding all unsigned operations to the pool
func NewCatchupper(rarimo *grpc.ClientConn, voter *Voter, log *logan.Entry) *Catchupper {
	return &Catchupper{
		rarimoClient: rarimotypes.NewQueryClient(rarimo),
		voter:        voter,
		log:          log,
	}
}

func (c *Catchupper) Run(ctx context.Context) {
	c.log.Infof("Starting catchup unvoted operations")

	var nextKey []byte

	for {
		operations, err := c.rarimoClient.OperationAll(ctx, &rarimotypes.QueryAllOperationRequest{
			Pagination: &query.PageRequest{
				Key: nextKey,
			},
		})
		if err != nil {
			panic(errors.Wrap(err, "failed to get operations"))
		}

		for _, op := range operations.Operation {
			if op.Status != rarimotypes.OpStatus_INITIALIZED {
				continue
			}

			c.log.WithField("index", op.Index).Info("New unapproved operation found")

			_, err := c.rarimoClient.Vote(ctx, &rarimotypes.QueryGetVoteRequest{
				Operation: op.Index,
				Validator: c.voter.Sender(),
			})

			if err == nil {
				c.log.WithField("index", op.Index).Info("Operation already voted")
				continue
			}

			if err := c.voter.Process(ctx, op); err != nil {
				c.log.WithError(err).WithField("index", op.Index).Error("failed to process operation")
			}
		}

		nextKey = operations.Pagination.NextKey
		if nextKey == nil {
			c.log.Infof("Finished catchup unvoted operations")
			return
		}
	}
}
