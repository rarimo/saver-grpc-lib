package verifiers

import (
	goerr "errors"

	"github.com/gogo/protobuf/proto"
	"gitlab.com/distributed_lab/logan/v3"
	rarimotypes "gitlab.com/rarimo/rarimo-core/x/rarimocore/types"
	"gitlab.com/rarimo/savers/saver-grpc-lib/voter"
)

var (
	ErrInvalidOperationType  = goerr.New("invalid operation type")
	ErrWrongOperationContent = goerr.New("wrong operation content")
	ErrUnsupportedNetwork    = goerr.New("unsupported network")
)

// ITransferOperator implements logic for transfer generation on every chain. Every saver should
// implement it based on its chain peculiarities
type ITransferOperator interface {
	VerifyTransfer(tx, eventId string, transfer *rarimotypes.Transfer) error
}

type TransferVerifier struct {
	ITransferOperator
	log *logan.Entry
}

func NewTransferVerifier(operator ITransferOperator, log *logan.Entry) voter.IVerifier {
	return &TransferVerifier{
		ITransferOperator: operator,
		log:               log,
	}
}

// Implements IVerifier
var _ voter.IVerifier = &TransferVerifier{}

func (t *TransferVerifier) Verify(operation rarimotypes.Operation) (rarimotypes.VoteType, error) {
	if operation.OperationType != rarimotypes.OpType_TRANSFER {
		return rarimotypes.VoteType_NO, ErrInvalidOperationType
	}

	transfer := new(rarimotypes.Transfer)
	if err := proto.Unmarshal(operation.Details.Value, transfer); err != nil {
		return rarimotypes.VoteType_NO, err
	}

	switch err := t.VerifyTransfer(transfer.Tx, transfer.EventId, transfer); err {
	case ErrUnsupportedNetwork:
		return rarimotypes.VoteType_NO, ErrUnsupportedNetwork
	case ErrWrongOperationContent:
		return rarimotypes.VoteType_NO, nil
	case nil:
		return rarimotypes.VoteType_YES, nil
	default:
		return rarimotypes.VoteType_NO, err
	}
}
