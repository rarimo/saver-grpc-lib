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
)

// ITransferOperator implements logic for transfer generation on every chain. Every saver should
// implement it based on its chain peculiarities
type ITransferOperator interface {
	GetOperation(tx, eventId string) (*rarimotypes.Transfer, error)
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

func (t *TransferVerifier) Verify(operation rarimotypes.Operation) error {
	if operation.OperationType != rarimotypes.OpType_TRANSFER {
		return ErrInvalidOperationType
	}

	transfer := new(rarimotypes.Transfer)
	if err := proto.Unmarshal(operation.Details.Value, transfer); err != nil {
		return err
	}

	confirmedTransfer, err := t.GetOperation(transfer.Tx, transfer.EventId)
	if err != nil {
		return err
	}

	if !proto.Equal(confirmedTransfer, transfer) {
		return ErrWrongOperationContent
	}

	return nil
}
