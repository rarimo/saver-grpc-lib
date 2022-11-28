package config

import (
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/kit/pgdb"
	"gitlab.com/rarimo/broadcaster-svc/internal/data"
)

type Config interface {
	comfig.Logger
	comfig.Listenerer
	pgdb.Databaser
	Keyer
	Cosmoser

	Storage() data.Storage
}

type config struct {
	comfig.Logger
	comfig.Listenerer
	pgdb.Databaser
	Keyer
	Cosmoser

	getter  kv.Getter
	storage comfig.Once
}

func New(getter kv.Getter) Config {
	logger := comfig.NewLogger(getter, comfig.LoggerOpts{})

	return &config{
		getter:     getter,
		Logger:     logger,
		Listenerer: comfig.NewListenerer(getter),
		Databaser:  pgdb.NewDatabaser(getter),
		Keyer:      NewKeyer(getter),
		Cosmoser:   NewCosmoser(getter),
	}
}
