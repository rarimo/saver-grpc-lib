package transactor

import (
	"context"
	"fmt"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/types/tx"
	txclient "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"gitlab.com/distributed_lab/logan/v3/errors"
	rarimocore "gitlab.com/rarify-protocol/rarimo-core/x/rarimocore/types"
	tokentypes "gitlab.com/rarify-protocol/rarimo-core/x/tokenmanager/types"
)

const (
	minGasPrice   = 1
	gasLimit      = 100_000_000
	SuccessTxCode = 0
)

type transactor struct {
	cfg      transactorConfig
	txConfig sdkclient.TxConfig
	txclient txclient.ServiceClient
	auth     authtypes.QueryClient
}

func (t *transactor) SubmitTransferOp(
	ctx context.Context,
	txHash string,
	eventId string,
	fromChain string,
	tokenType tokentypes.Type,
) error {
	msg := rarimocore.NewMsgCreateTransferOp(t.cfg.SenderAddress, txHash, eventId, fromChain, tokenType)

	tx, err := t.genTx(ctx, msg)
	if err != nil {
		return err
	}

	return t.broadcastTx(ctx, tx)
}

func (t *transactor) genTx(ctx context.Context, msg *rarimocore.MsgCreateTransferOp) ([]byte, error) {
	builder := t.txConfig.NewTxBuilder()
	err := builder.SetMsgs(msg)
	if err != nil {
		return nil, err
	}

	builder.SetGasLimit(gasLimit)
	builder.SetFeeAmount(types.Coins{types.NewInt64Coin("stake", int64(gasLimit*minGasPrice))})

	accountResp, err := t.auth.Account(ctx, &authtypes.QueryAccountRequest{Address: t.cfg.SenderAddress})
	if err != nil {
		panic(err)
	}

	account := authtypes.BaseAccount{}
	err = account.Unmarshal(accountResp.Account.Value)
	if err != nil {
		panic(err)
	}

	err = builder.SetSignatures(signing.SignatureV2{
		PubKey: t.cfg.Sender.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  t.txConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: account.Sequence,
	})
	if err != nil {
		return nil, err
	}

	signerData := xauthsigning.SignerData{
		ChainID:       t.cfg.ChainId,
		AccountNumber: account.AccountNumber,
		Sequence:      account.Sequence,
	}

	sigV2, err := clienttx.SignWithPrivKey(
		t.txConfig.SignModeHandler().DefaultMode(), signerData,
		builder, t.cfg.Sender, t.txConfig, account.Sequence,
	)

	err = builder.SetSignatures(sigV2)
	if err != nil {
		return nil, err
	}

	return t.txConfig.TxEncoder()(builder.GetTx())
}

func (t *transactor) broadcastTx(ctx context.Context, tx []byte) error {
	grpcRes, err := t.txclient.BroadcastTx(
		ctx,
		&client.BroadcastTxRequest{
			Mode:    client.BroadcastMode_BROADCAST_MODE_BLOCK,
			TxBytes: tx,
		},
	)
	if err != nil {
		return err
	}

	if grpcRes.TxResponse.Code != SuccessTxCode {
		return errors.New(fmt.Sprintf("Got error code: %d, info: %s", grpcRes.TxResponse.Code, grpcRes.TxResponse.Info))
	}

	return nil
}
