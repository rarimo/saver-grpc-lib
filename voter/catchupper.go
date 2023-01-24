package voter

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types/query"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	rarimotypes "gitlab.com/rarimo/rarimo-core/x/rarimocore/types"
	"google.golang.org/grpc"
)

// Catchupper catches up old unsigned operations from core.
type Catchupper struct {
	rarimo *grpc.ClientConn
	voter  *Voter
	log    *logan.Entry
}

// NewCatchupper creates the catchup instance for adding all unsigned operations to the pool
func NewCatchupper(rarimo *grpc.ClientConn, voter *Voter, log *logan.Entry) *Catchupper {
	return &Catchupper{
		rarimo: rarimo,
		voter:  voter,
		log:    log,
	}
}

func (o *Catchupper) Run() {
	var nextKey []byte

	for {
		operations, err := rarimotypes.NewQueryClient(o.rarimo).OperationAll(context.TODO(), &rarimotypes.QueryAllOperationRequest{Pagination: &query.PageRequest{Key: nextKey}})
		if err != nil {
			panic(err)
		}

		for _, op := range operations.Operation {
			if !op.Approved {
				o.log.Infof("New unapproved operation found index=%s", op.Index)

				_, err := rarimotypes.NewQueryClient(o.rarimo).Vote(context.TODO(), &rarimotypes.QueryGetVoteRequest{Operation: op.Index, Validator: o.voter.Sender()})
				if err == nil {
					o.log.Infof("Operation already voted, index=%s", op.Index)
					continue
				}

				if err := o.voter.Process(context.TODO(), op); err != nil {
					panic(errors.Wrap(err, "failed to process operation", map[string]interface{}{
						"index": op.Index,
					}))
				}
			}
		}

		nextKey = operations.Pagination.NextKey
		if nextKey == nil {
			return
		}
	}
}
