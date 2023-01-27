package voter

import (
	"context"

	"gitlab.com/distributed_lab/logan/v3"
	rarimotypes "gitlab.com/rarimo/rarimo-core/x/rarimocore/types"
	"gitlab.com/rarimo/savers/saver-grpc-lib/broadcaster"
)

type Verifier interface {
	Verify(operation rarimotypes.Operation) (rarimotypes.VoteType, error)
}

type Voter struct {
	verifiers   map[rarimotypes.OpType]Verifier
	broadcaster broadcaster.Broadcaster
	log         *logan.Entry
}

func NewVoter(log *logan.Entry, broadcaster broadcaster.Broadcaster, verifiers map[rarimotypes.OpType]Verifier) *Voter {
	return &Voter{
		verifiers:   verifiers,
		broadcaster: broadcaster,
		log:         log,
	}
}

func (v *Voter) Process(ctx context.Context, operation rarimotypes.Operation) error {
	v.log.Infof("Trying to verify operation: %s", operation.Index)

	if verifier, ok := v.verifiers[operation.OperationType]; ok {
		v.log.Infof("Found verifier for op type: %s", operation.OperationType.String())

		result, err := verifier.Verify(operation)
		if err != nil {
			v.log.WithError(err).Errorf("Verification failed for operation: %s", operation.Index)
			return nil
		}

		v.log.Infof("Verification result for operation %s %s", operation.Index, result.String())

		vote := &rarimotypes.MsgVote{
			Creator:   v.broadcaster.Sender(),
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
