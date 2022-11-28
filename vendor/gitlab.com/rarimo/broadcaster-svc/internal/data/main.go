package data

import (
	"context"

	"gitlab.com/distributed_lab/kit/pgdb"
)

//go:generate xo schema "postgres://postgres:postgres@localhost:5432/broadcaster?sslmode=disable" -o ./ --single=schema.xo.go --src templates
//go:generate xo schema "postgres://postgres:postgres@localhost:5432/broadcaster?sslmode=disable" -o pg --single=schema.xo.go --src=pg/templates --go-context=both
//go:generate goimports -w ./

type Storage interface {
	TransactionsQ() TransactionsQ

	DB() *pgdb.DB
}

type TransactionsQ interface {
	InsertCtx(ctx context.Context, t *Transaction) error
	DeleteCtx(ctx context.Context, t *Transaction) error
	Select(ctx context.Context) ([]Transaction, error)
}
type GorpMigrationQ interface{}
