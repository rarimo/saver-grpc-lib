package transactor

import (
	"context"
	"crypto/ecdsa"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	client "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	tokentypes "gitlab.com/rarify-protocol/rarimo-core/x/tokenmanager/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"time"
)

const AccountPrefix = "rarimo"

type Transactorer interface {
	Transactor() Transactor
}

type Transactor interface {
	SubmitTransferOp(
		ctx context.Context,
		creator string,
		txHash string,
		eventId string,
		fromChain string,
		tokenType tokentypes.Type,
	) error
}

type transactorer struct {
	getter kv.Getter
	once   comfig.Once
}

func New(getter kv.Getter) Transactorer {
	return &transactorer{
		getter: getter,
	}
}

type transactorConfig struct {
	RPC           string
	PrivateKey    *ecdsa.PrivateKey
	Sender        cryptotypes.PrivKey
	SenderAddress string
	ChainId       string
}

func (c *transactorer) Transactor() Transactor {
	return c.once.Do(func() interface{} {
		var config struct {
			RPC             string `fig:"rpc"`
			PrivateKeyHex   string `fig:"prv_key_hex"`
			SenderPrvKeyHex string `fig:"sender_prv_hex"`
			ChainId         string `fig:"chain_id"`
		}

		if err := figure.Out(&config).From(kv.MustGetStringMap(c.getter, "transactor")).Please(); err != nil {
			panic(err)
		}

		con, err := grpc.Dial(config.RPC, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    10 * time.Second, // wait time before ping if no activity
			Timeout: 20 * time.Second, // ping timeout
		}))
		if err != nil {
			panic(err)
		}

		prv, err := crypto.ToECDSA(hexutil.MustDecode(config.PrivateKeyHex))
		if err != nil {
			panic(err)
		}

		sender := &secp256k1.PrivKey{Key: hexutil.MustDecode(config.SenderPrvKeyHex)}

		address, err := bech32.ConvertAndEncode(AccountPrefix, sender.PubKey().Address().Bytes())
		if err != nil {
			panic(err)
		}

		return &transactor{
			cfg:      transactorConfig{RPC: config.RPC, PrivateKey: prv, Sender: sender, SenderAddress: address, ChainId: config.ChainId},
			txConfig: tx.NewTxConfig(codec.NewProtoCodec(codectypes.NewInterfaceRegistry()), []signing.SignMode{signing.SignMode_SIGN_MODE_DIRECT}),
			txclient: client.NewServiceClient(con),
		}
	}).(Transactor)
}
