package verifiers

import (
	"context"
	goerr "errors"

	"github.com/gogo/protobuf/proto"
	rarimotypes "github.com/rarimo/rarimo-core/x/rarimocore/types"
	"github.com/rarimo/saver-grpc-lib/voter"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

var (
	ErrInvalidOperationType  = goerr.New("invalid operation type")
	ErrWrongOperationContent = goerr.New("wrong operation content")
	ErrUnsupportedNetwork    = goerr.New("unsupported network")
)

// TransferOperator implements logic for transfer generation on every chain. Every saver should
// implement it based on its chain peculiarities
type TransferOperator interface {
	VerifyTransfer(ctx context.Context, tx, eventId string, transfer *rarimotypes.Transfer) error
}

type TransferVerifier struct {
	operator TransferOperator
	log      *logan.Entry
}

func NewTransferVerifier(operator TransferOperator, log *logan.Entry) *TransferVerifier {
	return &TransferVerifier{
		operator: operator,
		log:      log,
	}
}

// Implements Verifier
var _ voter.Verifier = &TransferVerifier{}

func (t *TransferVerifier) Verify(ctx context.Context, operation rarimotypes.Operation) (rarimotypes.VoteType, error) {
	if operation.OperationType != rarimotypes.OpType_TRANSFER {
		return rarimotypes.VoteType_NO, ErrInvalidOperationType
	}

	transfer := new(rarimotypes.Transfer)
	if err := proto.Unmarshal(operation.Details.Value, transfer); err != nil {
		return rarimotypes.VoteType_NO, err
	}

	if err := t.operator.VerifyTransfer(ctx, transfer.Tx, transfer.EventId, transfer); err != nil {
		switch errors.Cause(err) {
		case ErrUnsupportedNetwork:
			return rarimotypes.VoteType_NO, ErrUnsupportedNetwork
		case ErrWrongOperationContent:
			return rarimotypes.VoteType_NO, nil
		default:
			return rarimotypes.VoteType_NO, err
		}
	}

	return rarimotypes.VoteType_YES, nil
}
