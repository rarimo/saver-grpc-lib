package voter

import (
	"context"

	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/rarify-protocol/saver-grpc-lib/broadcaster"
	rarimotypes "gitlab.com/rarimo/rarimo-core/x/rarimocore/types"
	"google.golang.org/grpc"
)

type IVerifier interface {
	Verify(operation rarimotypes.Operation) (rarimotypes.VoteType, error)
}

type Voter struct {
	processors  map[rarimotypes.OpType]IVerifier
	rarimo      *grpc.ClientConn
	broadcaster broadcaster.Broadcaster
	log         *logan.Entry
}

func NewVoter(log *logan.Entry, broadcaster broadcaster.Broadcaster, processors map[rarimotypes.OpType]IVerifier) *Voter {
	return &Voter{
		processors:  processors,
		broadcaster: broadcaster,
		log:         log,
	}
}

func (v *Voter) Process(ctx context.Context, operation rarimotypes.Operation) error {
	v.log.Infof("Trying to verify operation: %s", operation.Index)

	if processor, ok := v.processors[operation.OperationType]; ok {
		v.log.Infof("Found verifier for op type: %s", operation.OperationType.String())

		result, err := processor.Verify(operation)
		if err != nil {
			v.log.WithError(err).Errorf("Verification failed for operation: %s", operation.Index)
			return nil
		}

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
