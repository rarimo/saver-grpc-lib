package broadcaster

import (
	"context"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	broadcasterclient "gitlab.com/rarimo/broadcaster-svc/grpc"
)

type broadcaster struct {
	cli      broadcasterclient.BroadcasterClient
	txConfig sdkclient.TxConfig
}

func (t *broadcaster) BroadcastTx(
	ctx context.Context,
	msgs ...sdk.Msg,
) error {
	builder := t.txConfig.NewTxBuilder()
	err := builder.SetMsgs(msgs...)
	if err != nil {
		return err
	}

	rawTx, err := t.txConfig.TxEncoder()(builder.GetTx())
	if err != nil {
		return err
	}

	_, err = t.cli.ScheduleBroadcastTx(ctx, &broadcasterclient.ScheduleBroadcastTxRequest{
		Tx: rawTx,
	})
	return err
}
