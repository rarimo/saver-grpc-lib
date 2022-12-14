package broadcaster

import (
	"context"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
	broadcasterclient "gitlab.com/rarimo/broadcaster-svc/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type Broadcasterer interface {
	Broadcaster() Broadcaster
}

type Broadcaster interface {
	BroadcastTx(
		ctx context.Context,
		msgs ...sdk.Msg,
	) error
	Sender() string
}

type broadcasterer struct {
	getter kv.Getter
	once   comfig.Once
}

func New(getter kv.Getter) Broadcasterer {
	return &broadcasterer{
		getter: getter,
	}
}

func (c *broadcasterer) Broadcaster() Broadcaster {
	return c.once.Do(func() interface{} {
		var config struct {
			Addr          string `fig:"addr"`
			SenderAccount string `fig:"sender_account"`
		}

		if err := figure.Out(&config).From(kv.MustGetStringMap(c.getter, "broadcaster")).Please(); err != nil {
			panic(errors.Wrap(err, "failed to figure out broadcaster config"))
		}

		con, err := grpc.Dial(config.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    10 * time.Second, // wait time before ping if no activity
			Timeout: 20 * time.Second, // ping timeout
		}))
		if err != nil {
			panic(errors.Wrap(err, "failed to dial broadcaster rpc"))
		}

		return &broadcaster{
			senderAccount: config.SenderAccount,
			txConfig:      tx.NewTxConfig(codec.NewProtoCodec(codectypes.NewInterfaceRegistry()), []signing.SignMode{signing.SignMode_SIGN_MODE_DIRECT}),
			cli:           broadcasterclient.NewBroadcasterClient(con),
		}
	}).(Broadcaster)
}
