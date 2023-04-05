package voter

import (
	"context"

	"gitlab.com/distributed_lab/logan/v3"
	oracletypes "gitlab.com/rarimo/rarimo-core/x/oraclemanager/types"
	rarimotypes "gitlab.com/rarimo/rarimo-core/x/rarimocore/types"
	"gitlab.com/rarimo/savers/saver-grpc-lib/broadcaster"
)

type Verifier interface {
	Verify(ctx context.Context, operation rarimotypes.Operation) (rarimotypes.VoteType, error)
}

type Voter struct {
	verifiers   map[rarimotypes.OpType]Verifier
	broadcaster broadcaster.Broadcaster
	chain       string
	log         *logan.Entry
}

func NewVoter(chain string, log *logan.Entry, broadcaster broadcaster.Broadcaster, verifiers map[rarimotypes.OpType]Verifier) *Voter {
	return &Voter{
		verifiers:   verifiers,
		broadcaster: broadcaster,
		chain:       chain,
		log:         log,
	}
}

func (v *Voter) Process(ctx context.Context, operation rarimotypes.Operation) error {
	v.log.Infof("Trying to verify operation: %s", operation.Index)

	if verifier, ok := v.verifiers[operation.OperationType]; ok {
		v.log.Infof("Found verifier for op type: %s", operation.OperationType.String())

		result, err := verifier.Verify(ctx, operation)
		if err != nil {
			v.log.WithError(err).Errorf("Verification failed for operation: %s", operation.Index)
			return nil
		}

		v.log.Infof("Verification result for operation %s %s", operation.Index, result.String())

		vote := &oracletypes.MsgVote{
			Index: &oracletypes.OracleIndex{
				Chain:   v.chain,
				Account: v.broadcaster.Sender(),
			},
			Operation: operation.Index,
			Vote:      result,
		}

		return v.broadcaster.BroadcastTx(ctx, vote)
	}

	v.log.Errorf("Verifier not found for operation type: %s", operation.OperationType.String())
	return nil
}

func (v *Voter) Sender() string {
	return v.broadcaster.Sender()
}
