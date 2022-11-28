package config

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

const AccountPrefix = "rarimo"

type KeyConf struct {
	Sender        cryptotypes.PrivKey
	SenderAddress string
	ChainId       string
}

type Keyer interface {
	Key() *KeyConf
}

type keyer struct {
	getter kv.Getter
	once   comfig.Once
}

func NewKeyer(getter kv.Getter) Keyer {
	return &keyer{
		getter: getter,
	}
}

func (k *keyer) Key() *KeyConf {
	return k.once.Do(func() interface{} {
		var config struct {
			SenderPrvKeyHex string `fig:"sender_prv_hex"`
			ChainId         string `fig:"chain_id"`
		}

		if err := figure.Out(&config).From(kv.MustGetStringMap(k.getter, "key")).Please(); err != nil {
			panic(err)
		}

		sender := &secp256k1.PrivKey{Key: hexutil.MustDecode(config.SenderPrvKeyHex)}

		address, err := bech32.ConvertAndEncode(AccountPrefix, sender.PubKey().Address().Bytes())
		if err != nil {
			panic(err)
		}

		return &KeyConf{
			Sender:        sender,
			SenderAddress: address,
			ChainId:       config.ChainId,
		}
	}).(*KeyConf)
}
